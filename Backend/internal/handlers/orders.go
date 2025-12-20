package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
)

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerEmail string      `json:"customer_email"`
	Items         []OrderItem `json:"items"`
}

// OrderItem represents an item in the order
type OrderItem struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

// CreateOrderResponse represents the order creation response
type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// CreateOrder creates a new pending order
// POST /api/v1/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.CustomerEmail == "" || len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "Email and items are required")
		return
	}

	// Get or create customer
	customerID, err := h.getOrCreateCustomer(req.CustomerEmail)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process customer")
		return
	}

	// Calculate total and prepare order items
	var totalAmount float64
	orderItems := make([]models.OrderItem, 0, len(req.Items))

	for _, item := range req.Items {
		// Get variant details
		var variantPrice float64
		var productName, variantName string

		err := h.db.QueryRow(`
			SELECT v.price, p.name, v.name
			FROM variants v
			JOIN products p ON v.product_id = p.id
			WHERE v.id = ? AND v.available = 1
		`, item.VariantID).Scan(&variantPrice, &productName, &variantName)

		if err == sql.ErrNoRows {
			respondError(w, http.StatusBadRequest, "Variant not available")
			return
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to fetch variant")
			return
		}

		itemTotal := variantPrice * float64(item.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, models.OrderItem{
			ID:          uuid.New().String(),
			OrderID:     "", // Will be set below
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			Quantity:    item.Quantity,
			UnitPrice:   variantPrice,
			TotalPrice:  itemTotal,
			ProductName: productName,
			VariantName: variantName,
			CreatedAt:   time.Now(),
		})
	}

	// Create order
	order := &models.Order{
		ID:          uuid.New().String(),
		CustomerID:  customerID,
		Status:      models.OrderStatusPending,
		TotalAmount: totalAmount,
		Currency:    "USD",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set order ID on items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}

	// Save to database
	if err := h.orderService.CreateOrder(order, orderItems); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	respondJSON(w, http.StatusCreated, CreateOrderResponse{
		OrderID: order.ID,
		Status:  order.Status,
	})
}

// GetOrder retrieves an order by ID
// GET /api/v1/orders/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	order, err := h.orderService.GetOrder(orderID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Order not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch order")
		return
	}

	items, err := h.orderService.GetOrderItems(orderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch order items")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"order": order,
		"items": items,
	})
}

// getOrCreateCustomer gets existing customer or creates new one
func (h *Handler) getOrCreateCustomer(email string) (string, error) {
	// Check if customer exists
	var customerID string
	err := h.db.QueryRow("SELECT id FROM customers WHERE email = ?", email).Scan(&customerID)
	if err == nil {
		return customerID, nil
	}

	// Create new customer
	customerID = uuid.New().String()
	now := time.Now()

	_, err = h.db.Exec(`
		INSERT INTO customers (id, email, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, customerID, email, now, now)

	if err != nil {
		return "", err
	}

	return customerID, nil
}
