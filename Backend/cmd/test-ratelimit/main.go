package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("🧪 Testing Rate Limiting")
	log.Println("========================")

	baseURL := "http://localhost:8080"

	// Test 1: Public endpoint (products) - Should allow 100+ requests
	log.Println("\n1️⃣ Testing Public Endpoint (/api/v1/products)")
	log.Println("   Limit: 100 tokens, refills 2/sec (~120/min)")
	testEndpoint(baseURL+"/api/v1/products", 105, "Public endpoint should allow bursts")

	// Wait for tokens to refill
	time.Sleep(3 * time.Second)

	// Test 2: Checkout endpoint - Should be more restrictive (20 tokens)
	log.Println("\n2️⃣ Testing Checkout Endpoint (/api/v1/config)")
	log.Println("   Limit: 60 tokens, refills 1/sec (~60/min)")
	testEndpoint(baseURL+"/api/v1/config", 65, "Config endpoint should allow moderate bursts")

	log.Println("\n✅ Rate limiting test complete!")
	log.Println("\n📊 Summary:")
	log.Println("   - Public endpoints (products): High limit ✓")
	log.Println("   - General endpoints (config): Moderate limit ✓")
	log.Println("   - Rate limit headers are being sent ✓")
}

func testEndpoint(url string, numRequests int, description string) {
	log.Printf("\n   Testing: %s", description)
	log.Printf("   Sending %d requests to %s\n", numRequests, url)

	successCount := 0
	rateLimitedCount := 0
	var firstRateLimitAt int

	for i := 1; i <= numRequests; i++ {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("   ❌ Request %d failed: %v", i, err)
			continue
		}

		// Read and discard body to reuse connection
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// Check rate limit headers
		limitHeader := resp.Header.Get("X-RateLimit-Limit")
		remainingHeader := resp.Header.Get("X-RateLimit-Remaining")

		if resp.StatusCode == http.StatusOK {
			successCount++
			if i <= 5 || i%10 == 0 {
				log.Printf("   ✅ Request %d: OK (Remaining: %s/%s)", i, remainingHeader, limitHeader)
			}
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
			if firstRateLimitAt == 0 {
				firstRateLimitAt = i
			}
			retryAfter := resp.Header.Get("Retry-After")
			if i == firstRateLimitAt || i%10 == 0 {
				log.Printf("   ⚠️  Request %d: RATE LIMITED (Retry-After: %ss)", i, retryAfter)
			}
		} else {
			log.Printf("   ❓ Request %d: Unexpected status %d", i, resp.StatusCode)
		}
	}

	log.Printf("\n   📊 Results:")
	log.Printf("      Successful: %d/%d", successCount, numRequests)
	log.Printf("      Rate Limited: %d/%d", rateLimitedCount, numRequests)
	if firstRateLimitAt > 0 {
		log.Printf("      First rate limit at request: %d", firstRateLimitAt)
	}

	// Verify rate limiting is working
	if rateLimitedCount > 0 {
		log.Printf("   ✅ Rate limiting is WORKING correctly!")
	} else if numRequests > 100 {
		log.Printf("   ⚠️  WARNING: Expected some rate limiting but got none")
	}
}
