package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TokenBucket represents a rate limiter using the token bucket algorithm
type TokenBucket struct {
	capacity     float64   // Maximum tokens in bucket
	tokens       float64   // Current token count
	refillRate   float64   // Tokens added per second
	lastRefill   time.Time // Last time tokens were refilled
	mu           sync.Mutex
}

// RateLimiter manages rate limiting for different IP addresses
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex

	// Default limits
	defaultCapacity   float64
	defaultRefillRate float64

	// Cleanup
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// NewRateLimiter creates a new rate limiter with token bucket algorithm
// capacity: maximum number of tokens (max burst size)
// refillRate: tokens added per second
func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:           make(map[string]*TokenBucket),
		defaultCapacity:   capacity,
		defaultRefillRate: refillRate,
		cleanupInterval:   5 * time.Minute,
		lastCleanup:       time.Now(),
	}

	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request should be allowed for the given IP
// Returns (allowed bool, remainingTokens int, retryAfter int)
func (rl *RateLimiter) Allow(ip string) (bool, int, int) {
	rl.mu.Lock()
	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = &TokenBucket{
			capacity:   rl.defaultCapacity,
			tokens:     rl.defaultCapacity,
			refillRate: rl.defaultRefillRate,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	timePassed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = min(bucket.capacity, bucket.tokens+(timePassed*bucket.refillRate))
	bucket.lastRefill = now

	// Check if we have at least 1 token
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true, int(bucket.tokens), 0
	}

	// Calculate retry-after (seconds until next token)
	retryAfter := int((1.0 - bucket.tokens) / bucket.refillRate)
	if retryAfter < 1 {
		retryAfter = 1
	}

	return false, 0, retryAfter
}

// cleanupLoop removes stale buckets to prevent memory leaks
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes buckets that haven't been accessed in a while
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	staleThreshold := 10 * time.Minute

	for ip, bucket := range rl.buckets {
		bucket.mu.Lock()
		timeSinceAccess := now.Sub(bucket.lastRefill)
		bucket.mu.Unlock()

		if timeSinceAccess > staleThreshold {
			delete(rl.buckets, ip)
		}
	}

	rl.lastCleanup = now
	log.Printf("Rate limiter cleanup: %d active IPs", len(rl.buckets))
}

// RateLimit creates a rate limiting middleware
// capacity: max tokens (burst size)
// refillRate: tokens per second
func RateLimit(capacity, refillRate float64) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(capacity, refillRate)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Check rate limit
			allowed, remaining, retryAfter := limiter.Allow(ip)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", capacity))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				response := fmt.Sprintf(`{"error":"Rate limit exceeded","message":"Too many requests. Please retry after %d seconds.","retry_after":%d}`, retryAfter, retryAfter)
				w.Write([]byte(response))

				log.Printf("Rate limit exceeded for IP %s (retry after %ds)", ip, retryAfter)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from the request
// Handles X-Forwarded-For, X-Real-IP headers (for proxies/load balancers)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (comma-separated list)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Get first IP in the list (original client)
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
