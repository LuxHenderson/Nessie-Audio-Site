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

// CartCheckoutRequest represents a cart-based checkout request
type CartCheckoutRequest struct {
	Items []CartCheckoutItem `json:"items"`
	Email string             `json:"email"`
}

// CartCheckoutItem represents a single item in the cart
type CartCheckoutItem struct {
	ProductID string `json:"product_id"` // UUID string
	VariantID string `json:"variant_id"` // UUID string
	Quantity  int    `json:"quantity"`
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

// CreateCartCheckout creates a Stripe checkout session directly from cart data
// POST /api/v1/cart/checkout
//
// Frontend contract:
// Request: { 
//   "items": [{"product_id": 1, "variant_id": 1, "quantity": 2}], 
//   "email": "customer@example.com" 
// }
// Response: { "session_id": "cs_test_..." }
func (h *Handler) CreateCartCheckout(w http.ResponseWriter, r *http.Request) {
	var req CartCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode cart checkout request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	log.Printf("Cart checkout request received: %d items, email: %s", len(req.Items), req.Email)

	// Email is optional - Stripe will collect it if not provided
	// Validate items
	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "Cart is empty")
		return
	}

	// Build line items by querying database for each cart item
	var lineItems []stripe.CheckoutLineItem
	for _, cartItem := range req.Items {
		// Get product name
		var productName string
		err := h.db.QueryRow("SELECT name FROM products WHERE id = ?", cartItem.ProductID).Scan(&productName)
		if err != nil {
			log.Printf("Failed to get product %d: %v", cartItem.ProductID, err)
			respondError(w, http.StatusBadRequest, "Invalid product")
			return
		}

		// Get variant name and price
		var variantName string
		var price float64
		err = h.db.QueryRow("SELECT name, price FROM variants WHERE id = ? AND product_id = ?", 
			cartItem.VariantID, cartItem.ProductID).Scan(&variantName, &price)
		if err != nil {
			log.Printf("Failed to get variant %d for product %d: %v", cartItem.VariantID, cartItem.ProductID, err)
			respondError(w, http.StatusBadRequest, "Invalid variant")
			return
		}

		lineItems = append(lineItems, stripe.CheckoutLineItem{
			ProductName: productName,
			VariantName: variantName,
			Quantity:    int64(cartItem.Quantity),
			UnitPrice:   int64(price * 100), // Convert to cents
		})
	}

	// Create Stripe checkout session
	sessionID, err := h.stripeClient.CreateCheckoutSession(&stripe.CheckoutSessionRequest{
		OrderID:       "", // No order created yet
		CustomerEmail: req.Email,
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
