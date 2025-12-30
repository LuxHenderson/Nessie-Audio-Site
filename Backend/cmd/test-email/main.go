package main

import (
	"log"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

func main() {
	log.Println("Email Test Utility")
	log.Println("==================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create email client
	emailClient := email.NewClient(cfg)

	// Test email data
	testData := email.OrderConfirmationData{
		OrderID:       "TEST-ORDER-12345",
		CustomerName:  "Test Customer",
		CustomerEmail: "jon.sanderson91@gmail.com", // Your email
		Items: []models.OrderItem{
			{
				ProductName: "Test Product",
				VariantName: "Large / Black",
				Quantity:    2,
				UnitPrice:   29.99,
				TotalPrice:  59.98,
			},
		},
		Total: 59.98,
		ShippingInfo: email.ShippingInfo{
			Name:    "Test Customer",
			Address: "123 Test Street",
			City:    "Test City",
			State:   "CA",
			Zip:     "12345",
			Country: "US",
		},
	}

	log.Println("\nSending test order confirmation email...")
	log.Printf("To: %s", testData.CustomerEmail)

	if err := emailClient.SendOrderConfirmation(testData); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	log.Println("✅ Email sent successfully!")
	log.Println("Check your inbox at:", testData.CustomerEmail)
}
