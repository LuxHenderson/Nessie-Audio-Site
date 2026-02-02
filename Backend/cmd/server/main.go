package main

import (
	"context"
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

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Nessie Audio eCommerce Backend (env: %s)", cfg.Env)

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

	// Serve frontend static files as catch-all AFTER API routes
	// gorilla/mux matches in registration order, so API routes take priority
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		// Serve homepage
		if urlPath == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "Nævermore.html"))
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

		// Serve the file if it exists
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}

		// Unknown path — serve homepage
		http.ServeFile(w, r, filepath.Join(staticDir, "Nævermore.html"))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
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
