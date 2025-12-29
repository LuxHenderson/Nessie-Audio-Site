package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
	"github.com/nessieaudio/ecommerce-backend/internal/handlers"
	"github.com/nessieaudio/ecommerce-backend/internal/middleware"
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
	defer db.Close()

	log.Println("Database initialized")

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

	// Initialize handlers
	handler := handlers.NewHandler(db, cfg, printfulClient, stripeClient, orderService, emailClient)

	// Setup router
	router := mux.NewRouter()

	// Apply middleware FIRST (before routes)
	router.Use(middleware.Recovery)
	router.Use(middleware.Logging)
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// Serve static files from Product Photos directory BEFORE registering API routes
	// This allows the frontend to load local product images
	productPhotosPath := "../Product Photos"
	if _, err := os.Stat(productPhotosPath); err == nil {
		fs := http.FileServer(http.Dir(".."))
		router.PathPrefix("/Product Photos/").Handler(fs)
		log.Println("Serving product images from local Product Photos directory")
	} else {
		log.Println("⚠️  WARNING: Product Photos directory not found at", productPhotosPath)
	}

	handler.RegisterRoutes(router)

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

	log.Println("Shutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
	log.Println("Server stopped")
}
