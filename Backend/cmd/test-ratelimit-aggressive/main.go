package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("🧪 Aggressive Rate Limit Test")
	log.Println("==============================")
	log.Println("This test will definitely trigger rate limits")

	baseURL := "http://localhost:8080"

	// Test with many more requests than the limit
	log.Println("\n📊 Testing /api/v1/config endpoint")
	log.Println("   Limit: 60 tokens burst")
	log.Println("   Sending 150 rapid requests...\n")

	successCount := 0
	rateLimitedCount := 0
	var firstRateLimitAt int

	for i := 1; i <= 150; i++ {
		resp, err := http.Get(baseURL + "/api/v1/config")
		if err != nil {
			log.Printf("❌ Request %d failed: %v", i, err)
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		limitHeader := resp.Header.Get("X-RateLimit-Limit")
		remainingHeader := resp.Header.Get("X-RateLimit-Remaining")
		retryAfter := resp.Header.Get("Retry-After")

		if resp.StatusCode == http.StatusOK {
			successCount++
			if i == 1 || i == 50 || i == 60 {
				log.Printf("✅ Request %d: SUCCESS (Limit: %s, Remaining: %s)", i, limitHeader, remainingHeader)
			}
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
			if firstRateLimitAt == 0 {
				firstRateLimitAt = i
				log.Printf("\n🚫 RATE LIMITED at request %d!", i)
				log.Printf("   Retry-After: %s seconds", retryAfter)
				log.Printf("   Limit: %s", limitHeader)
				log.Printf("   Remaining: %s\n", remainingHeader)
			}
		}

		// Small delay to show progress
		if i%10 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	log.Println("\n==================================================")
	log.Println("📊 FINAL RESULTS:")
	log.Println("==================================================")
	log.Printf("Total Requests:     150")
	log.Printf("Successful:         %d", successCount)
	log.Printf("Rate Limited:       %d", rateLimitedCount)
	if firstRateLimitAt > 0 {
		log.Printf("First Rate Limit:   Request #%d", firstRateLimitAt)
	}
	log.Println("==================================================")

	if rateLimitedCount > 0 {
		log.Println("\n✅ SUCCESS: Rate limiting is working!")
		log.Printf("   Token bucket correctly limited after ~%d requests", firstRateLimitAt-1)
	} else {
		log.Println("\n⚠️  WARNING: No rate limiting detected!")
		log.Println("   Make sure the server has been restarted with the new code.")
	}
}
