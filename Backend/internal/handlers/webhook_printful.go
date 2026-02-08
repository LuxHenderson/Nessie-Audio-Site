package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// PrintfulWebhookPayload represents a Printful webhook event
type PrintfulWebhookPayload struct {
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
	Order struct {
		ID             int64  `json:"id"`
		ExternalID     string `json:"external_id"`
		TrackingNumber string `json:"tracking_number"`
		TrackingURL    string `json:"tracking_url"`
	} `json:"order"`
}

// HandlePrintfulWebhook processes Printful webhook events
// POST /webhooks/printful/{token}
//
// Security: Uses secret token in URL since Printful doesn't provide signature verification
// Events we care about:
// - order_updated: Order status changed
// - shipment_created: Tracking info available
func (h *Handler) HandlePrintfulWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify webhook token from URL
	vars := mux.Vars(r)
	token := vars["token"]

	if h.config.PrintfulWebhookSecret == "" {
		log.Printf("WARNING: PRINTFUL_WEBHOOK_SECRET not configured - rejecting webhook")
		respondError(w, http.StatusUnauthorized, "Webhook not configured")
		return
	}

	if token != h.config.PrintfulWebhookSecret {
		log.Printf("Invalid Printful webhook token received")
		respondError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error reading request")
		return
	}

	// Parse webhook payload
	var payload PrintfulWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing Printful webhook: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	// Log event
	h.logPrintfulWebhookEvent(payload, body)

	// Process event
	switch payload.Type {
	case "order_updated":
		h.handlePrintfulOrderUpdated(payload)

	case "shipment_created":
		h.handlePrintfulShipmentCreated(payload)

	case "order_failed":
		h.handlePrintfulOrderFailed(payload)

	default:
		log.Printf("Unhandled Printful event: %s", payload.Type)
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// handlePrintfulOrderUpdated processes order status updates
func (h *Handler) handlePrintfulOrderUpdated(payload PrintfulWebhookPayload) {
	log.Printf("Printful order %d updated", payload.Order.ID)

	// Find order by Printful ID
	var orderID string
	err := h.db.QueryRow(`
		SELECT id FROM orders WHERE printful_order_id = ?
	`, payload.Order.ID).Scan(&orderID)

	if err != nil {
		log.Printf("Order not found for Printful ID %d: %v", payload.Order.ID, err)
		return
	}

	// Update order status based on Printful status
	// You may need to map Printful statuses to your statuses
	log.Printf("Order %s status updated from Printful", orderID)
}

// handlePrintfulShipmentCreated processes shipment creation
func (h *Handler) handlePrintfulShipmentCreated(payload PrintfulWebhookPayload) {
	log.Printf("Printful shipment created for order %d", payload.Order.ID)

	// Find order by Printful ID
	var orderID string
	err := h.db.QueryRow(`
		SELECT id FROM orders WHERE printful_order_id = ?
	`, payload.Order.ID).Scan(&orderID)

	if err != nil {
		log.Printf("Order not found for Printful ID %d: %v", payload.Order.ID, err)
		return
	}

	// Update tracking information
	if err := h.orderService.UpdateOrderTracking(
		orderID,
		payload.Order.TrackingNumber,
		payload.Order.TrackingURL,
	); err != nil {
		log.Printf("Failed to update tracking: %v", err)
		return
	}

	log.Printf("Order %s tracking updated: %s", orderID, payload.Order.TrackingNumber)

	// Send customer email with tracking info
	go func() {
		var customerEmail string
		err := h.db.QueryRow(`SELECT customer_email FROM orders WHERE id = ?`, orderID).Scan(&customerEmail)
		if err != nil || customerEmail == "" {
			log.Printf("Could not find customer email for order %s: %v", orderID, err)
			return
		}
		if err := h.emailClient.SendShippingNotification(
			customerEmail, orderID,
			payload.Order.TrackingNumber, payload.Order.TrackingURL,
		); err != nil {
			log.Printf("Failed to send shipping notification for order %s: %v", orderID, err)
		}
	}()
}

// handlePrintfulOrderFailed processes order failures
func (h *Handler) handlePrintfulOrderFailed(payload PrintfulWebhookPayload) {
	log.Printf("Printful order %d failed", payload.Order.ID)

	// Find order
	var orderID string
	err := h.db.QueryRow(`
		SELECT id FROM orders WHERE printful_order_id = ?
	`, payload.Order.ID).Scan(&orderID)

	if err != nil {
		log.Printf("Order not found for Printful ID %d: %v", payload.Order.ID, err)
		return
	}

	// Mark order as failed
	if err := h.orderService.UpdateOrderStatus(orderID, "failed"); err != nil {
		log.Printf("Failed to update order status: %v", err)
	}

	// Alert admin via email
	go func() {
		if h.config.AdminEmail == "" {
			log.Printf("WARNING: No admin email configured, cannot send failure alert for order %s", orderID)
			return
		}

		var customerEmail string
		h.db.QueryRow(`SELECT customer_email FROM orders WHERE id = ?`, orderID).Scan(&customerEmail)

		subject := fmt.Sprintf("ALERT: Printful Order Failed - #%s", orderID)
		body := fmt.Sprintf(
			"A Printful order has failed and may require manual intervention.\n\n"+
				"Order ID: %s\n"+
				"Printful Order ID: %d\n"+
				"Customer Email: %s\n\n"+
				"Please check the Printful dashboard and contact the customer if a refund is needed.",
			orderID, payload.Order.ID, customerEmail,
		)

		if err := h.emailClient.SendRawEmail(h.config.AdminEmail, subject, body); err != nil {
			log.Printf("Failed to send admin alert for failed order %s: %v", orderID, err)
		}
	}()
}

// logPrintfulWebhookEvent saves webhook event for audit
func (h *Handler) logPrintfulWebhookEvent(payload PrintfulWebhookPayload, body []byte) {
	orderID := ""
	if payload.Order.ExternalID != "" {
		orderID = payload.Order.ExternalID
	}

	_, err := h.db.Exec(`
		INSERT INTO printful_webhook_events (id, event_type, order_id, payload, processed, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), payload.Type, orderID, string(body), true, time.Now())

	if err != nil {
		log.Printf("Failed to log Printful webhook: %v", err)
	}
}
