package models

import "time"

// Product represents a product in the store (synced from Printful)
type Product struct {
	ID              string    `json:"id" db:"id"`
	PrintfulID      int64     `json:"printful_id" db:"printful_id"` // TODO: Get from Printful API
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Price           float64   `json:"price" db:"price"`           // Your price (can markup from Printful)
	Currency        string    `json:"currency" db:"currency"`
	ImageURL        string    `json:"image_url" db:"image_url"`
	ThumbnailURL    string    `json:"thumbnail_url" db:"thumbnail_url"`
	Category        string    `json:"category" db:"category"`
	Active          bool      `json:"active" db:"active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Variant represents a product variant (size, color, etc.)
type Variant struct {
	ID              string    `json:"id" db:"id"`
	ProductID       string    `json:"product_id" db:"product_id"`
	PrintfulVariantID int64   `json:"printful_variant_id" db:"printful_variant_id"` // TODO: From Printful
	Name            string    `json:"name" db:"name"` // e.g., "Large / Black"
	Size            string    `json:"size" db:"size"`
	Color           string    `json:"color" db:"color"`
	Price           float64   `json:"price" db:"price"` // Variant-specific price override
	Available       bool      `json:"available" db:"available"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Order represents a customer order
type Order struct {
	ID                    string    `json:"id" db:"id"`
	CustomerID            string    `json:"customer_id" db:"customer_id"`
	CustomerEmail         string    `json:"customer_email" db:"customer_email"`
	Status                string    `json:"status" db:"status"` // pending, paid, fulfilled, shipped, cancelled
	TotalAmount           float64   `json:"total_amount" db:"total_amount"`
	Currency              string    `json:"currency" db:"currency"`
	StripeSessionID       string    `json:"stripe_session_id" db:"stripe_session_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id" db:"stripe_payment_intent_id"`
	PrintfulOrderID       int64     `json:"printful_order_id,omitempty" db:"printful_order_id"` // Set after submission
	PrintfulRetryCount    int       `json:"printful_retry_count" db:"printful_retry_count"` // Number of retry attempts
	ShippingName          string    `json:"shipping_name" db:"shipping_name"`
	ShippingAddress1      string    `json:"shipping_address1" db:"shipping_address1"`
	ShippingAddress2      string    `json:"shipping_address2" db:"shipping_address2"`
	ShippingCity          string    `json:"shipping_city" db:"shipping_city"`
	ShippingState         string    `json:"shipping_state" db:"shipping_state"`
	ShippingZip           string    `json:"shipping_zip" db:"shipping_zip"`
	ShippingCountry       string    `json:"shipping_country" db:"shipping_country"`
	TrackingNumber        string    `json:"tracking_number,omitempty" db:"tracking_number"`
	TrackingURL           string    `json:"tracking_url,omitempty" db:"tracking_url"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
	ID                string    `json:"id" db:"id"`
	OrderID           string    `json:"order_id" db:"order_id"`
	ProductID         string    `json:"product_id" db:"product_id"`
	VariantID         string    `json:"variant_id" db:"variant_id"`
	PrintfulVariantID int64     `json:"printful_variant_id" db:"printful_variant_id"` // Fetched from variants table
	Quantity          int       `json:"quantity" db:"quantity"`
	UnitPrice         float64   `json:"unit_price" db:"unit_price"`
	TotalPrice        float64   `json:"total_price" db:"total_price"`
	ProductName       string    `json:"product_name" db:"product_name"`     // Snapshot at order time
	VariantName       string    `json:"variant_name" db:"variant_name"`     // Snapshot
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// Customer represents a customer
type Customer struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PrintfulWebhookEvent stores webhook events for audit
type PrintfulWebhookEvent struct {
	ID        string    `json:"id" db:"id"`
	EventType string    `json:"event_type" db:"event_type"`
	OrderID   string    `json:"order_id" db:"order_id"`
	Payload   string    `json:"payload" db:"payload"` // JSON blob
	Processed bool      `json:"processed" db:"processed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// StripeWebhookEvent stores Stripe webhook events for audit
type StripeWebhookEvent struct {
	ID        string    `json:"id" db:"id"`
	EventType string    `json:"event_type" db:"event_type"`
	EventID   string    `json:"event_id" db:"event_id"` // Stripe event ID
	Payload   string    `json:"payload" db:"payload"`   // JSON blob
	Processed bool      `json:"processed" db:"processed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PrintfulSubmissionFailure represents a failed Printful order submission attempt
type PrintfulSubmissionFailure struct {
	ID            string    `json:"id" db:"id"`
	OrderID       string    `json:"order_id" db:"order_id"`
	AttemptNumber int       `json:"attempt_number" db:"attempt_number"`
	ErrorMessage  string    `json:"error_message" db:"error_message"`
	ErrorDetails  string    `json:"error_details" db:"error_details"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Order status constants
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFulfilled = "fulfilled"
	OrderStatusShipped   = "shipped"
	OrderStatusCancelled = "cancelled"
)
