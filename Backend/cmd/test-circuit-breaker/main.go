package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/circuitbreaker"
	"github.com/nessieaudio/ecommerce-backend/internal/services/printful"
)

func main() {
	fmt.Println("=== Circuit Breaker & Timeout Testing ===\n")

	// Test 1: Circuit Breaker Basic Functionality
	testCircuitBreakerStates()

	// Test 2: Timeout Behavior
	testTimeoutBehavior()

	// Test 3: Printful Client with Circuit Breaker
	testPrintfulClientCircuitBreaker()

	fmt.Println("\n=== All Tests Complete ===")
}

// Test 1: Circuit Breaker State Transitions
func testCircuitBreakerStates() {
	fmt.Println("Test 1: Circuit Breaker State Transitions")
	fmt.Println("-------------------------------------------")

	cb := circuitbreaker.New(circuitbreaker.Config{
		Name:            "test",
		MaxFailures:     3,
		ResetTimeout:    2 * time.Second,
		HalfOpenMaxReqs: 1,
	})

	// Initially should be closed
	if cb.GetState() != 0 { // 0 = Closed
		log.Fatal("❌ FAIL: Circuit should start in Closed state")
	}
	fmt.Println("✅ Circuit starts in Closed state")

	// Trigger 3 failures to open circuit
	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return fmt.Errorf("simulated failure %d", i+1)
		})
	}

	if cb.GetState() != 1 { // 1 = Open
		log.Fatal("❌ FAIL: Circuit should be Open after 3 failures")
	}
	fmt.Println("✅ Circuit opens after 3 failures")

	// Try request while open - should fail immediately
	start := time.Now()
	err := cb.Execute(func() error {
		time.Sleep(100 * time.Millisecond) // This shouldn't execute
		return nil
	})
	elapsed := time.Since(start)

	if err != circuitbreaker.ErrCircuitOpen {
		log.Fatal("❌ FAIL: Should return ErrCircuitOpen")
	}
	if elapsed > 10*time.Millisecond {
		log.Fatalf("❌ FAIL: Should fail fast (took %v)", elapsed)
	}
	fmt.Printf("✅ Circuit fails fast when open (took %v)\n", elapsed)

	// Wait for reset timeout
	fmt.Println("⏳ Waiting 2 seconds for reset timeout...")
	time.Sleep(2100 * time.Millisecond)

	// Should now be half-open, accept one request
	successCount := 0
	err = cb.Execute(func() error {
		successCount++
		return nil // Success
	})

	if err != nil {
		log.Fatalf("❌ FAIL: Should allow request in half-open: %v", err)
	}
	if successCount != 1 {
		log.Fatal("❌ FAIL: Request should have executed")
	}
	fmt.Println("✅ Circuit transitions to Half-Open and allows test request")

	// After success, should be closed again
	if cb.GetState() != 0 { // 0 = Closed
		log.Fatal("❌ FAIL: Circuit should close after successful half-open request")
	}
	fmt.Println("✅ Circuit closes after successful half-open request")

	fmt.Println()
}

// Test 2: HTTP Client Timeout Behavior
func testTimeoutBehavior() {
	fmt.Println("Test 2: HTTP Client Timeout Behavior")
	fmt.Println("--------------------------------------")

	// Create a slow server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// Client with 1 second timeout
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	fmt.Println("⏳ Making request to slow server (3s delay, 1s timeout)...")
	start := time.Now()
	resp, err := client.Get(slowServer.URL)
	elapsed := time.Since(start)

	if err == nil {
		resp.Body.Close()
		log.Fatal("❌ FAIL: Should have timed out")
	}

	if elapsed > 1500*time.Millisecond {
		log.Fatalf("❌ FAIL: Timeout took too long: %v", elapsed)
	}

	fmt.Printf("✅ Request timed out after %v (expected ~1s)\n", elapsed)
	fmt.Println()
}

// Test 3: Printful Client with Circuit Breaker
func testPrintfulClientCircuitBreaker() {
	fmt.Println("Test 3: Printful Client Circuit Breaker Integration")
	fmt.Println("----------------------------------------------------")

	var requestCount int32
	failureCount := 0

	// Create test server that fails 5 times, then succeeds
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 5 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"simulated failure"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":200,"result":[]}`))
	}))
	defer testServer.Close()

	// Create Printful client pointing to test server
	client := printful.NewClient("test-key", testServer.URL)

	fmt.Println("Phase 1: Triggering 5 failures to open circuit...")
	for i := 0; i < 5; i++ {
		_, err := client.GetProducts()
		if err != nil {
			failureCount++
			fmt.Printf("  Attempt %d: Failed (expected) ✓\n", i+1)
		}
	}

	if failureCount != 5 {
		log.Fatalf("❌ FAIL: Expected 5 failures, got %d", failureCount)
	}
	fmt.Println("✅ All 5 requests failed as expected")

	// Next request should fail immediately (circuit open)
	fmt.Println("\nPhase 2: Testing fail-fast behavior...")
	start := time.Now()
	_, err := client.GetProducts()
	elapsed := time.Since(start)

	if err != circuitbreaker.ErrCircuitOpen {
		log.Fatalf("❌ FAIL: Expected circuit open error, got: %v", err)
	}

	if elapsed > 50*time.Millisecond {
		log.Fatalf("❌ FAIL: Should fail instantly, took %v", elapsed)
	}
	fmt.Printf("✅ Request failed instantly when circuit open (took %v)\n", elapsed)

	// Wait for circuit to go half-open
	fmt.Println("\n⏳ Waiting 60 seconds for circuit to reset...")
	fmt.Println("(This verifies the 60-second reset timeout is working)")

	// Show countdown
	for i := 60; i > 0; i -= 5 {
		fmt.Printf("  %d seconds remaining...\n", i)
		time.Sleep(5 * time.Second)
	}

	// Try again - should succeed now
	fmt.Println("\nPhase 3: Testing recovery after reset timeout...")
	atomic.StoreInt32(&requestCount, 5) // Reset to allow success
	products, err := client.GetProducts()

	if err != nil {
		log.Fatalf("❌ FAIL: Circuit should allow request after reset: %v", err)
	}

	if products == nil {
		log.Fatal("❌ FAIL: Should have received response")
	}

	fmt.Println("✅ Circuit allowed request after reset timeout")
	fmt.Println("✅ Request succeeded, circuit should now be closed")

	// Verify circuit is closed by making another successful request
	fmt.Println("\nPhase 4: Verifying circuit is fully closed...")
	_, err = client.GetProducts()
	if err != nil {
		log.Fatalf("❌ FAIL: Circuit should be closed and allow requests: %v", err)
	}
	fmt.Println("✅ Circuit is closed and accepting requests normally")

	fmt.Println()
}
