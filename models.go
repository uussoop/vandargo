// Package vandargo provides a secure integration with the Vandar payment gateway
// models.go contains data structures for requests, responses, and internal data
package vandargo

import (
	"fmt"
	"time"
)

// Transaction represents a payment transaction in the system
type Transaction struct {
	// ID is the unique identifier for the transaction
	ID string `json:"id"`

	// Token is the payment token from Vandar
	Token string `json:"token"`

	// Amount is the transaction amount in Rials
	Amount int64 `json:"amount"`

	// Status represents the current status of the transaction
	Status string `json:"status"`

	// Description is a description of what the payment is for
	Description string `json:"description"`

	// Metadata contains additional data about the transaction
	Metadata map[string]string `json:"metadata,omitempty"`

	// RefID is the reference ID received after successful payment
	TransactionID int64 `json:"transaction_id,omitempty"`

	// CID is the SHA256 hash of the card number
	CID string `json:"cid,omitempty"`

	// CardNumber is the masked card number used for payment (last 4 digits)
	CardNumber string `json:"card_number,omitempty"`

	// CardHash is the hashed card number
	CardHash string `json:"card_hash,omitempty"`

	// CreatedAt is when the transaction was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the transaction was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when the transaction was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// PaymentInitRequest represents a request to initialize a payment
type PaymentInitRequest struct {
	// Amount is the payment amount in Rials
	Amount int64 `json:"amount"`

	// CallbackURL is where the user will be redirected after payment
	CallbackURL string `json:"callback_url"`

	// Description is a description of what the payment is for
	Description string `json:"description,omitempty"`

	// Mobile is the customer's mobile number (optional)
	Mobile string `json:"mobile,omitempty"`

	// FactorNumber is an optional invoice/factor number
	FactorNumber string `json:"factorNumber,omitempty"`

	// ValidCardNumber is an optional allowed card number
	ValidCardNumber string `json:"valid_card_number,omitempty"`
}

// PaymentInitResponse represents a response to a payment initialization
type PaymentInitResponse struct {
	// Status indicates if the request was successful
	Status int `json:"status"`

	// Token is the payment token
	Token string `json:"token"`

	// Message contains any message from the API
	Message string `json:"message,omitempty"`

	// Errors contains any error messages
	Errors map[string]string `json:"errors,omitempty"`
}

// PaymentVerifyRequest represents a request to verify a payment
type PaymentVerifyRequest struct {
	// Token is the payment token received during initialization
	Token string `json:"token"`
}

// PaymentVerifyResponse represents a response to a payment verification
type PaymentVerifyResponse struct {
	// Status indicates if the verification was successful (0 or 1)
	Status int `json:"status"`

	// Amount is the verified payment amount
	Amount string `json:"amount,omitempty"`

	// RealAmount is the amount after deducting fees
	RealAmount int64 `json:"realAmount,omitempty"`

	// TransID is the unique payment identifier used for transaction tracking
	TransID int64 `json:"transId,omitempty"`

	// FactorNumber is the invoice/factor number
	FactorNumber string `json:"factorNumber,omitempty"`

	// Mobile is the customer's mobile number
	Mobile string `json:"mobile,omitempty"`

	// Description is the payment description
	Description string `json:"description,omitempty"`

	// CardNumber is the masked card number
	CardNumber string `json:"cardNumber,omitempty"`

	// PaymentDate is when the payment was completed
	PaymentDate string `json:"paymentDate,omitempty"`

	// CID is the SHA256 hash of the card number
	CID string `json:"cid,omitempty"`

	// Message contains the transaction status
	Message string `json:"message,omitempty"`

	// Errors contains any error messages
	Errors map[string]string `json:"errors,omitempty"`
}

// PaymentStatusRequest represents a request to check payment status
type PaymentStatusRequest struct {
	// Token is the payment token
	Token string `json:"token"`
}

// PaymentStatusResponse represents a response to a payment status check
type PaymentStatusResponse struct {
	// Status indicates if the request was successful
	Status bool `json:"status"`

	// Amount is the payment amount
	Amount int64 `json:"amount,omitempty"`

	// TransactionStatus is the status of the transaction
	TransactionStatus string `json:"transactionStatus,omitempty"`

	// RefID is the payment reference ID
	RefID string `json:"refId,omitempty"`

	// Message contains any message from the API
	Message string `json:"message,omitempty"`

	// Errors contains any error messages
	Errors map[string]string `json:"errors,omitempty"`
}

// RefundRequest represents a request to refund a payment
type RefundRequest struct {
	// TransactionID is the ID of the transaction to refund
	TransactionID string `json:"transaction_id"`

	// Amount is the amount to refund (optional, defaults to full amount)
	Amount int64 `json:"amount,omitempty"`
}

// RefundResponse represents a response to a refund request
type RefundResponse struct {
	// Status indicates if the refund was successful
	Status bool `json:"status"`

	// RefundID is the ID of the refund
	RefundID string `json:"refund_id,omitempty"`

	// Amount is the refunded amount
	Amount int64 `json:"amount,omitempty"`

	// Message contains any message from the API
	Message string `json:"message,omitempty"`

	// Errors contains any error messages
	Errors map[string]string `json:"errors,omitempty"`
}

// CallbackData represents the data received in a payment callback
type CallbackData struct {
	// Token is the payment token
	Token string `json:"token"`

	// Status indicates the status of the payment
	Status string `json:"status"`
}

// APIError represents an error returned by the Vandar API
type APIError struct {
	// Message is the error message
	Message string `json:"message"`

	// Code is the error code
	Code string `json:"code,omitempty"`

	// Errors contains detailed error information
	Errors map[string]string `json:"errors,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %s (code: %s)", e.Message, e.Code)
}

// TransactionInfoResponse represents the response from the transaction information endpoint
type TransactionInfoResponse struct {
	Status       int    `json:"status"`
	Amount       string `json:"amount"`
	Wage         string `json:"wage"`
	ShaparakWage string `json:"shaparakWage"`
	TransID      int64  `json:"transId"`
	RefNumber    string `json:"refnumber"`
	TrackingCode string `json:"trackingCode"`
	FactorNumber string `json:"factorNumber"`
	Mobile       string `json:"mobile"`
	Description  string `json:"description"`
	CardNumber   string `json:"cardNumber"`
	CID          string `json:"CID"`
	CreatedAt    string `json:"createdAt"`
	PaymentDate  string `json:"paymentDate"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}
