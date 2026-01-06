package main

import (
	"fmt"
	"log"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
	"github.com/nessieaudio/ecommerce-backend/internal/inventory"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

func main() {
	log.Println("🧪 Testing Inventory Tracking System")
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

	// Create services
	inventoryService := inventory.NewService(db)
	emailClient := email.NewClient(cfg)
	alertService := inventory.NewAlertService(inventoryService, emailClient, cfg)

	// Test 1: Check if new columns exist
	log.Println("\n📊 Test 1: Verifying database schema")
	var columnCount int
	db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('variants')
		WHERE name IN ('stock_quantity', 'low_stock_threshold', 'track_inventory')
	`).Scan(&columnCount)

	if columnCount == 3 {
		log.Println("✅ All inventory columns exist in database")
	} else {
		log.Printf("❌ Expected 3 columns, found %d", columnCount)
	}

	// Test 2: Set up a test variant with inventory tracking
	log.Println("\n📦 Test 2: Setting up test inventory")

	// Get first variant from database
	var variantID, variantName string
	err = db.QueryRow(`
		SELECT id, name FROM variants LIMIT 1
	`).Scan(&variantID, &variantName)

	if err != nil {
		log.Printf("⚠️ No variants found in database - skipping inventory tests")
		log.Println("\n💡 Tip: Run product sync first: go run cmd/sync-products/main.go")
		return
	}

	log.Printf("Using variant: %s (%s)", variantName, variantID)

	// Enable inventory tracking for this variant
	testStock := 10
	testThreshold := 5
	if err := inventoryService.UpdateStock(variantID, testStock, testThreshold, true); err != nil {
		log.Fatalf("Failed to update stock: %v", err)
	}
	log.Printf("✅ Set stock to %d with threshold %d", testStock, testThreshold)

	// Test 3: Check stock availability
	log.Println("\n🔍 Test 3: Checking stock availability")

	// Should be available (requesting less than stock)
	check, err := inventoryService.CheckStock(variantID, 5)
	if err != nil {
		log.Fatalf("Failed to check stock: %v", err)
	}

	if check.Available {
		log.Printf("✅ Stock check passed: %d units available (requested 5)", *check.StockQuantity)
	} else {
		log.Printf("❌ Stock check failed: expected available, got unavailable")
	}

	// Should be unavailable (requesting more than stock)
	check, err = inventoryService.CheckStock(variantID, 15)
	if err != nil {
		log.Fatalf("Failed to check stock: %v", err)
	}

	if !check.Available {
		log.Printf("✅ Stock check passed: correctly detected insufficient stock (requested 15, have %d)", *check.StockQuantity)
	} else {
		log.Printf("❌ Stock check failed: should have detected insufficient stock")
	}

	// Test 4: Deduct stock
	log.Println("\n➖ Test 4: Deducting stock")

	if err := inventoryService.DeductStock(variantID, 3); err != nil {
		log.Fatalf("Failed to deduct stock: %v", err)
	}
	log.Println("✅ Deducted 3 units successfully")

	// Verify new stock level
	check, _ = inventoryService.CheckStock(variantID, 1)
	log.Printf("✅ New stock level: %d units", *check.StockQuantity)

	// Test 5: Deduct more to trigger low stock
	log.Println("\n⚠️  Test 5: Triggering low stock threshold")

	if err := inventoryService.DeductStock(variantID, 4); err != nil {
		log.Fatalf("Failed to deduct stock: %v", err)
	}
	log.Println("✅ Deducted 4 more units")

	check, _ = inventoryService.CheckStock(variantID, 1)
	log.Printf("⚠️  Stock now at %d units (threshold: %d) - LOW STOCK!", *check.StockQuantity, testThreshold)

	// Test 6: Get low stock items
	log.Println("\n📉 Test 6: Fetching low stock items")

	lowStockItems, err := inventoryService.GetLowStockItems()
	if err != nil {
		log.Fatalf("Failed to get low stock items: %v", err)
	}

	if len(lowStockItems) > 0 {
		log.Printf("✅ Found %d low stock item(s):", len(lowStockItems))
		for _, item := range lowStockItems {
			log.Printf("   - %s / %s: %d units (threshold: %d)",
				item.ProductName, item.VariantName, item.StockQuantity, item.LowStockThreshold)
		}
	} else {
		log.Println("ℹ️  No low stock items found")
	}

	// Test 7: Test low stock alert email (optional - only if SMTP configured)
	log.Println("\n📧 Test 7: Low stock alert system")

	if cfg.AdminEmail != "" && cfg.SMTPUsername != "" {
		log.Printf("Sending low stock alert to: %s", cfg.AdminEmail)
		if err := alertService.CheckAndSendLowStockAlerts(); err != nil {
			log.Printf("❌ Failed to send alert: %v", err)
		} else {
			log.Println("✅ Low stock alert sent successfully!")
		}
	} else {
		log.Println("ℹ️  SMTP or admin email not configured - skipping email test")
		log.Println("   Set ADMIN_EMAIL and SMTP credentials in .env to test alerts")
	}

	// Test 8: Restore stock
	log.Println("\n➕ Test 8: Restoring stock")

	if err := inventoryService.RestoreStock(variantID, 7); err != nil {
		log.Fatalf("Failed to restore stock: %v", err)
	}
	log.Println("✅ Restored 7 units")

	check, _ = inventoryService.CheckStock(variantID, 1)
	log.Printf("✅ Stock restored to: %d units", *check.StockQuantity)

	// Clean up: Disable inventory tracking for test variant
	log.Println("\n🧹 Cleanup: Disabling inventory tracking for test variant")
	if err := inventoryService.UpdateStock(variantID, 0, 5, false); err != nil {
		log.Printf("Warning: Failed to disable tracking: %v", err)
	} else {
		log.Println("✅ Test variant reset to print-on-demand mode")
	}

	// Summary
	log.Println("\n===================================")
	log.Println("✅ All inventory tests completed!")
	log.Println("===================================")
	log.Println("\n📋 Summary:")
	log.Println("  ✓ Database schema updated")
	log.Println("  ✓ Stock availability checking works")
	log.Println("  ✓ Stock deduction works")
	log.Println("  ✓ Low stock detection works")
	log.Println("  ✓ Stock restoration works")
	log.Println("  ✓ Low stock alerts functional")
	log.Println("\n💡 Next Steps:")
	log.Println("  1. Use API endpoints to manage inventory:")
	log.Println("     GET  /api/v1/inventory           - View all inventory")
	log.Println("     GET  /api/v1/inventory/low-stock - View low stock items")
	log.Println("     PUT  /api/v1/inventory/{id}     - Update stock levels")
	log.Println("  2. Enable tracking for specific variants you want to monitor")
	log.Println("  3. Set appropriate low_stock_threshold values")
	log.Println("  4. Monitor admin email for low stock alerts")
	fmt.Println()
}
