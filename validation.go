// Package vandargo provides a secure integration with the Vandar payment gateway
// validation.go implements input validation utilities
package vandargo

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Constants for validation
const (
	// MinAmount is the minimum amount in Rials (100 Rials = 1 Toman)
	MinAmount = 10000 // 10,000 Rials

	// MaxAmount is the maximum amount in Rials
	MaxAmount = 5000000000 // 5 billion Rials

	// MaxDescriptionLength is the maximum length for description
	MaxDescriptionLength = 255

	// MinCallbackURLLength is the minimum length for callback URL
	MinCallbackURLLength = 5
)

var (
	// Regular expressions for validation
	cardNumberRegex = regexp.MustCompile(`^[0-9]{16}$`)
	mobileRegex     = regexp.MustCompile(`^09[0-9]{9}$`)
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	ibanRegex       = regexp.MustCompile(`^IR[0-9]{24}$`)
	urlRegex        = regexp.MustCompile(`^https?://[a-zA-Z0-9][-a-zA-Z0-9_.]+\.[a-zA-Z0-9][-a-zA-Z0-9_]+(/[-a-zA-Z0-9_%$.~#&=]*)?$`)
)

// ValidatePaymentInitRequest validates a payment initialization request
func ValidatePaymentInitRequest(req *PaymentInitRequest) error {
	var errors ValidationErrors

	// Validate amount
	if req.Amount < MinAmount {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: fmt.Sprintf("amount must be at least %d Rials", MinAmount),
		})
	}

	if req.Amount > MaxAmount {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: fmt.Sprintf("amount must be at most %d Rials", MaxAmount),
		})
	}

	// Validate callback URL
	if req.CallbackURL == "" {
		errors = append(errors, ValidationError{
			Field:   "callback_url",
			Message: "callback URL is required",
		})
	} else if !urlRegex.MatchString(req.CallbackURL) {
		errors = append(errors, ValidationError{
			Field:   "callback_url",
			Message: "callback URL must be a valid HTTP(S) URL",
		})
	}

	// Validate description (optional)
	if len(req.Description) > MaxDescriptionLength {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength),
		})
	}

	// Validate mobile (optional)
	if req.Mobile != "" && !mobileRegex.MatchString(req.Mobile) {
		errors = append(errors, ValidationError{
			Field:   "mobile",
			Message: "mobile must be a valid Iranian mobile number (e.g., 09123456789)",
		})
	}

	// Validate valid card number (optional)
	if req.ValidCardNumber != "" {
		cleanCard := sanitizeCardNumber(req.ValidCardNumber)
		if !cardNumberRegex.MatchString(cleanCard) {
			errors = append(errors, ValidationError{
				Field:   "valid_card_number",
				Message: "valid card number must be a 16-digit number",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidatePaymentVerifyRequest validates a payment verification request
func ValidatePaymentVerifyRequest(req *PaymentVerifyRequest) error {
	if req.Token == "" {
		return NewValidationError("token", "token is required")
	}

	return nil
}

// ValidatePaymentStatusRequest validates a payment status request
func ValidatePaymentStatusRequest(req *PaymentStatusRequest) error {
	if req.Token == "" {
		return NewValidationError("token", "token is required")
	}

	return nil
}

// ValidateRefundRequest validates a refund request
func ValidateRefundRequest(req *RefundRequest) error {
	var errors ValidationErrors

	if req.TransactionID == "" {
		errors = append(errors, ValidationError{
			Field:   "transaction_id",
			Message: "transaction ID is required",
		})
	}

	if req.Amount < 0 {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "amount must be a positive number",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateCallbackData validates data received in a callback
func ValidateCallbackData(data *CallbackData) error {
	if data.Token == "" {
		return NewValidationError("token", "token is required")
	}

	return nil
}

// ValidateIBAN validates an IBAN (International Bank Account Number)
func ValidateIBAN(iban string) error {
	if !ibanRegex.MatchString(iban) {
		return errors.New("invalid IBAN format, must start with IR followed by 24 digits")
	}

	return nil
}

// SanitizeInput sanitizes a string input to prevent injection attacks
func SanitizeInput(input string) string {
	// Remove any control characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ValidateAmount validates that a string represents a valid amount
func ValidateAmount(amount string) (int64, error) {
	// Remove any non-digit characters (like commas)
	cleanAmount := ""
	for _, r := range amount {
		if r >= '0' && r <= '9' {
			cleanAmount += string(r)
		}
	}

	// Convert to int64
	amountInt, err := strconv.ParseInt(cleanAmount, 10, 64)
	if err != nil {
		return 0, errors.New("invalid amount format")
	}

	// Check range
	if amountInt < MinAmount {
		return 0, fmt.Errorf("amount must be at least %d Rials", MinAmount)
	}

	if amountInt > MaxAmount {
		return 0, fmt.Errorf("amount must be at most %d Rials", MaxAmount)
	}

	return amountInt, nil
}
