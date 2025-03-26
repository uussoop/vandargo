// Package vandargo provides a secure integration with the Vandar payment gateway
// client.go implements the HTTP client for Vandar API communication
package vandargo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the main Vandar API client
type Client struct {
	config     ConfigInterface
	httpClient HTTPClientInterface
	logger     LoggerInterface
	storage    StorageInterface
}

// NewClient creates a new Vandar API client
func NewClient(config ConfigInterface, storage StorageInterface, logger LoggerInterface) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if storage == nil {
		return nil, fmt.Errorf("storage cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Create HTTP client with appropriate timeouts
	httpClient := &http.Client{
		Timeout: time.Duration(config.GetTimeout()) * time.Second,
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		storage:    storage,
	}, nil
}

// WithHTTPClient allows setting a custom HTTP client
func (c *Client) WithHTTPClient(httpClient HTTPClientInterface) *Client {
	c.httpClient = httpClient
	return c
}

// InitiatePayment starts a new payment transaction
func (c *Client) InitiatePayment(ctx context.Context, amount int64, description string, metadata map[string]string) (*PaymentInitResponse, error) {
	// Create payment init request
	req := &PaymentInitRequest{
		Amount:      amount,
		CallbackURL: c.config.GetCallbackURL(),
		Description: description,
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"api_key":      c.config.GetAPIKey(),
		"amount":       req.Amount,
		"callback_url": req.CallbackURL,
	}

	if req.Description != "" {
		apiReq["description"] = req.Description
	}

	// Add metadata if provided
	if metadata != nil {
		for key, value := range metadata {
			apiReq[key] = value
		}
	}

	// Make API request
	respBody, _, err := c.makeRequest(ctx, http.MethodPost, "/api/v4/send", apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize payment: %w", err)
	}

	// Parse API response
	var apiResp PaymentInitResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check if payment initialization was successful
	if apiResp.Status != 1 {
		return &apiResp, fmt.Errorf("payment initialization failed: %s", apiResp.Message)
	}

	// Create transaction record
	transaction := &Transaction{
		ID:          generateRequestID(),
		Token:       apiResp.Token,
		Amount:      req.Amount,
		Status:      "INIT",
		Description: req.Description,
		Metadata:    metadata,
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

	return &apiResp, nil
}

// VerifyPayment verifies a payment transaction
func (c *Client) VerifyPayment(ctx context.Context, token string) (*PaymentVerifyResponse, error) {
	// Create verify request
	req := &PaymentVerifyRequest{
		Token: token,
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"api_key": c.config.GetAPIKey(),
		"token":   req.Token,
	}

	// Make API request
	respBody, _, err := c.makeRequest(ctx, http.MethodPost, "/api/v4/verify", apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to verify payment: %w", err)
	}

	// Parse API response
	var apiResp PaymentVerifyResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check if payment verification was successful
	if apiResp.Status != 1 {
		return &apiResp, fmt.Errorf("payment verification failed: %s", apiResp.Message)
	}

	// Get transaction from storage
	transaction, err := c.storage.GetTransaction(ctx, token)
	if err == nil {
		// Update transaction status
		transaction.Status = "PAID"
		transaction.TransactionID = apiResp.TransID
		transaction.CardNumber = apiResp.CardNumber
		transaction.CID = apiResp.CID
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
			"token": token,
		})
		// Continue with the response even if transaction is not found
	}

	return &apiResp, nil
}

// GetTransactionInfo retrieves detailed information about a transaction
func (c *Client) GetTransactionInfo(ctx context.Context, token string) (*TransactionInfoResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"api_key": c.config.GetAPIKey(),
		"token":   token,
	}

	// Make API request
	respBody, _, err := c.makeRequest(ctx, http.MethodPost, "/api/v4/transaction", apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction info: %w", err)
	}

	// Parse API response
	var apiResp TransactionInfoResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &apiResp, nil
}

// RefundPayment initiates a refund for a transaction
func (c *Client) RefundPayment(ctx context.Context, transactionID string, amount int64) (*RefundResponse, error) {
	// Create refund request
	req := &RefundRequest{
		TransactionID: transactionID,
		Amount:        amount,
	}

	// Prepare API request body
	apiReq := map[string]interface{}{
		"api_key":        c.config.GetAPIKey(),
		"transaction_id": req.TransactionID,
	}

	if req.Amount > 0 {
		apiReq["amount"] = req.Amount
	}

	// Make API request
	respBody, _, err := c.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/v3/business/%s/transaction/%s/refund", "business", req.TransactionID),
		apiReq,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to refund payment: %w", err)
	}

	// Parse API response
	var apiResp RefundResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check if refund was successful
	if !apiResp.Status {
		return &apiResp, fmt.Errorf("payment refund failed: %s", apiResp.Message)
	}

	return &apiResp, nil
}

// makeRequest creates and executes an HTTP request to the Vandar API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, int, error) {
	url := c.config.GetBaseURL() + endpoint

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.GetAPIKey())

	// Add tracking information
	requestID := generateRequestID()
	req.Header.Set("X-Request-ID", requestID)

	// Log the request (without sensitive data)
	c.logger.Debug(ctx, "Making API request", map[string]interface{}{
		"method":     method,
		"endpoint":   endpoint,
		"request_id": requestID,
	})

	// Execute the request with retry mechanism
	var resp *http.Response
	var respErr error

	// Execute request
	resp, respErr = c.httpClient.Do(req)
	if respErr != nil {
		c.logger.Error(ctx, "API request failed", respErr, map[string]interface{}{
			"method":     method,
			"endpoint":   endpoint,
			"request_id": requestID,
		})
		return nil, 0, fmt.Errorf("api request failed: %w", respErr)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response (without sensitive data)
	c.logger.Debug(ctx, "Received API response", map[string]interface{}{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": resp.StatusCode,
		"request_id":  requestID,
	})

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			// If can't parse as APIError, create a generic one
			apiErr = APIError{
				Message: string(respBody),
				Code:    fmt.Sprintf("%d", resp.StatusCode),
			}
		}

		return nil, resp.StatusCode, &apiErr
	}

	return respBody, resp.StatusCode, nil
}

// generateRequestID creates a unique ID for request tracking
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
