package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/logger"
	"github.com/nessieaudio/ecommerce-backend/internal/middleware"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
	"github.com/nessieaudio/ecommerce-backend/internal/services/order"
	"github.com/nessieaudio/ecommerce-backend/internal/services/printful"
	"github.com/nessieaudio/ecommerce-backend/internal/services/stripe"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	db            *sql.DB
	config        *config.Config
	printfulClient *printful.Client
	stripeClient   *stripe.Client
	orderService   *order.Service
	emailClient    *email.Client
	logger         *logger.Logger
}

// NewHandler creates a new handler with dependencies
func NewHandler(
	db *sql.DB,
	cfg *config.Config,
	printfulClient *printful.Client,
	stripeClient *stripe.Client,
	orderService *order.Service,
	emailClient *email.Client,
	appLogger *logger.Logger,
) *Handler {
	return &Handler{
		db:             db,
		config:         cfg,
		printfulClient: printfulClient,
		stripeClient:   stripeClient,
		orderService:   orderService,
		emailClient:    emailClient,
		logger:         appLogger,
	}
}

// RegisterRoutes registers all API routes with appropriate rate limiting
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Rate limiters for different endpoint types
	// Token Bucket: (capacity, refillRate per second)

	// Public read endpoints - generous limits (100 tokens, refill 2/sec = ~120/min)
	publicLimiter := middleware.RateLimit(100, 2.0)

	// Checkout/Orders - stricter limits (20 tokens, refill 0.33/sec = ~20/min)
	checkoutLimiter := middleware.RateLimit(20, 0.33)

	// General API - moderate limits (60 tokens, refill 1/sec = ~60/min)
	generalLimiter := middleware.RateLimit(60, 1.0)

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Products - Public read endpoints (higher limit)
	api.Handle("/products", publicLimiter(http.HandlerFunc(h.GetProducts))).Methods("GET")
	api.Handle("/products/{id}", publicLimiter(http.HandlerFunc(h.GetProduct))).Methods("GET")

	// Orders - Moderate limits
	api.Handle("/orders", checkoutLimiter(http.HandlerFunc(h.CreateOrder))).Methods("POST")
	api.Handle("/orders/{id}", generalLimiter(http.HandlerFunc(h.GetOrder))).Methods("GET")

	// Checkout - Strict limits (most important to protect)
	api.Handle("/checkout", checkoutLimiter(http.HandlerFunc(h.CreateCheckout))).Methods("POST", "OPTIONS")
	api.Handle("/cart/checkout", checkoutLimiter(http.HandlerFunc(h.CreateCartCheckout))).Methods("POST", "OPTIONS")

	// Config - General limits
	api.Handle("/config", generalLimiter(http.HandlerFunc(h.GetConfig))).Methods("GET")

	// Webhooks - NO rate limiting (Stripe/Printful need reliable delivery)
	r.HandleFunc("/webhooks/stripe", h.HandleStripeWebhook).Methods("POST")
	r.HandleFunc("/webhooks/printful/{token}", h.HandlePrintfulWebhook).Methods("POST")

	// Health check - NO rate limiting (used for monitoring)
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// GetConfig returns public configuration for frontend
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"stripe_publishable_key": h.config.StripePublishableKey,
	})
}

// HealthCheck returns server health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "nessie-audio-ecommerce",
	})
}
