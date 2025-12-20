package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
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
}

// NewHandler creates a new handler with dependencies
func NewHandler(
	db *sql.DB,
	cfg *config.Config,
	printfulClient *printful.Client,
	stripeClient *stripe.Client,
	orderService *order.Service,
) *Handler {
	return &Handler{
		db:             db,
		config:         cfg,
		printfulClient: printfulClient,
		stripeClient:   stripeClient,
		orderService:   orderService,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Products
	api.HandleFunc("/products", h.GetProducts).Methods("GET")
	api.HandleFunc("/products/{id}", h.GetProduct).Methods("GET")

	// Orders
	api.HandleFunc("/orders", h.CreateOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", h.GetOrder).Methods("GET")

	// Checkout
	api.HandleFunc("/checkout", h.CreateCheckout).Methods("POST")

	// Webhooks (no auth required)
	r.HandleFunc("/webhooks/stripe", h.HandleStripeWebhook).Methods("POST")
	r.HandleFunc("/webhooks/printful", h.HandlePrintfulWebhook).Methods("POST")

	// Health check
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

// HealthCheck returns server health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "nessie-audio-ecommerce",
	})
}
