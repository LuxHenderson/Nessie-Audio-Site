package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/backup"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
	"github.com/nessieaudio/ecommerce-backend/internal/handlers"
	"github.com/nessieaudio/ecommerce-backend/internal/logger"
	"github.com/nessieaudio/ecommerce-backend/internal/middleware"
	"github.com/nessieaudio/ecommerce-backend/internal/migrations"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
	"github.com/nessieaudio/ecommerce-backend/internal/services/order"
	"github.com/nessieaudio/ecommerce-backend/internal/services/printful"
	"github.com/nessieaudio/ecommerce-backend/internal/services/stripe"
)

// seedProductsIfEmpty checks if the products table is empty and seeds it with initial products
func seedProductsIfEmpty(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check products count: %w", err)
	}

	if count > 0 {
		log.Printf("Products table already populated with %d products", count)
		return nil
	}

	log.Println("📦 Products table is empty - seeding with initial products...")

	products := []struct {
		id          string
		printfulID  int
		name        string
		description string
		price       float64
		imageURL    string
	}{
		{
			id:          "4f92e8f5-dc35-4c67-ae47-2e41f959680f",
			printfulID:  408670865,
			name:        "Nessie Audio Eco Tote Bag",
			description: "There's nothing trendier than being eco-friendly!\n\n- 100% certified organic cotton 3/1 twill\n- Fabric weight: 8 oz/yd² (272 g/m²)\n- Dimensions: 16″ × 14 ½″ × 5″\n- Weight limit: 30 lbs (13.6 kg)\n- OEKO-TEX Standard 100 certified and PETA-Approved Vegan",
			price:       25.0,
			imageURL:    "/Product Photos/Nessie Audio Eco Tote Bag/eco-tote-bag-black-front-694707a54ec5c.jpg",
		},
		{
			id:          "7eb5405b-ba58-4564-a395-b0d17e8d45e9",
			printfulID:  408670806,
			name:        "Hardcover bound Nessie Audio notebook",
			description: "Whether crafting a masterpiece or brainstorming the next big idea, the Hardcover Bound Notebook will inspire your inner wordsmith.\n\n- Cover material: UltraHyde hardcover paper\n- Size: 5.5\" × 8.5\" (13.97 cm × 21.59 cm)\n- 80 pages of lined, cream-colored paper\n- Matching elastic closure and ribbon marker",
			price:       20.0,
			imageURL:    "/Product Photos/Hardcover bound Nessie Audio notebook/hardcover-bound-notebook-black-front-6947075450efd.jpg",
		},
		{
			id:          "bd45da14-cd20-4840-8095-29a0547c6f6f",
			printfulID:  408670774,
			name:        "Nessie Audio Bubble-free stickers",
			description: "Available in four sizes and there are no order minimums.\n\n- High opacity film that's impossible to see through\n- Durable vinyl\n- 95µ thickness\n- Fast and easy bubble-free application",
			price:       5.0,
			imageURL:    "/Product Photos/Nessie Audio Bubble-free stickers/kiss-cut-stickers-white-3x3-default-6947069ac72f0.jpg",
		},
		{
			id:          "331ff894-0eaa-43f9-bd8b-626eb29656fc",
			printfulID:  408670710,
			name:        "Nessie Audio Black Glossy Mug",
			description: "Sturdy and sleek in glossy black—this mug is a cupboard essential.\n\n- Ceramic\n- 11 oz mug dimensions: 3.85″ × 3.35″\n- 15 oz mug dimensions: 4.7″ × 3.35″\n- Lead and BPA-free\n- Dishwasher and microwave safe",
			price:       15.0,
			imageURL:    "/Product Photos/Nessie Audio Black Glossy Mug/black-glossy-mug-black-11-oz-handle-on-right-694706e20d560.jpg",
		},
		{
			id:          "b33c14d3-dadd-41f0-b404-f055f0d406fa",
			printfulID:  408670639,
			name:        "Nessie Audio Unisex Champion hoodie",
			description: "A classic hoodie that combines Champion's signature quality with everyday comfort.\n\nDisclaimer: Size up for a looser fit.",
			price:       40.0,
			imageURL:    "/Product Photos/Nessie Audio Unisex Champion hoodie/unisex-champion-hoodie-black-back-694705e44574e.png",
		},
		{
			id:          "86ebaeb1-4889-4f79-83f3-b3ad22e8652e",
			printfulID:  408670558,
			name:        "Nessie Audio Unisex t-shirt",
			description: "The Unisex Staple T-Shirt feels soft and light with just the right amount of stretch.\n\nDisclaimer: The fabric is slightly sheer and may appear see-through in lighter colors.",
			price:       15.0,
			imageURL:    "/Product Photos/Nessie Audio Unisex t-shirt/unisex-staple-t-shirt-black-back-6947058beaf9f.jpg",
		},
	}

	for _, p := range products {
		_, err := db.Exec(`
			INSERT INTO products (id, printful_id, name, description, price, currency, image_url, category, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'USD', ?, 'merch', 1, datetime('now'), datetime('now'))
		`, p.id, p.printfulID, p.name, p.description, p.price, p.imageURL)

		if err != nil {
			log.Printf("  ⚠️  Failed to seed product: %s - %v", p.name, err)
		} else {
			log.Printf("  ✓ Seeded: %s", p.name)
		}
	}

	log.Println("✅ Product seeding complete!")

	// Also seed variants for the newly created products
	return seedVariantsIfEmpty(db)
}

