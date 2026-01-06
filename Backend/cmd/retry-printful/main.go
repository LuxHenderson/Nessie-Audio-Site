package main

import (
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

	body := `
URGENT: Printful Order Submission Failure

An order has failed to submit to Printful after 24 hours of retries and requires manual intervention.

Order Details:
- Order ID: ` + order.ID + `
- Customer Email: ` + order.CustomerEmail + `
- Total Amount: $` + formatMoney(order.TotalAmount) + `
- Created: ` + order.CreatedAt.Format(time.RFC1123) + `
- Retry Attempts: ` + formatInt(attemptNumber) + `
- Stripe Session: ` + order.StripeSessionID + `

Shipping Information:
- Name: ` + order.ShippingName + `
- Address: ` + order.ShippingAddress1 + `
` + formatAddress2(order.ShippingAddress2) + `
- City: ` + order.ShippingCity + `
- State: ` + order.ShippingState + `
- Zip: ` + order.ShippingZip + `
- Country: ` + order.ShippingCountry + `

Last Error:
` + errorMsg + `

Action Required:
1. Check Printful dashboard for any service issues
2. Verify API key permissions
3. Manually submit order to Printful if necessary
4. Contact customer if further delays expected

This is an automated alert from Nessie Audio Backend.
`

	return emailClient.SendRawEmail(cfg.AdminEmail, subject, body)
}

// Helper functions for email formatting
func formatMoney(amount float64) string {
	return formatFloat(amount, 2)
}

func formatFloat(f float64, decimals int) string {
	switch decimals {
	case 2:
		return formatTwoDecimals(f)
	default:
		return formatTwoDecimals(f)
	}
}

func formatTwoDecimals(f float64) string {
	// Format float to 2 decimal places
	whole := int(f)
	fraction := int((f - float64(whole)) * 100)
	return intToString(whole) + "." + padZero(fraction, 2)
}

func formatInt(i int) string {
	return intToString(i)
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var result string
	for i > 0 {
		digit := i % 10
		result = string(rune('0'+digit)) + result
		i /= 10
	}

	if negative {
		result = "-" + result
	}

	return result
}

func padZero(i int, width int) string {
	s := intToString(i)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

func formatAddress2(addr string) string {
	if addr == "" {
		return ""
	}
	return "- Address 2: " + addr + "\n"
}
