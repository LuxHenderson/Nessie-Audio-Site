# Comprehensive Error Handling Documentation

## Overview

The Nessie Audio backend now has a robust, production-ready error handling system that provides:
- Standardized error responses across all endpoints
- Request ID tracking for debugging
- Proper error logging without exposing internal details
- Panic recovery with graceful error responses
- Input validation helpers
- Circuit breaker error handling for external APIs

## Architecture

### 1. Error Response Format

All API errors follow this standardized JSON structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "request_id": "uuid-for-tracking",
  "timestamp": "2026-01-12T15:04:40-06:00",
  "details": {}  // Optional, for validation errors
}
```

**Key Features:**
- `error`: User-friendly message (never exposes internal details)
- `code`: Machine-readable error code for client-side handling
- `request_id`: Unique ID for tracking this request in logs
- `timestamp`: When the error occurred (ISO 8601 format)
- `details`: Additional context (e.g., validation errors)

### 2. Error Codes

Located in `internal/errors/errors.go`:

**Client Errors (4xx):**
- `BAD_REQUEST` - Invalid request format or missing required fields
- `VALIDATION_ERROR` - Input validation failed (includes field-specific details)
- `NOT_FOUND` - Resource doesn't exist
- `CONFLICT` - Business logic violation (e.g., out of stock)
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions

**Server Errors (5xx):**
- `INTERNAL_ERROR` - Unexpected server error (hides internal details)
- `SERVICE_UNAVAILABLE` - External service down or circuit breaker open
- `DATABASE_ERROR` - Database connectivity issue
- `EXTERNAL_API_ERROR` - Stripe/Printful API failure

### 3. Request ID Tracking

Every request automatically gets a unique UUID that:
- Appears in the `X-Request-ID` response header
- Is included in all error responses
- Is logged with all error messages
- Allows tracing a request through the entire system

**Middleware:** `internal/middleware/middleware.go:RequestID()`

Example usage in handlers:
```go
requestID := middleware.GetRequestID(r.Context())
h.logger.Error("Failed to query database [request_id: "+requestID+"]", err)
```

### 4. Panic Recovery

The recovery middleware catches all panics and returns a safe error response:

**Before panic recovery:**
```
Server crashes with stack trace exposed to client
```

**After panic recovery:**
```json
{
  "error": "An unexpected error occurred",
  "code": "INTERNAL_ERROR",
  "request_id": "...",
  "timestamp": "..."
}
```

The full stack trace is logged server-side for debugging.

**Middleware:** `internal/middleware/middleware.go:Recovery()`

### 5. Input Validation

The validation package provides reusable validators:

**Location:** `internal/validation/validation.go`

**Example usage:**
```go
import "github.com/nessieaudio/ecommerce-backend/internal/validation"

v := validation.NewValidator()
v.RequireString("name", name)
v.RequireEmail("email", email)
v.RequirePositiveInt("quantity", quantity)

if v.HasErrors() {
    apierrors.RespondValidationError(w, v.Errors(), requestID)
    return
}
```

**Available validators:**
- `RequireString()` - Non-empty string
- `RequireEmail()` - Valid email format
- `RequirePhone()` - Valid phone format
- `RequirePositiveInt()` - Positive integer
- `RequirePositiveFloat()` - Positive float
- `RequireMinLength()` - Minimum string length
- `RequireMaxLength()` - Maximum string length
- `RequireOneOf()` - Value from allowed list
- `ValidateAddress()` - Complete shipping address

### 6. Logging Strategy

**Internal errors are logged, not exposed:**

```go
// ❌ BAD - Exposes internal details
if err != nil {
    http.Error(w, err.Error(), 500)
    return
}

// ✅ GOOD - Logs internally, safe message to user
if err != nil {
    h.logger.Error("Failed to query products [request_id: "+requestID+"]", err)
    apierrors.RespondInternalError(w, requestID)
    return
}
```

**Log levels:**
- `Info()` - Normal operations
- `Warning()` - Recoverable issues
- `Error()` - Errors that were handled
- `Critical()` - Severe errors (sends email alert to admin)

## Implementation Examples

### Example 1: Basic Error Handling

```go
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())

    productID := mux.Vars(r)["id"]
    if productID == "" {
        apierrors.RespondError(w, http.StatusBadRequest,
            "Product ID is required",
            apierrors.ErrCodeBadRequest, nil, requestID)
        return
    }

    product, err := h.db.GetProduct(productID)
    if err == sql.ErrNoRows {
        apierrors.RespondNotFound(w, "Product", requestID)
        return
    }
    if err != nil {
        h.logger.Error("Database query failed [request_id: "+requestID+"]", err)
        apierrors.RespondInternalError(w, requestID)
        return
    }

    apierrors.RespondJSON(w, http.StatusOK, product)
}
```

### Example 2: Validation Errors

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())

    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        apierrors.RespondError(w, http.StatusBadRequest,
            "Invalid JSON", apierrors.ErrCodeBadRequest, nil, requestID)
        return
    }

    // Validate input
    v := validation.NewValidator()
    v.RequireEmail("email", req.Email)
    v.RequireString("name", req.Name)
    v.RequirePositiveInt("quantity", req.Quantity)

    if v.HasErrors() {
        apierrors.RespondValidationError(w, v.Errors(), requestID)
        return
    }

    // Process order...
}
```