// seedVariantsIfEmpty populates the variants table with Printful variant data
func seedVariantsIfEmpty(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM variants").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check variants count: %w", err)
	}

	if count > 0 {
		log.Printf("Variants table already populated with %d variants", count)
		return nil
	}

	log.Println("📦 Variants table is empty - seeding with Printful variant data...")

	type variant struct {
		id         string
		productID  string
		printfulID int
		name       string
		price      float64
	}

	variants := []variant{
		// Eco Tote Bag (1 variant)
		{"b06c0f89-98d2-4171-b416-8b471f1e591b", "4f92e8f5-dc35-4c67-ae47-2e41f959680f", 5117581114, "Nessie Audio Eco Tote Bag", 25.0},
		// Notebook (1 variant)
		{"d1e37055-22ef-4e50-82fd-712145ec0b70", "7eb5405b-ba58-4564-a395-b0d17e8d45e9", 5117580723, "Hardcover bound Nessie Audio notebook / Black", 20.0},
		// Stickers (4 variants)
		{"6bdaef77-07e7-4b4a-8284-d8366e21c467", "bd45da14-cd20-4840-8095-29a0547c6f6f", 5117580378, "Nessie Audio Bubble-free stickers / 3″×3″", 5.0},
		{"ec2a30f3-f328-4322-acfb-fabd22bac612", "bd45da14-cd20-4840-8095-29a0547c6f6f", 5117580379, "Nessie Audio Bubble-free stickers / 4″×4″", 6.0},
		{"f77df16f-b273-4768-9715-f2b011a11738", "bd45da14-cd20-4840-8095-29a0547c6f6f", 5117580380, "Nessie Audio Bubble-free stickers / 5.5″×5.5″", 7.0},
		{"9592ae91-af5d-4435-be2d-df1694bb5b16", "bd45da14-cd20-4840-8095-29a0547c6f6f", 5117580381, "Nessie Audio Bubble-free stickers / 15″×3.75″", 8.0},
		// Mug (2 variants)
		{"16dc11bb-bc42-4f24-8d40-6fdd779abb6f", "331ff894-0eaa-43f9-bd8b-626eb29656fc", 5117579999, "Nessie Audio Black Glossy Mug / 11 oz", 15.0},
		{"0f207194-b7c1-4cf5-b7aa-597e05405e01", "331ff894-0eaa-43f9-bd8b-626eb29656fc", 5117580000, "Nessie Audio Black Glossy Mug / 15 oz", 18.0},
		// Hoodie (6 variants)
		{"f517b811-a52f-4d49-b26e-f4ae19d247f3", "b33c14d3-dadd-41f0-b404-f055f0d406fa", 5117579650, "Nessie Audio Unisex Champion hoodie / S", 40.0},
		{"24b3f297-0fef-4d4b-9ab2-0eb8b9458be7", "b33c14d3-dadd-41f0-b404-f055f0d406fa", 5117579651, "Nessie Audio Unisex Champion hoodie / M", 40.0},
		{"ab240853-b5b3-46ef-8c53-54a3b6c6dfef", "b33c14d3-dadd-41f0-b404-f055f0d406fa", 5117579652, "Nessie Audio Unisex Champion hoodie / L", 45.0},
		{"2c14ac74-f023-433e-a49b-1d189ee2ad0c", "b33c14d3-dadd-41f0-b404-f055f0d406fa", 5117579653, "Nessie Audio Unisex Champion hoodie / XL", 45.0},
		{"0157802d-f98a-4b15-a5a8-7494b8b42e3e", "b33c14d3-dadd-41f0-b404-f055f0d406fa", 5117579654, "Nessie Audio Unisex Champion hoodie / 2XL", 50.0},
		{"1f7762b1-0776-4d63-a5b2-25f9a120b995", "b33c14d3-dadd-41f0-b404-f055f0d406fa", 5117579655, "Nessie Audio Unisex Champion hoodie / 3XL", 50.0},
		// T-shirt (9 variants)
		{"ed77214c-43fc-4a3a-baa0-30dc9de85199", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578987, "Nessie Audio Unisex t-shirt / XS", 15.0},
		{"bc8ae324-b794-4a08-bf65-5dfa55c31457", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578988, "Nessie Audio Unisex t-shirt / S", 15.0},
		{"811ae62b-3ff3-4276-995f-6ca6803a72ee", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578989, "Nessie Audio Unisex t-shirt / M", 15.0},
		{"e599896e-a602-4f49-95b3-fe835ac8f7f9", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578990, "Nessie Audio Unisex t-shirt / L", 20.0},
		{"cb582a1d-9e23-4136-991e-3d64abbb52c2", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578991, "Nessie Audio Unisex t-shirt / XL", 20.0},
		{"567a02f6-51d7-487b-af02-f7cd9f878c39", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578992, "Nessie Audio Unisex t-shirt / 2XL", 20.0},
		{"a1e6cf12-635f-46c2-b75a-b2b31a69bbb2", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578993, "Nessie Audio Unisex t-shirt / 3XL", 25.0},
		{"f86c585b-7725-48d7-aedb-f65a86c47201", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578994, "Nessie Audio Unisex t-shirt / 4XL", 25.0},
		{"67dbc086-dd1c-409f-a125-bd73c1ca054f", "86ebaeb1-4889-4f79-83f3-b3ad22e8652e", 5117578995, "Nessie Audio Unisex t-shirt / 5XL", 25.0},
	}

	for _, v := range variants {
		_, err := db.Exec(`
			INSERT INTO variants (id, product_id, printful_variant_id, name, price, available, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 1, datetime('now'), datetime('now'))
		`, v.id, v.productID, v.printfulID, v.name, v.price)

		if err != nil {
			log.Printf("  ⚠️  Failed to seed variant: %s - %v", v.name, err)
		} else {
			log.Printf("  ✓ Seeded: %s", v.name)
		}
	}

	log.Println("✅ Variant seeding complete!")
	return nil
}

