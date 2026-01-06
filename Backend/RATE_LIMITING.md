# Rate Limiting Implementation

## Overview

Implemented **Token Bucket** rate limiting to protect the Nessie Audio API from abuse and ensure fair resource usage.

## Algorithm: Token Bucket

The token bucket algorithm is the industry standard used by AWS, Stripe, and GitHub. It provides:
- **Burst tolerance**: Allows legitimate users to make quick bursts of requests
- **Smooth rate control**: Maintains average request rate over time
- **Memory efficiency**: Stores only bucket state per IP, not every request

### How It Works

1. Each IP address gets a "bucket" with tokens
2. Each request consumes 1 token
3. Tokens refill at a constant rate
4. If no tokens available → Request denied with 429 status

## Rate Limits by Endpoint

| Endpoint | Capacity | Refill Rate | Effective Limit | Use Case |
|----------|----------|-------------|-----------------|----------|
| `/api/v1/products` | 100 tokens | 2/sec | ~120/min | Public browsing |
| `/api/v1/products/{id}` | 100 tokens | 2/sec | ~120/min | Public browsing |
| `/api/v1/config` | 60 tokens | 1/sec | ~60/min | General API |
| `/api/v1/orders` (GET) | 60 tokens | 1/sec | ~60/min | Order lookup |
| `/api/v1/orders` (POST) | 20 tokens | 0.33/sec | ~20/min | Order creation |
| `/api/v1/checkout` | 20 tokens | 0.33/sec | ~20/min | Checkout |
| `/api/v1/cart/checkout` | 20 tokens | 0.33/sec | ~20/min | Cart checkout |
| `/webhooks/*` | **No limit** | - | Unlimited | Stripe/Printful |
| `/health` | **No limit** | - | Unlimited | Monitoring |

## Response Headers

Every rate-limited response includes:

```http
X-RateLimit-Limit: 60           # Maximum tokens in bucket
X-RateLimit-Remaining: 42       # Tokens remaining
```

When rate limited (429 status):

```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
Retry-After: 3                  # Seconds until next token available

{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please retry after 3 seconds.",
  "retry_after": 3
}
```

## Implementation Details

### Files Created/Modified

1. **`internal/middleware/ratelimit.go`** (new)
   - Token bucket implementation
   - IP-based rate limiting
   - Automatic cleanup to prevent memory leaks

2. **`internal/handlers/handlers.go`** (modified)
   - Applied rate limiters to routes
   - Different limits for different endpoint types

### Key Features

**IP Detection**: Properly handles proxy headers
```go
X-Forwarded-For  → Original client IP (proxy/load balancer)
X-Real-IP        → Real client IP
RemoteAddr       → Fallback to direct connection IP
```

**Memory Management**: Auto-cleanup every 5 minutes
- Removes stale buckets (inactive >10 minutes)
- Prevents memory leaks from one-time visitors

**Thread-Safe**: Uses sync.RWMutex for concurrent access

## Testing

### Test Scripts

1. **`cmd/test-ratelimit/main.go`**
   - Basic rate limit testing
   - Tests multiple endpoints
   - Verifies headers

2. **`cmd/test-ratelimit-aggressive/main.go`**
   - Aggressive testing (150 requests)
   - Guaranteed to trigger rate limits
   - Shows exact limit thresholds

### Running Tests

**After restarting the server**, run:

```bash
# Basic test
go run cmd/test-ratelimit/main.go

# Aggressive test (will definitely hit limits)
go run cmd/test-ratelimit-aggressive/main.go
```

## Example Usage Scenarios

### Scenario 1: Normal Customer Browsing
```
Customer views products page     → 1 token  (99 remaining)
Clicks on product detail         → 1 token  (98 remaining)
Views another product            → 1 token  (97 remaining)
... continues browsing ...
Tokens refill 2/sec in background
```
**Result**: ✅ Smooth experience, never hits limit

### Scenario 2: Attacker Scraping
```
Bot makes 100 rapid requests     → 100 tokens (0 remaining)
Bot makes 101st request          → ❌ 429 Too Many Requests
Bot makes 102nd request          → ❌ 429 Too Many Requests
... after 30 seconds ...
30 tokens have refilled          → 30 more requests allowed
```
**Result**: ✅ Bot is throttled, legitimate users unaffected

### Scenario 3: Accidental Double-Click
```
User clicks "Add to Cart" twice quickly
Request 1: ✅ Processed (19 tokens left)
Request 2: ✅ Processed (18 tokens left)
```
**Result**: ✅ Both requests succeed (burst tolerance)

## Security Benefits

1. **DDoS Protection**: Prevents overwhelming the server
2. **Brute Force Prevention**: Limits checkout attempts
3. **Cost Control**: Limits expensive Stripe/Printful API calls
4. **Fair Usage**: Ensures all customers get equal access

## Production Deployment

### Current Setup (Single Server)
- ✅ In-memory storage (fast, simple)
- ✅ Automatic cleanup
- ✅ No external dependencies

### Future Scaling (Multiple Servers)
If you add multiple backend servers, upgrade to Redis:

```go
// Instead of in-memory map, use Redis
import "github.com/go-redis/redis/v8"

// Redis-backed rate limiter (shared across servers)
```

## Monitoring

Rate limit events are logged:

```
Rate limit exceeded for IP 192.168.1.100 (retry after 5s)
Rate limiter cleanup: 42 active IPs
```

Monitor these logs to:
- Detect attacks (many IPs hitting limits)
- Identify legitimate users needing higher limits
- Tune rate limit thresholds

## Exemptions

These endpoints are **NOT** rate limited:
- `/webhooks/stripe` - Stripe requires reliable webhook delivery
- `/webhooks/printful/{token}` - Printful requires reliable webhook delivery
- `/health` - Used by monitoring tools

## Configuration

To adjust limits, modify `internal/handlers/handlers.go`:

```go
// Example: Increase products limit to 200/min
publicLimiter := middleware.RateLimit(200, 3.33)  // 200 tokens, 3.33/sec refill
```

## Best Practices

1. **Monitor logs** for rate limit events
2. **Adjust limits** based on real usage patterns
3. **Whitelist trusted IPs** if needed (future enhancement)
4. **Use Redis** if deploying multiple servers
5. **Test thoroughly** after any limit changes
