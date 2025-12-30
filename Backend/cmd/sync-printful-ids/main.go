package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
)

type PrintfulProductList struct {
	Code   int `json:"code"`
	Result []struct {
		ID         int64  `json:"id"`
		ExternalID string `json:"external_id"`
		Name       string `json:"name"`
	} `json:"result"`
}

type PrintfulProductDetail struct {
	Code   int `json:"code"`
	Result struct {
		SyncVariants []struct {
			ID         int64  `json:"id"`
			ExternalID string `json:"external_id"`
			Name       string `json:"name"`
			VariantID  int64  `json:"variant_id"` // This is the Printful variant ID we need
			Size       string `json:"size"`
			Color      string `json:"color"`
		} `json:"sync_variants"`
	} `json:"result"`
}

func main() {
	log.Println("Printful Variant ID Sync Utility")
	log.Println("=================================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Fetch Printful products
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.printful.com/store/products", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.PrintfulAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to fetch products: %v", err)
	}
	defer resp.Body.Close()

	var productList PrintfulProductList
	if err := json.NewDecoder(resp.Body).Decode(&productList); err != nil {
		log.Fatalf("Failed to decode products: %v", err)
	}

	log.Printf("Found %d products in Printful\n", len(productList.Result))

	// For each product, fetch details and update variants
	updateCount := 0
	for _, product := range productList.Result {
		log.Printf("\nProcessing: %s (ID: %d)", product.Name, product.ID)

		// Fetch product details
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.printful.com/store/products/%d", product.ID), nil)
		req.Header.Set("Authorization", "Bearer "+cfg.PrintfulAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("  Error fetching details: %v", err)
			continue
		}

		var details PrintfulProductDetail
		if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
			log.Printf("  Error decoding details: %v", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Update each variant
		for _, syncVariant := range details.Result.SyncVariants {
			// Try multiple matching strategies
			var dbVariantID string
			var currentPrintfulID int64
			var err error

			// Strategy 1: Match on size in variant name
			if syncVariant.Size != "" {
				err = db.QueryRow(`
					SELECT v.id, v.printful_variant_id
					FROM variants v
					JOIN products p ON v.product_id = p.id
					WHERE p.name = ? AND v.name LIKE ?
				`, product.Name, "%"+syncVariant.Size+"%").Scan(&dbVariantID, &currentPrintfulID)
			}

			// Strategy 2: If size didn't work, try exact variant name match
			if err != nil {
				err = db.QueryRow(`
					SELECT v.id, v.printful_variant_id
					FROM variants v
					JOIN products p ON v.product_id = p.id
					WHERE p.name = ? AND v.name = ?
				`, product.Name, syncVariant.Name).Scan(&dbVariantID, &currentPrintfulID)
			}

			// Strategy 3: If only one variant, just match on product
			if err != nil && len(details.Result.SyncVariants) == 1 {
				err = db.QueryRow(`
					SELECT v.id, v.printful_variant_id
					FROM variants v
					JOIN products p ON v.product_id = p.id
					WHERE p.name = ?
					LIMIT 1
				`, product.Name).Scan(&dbVariantID, &currentPrintfulID)
			}

			if err != nil {
				log.Printf("  ⚠️  No DB match for: %s (size: %s, color: %s)", syncVariant.Name, syncVariant.Size, syncVariant.Color)
				continue
			}

			// Update the variant with Printful ID
			_, err = db.Exec(`
				UPDATE variants
				SET printful_variant_id = ?, updated_at = datetime('now')
				WHERE id = ?
			`, syncVariant.VariantID, dbVariantID)

			if err != nil {
				log.Printf("  ❌ Failed to update variant %s: %v", dbVariantID, err)
			} else {
				log.Printf("  ✅ Updated: %s → Printful ID %d", syncVariant.Name, syncVariant.VariantID)
				updateCount++
			}
		}
	}

	log.Printf("\n\n=================================")
	log.Printf("✅ Sync complete! Updated %d variants", updateCount)
	log.Printf("=================================")
}
