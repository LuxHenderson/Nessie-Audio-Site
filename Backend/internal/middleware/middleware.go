package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	apierrors "github.com/nessieaudio/ecommerce-backend/internal/errors"
)

// CORS middleware allows cross-origin requests
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if wildcard is allowed
			if strings.TrimSpace(allowedOrigins) == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// Check if origin is allowed
				allowed := false
				for _, o := range origins {
					if strings.TrimSpace(o) == origin {
						allowed = true
						break
					}
				}

				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Logging middleware logs all requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Call next handler
		next.ServeHTTP(w, r)

		// Log duration
		log.Printf("Completed in %v", time.Since(start))
	})
}

// RequestIDKey is the context key for request IDs
type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestID middleware adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add to request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add to response headers for debugging
		w.Header().Set("X-Request-ID", requestID)

		// Continue with modified request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// Recovery middleware recovers from panics with detailed logging
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())

				// Log panic with stack trace
				log.Printf("PANIC [request_id=%s]: %v\n%s", requestID, err, debug.Stack())

				// Return standardized error response
				apierrors.RespondError(
					w,
					http.StatusInternalServerError,
					"An unexpected error occurred",
					apierrors.ErrCodeInternal,
					nil,
					requestID,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