**Validation error response:**
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": [
    {
      "field": "email",
      "message": "email must be a valid email address"
    },
    {
      "field": "quantity",
      "message": "quantity must be a positive number"
    }
  ],
  "request_id": "...",
  "timestamp": "..."
}
```

### Example 3: External API Errors

```go
func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())

    sessionID, err := h.stripeClient.CreateCheckoutSession(req)
    if err != nil {
        // Check if circuit breaker is open
        if err == circuitbreaker.ErrCircuitOpen {
            h.logger.Warning("Stripe circuit breaker open", err)
            apierrors.RespondServiceUnavailable(w, "Payment service", requestID)
            return
        }

        h.logger.Error("Stripe checkout failed [request_id: "+requestID+"]", err)
        apierrors.RespondError(w, http.StatusBadGateway,
            "Unable to process payment at this time",
            apierrors.ErrCodeExternalAPIError, nil, requestID)
        return
    }

    apierrors.RespondJSON(w, http.StatusOK, map[string]string{
        "session_id": sessionID,
    })
}
```

## Testing Error Handling

### Manual Testing

```bash
# Test successful request
curl -i http://localhost:8080/api/v1/products

# Test 404 error
curl http://localhost:8080/api/v1/products/nonexistent-id | jq

# Test validation error
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid","quantity":-1}' | jq
```

### Automated Testing

```bash
go run cmd/test-error-handling/main.go
```

This runs comprehensive tests covering:
- Successful requests with request IDs
- 404 error responses
- Request ID uniqueness
- Health check responses

## Benefits

### For Development
- ✅ Consistent error handling across all endpoints
- ✅ Easy debugging with request ID tracking
- ✅ Reusable validation helpers reduce boilerplate
- ✅ Panic recovery prevents server crashes

### For Production
- ✅ Never exposes internal error details to clients
- ✅ Proper HTTP status codes for client-side error handling
- ✅ Complete error logging for troubleshooting
- ✅ Request tracing across the entire request lifecycle

### For Users
- ✅ Clear, actionable error messages
- ✅ Field-specific validation feedback
- ✅ Graceful degradation when services are unavailable

## Error Response Examples

### 404 Not Found
```json
{
  "error": "Product not found",
  "code": "NOT_FOUND",
  "request_id": "853a310e-881d-46bf-8198-78b00bac569e",
  "timestamp": "2026-01-12T15:03:35-06:00"
}
```

### 400 Validation Error
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": [
    {"field": "email", "message": "email must be a valid email address"},
    {"field": "quantity", "message": "quantity must be a positive number"}
  ],
  "request_id": "...",
  "timestamp": "..."
}
```

### 500 Internal Error
```json
{
  "error": "An internal error occurred. Please try again later.",
  "code": "INTERNAL_ERROR",
  "request_id": "7b3865cb-76f3-4d1c-a787-f630f7b48847",
  "timestamp": "2026-01-12T15:04:40-06:00"
}
```

### 503 Service Unavailable
```json
{
  "error": "Payment service is temporarily unavailable",
  "code": "SERVICE_UNAVAILABLE",
  "request_id": "...",
  "timestamp": "..."
}
```

## Maintenance

### Adding New Error Types

1. Add error code to `internal/errors/errors.go`:
```go
const (
    ErrCodeRateLimited ErrorCode = "RATE_LIMITED"
)
```

2. Create helper function (optional):
```go
func RespondRateLimited(w http.ResponseWriter, requestID string) {
    RespondError(w, http.StatusTooManyRequests,
        "Rate limit exceeded. Please try again later.",
        ErrCodeRateLimited, nil, requestID)
}
```

3. Use in handlers:
```go
if rateLimited {
    apierrors.RespondRateLimited(w, requestID)
    return
}
```

### Updating Existing Handlers

Pattern to follow:
1. Get request ID: `requestID := middleware.GetRequestID(r.Context())`
2. Validate input with validation helpers
3. Log errors: `h.logger.Error("message [request_id: "+requestID+"]", err)`
4. Return safe errors: `apierrors.Respond*(w, requestID)`
5. Use `apierrors.RespondJSON()` for success responses

## Summary

The comprehensive error handling system provides:
- **Standardization** - All errors follow the same format
- **Traceability** - Every request has a unique ID
- **Security** - Internal details never exposed to clients
- **Reliability** - Panic recovery prevents crashes
- **Debuggability** - Complete error logging with context
- **Usability** - Clear, actionable error messages for users

This system is production-ready and provides the foundation for maintaining and debugging the Nessie Audio eCommerce platform.
