package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
)

type PrintfulProduct struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail_url"`
	Currency  string `json:"currency"`
	Variants  int    `json:"variants"`
}

type PrintfulVariant struct {
	ID            int     `json:"id"`
	SyncVariantID int     `json:"sync_variant_id"`
	Name          string  `json:"name"`
	Price         string  `json:"retail_price"`
	Currency      string  `json:"currency"`
	ProductID     int     `json:"product_id"`
	Files         []struct {
		Type string `json:"type"`
	} `json:"files"`
}

type PrintfulProductDetail struct {
	SyncProduct  PrintfulProduct   `json:"sync_product"`
	SyncVariants []PrintfulVariant `json:"sync_variants"`
}

type PrintfulListResponse struct {
	Code   int               `json:"code"`
	Result []PrintfulProduct `json:"result"`
}

type PrintfulDetailResponse struct {
	Code   int                    `json:"code"`
	Result PrintfulProductDetail `json:"result"`
}

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

	log.Println("Fetching products from Printful...")

	// Fetch products list from Printful
	req, err := http.NewRequest("GET", cfg.PrintfulAPIURL+"/sync/products", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.PrintfulAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to fetch products: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Printful API returned status %d. Check your API key in .env", resp.StatusCode)
	}

	var listResp PrintfulListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	log.Printf("Found %d products from Printful", len(listResp.Result))

	// Fetch details for each product to get variants
	for _, product := range listResp.Result {
		// Fetch product detail
		detailReq, err := http.NewRequest("GET", cfg.PrintfulAPIURL+"/sync/products/"+fmt.Sprintf("%d", product.ID), nil)
		if err != nil {
			log.Printf("Failed to create request for product %d: %v", product.ID, err)
			continue
		}
		detailReq.Header.Set("Authorization", "Bearer "+cfg.PrintfulAPIKey)

		detailResp, err := client.Do(detailReq)
		if err != nil {
			log.Printf("Failed to fetch product %d: %v", product.ID, err)
			continue
		}
		defer detailResp.Body.Close()

		var detail PrintfulDetailResponse
		if err := json.NewDecoder(detailResp.Body).Decode(&detail); err != nil {
			log.Printf("Failed to decode product %d: %v", product.ID, err)
			continue
		}

		item := detail.Result
		
		// Generate UUID for product
		productID := uuid.New().String()

		// Determine category from product name (simple heuristic)
		category := "merch"
		
		// Calculate base price from first variant
		price := "0.00"
		currency := "USD"
		if len(item.SyncVariants) > 0 {
			price = item.SyncVariants[0].Price
			currency = item.SyncVariants[0].Currency
		}

		// Insert product
		_, err = db.Exec(`
			INSERT INTO products (
				id, name, description, printful_id, 
				price, currency, category, image_url, active, 
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		`, productID, item.SyncProduct.Name, item.SyncProduct.Name, item.SyncProduct.ID, 
			price, currency, category, item.SyncProduct.Thumbnail, 1)

		if err != nil {
			log.Printf("Failed to insert product %s: %v", item.SyncProduct.Name, err)
			continue
		}

		log.Printf("✓ Added product: %s (ID: %s)", item.SyncProduct.Name, productID)

		// Insert variants
		for _, variant := range item.SyncVariants {
			variantID := uuid.New().String()

			_, err := db.Exec(`
				INSERT INTO variants (
					id, product_id, printful_variant_id, name,
					price, available,
					created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
			`, variantID, productID, variant.SyncVariantID, variant.Name,
				variant.Price, 1)

			if err != nil {
				log.Printf("  Failed to insert variant %s: %v", variant.Name, err)
				continue
			}

			log.Printf("  ✓ Added variant: %s", variant.Name)
		}
	}

	// Show summary
	var productCount int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&productCount)
	
	var variantCount int
	db.QueryRow("SELECT COUNT(*) FROM variants").Scan(&variantCount)

	log.Printf("\n✅ Sync complete!")
	log.Printf("Total products in database: %d", productCount)
	log.Printf("Total variants in database: %d", variantCount)
}
