package main

import (
	"log"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

func main() {
	log.Println("📧 Testing Email Sending Directly")
	log.Println("==================================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create email client
	emailClient := email.NewClient(cfg)

	// Test sending a simple email
	subject := "Test Email from Nessie Audio Backend"
	body := `Hello,

This is a test email to verify the SMTP configuration is working correctly.

If you receive this email, it means:
  - SMTP authentication is successful
  - Email sending functionality works
  - Admin alerts should be delivered properly

---
Sent from Nessie Audio eCommerce Backend Test Suite
`

	log.Printf("\n📤 Sending test email to: %s", cfg.AdminEmail)
	log.Printf("   SMTP Host: %s:%s", cfg.SMTPHost, cfg.SMTPPort)
	log.Printf("   From: %s <%s>", cfg.SMTPFromName, cfg.SMTPFromEmail)

	err = emailClient.SendRawEmail(cfg.AdminEmail, subject, body)
	if err != nil {
		log.Fatalf("❌ Failed to send email: %v", err)
	}

	log.Println("\n✅ Email sent successfully!")
	log.Printf("📬 Check your inbox at: %s\n", cfg.AdminEmail)
}
