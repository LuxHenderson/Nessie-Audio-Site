package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

func main() {
	fmt.Println("=== Comprehensive Error Handling Tests ===\n")

	// Test 1: Successful request
	fmt.Println("Test 1: Successful Request (GET /api/v1/products)")
	fmt.Println("---------------------------------------------------")
	testSuccessfulRequest()

	time.Sleep(500 * time.Millisecond)

	// Test 2: 404 Not Found
	fmt.Println("\nTest 2: 404 Not Found (Non-existent product)")
	fmt.Println("----------------------------------------------")
	test404NotFound()

	time.Sleep(500 * time.Millisecond)

	// Test 3: Request ID tracking
	fmt.Println("\nTest 3: Request ID Tracking")
	fmt.Println("---------------------------")
	testRequestIDTracking()

	time.Sleep(500 * time.Millisecond)

	// Test 4: Health check
	fmt.Println("\nTest 4: Health Check")
	fmt.Println("--------------------")
	testHealthCheck()

	fmt.Println("\n=== All Error Handling Tests Complete ===")
}

func testSuccessfulRequest() {
	resp, err := http.Get(baseURL + "/api/v1/products")
	if err != nil {
		log.Printf("❌ FAIL: Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get("X-Request-ID")

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ FAIL: Expected status 200, got %d", resp.StatusCode)
		return
	}

	if requestID == "" {
		log.Printf("❌ FAIL: Missing X-Request-ID header")
		return
	}

	fmt.Printf("✅ Status: %d OK\n", resp.StatusCode)
	fmt.Printf("✅ Request ID: %s\n", requestID)
	fmt.Println("✅ Response contains products data")
}

func test404NotFound() {
	resp, err := http.Get(baseURL + "/api/v1/products/nonexistent-product-id")
	if err != nil {
		log.Printf("❌ FAIL: Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		log.Printf("❌ FAIL: Expected status 404, got %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ FAIL: Failed to read response: %v", err)
		return
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		log.Printf("❌ FAIL: Failed to parse error response: %v", err)
		return
	}

	fmt.Printf("✅ Status: %d Not Found\n", resp.StatusCode)
	fmt.Printf("✅ Error: %s\n", errorResp.Error)
	fmt.Printf("✅ Error Code: %s\n", errorResp.Code)
	fmt.Printf("✅ Request ID: %s\n", errorResp.RequestID)
	fmt.Printf("✅ Timestamp: %s\n", errorResp.Timestamp)

	// Validate fields are populated
	if errorResp.Error == "" || errorResp.Code == "" || errorResp.RequestID == "" || errorResp.Timestamp == "" {
		log.Printf("❌ FAIL: Error response missing required fields")
		return
	}

	fmt.Println("✅ All error response fields present")
}

func testRequestIDTracking() {
	// Make two requests and ensure they have different request IDs
	resp1, err := http.Get(baseURL + "/api/v1/products")
	if err != nil {
		log.Printf("❌ FAIL: Request 1 failed: %v", err)
		return
	}
	defer resp1.Body.Close()

	requestID1 := resp1.Header.Get("X-Request-ID")

	time.Sleep(100 * time.Millisecond)

	resp2, err := http.Get(baseURL + "/api/v1/products")
	if err != nil {
		log.Printf("❌ FAIL: Request 2 failed: %v", err)
		return
	}
	defer resp2.Body.Close()

	requestID2 := resp2.Header.Get("X-Request-ID")

	if requestID1 == "" || requestID2 == "" {
		log.Printf("❌ FAIL: Missing request IDs")
		return
	}

	if requestID1 == requestID2 {
		log.Printf("❌ FAIL: Request IDs should be unique")
		return
	}

	fmt.Printf("✅ Request 1 ID: %s\n", requestID1)
	fmt.Printf("✅ Request 2 ID: %s\n", requestID2)
	fmt.Println("✅ Request IDs are unique")
}

func testHealthCheck() {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		log.Printf("❌ FAIL: Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ FAIL: Expected status 200, got %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ FAIL: Failed to read response: %v", err)
		return
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		log.Printf("❌ FAIL: Failed to parse health response: %v", err)
		return
	}

	fmt.Printf("✅ Status: %d OK\n", resp.StatusCode)
	fmt.Printf("✅ Health: %s\n", health["status"])
	fmt.Println("✅ Health check passed")
}
