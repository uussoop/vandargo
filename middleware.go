// Package vandargo provides a secure integration with the Vandar payment gateway
// middleware.go implements security middleware for the API
package vandargo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Middleware represents a function that wraps an HTTP handler
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain applies multiple middleware to a handler in sequence
func Chain(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// LoggingMiddleware logs request information
func LoggingMiddleware(logger LoggerInterface) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response wrapper to capture status code
			rw := newResponseWriter(w)

			// Process request
			next(rw, r)

			// Log request details
			duration := time.Since(start)
			logger.Info(r.Context(), "HTTP Request", map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     rw.status,
				"duration":   duration.Milliseconds(),
				"user_agent": r.UserAgent(),
				"remote_ip":  getClientIP(r),
			})
		}
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			next(w, r)
		}
	}
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(limit int, window time.Duration) Middleware {
	// A simple in-memory rate limiter
	type client struct {
		count    int
		lastSeen time.Time
	}

	clients := make(map[string]*client)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			now := time.Now()

			// Get or create client record
			c, exists := clients[ip]
			if !exists || now.Sub(c.lastSeen) > window {
				// Reset count if window has passed
				clients[ip] = &client{
					count:    1,
					lastSeen: now,
				}
				next(w, r)
				return
			}

			// Update client record
			c.lastSeen = now
			c.count++

			// Check if rate limit is exceeded
			if c.count > limit {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

// IPFilterMiddleware filters requests by IP allowlist
func IPFilterMiddleware(config ConfigInterface) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// If allowlist is empty, allow all IPs
			if len(config.(*configImpl).config.IPAllowList) == 0 {
				next(w, r)
				return
			}

			// Get client IP
			ip := getClientIP(r)

			// Check if IP is allowed
			allowed := false
			for _, allowedIP := range config.(*configImpl).config.IPAllowList {
				if ip == allowedIP {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

// AuthMiddleware validates API key
func AuthMiddleware(config ConfigInterface) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			// Check if Authorization header exists
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if Authorization header format is valid
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			// Check if API key is valid
			if parts[1] != config.GetAPIKey() {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}

// ValidateSignatureMiddleware validates request signature
func ValidateSignatureMiddleware(config ConfigInterface) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST and PUT requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				next(w, r)
				return
			}

			// Get signature from header
			signature := r.Header.Get("X-Signature")
			if signature == "" {
				http.Error(w, "Missing signature", http.StatusUnauthorized)
				return
			}

			// Get timestamp from header
			timestamp := r.Header.Get("X-Timestamp")
			if timestamp == "" {
				http.Error(w, "Missing timestamp", http.StatusUnauthorized)
				return
			}

			// Verify timestamp is recent (within 5 minutes)
			timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
				return
			}

			now := time.Now().Unix()
			if now-timestampInt > 300 || timestampInt-now > 300 {
				http.Error(w, "Timestamp expired", http.StatusUnauthorized)
				return
			}

			// Create signature string
			signatureData := fmt.Sprintf("%s:%s:%s", r.URL.Path, timestamp, config.GetAPIKey())

			// Verify signature
			if !VerifySignature(signature, signatureData, config.GetAPIKey()) {
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}

// RequestIDMiddleware adds a request ID to each request context
func RequestIDMiddleware() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate a new one
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)

			// Call next handler with updated context
			next(w, r.WithContext(ctx))
		}
	}
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// newResponseWriter creates a new response writer
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader captures the status code before writing it
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP gets the client IP from the request
func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first (for clients behind proxies)
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Get the first IP in the list
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Try X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Finally, use RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