func main() {
	// Ensure logging goes to stdout for Railway
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags)

	log.Println("🚀 Starting Nessie Audio Backend...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Nessie Audio eCommerce Backend (env: %s)", cfg.Env)
	log.Printf("Database path: %s", cfg.DatabasePath)

	// Initialize database
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Database initialized")

	// Run database migrations
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize services
	printfulClient := printful.NewClient(cfg.PrintfulAPIKey, cfg.PrintfulAPIURL)
	stripeClient := stripe.NewClient(
		cfg.StripeSecretKey,
		cfg.StripePublishableKey,
		cfg.StripeSuccessURL,
		cfg.StripeCancelURL,
	)
	orderService := order.NewService(db)
	emailClient := email.NewClient(cfg)

	// Initialize logger
	appLogger, err := logger.New("logs/error.log", emailClient, cfg.AdminEmail)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.Info("Nessie Audio eCommerce Backend started")

	// Initialize backup manager
	backupManager, err := backup.NewManager(db, backup.Config{
		BackupDir:        "backups",
		DatabasePath:     cfg.DatabasePath,
		DailyRetention:   30,
		MonthlyRetention: 12,
		CompressBackups:  true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize backup manager: %v", err)
	}

	// Start scheduled backups (daily at 3:00 AM)
	backupManager.StartScheduledBackups()
	appLogger.Info("Backup system initialized - daily backups at 3:00 AM")

	// Create initial backup on startup
	go func() {
		if err := backupManager.CreateBackup("daily"); err != nil {
			appLogger.Warning("Failed to create startup backup", err)
		} else {
			log.Println("Initial backup created successfully")
		}
	}()

	// Auto-seed products and variants if database is empty (non-blocking)
	go func() {
		if err := seedProductsIfEmpty(db); err != nil {
			log.Printf("⚠️  Failed to seed products: %v", err)
		}
		// Also check variants independently (products may already exist from a previous deploy)
		if err := seedVariantsIfEmpty(db); err != nil {
			log.Printf("⚠️  Failed to seed variants: %v", err)
		}
	}()

	// Initialize handlers
	handler := handlers.NewHandler(db, cfg, printfulClient, stripeClient, orderService, emailClient, appLogger)

	// Setup router
	router := mux.NewRouter()

	// Apply middleware FIRST (before routes)
	// Order matters: Recovery → RequestID → HTTPS Redirect → Security Headers → Logging → CORS
	router.Use(middleware.Recovery)
	router.Use(middleware.RequestID)
	router.Use(middleware.HTTPSRedirect(cfg.Env))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Logging)
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// Determine static files directory
	// In Docker (production): /app/static
	// In local dev (CWD is Backend/): ../
	staticDir := "/app/static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = ".."
	}
	log.Printf("Serving static files from: %s", staticDir)

	// Serve Product Photos BEFORE registering API routes
	productPhotosDir := filepath.Join(staticDir, "Product Photos")
	if _, err := os.Stat(productPhotosDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		router.PathPrefix("/Product Photos/").Handler(fs)
		log.Println("Serving product images from", productPhotosDir)
	} else {
		log.Println("⚠️  WARNING: Product Photos directory not found at", productPhotosDir)
	}

	handler.RegisterRoutes(router)

	// serveFileWithCache sets Cache-Control headers based on file type, then serves the file.
	// HTML: no-cache (always revalidate with origin - prevents Cloudflare serving stale pages)
	// Static assets: cached for 1 hour
	serveFileWithCache := func(w http.ResponseWriter, r *http.Request, filePath string) {
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == ".html" || ext == "" {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		http.ServeFile(w, r, filePath)
	}

	// Serve frontend static files as catch-all AFTER API routes
	// gorilla/mux matches in registration order, so API routes take priority
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		// Serve homepage
		if urlPath == "/" {
			serveFileWithCache(w, r, filepath.Join(staticDir, "home.html"))
			return
		}

		// Resolve and validate the file path (prevent directory traversal)
		filePath := filepath.Join(staticDir, filepath.Clean(urlPath))
		absStaticDir, _ := filepath.Abs(staticDir)
		absFilePath, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFilePath, absStaticDir) {
			http.NotFound(w, r)
			return
		}

		// Try exact file path first
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			serveFileWithCache(w, r, filePath)
			return
		}

		// If path doesn't end with .html, try adding .html extension
		if !strings.HasSuffix(urlPath, ".html") {
			htmlPath := filePath + ".html"
			if info, err := os.Stat(htmlPath); err == nil && !info.IsDir() {
				serveFileWithCache(w, r, htmlPath)
				return
			}
		}

		// Unknown path — serve homepage
		serveFileWithCache(w, r, filepath.Join(staticDir, "home.html"))
	})

	// Create HTTP server
	addr := "0.0.0.0:" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on %s", addr)
		log.Printf("API endpoints:")
		log.Printf("  - GET  /health")
		log.Printf("  - GET  /api/v1/products")
		log.Printf("  - GET  /api/v1/products/{id}")
		log.Printf("  - POST /api/v1/orders")
		log.Printf("  - GET  /api/v1/orders/{id}")
		log.Printf("  - POST /api/v1/checkout")
		log.Printf("  - POST /api/v1/cart/checkout")
		log.Printf("  - GET  /api/v1/inventory")
		log.Printf("  - GET  /api/v1/inventory/low-stock")
		log.Printf("  - PUT  /api/v1/inventory/{variant_id}")
		log.Printf("  - POST /webhooks/stripe")
		log.Printf("  - POST /webhooks/printful/{token}")
		log.Println()

		if cfg.PrintfulAPIKey == "" {
			log.Println("⚠️  WARNING: PRINTFUL_API_KEY not set")
		}
		if cfg.StripeSecretKey == "" {
			log.Println("⚠️  WARNING: STRIPE_SECRET_KEY not set")
		}
		log.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutdown signal received, initiating graceful shutdown...")
	appLogger.Info("Server shutdown initiated")

	// Create shutdown context with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop accepting new requests and wait for existing ones to complete
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		appLogger.Critical("Forced server shutdown", err, nil)
	} else {
		log.Println("✅ All active requests completed")
	}

	// Close database connections
	log.Println("🔌 Closing database connections...")
	if err := db.Close(); err != nil {
		log.Printf("⚠️  Error closing database: %v", err)
	}

	// Close logger
	log.Println("📝 Closing logger...")
	appLogger.Close()

	log.Println("✅ Server shutdown complete")
}
