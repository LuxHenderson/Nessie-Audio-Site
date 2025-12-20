package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nessieaudio/ecommerce-backend/internal/services/stripe"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
	stripeLib "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// HandleStripeWebhook processes Stripe webhook events
// POST /webhooks/stripe
//
// Critical: This endpoint MUST verify webhook signatures
// This is where you confirm payment and submit to Printful
func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		respondError(w, http.StatusBadRequest, "Error reading request")
		return
	}

	// Verify webhook signature
	// TODO: Ensure STRIPE_WEBHOOK_SECRET is set in your .env
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"),
		h.config.StripeWebhookSecret)

	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// Log webhook event
	if err := h.logStripeWebhookEvent(event); err != nil {
		log.Printf("Failed to log webhook event: %v", err)
	}

	// Process based on event type
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(event)

	case "payment_intent.succeeded":
		log.Printf("PaymentIntent succeeded: %s", event.ID)

	case "payment_intent.payment_failed":
		log.Printf("PaymentIntent failed: %s", event.ID)

	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// handleCheckoutSessionCompleted processes successful checkout
// This is where payment is confirmed and order should be submitted to Printful
func (h *Handler) handleCheckoutSessionCompleted(event stripeLib.Event) {
	var session stripeLib.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session: %v", err)
		return
	}

	// Get order ID from metadata
	orderID, ok := session.Metadata["order_id"]
	if !ok {
		log.Printf("No order_id in session metadata")
		return
	}

	// Retrieve full session with shipping details
	fullSession, err := h.stripeClient.GetSession(session.ID)
	if err != nil {
		log.Printf("Failed to get full session: %v", err)
		return
	}

	// Get order
	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		log.Printf("Failed to get order %s: %v", orderID, err)
		return
	}

	// Update order with Stripe and shipping details
	stripe.UpdateOrderFromSession(order, fullSession)

	if err := h.orderService.UpdateOrderWithStripeSession(order); err != nil {
		log.Printf("Failed to update order: %v", err)
		return
	}

	log.Printf("Order %s marked as paid", orderID)

	// ====== SUBMIT TO PRINTFUL ======
	// This is the critical step - only submit after payment is confirmed
	go h.submitOrderToPrintful(orderID)
}

// submitOrderToPrintful submits a paid order to Printful for fulfillment
// Runs asynchronously to not block webhook response
func (h *Handler) submitOrderToPrintful(orderID string) {
	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		log.Printf("Failed to get order for Printful: %v", err)
		return
	}

	items, err := h.orderService.GetOrderItems(orderID)
	if err != nil {
		log.Printf("Failed to get order items for Printful: %v", err)
		return
	}

	// Submit to Printful
	printfulOrderID, err := h.printfulClient.CreateOrder(order, items)
	if err != nil {
		log.Printf("Failed to create Printful order: %v", err)
		// TODO: Implement retry logic or alert system
		return
	}

	// Confirm the order with Printful
	if err := h.printfulClient.ConfirmOrder(printfulOrderID); err != nil {
		log.Printf("Failed to confirm Printful order: %v", err)
		return
	}

	// Update order with Printful ID
	if err := h.orderService.UpdateOrderWithPrintful(orderID, printfulOrderID); err != nil {
		log.Printf("Failed to update order with Printful ID: %v", err)
		return
	}

	// Update status to fulfilled
	if err := h.orderService.UpdateOrderStatus(orderID, models.OrderStatusFulfilled); err != nil {
		log.Printf("Failed to update order status: %v", err)
		return
	}

	log.Printf("Order %s submitted to Printful (ID: %d)", orderID, printfulOrderID)
}

// logStripeWebhookEvent saves webhook event for audit
func (h *Handler) logStripeWebhookEvent(event stripeLib.Event) error {
	payload, _ := json.Marshal(event)

	_, err := h.db.Exec(`
		INSERT INTO stripe_webhook_events (id, event_type, event_id, payload, processed, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), event.Type, event.ID, string(payload), true, time.Now())

	return err
}
