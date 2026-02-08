package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
	"github.com/nessieaudio/ecommerce-backend/internal/services/order"
	"github.com/nessieaudio/ecommerce-backend/internal/services/printful"
)

func main() {
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

	// Initialize services
	orderService := order.NewService(db)
	printfulClient := printful.NewClient(cfg.PrintfulAPIKey, cfg.PrintfulAPIURL)
	emailClient := email.NewClient(cfg)

	log.Println("🔄 Starting Printful retry job...")

	// Get failed orders
	failedOrders, err := orderService.GetFailedPrintfulOrders()
	if err != nil {
		log.Fatalf("Failed to get failed orders: %v", err)
	}

	if len(failedOrders) == 0 {
		log.Println("No failed orders to retry")
		return
	}

	log.Printf("Found %d orders to retry", len(failedOrders))

	for _, order := range failedOrders {
		log.Printf("Retrying order %s (attempt #%d)...", order.ID, order.PrintfulRetryCount+1)

		// Get order items
		items, err := orderService.GetOrderItems(order.ID)
		if err != nil {
			log.Printf("Failed to get order items for %s: %v", order.ID, err)
			continue
		}

		// Attempt to submit to Printful
		printfulOrderID, err := printfulClient.CreateOrder(&order, items)
		if err != nil {
			log.Printf("Retry failed for order %s: %v", order.ID, err)

			// Increment retry count
			if err := orderService.IncrementPrintfulRetryCount(order.ID); err != nil {
				log.Printf("Failed to increment retry count: %v", err)
			}

			// Record failure
			attemptNumber := order.PrintfulRetryCount + 1
			if err := orderService.RecordPrintfulFailure(order.ID, attemptNumber, err.Error(), ""); err != nil {
				log.Printf("Failed to record failure: %v", err)
			}

			// Check if order is older than 24 hours - if so, send alert
			if time.Since(order.CreatedAt) > 24*time.Hour {
				log.Printf("⚠️  Order %s has exceeded 24 hour retry window. Sending alert...", order.ID)
				if err := sendFailureAlert(emailClient, cfg, &order, attemptNumber, err.Error()); err != nil {
					log.Printf("Failed to send alert email: %v", err)
				}
			}

			continue
		}

		// Success! Confirm the order
		if err := printfulClient.ConfirmOrder(printfulOrderID); err != nil {
			log.Printf("Failed to confirm Printful order %d: %v", printfulOrderID, err)

			// Record this failure too
			attemptNumber := order.PrintfulRetryCount + 1
			if err := orderService.IncrementPrintfulRetryCount(order.ID); err != nil {
				log.Printf("Failed to increment retry count: %v", err)
			}
			if err := orderService.RecordPrintfulFailure(order.ID, attemptNumber, "Confirm failed: "+err.Error(), ""); err != nil {
				log.Printf("Failed to record failure: %v", err)
			}

			// Check if order is older than 24 hours
			if time.Since(order.CreatedAt) > 24*time.Hour {
				log.Printf("⚠️  Order %s has exceeded 24 hour retry window. Sending alert...", order.ID)
				if err := sendFailureAlert(emailClient, cfg, &order, attemptNumber, "Confirm failed: "+err.Error()); err != nil {
					log.Printf("Failed to send alert email: %v", err)
				}
			}

			continue
		}

		// Update order with Printful ID
		if err := orderService.UpdateOrderWithPrintful(order.ID, printfulOrderID); err != nil {
			log.Printf("Failed to update order with Printful ID: %v", err)
			continue
		}

		// Update status to fulfilled
		if err := orderService.UpdateOrderStatus(order.ID, models.OrderStatusFulfilled); err != nil {
			log.Printf("Failed to update order status: %v", err)
			continue
		}

		log.Printf("✅ Order %s submitted successfully to Printful (ID: %d) after %d retries", order.ID, printfulOrderID, order.PrintfulRetryCount+1)
	}

	log.Println("✅ Retry job complete")
}

// sendFailureAlert sends an email alert for orders that failed after 24 hours
func sendFailureAlert(emailClient *email.Client, cfg *config.Config, order *models.Order, attemptNumber int, errorMsg string) error {
	subject := "🚨 Printful Order Submission Failed - Manual Intervention Required"

	shippingAddr := order.ShippingAddress1
	if order.ShippingAddress2 != "" {
		shippingAddr += ", " + order.ShippingAddress2
	}

	contentHTML := fmt.Sprintf(`<p style="font-size:16px;">An order has failed to submit to Printful after 24 hours of retries and requires manual intervention.</p>%s%s%s%s`,
		email.InfoBox("Order Details",
			email.DetailRow("Order ID:", fmt.Sprintf("#%s", order.ID))+
				email.DetailRow("Customer Email:", order.CustomerEmail)+
				email.DetailRow("Total Amount:", fmt.Sprintf("$%.2f", order.TotalAmount))+
				email.DetailRow("Created:", order.CreatedAt.Format(time.RFC1123))+
				email.DetailRow("Retry Attempts:", fmt.Sprintf("%d", attemptNumber))+
				email.DetailRow("Stripe Session:", order.StripeSessionID)),
		email.InfoBox("Shipping Information",
			email.DetailRow("Name:", order.ShippingName)+
				email.DetailRow("Address:", shippingAddr)+
				email.DetailRow("City:", order.ShippingCity)+
				email.DetailRow("State:", order.ShippingState)+
				email.DetailRow("Zip:", order.ShippingZip)+
				email.DetailRow("Country:", order.ShippingCountry)),
		email.NoteBox(fmt.Sprintf("<strong>Last Error:</strong><br>%s", errorMsg), true),
		email.NoteBox("<strong>Action Required:</strong><br>&bull; Check Printful dashboard for service issues<br>&bull; Verify API key permissions<br>&bull; Manually submit order if necessary<br>&bull; Contact customer if further delays expected", true),
	)
	htmlBody := email.EmailLayout("Printful Submission Failed", "&#128680;", contentHTML, true)

	return emailClient.SendHTMLEmail(cfg.AdminEmail, subject, htmlBody)
}
