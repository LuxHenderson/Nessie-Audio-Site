package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

func main() {
	log.Println("📧 Sending Test Admin Alert Emails")
	log.Println("===================================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create email client
	emailClient := email.NewClient(cfg)

	log.Printf("\n📤 Sending to: %s\n", cfg.AdminEmail)

	// Test 1: Payment Failed Alert
	log.Println("\n1️⃣ Sending Payment Failure Alert...")
	subject1 := "⚠️ Payment Failed: $49.99"
	body1 := fmt.Sprintf(`PAYMENT FAILURE ALERT
========================

A payment attempt has failed on the Nessie Audio store.

Payment Details:
  Payment Intent ID: pi_test_failed_12345
  Amount: $49.99 usd
  Status: requires_payment_method

Error:
  Your card was declined. Please try another payment method.

Customer Information:
  Email: test.customer@example.com
  Name: Test Customer

Timestamp: %s

---
This is an automated alert from Nessie Audio eCommerce Backend.
`, time.Now().Format("2006-01-02 15:04:05 MST"))

	if err := emailClient.SendRawEmail(cfg.AdminEmail, subject1, body1); err != nil {
		log.Printf("❌ Failed: %v", err)
	} else {
		log.Println("✅ Sent successfully")
	}

	time.Sleep(1 * time.Second)

	// Test 2: Payment Canceled Alert
	log.Println("\n2️⃣ Sending Payment Cancellation Alert...")
	subject2 := "🚫 Payment Canceled: $29.99"
	body2 := fmt.Sprintf(`PAYMENT CANCELLATION ALERT
==========================

A payment has been canceled on the Nessie Audio store.

Payment Details:
  Payment Intent ID: pi_test_canceled_12345
  Amount: $29.99 usd
  Status: canceled
  Cancellation Reason: abandoned

Customer Information:
  Email: abandoned.cart@example.com
  Name: Abandoned User

Timestamp: %s

---
This is an automated alert from Nessie Audio eCommerce Backend.
`, time.Now().Format("2006-01-02 15:04:05 MST"))

	if err := emailClient.SendRawEmail(cfg.AdminEmail, subject2, body2); err != nil {
		log.Printf("❌ Failed: %v", err)
	} else {
		log.Println("✅ Sent successfully")
	}

	time.Sleep(1 * time.Second)

	// Test 3: Checkout Expired Alert
	log.Println("\n3️⃣ Sending Checkout Expiration Alert...")
	subject3 := "⏱️ Checkout Expired: $79.98"
	body3 := fmt.Sprintf(`CHECKOUT SESSION EXPIRED ALERT
==============================

A checkout session has expired without completion on the Nessie Audio store.

Session Details:
  Session ID: cs_test_expired_12345
  Amount: $79.98 usd
  Status: expired

Customer Information:
  Email: timeout.customer@example.com
  Name: Timeout Customer

Possible Reasons:
  - Customer abandoned cart
  - Session timeout (24 hours)
  - Customer did not complete payment

Timestamp: %s

---
This is an automated alert from Nessie Audio eCommerce Backend.
`, time.Now().Format("2006-01-02 15:04:05 MST"))

	if err := emailClient.SendRawEmail(cfg.AdminEmail, subject3, body3); err != nil {
		log.Printf("❌ Failed: %v", err)
	} else {
		log.Println("✅ Sent successfully")
	}

	log.Println("\n✅ All test alerts sent!")
	log.Printf("📬 Check your inbox at: %s\n", cfg.AdminEmail)
	log.Println("\n💡 Note: These emails match exactly what the webhook handlers would send.")
}
