package main

import (
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/middleware"
)

func main() {
	log.Println("🔒 Testing HTTPS Redirect Logic")
	log.Println("================================")

	// Test in development mode
	log.Println("\n1️⃣ Testing Development Mode (HTTP allowed)")
	testRedirect("development", false)

	// Test in production mode
	log.Println("\n2️⃣ Testing Production Mode (HTTP → HTTPS redirect)")
	testRedirect("production", true)

	log.Println("\n✅ HTTPS redirect tests complete!")
}

func testRedirect(env string, expectRedirect bool) {
	// Create test router with HTTPS redirect middleware
	router := mux.NewRouter()
	router.Use(middleware.HTTPSRedirect(env))

	// Add test handler
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create test HTTP request (no TLS)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Host = "example.com" // Set host separately
	rr := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(rr, req)

	// Check result
	if expectRedirect {
		if rr.Code == http.StatusMovedPermanently {
			location := rr.Header().Get("Location")
			log.Printf("   ✅ Redirected to: %s", location)
			log.Printf("   ✅ Status: 301 Moved Permanently")
			if location != "https://example.com/test" {
				log.Printf("   ⚠️  WARNING: Expected https://example.com/test, got %s", location)
			}
		} else {
			log.Printf("   ❌ FAILED: Expected 301 redirect, got %d", rr.Code)
		}
	} else {
		if rr.Code == http.StatusOK {
			log.Printf("   ✅ HTTP request allowed (development mode)")
			log.Printf("   ✅ Status: 200 OK")
		} else {
			log.Printf("   ❌ FAILED: Expected 200 OK, got %d", rr.Code)
		}
	}

	// Test HTTPS request (should always pass)
	log.Println("\n   Testing HTTPS request (should always work):")
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Host = "example.com"
	req2.Header.Set("X-Forwarded-Proto", "https") // Simulate proxy
	rr2 := httptest.NewRecorder()

	router.ServeHTTP(rr2, req2)

	if rr2.Code == http.StatusOK {
		log.Printf("   ✅ HTTPS request allowed")
		log.Printf("   ✅ Status: 200 OK")
	} else {
		log.Printf("   ❌ FAILED: HTTPS should always work, got %d", rr2.Code)
	}
}
