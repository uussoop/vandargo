// Package vandargo provides a secure integration with the Vandar payment gateway
// crypto.go implements cryptographic utilities for secure communication
package vandargo

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// SignData signs data using HMAC-SHA256
func SignData(data string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies that a signature is valid for the given data
func VerifySignature(signature, data, key string) bool {
	expectedSignature := SignData(data, key)
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) == 1
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("number of bytes must be positive")
	}

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(n int) (string, error) {
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes)[:n], nil
}

// GenerateNonce generates a random nonce for API requests
func GenerateNonce() string {
	nonce, err := GenerateRandomString(16)
	if err != nil {
		// If random generation fails, use a timestamp-based approach as fallback
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return nonce
}

// HashCardNumber securely hashes a card number using SHA-256
func HashCardNumber(cardNumber string) string {
	// Remove spaces and non-digit characters
	cleanCard := sanitizeCardNumber(cardNumber)

	// Hash the card number
	hash := sha256.Sum256([]byte(cleanCard))
	return hex.EncodeToString(hash[:])
}

// MaskCardNumber masks a card number showing only the last 4 digits
func MaskCardNumber(cardNumber string) string {
	// Remove spaces and non-digit characters
	cleanCard := sanitizeCardNumber(cardNumber)

	if len(cleanCard) < 4 {
		return "****"
	}

	masked := ""
	for i := 0; i < len(cleanCard)-4; i++ {
		masked += "*"
	}

	return masked + cleanCard[len(cleanCard)-4:]
}

// sanitizeCardNumber removes spaces and non-digit characters from a card number
func sanitizeCardNumber(cardNumber string) string {
	var clean []rune

	for _, r := range cardNumber {
		if r >= '0' && r <= '9' {
			clean = append(clean, r)
		}
	}

	return string(clean)
}

// VerifyCallbackIP checks if the IP is in the allowed list
func VerifyCallbackIP(ip string, allowList []string) bool {
	if len(allowList) == 0 {
		return true // No restrictions if list is empty
	}

	for _, allowedIP := range allowList {
		if ip == allowedIP {
			return true
		}
	}

	return false
}
