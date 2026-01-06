package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
)

func main() {
	log.Println("🧪 Testing Printful Retry Logic")
	log.Println("================================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a test order that simulates a paid order with failed Printful submission
	orderID := uuid.New().String()
	customerID := uuid.New().String()

	log.Printf("\n📝 Creating test order: %s", orderID)

	_, err = db.Exec(`
		INSERT INTO orders (
			id, customer_id, customer_email, status, total_amount, currency,
			stripe_session_id, stripe_payment_intent_id,
			shipping_name, shipping_address1, shipping_city, shipping_state,
			shipping_zip, shipping_country,
			printful_retry_count,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, orderID, customerID, "test@example.com", models.OrderStatusPaid, 29.99, "USD",
		"cs_test_123", "pi_test_123",
		"Test Customer", "123 Test St", "Test City", "CA",
		"12345", "US",
		1, // Already failed once to trigger retry
		time.Now().Add(-30*time.Minute), // Created 30 mins ago
		time.Now())

	if err != nil {
		log.Fatalf("Failed to create test order: %v", err)
	}

	// Get a real variant ID from the database
	var variantID string
	err = db.QueryRow(`SELECT id FROM variants LIMIT 1`).Scan(&variantID)
	if err != nil {
		log.Fatalf("Failed to get variant ID: %v", err)
	}

	// Add a test order item
	itemID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO order_items (
			id, order_id, product_id, variant_id,
			quantity, unit_price, total_price,
			product_name, variant_name, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, itemID, orderID, uuid.New().String(), variantID,
		1, 29.99, 29.99,
		"Test Product", "Size M", time.Now())

	if err != nil {
		log.Fatalf("Failed to create test order item: %v", err)
	}

	// Record a fake failure
	_, err = db.Exec(`
		INSERT INTO printful_submission_failures (
			id, order_id, attempt_number, error_message, error_details, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`, orderID+"-1", orderID, 1, "Simulated failure for testing", "", time.Now().Add(-25*time.Minute))

	if err != nil {
		log.Fatalf("Failed to record test failure: %v", err)
	}

	log.Println("✅ Test order created successfully")
	log.Println("\n📊 Order Details:")
	log.Printf("  Order ID: %s", orderID)
	log.Printf("  Customer: test@example.com")
	log.Printf("  Status: paid")
	log.Printf("  Total: $29.99")
	log.Printf("  Retry Count: 1")
	log.Printf("  Printful Order ID: NULL (not submitted)")

	log.Println("\n🔍 Verifying order is in retry queue...")

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE status = ?
			AND printful_order_id IS NULL
			AND created_at > datetime('now', '-24 hours')
			AND printful_retry_count > 0
	`, models.OrderStatusPaid).Scan(&count)

	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	if count > 0 {
		log.Printf("✅ Found %d order(s) in retry queue", count)
	} else {
		log.Println("❌ Order not found in retry queue")
	}

	log.Println("\n📋 Next Steps:")
	log.Println("  1. The retry worker will pick this up in the next cycle (within 15 min)")
	log.Println("  2. It will attempt to submit to Printful")
	log.Println("  3. If it fails, it will increment retry_count")
	log.Println("  4. After 24 hours of failures, you'll receive an email alert")
	log.Println("\n💡 To manually trigger retry now, run:")
	log.Println("     go run cmd/retry-printful/main.go")
	log.Println("\n🧹 To clean up this test order, run:")
	log.Printf("     DELETE FROM order_items WHERE order_id = '%s';", orderID)
	log.Printf("     DELETE FROM printful_submission_failures WHERE order_id = '%s';", orderID)
	log.Printf("     DELETE FROM orders WHERE id = '%s';", orderID)
}
