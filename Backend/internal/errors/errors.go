package errors

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"

	// Server errors (5xx)
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPIError   ErrorCode = "EXTERNAL_API_ERROR"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Error     string      `json:"error"`
	Code      string      `json:"code,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// RespondJSON writes a JSON response
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// RespondError writes a standardized error response
func RespondError(w http.ResponseWriter, status int, message string, code ErrorCode, details interface{}, requestID string) {
	errorResponse := ErrorResponse{
		Error:     message,
		Code:      string(code),
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	RespondJSON(w, status, errorResponse)
}

// RespondValidationError writes validation errors
func RespondValidationError(w http.ResponseWriter, errors []ValidationError, requestID string) {
	errorResponse := ErrorResponse{
		Error:     "Validation failed",
		Code:      string(ErrCodeValidation),
		Details:   errors,
		RequestID: requestID,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	RespondJSON(w, http.StatusBadRequest, errorResponse)
}

// RespondNotFound writes a 404 error
func RespondNotFound(w http.ResponseWriter, resource string, requestID string) {
	RespondError(w, http.StatusNotFound, resource+" not found", ErrCodeNotFound, nil, requestID)
}

// RespondInternalError writes a 500 error (hides internal details from user)
func RespondInternalError(w http.ResponseWriter, requestID string) {
	RespondError(w, http.StatusInternalServerError, "An internal error occurred. Please try again later.", ErrCodeInternal, nil, requestID)
}

// RespondServiceUnavailable writes a 503 error
func RespondServiceUnavailable(w http.ResponseWriter, service string, requestID string) {
	RespondError(w, http.StatusServiceUnavailable, service+" is temporarily unavailable", ErrCodeServiceUnavailable, nil, requestID)
}
