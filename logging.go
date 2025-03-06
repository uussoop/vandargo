// Package vandargo provides a secure integration with the Vandar payment gateway
// logging.go implements logging utilities
package vandargo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// defaultLogger is a simple implementation of LoggerInterface
type defaultLogger struct {
	// logLevel defines the minimum level of logs to output
	logLevel string
}

// LogLevel represents log severity levels
type LogLevel int

const (
	// Debug is for detailed information, typically useful only when diagnosing problems
	Debug LogLevel = iota
	// Info is for confirmation that things are working as expected
	Info
	// Warn is for situations that might cause problems
	Warn
	// Error is for error situations
	Error
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

// NewDefaultLogger creates a new default logger with the specified log level
func NewDefaultLogger(level string) LoggerInterface {
	return &defaultLogger{
		logLevel: level,
	}
}

// shouldLog determines whether the log should be output based on log level
func (l *defaultLogger) shouldLog(level LogLevel) bool {
	switch l.logLevel {
	case "DEBUG":
		return true
	case "INFO":
		return level >= Info
	case "WARN":
		return level >= Warn
	case "ERROR":
		return level >= Error
	default:
		return level >= Info // Default to INFO level
	}
}

// formatLog formats a log entry as JSON
func (l *defaultLogger) formatLog(ctx context.Context, level LogLevel, message string, err error, fields map[string]interface{}) string {
	// Create log entry
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level.String(),
		"message":   message,
	}

	// Add request ID if available
	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			entry["request_id"] = requestID
		}
	}

	// Add error if available
	if err != nil {
		entry["error"] = err.Error()
	}

	// Add fields if available
	if fields != nil {
		// Sanitize sensitive data
		sanitizedFields := l.sanitizeSensitiveData(fields)
		for k, v := range sanitizedFields {
			entry[k] = v
		}
	}

	// Marshal to JSON
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf("Failed to marshal log entry: %v", err)
	}

	return string(jsonEntry)
}

// sanitizeSensitiveData masks sensitive information in log fields
func (l *defaultLogger) sanitizeSensitiveData(fields map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	sensitiveKeys := []string{
		"card_number", "cardNumber", "card",
		"password", "secret", "token", "api_key",
		"authorization", "auth", "api_secret",
		"credit_card", "cvv", "cvc", "pin",
	}

	for k, v := range fields {
		// Check if this is a sensitive field
		isSensitive := false
		for _, sensitiveKey := range sensitiveKeys {
			if k == sensitiveKey {
				isSensitive = true
				break
			}
		}

		// Handle sensitive fields
		if isSensitive {
			switch value := v.(type) {
			case string:
				if len(value) > 4 {
					sanitized[k] = MaskCardNumber(value)
				} else {
					sanitized[k] = "****"
				}
			default:
				sanitized[k] = "****"
			}
		} else {
			// For non-sensitive fields, check if it's a nested map
			if nestedMap, ok := v.(map[string]interface{}); ok {
				sanitized[k] = l.sanitizeSensitiveData(nestedMap)
			} else {
				sanitized[k] = v
			}
		}
	}

	return sanitized
}

// Debug logs debug level messages
func (l *defaultLogger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	if !l.shouldLog(Debug) {
		return
	}

	fmt.Fprintln(os.Stdout, l.formatLog(ctx, Debug, message, nil, fields))
}

// Info logs informational messages
func (l *defaultLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	if !l.shouldLog(Info) {
		return
	}

	fmt.Fprintln(os.Stdout, l.formatLog(ctx, Info, message, nil, fields))
}

// Warn logs warning messages
func (l *defaultLogger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	if !l.shouldLog(Warn) {
		return
	}

	fmt.Fprintln(os.Stderr, l.formatLog(ctx, Warn, message, nil, fields))
}

// Error logs error messages
func (l *defaultLogger) Error(ctx context.Context, message string, err error, fields map[string]interface{}) {
	if !l.shouldLog(Error) {
		return
	}

	fmt.Fprintln(os.Stderr, l.formatLog(ctx, Error, message, err, fields))
}
