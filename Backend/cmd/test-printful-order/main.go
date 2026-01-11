package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
	"github.com/nessieaudio/ecommerce-backend/internal/services/printful"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create Printful client
	printfulClient := printful.NewClient(cfg.PrintfulAPIKey, cfg.PrintfulAPIURL)

	fmt.Println("=== Printful Order Test ===")
	fmt.Println("Creating test order with valid data...")
	fmt.Println()

	// Create a test order with realistic data
	orderID := uuid.New().String()
	customerID := uuid.New().String()

	testOrder := &models.Order{
		ID:                 orderID,
		CustomerID:         customerID,
		CustomerEmail:      "test@nessieaudio.com",
		Status:             models.OrderStatusPaid,
		TotalAmount:        25.00, // Eco Tote Bag price
		Currency:           "USD",
		ShippingName:       "John Doe",
		ShippingAddress1:   "123 Main Street",
		ShippingAddress2:   "Apt 4B",
		ShippingCity:       "Los Angeles",
		ShippingState:      "CA",
		ShippingZip:        "90001",
		ShippingCountry:    "US",
		StripeSessionID:    "cs_test_" + uuid.New().String(),
		StripePaymentIntentID: "pi_test_" + uuid.New().String(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Insert test order into database
	_, err = db.Exec(`
		INSERT INTO orders (
			id, customer_id, customer_email, status, total_amount, currency,
			stripe_session_id, stripe_payment_intent_id,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, orderID, customerID, testOrder.CustomerEmail, testOrder.Status, testOrder.TotalAmount, testOrder.Currency,
		testOrder.StripeSessionID, testOrder.StripePaymentIntentID,
		testOrder.ShippingName, testOrder.ShippingAddress1, testOrder.ShippingAddress2,
		testOrder.ShippingCity, testOrder.ShippingState, testOrder.ShippingZip, testOrder.ShippingCountry,
		testOrder.CreatedAt, testOrder.UpdatedAt)

	if err != nil {
		log.Fatalf("Failed to create test order: %v", err)
	}

	fmt.Printf("✓ Created test order: %s\n", orderID)

	// Get a variant with a valid Printful ID (Eco Tote Bag)
	var variantID string
	var printfulVariantID int64
	var productName string
	var variantName string
	var price float64

	err = db.QueryRow(`
		SELECT v.id, v.printful_variant_id, p.name, v.name, v.price
		FROM variants v
		JOIN products p ON v.product_id = p.id
		WHERE v.printful_variant_id = 5117581114
		LIMIT 1
	`).Scan(&variantID, &printfulVariantID, &productName, &variantName, &price)

	if err != nil {
		log.Fatalf("Failed to get test variant: %v", err)
	}

	fmt.Printf("✓ Using variant: %s (Printful ID: %d)\n", variantName, printfulVariantID)

	// Create order item
	itemID := uuid.New().String()
	testItem := models.OrderItem{
		ID:                itemID,
		OrderID:           orderID,
		ProductID:         "",
		VariantID:         variantID,
		ProductName:       productName,
		VariantName:       variantName,
		Quantity:          1,
		UnitPrice:         price,
		TotalPrice:        price,
		PrintfulVariantID: printfulVariantID,
		CreatedAt:         time.Now(),
	}

	_, err = db.Exec(`
		INSERT INTO order_items (
			id, order_id, product_id, variant_id,
			product_name, variant_name,
			quantity, unit_price, total_price, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, itemID, orderID, testItem.ProductID, variantID,
		testItem.ProductName, testItem.VariantName,
		testItem.Quantity, testItem.UnitPrice, testItem.TotalPrice, testItem.CreatedAt)

	if err != nil {
		log.Fatalf("Failed to create test order item: %v", err)
	}

	fmt.Printf("✓ Created order item: %s x%d @ $%.2f\n", variantName, testItem.Quantity, price)
	fmt.Println()

	// Now submit to Printful
	fmt.Println("Submitting order to Printful...")
	fmt.Println()

	items := []models.OrderItem{testItem}
	printfulOrderID, err := printfulClient.CreateOrder(testOrder, items)
	if err != nil {
		log.Printf("❌ Failed to create Printful order: %v", err)

		// Clean up test data
		db.Exec("DELETE FROM order_items WHERE order_id = ?", orderID)
		db.Exec("DELETE FROM orders WHERE id = ?", orderID)

		log.Fatalf("Test failed")
	}

	fmt.Printf("✓ Printful draft order created: ID %d\n", printfulOrderID)
	fmt.Println()

	// Confirm the order
	fmt.Println("Confirming order with Printful...")
	err = printfulClient.ConfirmOrder(printfulOrderID)
	if err != nil {
		log.Printf("❌ Failed to confirm Printful order: %v", err)

		// Clean up test data
		db.Exec("DELETE FROM order_items WHERE order_id = ?", orderID)
		db.Exec("DELETE FROM orders WHERE id = ?", orderID)

		log.Fatalf("Test failed")
	}

	fmt.Printf("✓ Order confirmed with Printful!\n")
	fmt.Println()

	// Update order with Printful ID
	_, err = db.Exec(`
		UPDATE orders
		SET printful_order_id = ?, status = ?, updated_at = ?
		WHERE id = ?
	`, printfulOrderID, models.OrderStatusFulfilled, time.Now(), orderID)

	if err != nil {
		log.Printf("Warning: Failed to update order with Printful ID: %v", err)
	}

	// Print summary
	fmt.Println("=== TEST SUCCESSFUL ===")
	fmt.Printf("Order ID: %s\n", orderID)
	fmt.Printf("Printful Order ID: %d\n", printfulOrderID)
	fmt.Printf("Customer: %s (%s)\n", testOrder.ShippingName, testOrder.CustomerEmail)
	fmt.Printf("Shipping: %s, %s, %s %s, %s\n",
		testOrder.ShippingAddress1,
		testOrder.ShippingCity,
		testOrder.ShippingState,
		testOrder.ShippingZip,
		testOrder.ShippingCountry)
	fmt.Printf("Item: %s x%d\n", variantName, testItem.Quantity)
	fmt.Printf("Total: $%.2f\n", testOrder.TotalAmount)
	fmt.Println()
	fmt.Println("Check your Printful dashboard to verify the order:")
	fmt.Println("https://www.printful.com/dashboard/default/orders")
	fmt.Println()

	// Clean up prompt
	fmt.Print("Clean up test order from database? (y/n): ")
	var cleanup string
	fmt.Scanln(&cleanup)

	if cleanup == "y" || cleanup == "Y" {
		db.Exec("DELETE FROM order_items WHERE order_id = ?", orderID)
		db.Exec("DELETE FROM orders WHERE id = ?", orderID)
		fmt.Println("✓ Test data cleaned up")
	} else {
		fmt.Println("Test data kept in database")
	}
}
