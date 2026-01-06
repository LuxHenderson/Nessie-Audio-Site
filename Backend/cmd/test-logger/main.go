package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/logger"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

func main() {
	log.Println("🧪 Testing Error Logging & Email Alerts")
	log.Println("=========================================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize email client
	emailClient := email.NewClient(cfg)

	// Initialize logger
	appLogger, err := logger.New("logs/test-error.log", emailClient, cfg.AdminEmail)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	log.Println("\n✅ Logger initialized successfully")
	log.Printf("   Log file: logs/test-error.log")
	log.Printf("   Admin email: %s\n", cfg.AdminEmail)

	// Test 1: Info log
	log.Println("\n📝 Test 1: Info Log")
	appLogger.Info("Test info message - logger system initialized")
	log.Println("   ✓ Info log written")

	// Test 2: Warning log
	log.Println("\n⚠️  Test 2: Warning Log")
	appLogger.Warning("Test warning message", fmt.Errorf("sample warning error"))
	log.Println("   ✓ Warning log written")

	// Test 3: Error log (with stack trace)
	log.Println("\n❌ Test 3: Error Log (with stack trace)")
	appLogger.Error("Test error message", fmt.Errorf("sample error occurred"))
	log.Println("   ✓ Error log written with stack trace")

	// Test 4: Critical error (triggers email alert)
	log.Println("\n🚨 Test 4: Critical Error (triggers email alert)")
	log.Println("   This will send an email to:", cfg.AdminEmail)
	log.Println("   Sending critical error notification...")

	appLogger.Critical("Test critical error - Payment processing failure", fmt.Errorf("stripe payment failed"), map[string]interface{}{
		"payment_intent_id": "pi_test_1234567890",
		"amount":            49.99,
		"currency":          "USD",
		"customer_email":    "test@example.com",
		"error_code":        "card_declined",
	})

	log.Println("   ✓ Critical error logged")
	log.Println("   ✓ Email alert sent (check your inbox!)")

	// Test 5: Critical error with full context
	log.Println("\n🚨 Test 5: Critical Error with Full Context")
	appLogger.CriticalWithContext(logger.ErrorContext{
		Message:  "Webhook signature verification failed",
		Error:    fmt.Errorf("invalid signature"),
		UserIP:   "192.168.1.100",
		Endpoint: "/webhooks/stripe",
		Details: map[string]interface{}{
			"signature_header": "stripe-signature-xyz",
			"attempt":          3,
			"potential_attack": true,
		},
	})

	log.Println("   ✓ Critical error with context logged")
	log.Println("   ✓ Email alert queued for sending")

	// Wait for email goroutines to complete
	log.Println("\n⏳ Waiting for email alerts to send...")
	time.Sleep(3 * time.Second)

	// Summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("✅ All logging tests completed!")
	log.Println(strings.Repeat("=", 60))
	log.Println("\n📁 Check the log file:")
	log.Println("   cat logs/test-error.log")
	log.Println("\n📧 Check your email for 2 critical error alerts")
	log.Println("   Subject: [CRITICAL] Nessie Audio - ...")
	log.Println("\n💡 Note: In production, only CRITICAL errors send emails")
	log.Println("   INFO, WARNING, ERROR levels only write to the log file")
}
