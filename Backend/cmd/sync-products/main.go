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

// Local product image mapping
func getLocalImagePath(productName string) string {
	imageMap := map[string]string{
		"Nessie Audio Unisex t-shirt":          "http://localhost:8080/Product Photos/Nessie Audio Unisex t-shirt/unisex-staple-t-shirt-black-back-6947058beaf9f.jpg",
		"Nessie Audio Unisex Champion hoodie":  "http://localhost:8080/Product Photos/Nessie Audio Unisex Champion hoodie/unisex-champion-hoodie-black-back-694705e44574e.png",
		"Nessie Audio Black Glossy Mug":        "http://localhost:8080/Product Photos/Nessie Audio Black Glossy Mug/black-glossy-mug-black-11-oz-handle-on-right-694706e20d560.jpg",
		"Hardcover bound Nessie Audio notebook": "http://localhost:8080/Product Photos/Hardcover bound Nessie Audio notebook/hardcover-bound-notebook-black-front-6947075450efd.jpg",
		"Nessie Audio Eco Tote Bag":            "http://localhost:8080/Product Photos/Nessie Audio Eco Tote Bag/eco-tote-bag-black-front-694707a54ec5c.jpg",
		"Nessie Audio Bubble-free stickers":    "http://localhost:8080/Product Photos/Nessie Audio Bubble-free stickers/kiss-cut-stickers-white-3x3-default-6947069ac72f0.jpg",
	}
	
	if path, ok := imageMap[productName]; ok {
		return path
	}
	return "" // Fallback to Printful CDN if no local image
}

// Product description mapping
func getProductDescription(productName string) string {
	descriptionMap := map[string]string{
		"Nessie Audio Unisex t-shirt": `The Unisex Staple T-Shirt feels soft and light with just the right amount of stretch. It's comfortable and flattering for all. We can't compliment this shirt enough–it's one of our crowd favorites, and it's sure to be your next favorite too!

Disclaimer: The fabric is slightly sheer and may appear see-through, especially in lighter colors or under certain lighting conditions.`,
		"Nessie Audio Unisex Champion hoodie": `A classic hoodie that combines Champion's signature quality with everyday comfort. The cotton-poly blend makes it soft and durable, while the two-ply hood and snug rib-knit cuffs lock in warmth. Champion's double Dry® technology keeps the wearer dry on the move, and the kangaroo pocket keeps essentials handy.

Disclaimer: Size up for a looser fit.`,
		"Nessie Audio Black Glossy Mug": `Sturdy and sleek in glossy black—this mug is a cupboard essential for a morning java or afternoon tea. 

- Ceramic
- 11 oz mug dimensions: 3.85″ × 3.35″ (9.8 cm × 8.5 cm)
- 15 oz mug dimensions: 4.7″ × 3.35″ (12 cm × 8.5 cm)
- Lead and BPA-free material
- Dishwasher and microwave safe`,
		"Hardcover bound Nessie Audio notebook": `Whether crafting a masterpiece or brainstorming the next big idea, the Hardcover Bound Notebook will inspire your inner wordsmith. The notebook features 80 lined, cream-colored pages, a built-in elastic closure, and a matching ribbon page marker. Plus, the expandable inner pocket is perfect for storing loose notes and business cards to never lose track of important information. 

- Cover material: UltraHyde hardcover paper
- Size: 5.5" × 8.5" (13.97 cm × 21.59 cm)
- Weight: 10.9 oz (309 g)
- 80 pages of lined, cream-colored paper
- Matching elastic closure and ribbon marker
- Expandable inner pocket`,
		"Nessie Audio Eco Tote Bag": `There's nothing trendier than being eco-friendly! 

- 100% certified organic cotton 3/1 twill
- Fabric weight: 8 oz/yd² (272 g/m²)
- Dimensions: 16″ × 14 ½″ × 5″ (40.6 cm × 35.6 cm × 12.7 cm)
- Weight limit: 30 lbs (13.6 kg)
- 1″ (2.5 cm) wide dual straps, 24.5″ (62.2 cm) length
- Open main compartment
- The fabric of this product holds certifications for its organic cotton content under GOTS (Global Organic Textile Standard) and OCS (Organic Content Standard)
- The fabric of this product is OEKO-TEX Standard 100 certified and PETA-Approved Vegan`,
		"Nessie Audio Bubble-free stickers": `Available in four sizes and there are no order minimums, so you can get a single sticker or a whole stack — the world is your oyster.

- High opacity film that's impossible to see through
- Durable vinyl
- 95µ thickness
- Fast and easy bubble-free application`,
	}
	
	if desc, ok := descriptionMap[productName]; ok {
		return desc
	}
	return productName // Fallback to product name if no custom description
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

		// Use local image path instead of Printful CDN
		imageURL := getLocalImagePath(item.SyncProduct.Name)
		if imageURL == "" {
			imageURL = item.SyncProduct.Thumbnail // Fallback to Printful
		}

		// Use custom description if available
		description := getProductDescription(item.SyncProduct.Name)

		// Insert product
		_, err = db.Exec(`
			INSERT INTO products (
				id, name, description, printful_id, 
				price, currency, category, image_url, active, 
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		`, productID, item.SyncProduct.Name, description, item.SyncProduct.ID, 
			price, currency, category, imageURL, 1)

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
