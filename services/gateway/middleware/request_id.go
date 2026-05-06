package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use existing X-Request-ID from upstream (e.g., Nginx) or generate new
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set on response header for client tracing
		w.Header().Set("X-Request-ID", requestID)

		// Store in context for downstream handlers
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts request ID from context for logging
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// LogWithRequestID returns a slog logger with request_id attribute
func LogWithRequestID(ctx context.Context) *slog.Logger {
	return slog.With("request_id", GetRequestID(ctx))
}

func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to simple counter if crypto/rand fails
		return "fallback"
	}
	return hex.EncodeToString(b)
}
