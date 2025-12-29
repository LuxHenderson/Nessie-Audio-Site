package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/services/printful"
)

func main() {
	log.Println("Printful Webhook Setup Tool")
	log.Println("============================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required config
	if cfg.PrintfulAPIKey == "" {
		log.Fatal("PRINTFUL_API_KEY not set in .env")
	}
	if cfg.PrintfulWebhookSecret == "" {
		log.Fatal("PRINTFUL_WEBHOOK_SECRET not set in .env")
	}

	// Get production domain or ask for it
	domain := cfg.ProductionDomain
	if domain == "" {
		fmt.Print("\nEnter your production domain (e.g., api.yourdomain.com): ")
		fmt.Scanln(&domain)
		if domain == "" {
			log.Fatal("Domain is required")
		}
	}

	// Create Printful client
	client := printful.NewClient(cfg.PrintfulAPIKey, cfg.PrintfulAPIURL)

	// Check if webhook already exists
	log.Println("\nChecking for existing webhook...")
	existing, err := client.GetWebhook()
	if err == nil && existing != nil {
		log.Printf("Found existing webhook: %s", existing.URL)
		log.Printf("Events: %v", existing.Types)

		fmt.Print("\nDo you want to replace it? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			log.Println("Cancelled. Webhook not modified.")
			return
		}
	}

	// Build webhook URL with secret token
	webhookURL := fmt.Sprintf("https://%s/webhooks/printful/%s", domain, cfg.PrintfulWebhookSecret)

	// Event types we want to receive
	eventTypes := []string{
		"package_shipped",
		"order_created",
		"order_updated",
		"order_failed",
		"order_canceled",
	}

	log.Println("\nRegistering webhook...")
	log.Printf("URL: %s", webhookURL)
	log.Printf("Events: %v", eventTypes)

	// Setup webhook
	if err := client.SetupWebhook(webhookURL, eventTypes); err != nil {
		log.Fatalf("Failed to setup webhook: %v", err)
	}

	log.Println("\n✅ Webhook successfully registered with Printful!")
	log.Println("\nYour webhook will now receive the following events:")
	for _, eventType := range eventTypes {
		log.Printf("  - %s", eventType)
	}

	log.Println("\nIMPORTANT:")
	log.Println("1. Make sure your server is running at:", fmt.Sprintf("https://%s", domain))
	log.Println("2. Your webhook endpoint must be accessible via HTTPS")
	log.Println("3. Keep your PRINTFUL_WEBHOOK_SECRET secure - it's embedded in the URL")

	// Verify registration
	log.Println("\nVerifying registration...")
	info, err := client.GetWebhook()
	if err != nil {
		log.Printf("Warning: Could not verify webhook: %v", err)
		os.Exit(0)
	}

	log.Printf("✓ Confirmed: Webhook is active at %s", info.URL)
}
