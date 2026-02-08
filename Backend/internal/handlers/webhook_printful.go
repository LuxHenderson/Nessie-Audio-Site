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

	// Extract carrier from Printful shipment data
	carrier := ""
	if shipmentData, ok := payload.Data["shipment"].(map[string]interface{}); ok {
		if c, ok := shipmentData["carrier"].(string); ok {
			carrier = c
		}
	}

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
			payload.Order.TrackingNumber, payload.Order.TrackingURL, carrier,
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
		htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Order Failed Alert</title>
    <style>
        body { font-family: 'Montserrat', Arial, sans-serif; background-color: #020202; color: #ffffff; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; background-color: #0f0f0f; border: 1px solid #2a2a2a; border-radius: 8px; overflow: hidden; }
        .header { background: linear-gradient(135deg, #0f0f0f 0%%, #1a1a1a 100%%); padding: 40px 20px; text-align: center; border-bottom: 2px solid rgba(255, 255, 255, 0.15); }
        .header h1 { margin: 0; font-family: 'Montserrat', sans-serif; font-size: 32px; color: #ffffff; text-transform: uppercase; letter-spacing: 2px; font-weight: 700; }
        .alert-icon { font-size: 48px; margin-bottom: 10px; }
        .content { padding: 40px 20px; }
        .order-info { background-color: #252525; border: 1px solid #333; border-radius: 6px; padding: 20px; margin: 20px 0; }
        .order-info h2 { margin: 0 0 15px 0; font-size: 16px; color: #ffffff; text-transform: uppercase; letter-spacing: 1px; font-weight: 600; }
        .order-detail { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #333; }
        .order-detail:last-child { border-bottom: none; }
        .order-detail strong { color: #fff; }
        .note { background-color: #252525; border-left: 3px solid #ff4444; padding: 15px; margin: 20px 0; font-size: 14px; color: #ccc; }
        .cta-button { display: inline-block; background-color: #00ff88; color: #0a0a0a; padding: 14px 32px; text-decoration: none; border-radius: 4px; font-weight: bold; text-transform: uppercase; letter-spacing: 1px; margin: 20px 0; font-size: 14px; }
        .footer { background-color: #0a0a0a; padding: 30px 20px; text-align: center; border-top: 1px solid #333; }
        .footer p { margin: 5px 0; font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="alert-icon">⚠</div>
            <h1>Order Failed</h1>
        </div>
        <div class="content">
            <p style="font-size: 18px; color: #fff;">A Printful order has failed and may require manual intervention.</p>

            <div class="order-info">
                <h2>Order Details</h2>
                <div class="order-detail">
                    <span>Order ID:</span>
                    <strong>#%s</strong>
                </div>
                <div class="order-detail">
                    <span>Printful Order ID:</span>
                    <strong>%d</strong>
                </div>
                <div class="order-detail">
                    <span>Customer Email:</span>
                    <strong>%s</strong>
                </div>
            </div>

            <div class="note">
                <strong>Action Required</strong><br>
                Please check the Printful dashboard and contact the customer if a refund is needed.
            </div>

            <div style="text-align: center;">
                <a href="https://www.printful.com/dashboard/orders" class="cta-button">Open Printful Dashboard</a>
            </div>
        </div>
        <div class="footer">
            <p><strong>Nessie Audio</strong> — Admin Alert</p>
            <p>This is an automated notification from your store.</p>
        </div>
    </div>
</body>
</html>
`, orderID, payload.Order.ID, customerEmail)

		if err := h.emailClient.SendHTMLEmail(h.config.AdminEmail, subject, htmlBody); err != nil {
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
