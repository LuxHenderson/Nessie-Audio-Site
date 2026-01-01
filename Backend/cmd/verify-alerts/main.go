package main

import (
	"fmt"
	"log"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
)

func main() {
	log.Println("📊 Admin Alert System Verification")
	log.Println("===================================")

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

	// Check webhook events
	log.Println("\n📥 Recent Webhook Events (Last 10):")
	log.Println("=====================================")

	rows, err := db.Query(`
		SELECT event_type, event_id, created_at
		FROM stripe_webhook_events
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		log.Fatalf("Failed to query webhooks: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var eventType, eventID, createdAt string
		if err := rows.Scan(&eventType, &eventID, &createdAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		emoji := "📝"
		switch eventType {
		case "payment_intent.payment_failed":
			emoji = "⚠️"
		case "payment_intent.canceled":
			emoji = "🚫"
		case "checkout.session.expired":
			emoji = "⏱️"
		case "checkout.session.completed":
			emoji = "✅"
		}

		fmt.Printf("  %s %s\n", emoji, eventType)
		fmt.Printf("     ID: %s\n", eventID)
		fmt.Printf("     Time: %s\n\n", createdAt)
		count++
	}

	if count == 0 {
		log.Println("  No webhook events found")
	}

	// Summary
	log.Println("\n📬 Email Configuration:")
	log.Println("=======================")
	fmt.Printf("  Admin Email: %s\n", cfg.AdminEmail)
	fmt.Printf("  SMTP Host: %s:%s\n", cfg.SMTPHost, cfg.SMTPPort)
	fmt.Printf("  SMTP Username: %s\n", cfg.SMTPUsername)
	if cfg.SMTPPassword != "" {
		fmt.Printf("  SMTP Password: *** (configured)\n")
	} else {
		fmt.Printf("  SMTP Password: ⚠️  NOT CONFIGURED\n")
	}

	log.Println("\n✅ Verification Complete")
	log.Println("\nℹ️  Test Events to Trigger:")
	log.Println("   - payment_intent.payment_failed → ⚠️ Payment Failed email")
	log.Println("   - payment_intent.canceled → 🚫 Payment Canceled email")
	log.Println("   - checkout.session.expired → ⏱️ Checkout Expired email")
	log.Println("\n📧 Check your inbox at:", cfg.AdminEmail)
}
