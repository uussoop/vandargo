// Package vandargo provides a secure integration with the Vandar payment gateway
// interfaces.go contains all the interfaces needed for package integration points
package vandargo

import (
	"context"
	"net/http"
)

// StorageInterface defines methods for data persistence operations
type StorageInterface interface {
	// StoreTransaction saves a new transaction to storage
	StoreTransaction(ctx context.Context, transaction *Transaction) error

	// GetTransaction retrieves a transaction by ID
	GetTransaction(ctx context.Context, id string) (*Transaction, error)

	// UpdateTransaction updates an existing transaction
	UpdateTransaction(ctx context.Context, transaction *Transaction) error

	// GetTransactionsByStatus retrieves transactions by their status
	GetTransactionsByStatus(ctx context.Context, status string) ([]*Transaction, error)
}

// LoggerInterface defines methods for logging operations
type LoggerInterface interface {
	// Debug logs debug level messages
	Debug(ctx context.Context, message string, fields map[string]interface{})

	// Info logs informational messages
	Info(ctx context.Context, message string, fields map[string]interface{})

	// Warn logs warning messages
	Warn(ctx context.Context, message string, fields map[string]interface{})

	// Error logs error messages
	Error(ctx context.Context, message string, err error, fields map[string]interface{})
}

// ConfigInterface defines methods for configuration operations
type ConfigInterface interface {
	// GetAPIKey returns the Vandar API key
	GetAPIKey() string

	// GetBaseURL returns the base URL for the Vandar API
	GetBaseURL() string

	// IsSandboxMode returns whether the integration is in sandbox mode
	IsSandboxMode() bool

	// GetTimeout returns the HTTP client timeout duration
	GetTimeout() int

	// GetCallbackURL returns the URL for payment callbacks
	GetCallbackURL() string
}

// HTTPClientInterface defines methods for making HTTP requests
type HTTPClientInterface interface {
	// Do executes an HTTP request and returns an HTTP response
	Do(req *http.Request) (*http.Response, error)
}

// RouterInterface defines methods for registering HTTP routes
type RouterInterface interface {
	// POST registers a POST route with a handler
	POST(path string, handler http.HandlerFunc)

	// GET registers a GET route with a handler
	GET(path string, handler http.HandlerFunc)
}

// PaymentServiceInterface defines methods for payment operations
type PaymentServiceInterface interface {
	// InitiatePayment starts a new payment transaction
	InitiatePayment(ctx context.Context, amount int, description string, metadata map[string]string) (*PaymentInitResponse, error)

	// VerifyPayment verifies a payment transaction
	VerifyPayment(ctx context.Context, token string) (*PaymentVerifyResponse, error)

	// GetPaymentStatus checks the status of a payment
	GetPaymentStatus(ctx context.Context, token string) (*PaymentStatusResponse, error)

	// RefundPayment initiates a refund for a transaction
	RefundPayment(ctx context.Context, transactionID string, amount int) (*RefundResponse, error)
}
