package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
)

func main() {
	log.Println("🔧 Testing Environment Auto-Detection")
	log.Println("======================================")

	// Test 1: Default detection (should be development)
	log.Println("\n📍 Test 1: Default Detection (no ENV set)")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	printConfig(cfg)

	// Test 2: Explicit ENV variable
	log.Println("\n📍 Test 2: Explicit ENV=staging")
	os.Setenv("ENV", "staging")
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	printConfig(cfg)

	// Test 3: Back to development
	log.Println("\n📍 Test 3: Explicit ENV=development")
	os.Setenv("ENV", "development")
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	printConfig(cfg)

	// Summary
	log.Println("\n======================================")
	log.Println("✅ Environment detection tests complete!")
	log.Println("======================================")
}

func printConfig(cfg *config.Config) {
	fmt.Printf("   Environment: %s\n", cfg.Env)
	fmt.Printf("   Port: %s\n", cfg.Port)
	fmt.Printf("   Database: %s\n", cfg.DatabasePath)
	fmt.Printf("   Stripe Mode: %s\n", getStripeMode(cfg.StripeSecretKey))
	fmt.Printf("   Success URL: %s\n", cfg.StripeSuccessURL)
	fmt.Printf("   Cancel URL: %s\n", cfg.StripeCancelURL)
	fmt.Printf("   CORS Origins: %s\n", cfg.AllowedOrigins)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   IsDevelopment: %v\n", cfg.IsDevelopment())
	fmt.Printf("   IsStaging: %v\n", cfg.IsStaging())
	fmt.Printf("   IsProduction: %v\n", cfg.IsProduction())
}

func getStripeMode(secretKey string) string {
	if len(secretKey) > 7 {
		if secretKey[3:7] == "test" {
			return "TEST (safe)"
		}
		if secretKey[3:7] == "live" {
			return "LIVE (real charges!)"
		}
	}
	return "unknown"
}
