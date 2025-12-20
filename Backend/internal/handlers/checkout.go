package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nessieaudio/ecommerce-backend/internal/services/stripe"
)

// CreateCheckoutRequest represents checkout initiation request
type CreateCheckoutRequest struct {
	OrderID string `json:"order_id"`
}

// CreateCheckoutResponse contains the Stripe session ID
type CreateCheckoutResponse struct {
	SessionID string `json:"session_id"`
}

// CreateCheckout initiates a Stripe checkout session
// POST /api/v1/checkout
//
// Frontend contract:
// Request: { "order_id": "uuid" }
// Response: { "session_id": "cs_test_..." }
// Frontend should redirect to Stripe using this session ID
func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Get order
	order, err := h.orderService.GetOrder(req.OrderID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Get order items
	items, err := h.orderService.GetOrderItems(req.OrderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get order items")
		return
	}

	// Get customer email
	var customerEmail string
	err = h.db.QueryRow("SELECT email FROM customers WHERE id = ?", order.CustomerID).Scan(&customerEmail)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get customer")
		return
	}

	// Build line items for Stripe
	var lineItems []stripe.CheckoutLineItem
	for _, item := range items {
		lineItems = append(lineItems, stripe.CheckoutLineItem{
			ProductName: item.ProductName,
			VariantName: item.VariantName,
			Quantity:    int64(item.Quantity),
			UnitPrice:   int64(item.UnitPrice * 100), // Convert to cents
		})
	}

	// Create Stripe checkout session
	sessionID, err := h.stripeClient.CreateCheckoutSession(&stripe.CheckoutSessionRequest{
		OrderID:       order.ID,
		CustomerEmail: customerEmail,
		LineItems:     lineItems,
	})

	if err != nil {
		log.Printf("Stripe checkout error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	respondJSON(w, http.StatusOK, CreateCheckoutResponse{
		SessionID: sessionID,
	})
}
