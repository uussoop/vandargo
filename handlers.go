// Package vandargo provides a secure integration with the Vandar payment gateway
// handlers.go implements HTTP handlers for different payment operations
package vandargo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RegisterRoutes registers all the handlers with the provided router
func (c *Client) RegisterRoutes(router RouterInterface) {
	// Payment initialization
	router.POST("/payments/init", Chain(
		c.handlePaymentInit,
		RequestIDMiddleware(),
		LoggingMiddleware(c.logger),
		SecurityHeadersMiddleware(),
		RateLimitMiddleware(10, 60),
		AuthMiddleware(c.config),
	))

	// Payment verification
	router.POST("/payments/verify", Chain(
		c.handlePaymentVerify,
		RequestIDMiddleware(),
		LoggingMiddleware(c.logger),
		SecurityHeadersMiddleware(),
		RateLimitMiddleware(10, 60),
		AuthMiddleware(c.config),
	))

	// Payment status check
	router.GET("/payments/status", Chain(
		c.handlePaymentStatus,
		RequestIDMiddleware(),
		LoggingMiddleware(c.logger),
		SecurityHeadersMiddleware(),
		RateLimitMiddleware(20, 60),
		AuthMiddleware(c.config),
	))

	// Refund
	router.POST("/payments/refund", Chain(
		c.handleRefund,
		RequestIDMiddleware(),
		LoggingMiddleware(c.logger),
		SecurityHeadersMiddleware(),
		RateLimitMiddleware(5, 60),
		AuthMiddleware(c.config),
	))

	// Callback
	router.POST("/payments/callback", Chain(
		c.handleCallback,
		RequestIDMiddleware(),
		LoggingMiddleware(c.logger),
		SecurityHeadersMiddleware(),
		IPFilterMiddleware(c.config),
	))
}

// handlePaymentInit handles payment initialization requests
func (c *Client) handlePaymentInit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req PaymentInitRequest
	if err := parseJSONBody(r, &req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Validate request
	if err := ValidatePaymentInitRequest(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Set callback URL from config if not provided
	if req.CallbackURL == "" {
		req.CallbackURL = c.config.GetCallbackURL()
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"amount":       req.Amount,
		"callback_url": req.CallbackURL,
	}

	if req.Description != "" {
		apiReq["description"] = req.Description
	}

	if req.Mobile != "" {
		apiReq["mobile"] = req.Mobile
	}

	if req.FactorNumber != "" {
		apiReq["factorNumber"] = req.FactorNumber
	}

	if req.ValidCardNumber != "" {
		apiReq["valid_card_number"] = req.ValidCardNumber
	}

	// Make API request
	respBody, statusCode, err := c.makeRequest(ctx, http.MethodPost, "/api/v4/send", apiReq)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to initialize payment")
		c.logger.Error(ctx, "Failed to initialize payment", err, map[string]interface{}{
			"request": req,
		})
		return
	}

	// Parse API response
	var apiResp PaymentInitResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to parse API response")
		c.logger.Error(ctx, "Failed to parse API response", err, map[string]interface{}{
			"response_body": string(respBody),
		})
		return
	}

	// Check if payment initialization was successful
	if apiResp.Status != 1 {
		c.respondWithError(w, statusCode, ErrPaymentFailed, apiResp.Message)
		return
	}

	// Create transaction record
	transaction := &Transaction{
		ID:          generateRequestID(),
		Token:       apiResp.Token,
		Amount:      req.Amount,
		Status:      "INIT",
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store transaction
	err = c.storage.StoreTransaction(ctx, transaction)
	if err != nil {
		c.logger.Error(ctx, "Failed to store transaction", err, map[string]interface{}{
			"transaction": transaction,
		})
		// Continue with the response even if storage fails
	}

	// Respond with success
	c.respondWithJSON(w, http.StatusOK, apiResp)
}

// handlePaymentVerify handles payment verification requests
func (c *Client) handlePaymentVerify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req PaymentVerifyRequest
	if err := parseJSONBody(r, &req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Validate request
	if err := ValidatePaymentVerifyRequest(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"token": req.Token,
	}

	// Make API request
	respBody, statusCode, err := c.makeRequest(ctx, http.MethodPost, "/api/v4/verify", apiReq)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to verify payment")
		c.logger.Error(ctx, "Failed to verify payment", err, map[string]interface{}{
			"token": req.Token,
		})
		return
	}

	// Parse API response
	var apiResp PaymentVerifyResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to parse API response")
		c.logger.Error(ctx, "Failed to parse API response", err, map[string]interface{}{
			"response_body": string(respBody),
		})
		return
	}

	// Check if payment verification was successful
	if !apiResp.Status {
		c.respondWithError(w, statusCode, ErrVerificationFailed, apiResp.Message)
		return
	}

	// Get transaction from storage
	transaction, err := c.storage.GetTransaction(ctx, req.Token)
	if err == nil {
		// Update transaction status
		transaction.Status = "PAID"
		transaction.RefID = apiResp.RefID
		transaction.CardNumber = apiResp.CardNumber
		transaction.CardHash = apiResp.CardHash
		transaction.UpdatedAt = time.Now()

		completedAt := time.Now()
		transaction.CompletedAt = &completedAt

		// Store updated transaction
		err = c.storage.UpdateTransaction(ctx, transaction)
		if err != nil {
			c.logger.Error(ctx, "Failed to update transaction", err, map[string]interface{}{
				"transaction": transaction,
			})
			// Continue with the response even if storage fails
		}
	} else {
		c.logger.Warn(ctx, "Transaction not found in storage", map[string]interface{}{
			"token": req.Token,
		})
		// Continue with the response even if transaction is not found
	}

	// Respond with success
	c.respondWithJSON(w, http.StatusOK, apiResp)
}

// handlePaymentStatus handles payment status check requests
func (c *Client) handlePaymentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, "Token is required")
		return
	}

	// Create request
	req := PaymentStatusRequest{
		Token: token,
	}

	// Validate request
	if err := ValidatePaymentStatusRequest(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Make API request
	respBody, statusCode, err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v4/%s", token), nil)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to check payment status")
		c.logger.Error(ctx, "Failed to check payment status", err, map[string]interface{}{
			"token": token,
		})
		return
	}

	// Parse API response
	var apiResp PaymentStatusResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to parse API response")
		c.logger.Error(ctx, "Failed to parse API response", err, map[string]interface{}{
			"response_body": string(respBody),
		})
		return
	}

	// Respond with the status
	c.respondWithJSON(w, statusCode, apiResp)
}

// handleRefund handles refund requests
func (c *Client) handleRefund(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req RefundRequest
	if err := parseJSONBody(r, &req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Validate request
	if err := ValidateRefundRequest(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"transaction_id": req.TransactionID,
	}

	if req.Amount > 0 {
		apiReq["amount"] = req.Amount
	}

	// Make API request
	respBody, statusCode, err := c.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/v3/business/%s/transaction/%s/refund", "business", req.TransactionID),
		apiReq,
	)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to refund payment")
		c.logger.Error(ctx, "Failed to refund payment", err, map[string]interface{}{
			"transaction_id": req.TransactionID,
			"amount":         req.Amount,
		})
		return
	}

	// Parse API response
	var apiResp RefundResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, ErrInternalError, "Failed to parse API response")
		c.logger.Error(ctx, "Failed to parse API response", err, map[string]interface{}{
			"response_body": string(respBody),
		})
		return
	}

	// Check if refund was successful
	if !apiResp.Status {
		c.respondWithError(w, statusCode, ErrRefundFailed, apiResp.Message)
		return
	}

	// Respond with success
	c.respondWithJSON(w, http.StatusOK, apiResp)
}

// handleCallback handles callbacks from Vandar after payment
func (c *Client) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse callback data
	err := r.ParseForm()
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, "Invalid form data")
		c.logger.Error(ctx, "Failed to parse callback form data", err, nil)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, "Token is required")
		return
	}

	// Create callback data
	callbackData := &CallbackData{
		Token:  token,
		Status: r.FormValue("status"),
	}

	// Validate callback data
	if err := ValidateCallbackData(callbackData); err != nil {
		c.respondWithError(w, http.StatusBadRequest, ErrInvalidRequest, err.Error())
		return
	}

	// Log callback details
	c.logger.Info(ctx, "Received payment callback", map[string]interface{}{
		"token":  token,
		"status": callbackData.Status,
	})

	// Get transaction from storage
	transaction, err := c.storage.GetTransaction(ctx, token)
	if err != nil {
		c.logger.Warn(ctx, "Transaction not found for callback", map[string]interface{}{
			"token": token,
		})
		// Continue with the response even if transaction is not found
	} else {
		// Update transaction status based on callback status
		transaction.Status = callbackData.Status
		transaction.UpdatedAt = time.Now()

		// Store updated transaction
		err = c.storage.UpdateTransaction(ctx, transaction)
		if err != nil {
			c.logger.Error(ctx, "Failed to update transaction from callback", err, map[string]interface{}{
				"transaction": transaction,
			})
			// Continue with the response even if storage fails
		}
	}

	// Respond with success
	c.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Callback received successfully",
	})
}

// parseJSONBody parses a JSON request body into the given struct
func parseJSONBody(r *http.Request, v interface{}) error {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("Content-Type must be application/json")
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	// Parse JSON
	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// respondWithJSON responds with a JSON payload
func (c *Client) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Marshal payload to JSON
	response, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error(context.Background(), "Failed to marshal JSON response", err, map[string]interface{}{
			"payload": payload,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set status code and write response
	w.WriteHeader(statusCode)
	_, err = w.Write(response)
	if err != nil {
		c.logger.Error(context.Background(), "Failed to write response", err, nil)
	}
}

// respondWithError responds with an error message
func (c *Client) respondWithError(w http.ResponseWriter, statusCode int, err error, message string) {
	errorResponse := APIErrorResponse(err)
	if message != "" {
		errorResponse["message"] = message
	}

	c.respondWithJSON(w, statusCode, errorResponse)
}
