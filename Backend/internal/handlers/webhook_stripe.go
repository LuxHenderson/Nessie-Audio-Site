package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
	"github.com/nessieaudio/ecommerce-backend/internal/services/stripe"
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
	event, err := webhook.ConstructEventWithOptions(
		payload,
		r.Header.Get("Stripe-Signature"),
		h.config.StripeWebhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)

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

	// Retrieve full session with shipping details
	fullSession, err := h.stripeClient.GetSession(session.ID)
	if err != nil {
		log.Printf("Failed to get full session: %v", err)
		return
	}

	// Get order ID from metadata (may be empty for cart-based checkouts)
	orderID, ok := session.Metadata["order_id"]

	var order *models.Order
	if !ok || orderID == "" {
		// No pre-existing order - this is a cart-based checkout
		// Create order from session data
		order, err = h.createOrderFromSession(fullSession)
		if err != nil {
			log.Printf("Failed to create order from session: %v", err)
			return
		}
		orderID = order.ID
		log.Printf("Created new order %s from cart checkout", orderID)
	} else {
		// Get existing order
		order, err = h.orderService.GetOrder(orderID)
		if err != nil {
			log.Printf("Failed to get order %s: %v", orderID, err)
			return
		}
	}

	// Update order with Stripe and shipping details
	stripe.UpdateOrderFromSession(order, fullSession)

	if err := h.orderService.UpdateOrderWithStripeSession(order); err != nil {
		log.Printf("Failed to update order: %v", err)
		return
	}

	log.Printf("Order %s marked as paid", orderID)

	// ====== SEND ORDER CONFIRMATION EMAIL ======
	go h.sendOrderConfirmationEmail(orderID, fullSession)

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

// sendOrderConfirmationEmail sends order confirmation email to customer
// Runs asynchronously to not block webhook response
func (h *Handler) sendOrderConfirmationEmail(orderID string, session *stripeLib.CheckoutSession) {
	// Get order details
	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		log.Printf("Failed to get order for email: %v", err)
		return
	}

	// Get order items
	items, err := h.orderService.GetOrderItems(orderID)
	if err != nil {
		log.Printf("Failed to get order items for email: %v", err)
		return
	}

	// Extract customer info from session
	customerName := ""
	customerEmail := ""
	if session.CustomerDetails != nil {
		if session.CustomerDetails.Name != "" {
			customerName = session.CustomerDetails.Name
		}
		if session.CustomerDetails.Email != "" {
			customerEmail = session.CustomerDetails.Email
		}
	}

	// Default customer name if not provided
	if customerName == "" {
		customerName = "Valued Customer"
	}

	// Default email if not in session (shouldn't happen, but safe fallback)
	if customerEmail == "" {
		log.Printf("WARNING: No customer email found for order %s", orderID)
		return // Can't send email without an address
	}

	// Extract shipping info from order
	shippingAddress := order.ShippingAddress1
	if order.ShippingAddress2 != "" {
		shippingAddress += ", " + order.ShippingAddress2
	}

	shippingInfo := email.ShippingInfo{
		Name:    order.ShippingName,
		Address: shippingAddress,
		City:    order.ShippingCity,
		State:   order.ShippingState,
		Zip:     order.ShippingZip,
		Country: order.ShippingCountry,
	}

	// If we have session shipping details, use those (more complete)
	if session.ShippingDetails != nil {
		if session.ShippingDetails.Name != "" {
			shippingInfo.Name = session.ShippingDetails.Name
		}
		if session.ShippingDetails.Address != nil {
			if session.ShippingDetails.Address.Line1 != "" {
				shippingInfo.Address = session.ShippingDetails.Address.Line1
				if session.ShippingDetails.Address.Line2 != "" {
					shippingInfo.Address += ", " + session.ShippingDetails.Address.Line2
				}
			}
			if session.ShippingDetails.Address.City != "" {
				shippingInfo.City = session.ShippingDetails.Address.City
			}
			if session.ShippingDetails.Address.State != "" {
				shippingInfo.State = session.ShippingDetails.Address.State
			}
			if session.ShippingDetails.Address.PostalCode != "" {
				shippingInfo.Zip = session.ShippingDetails.Address.PostalCode
			}
			if session.ShippingDetails.Address.Country != "" {
				shippingInfo.Country = session.ShippingDetails.Address.Country
			}
		}
	}

	// Prepare email data
	emailData := email.OrderConfirmationData{
		OrderID:       orderID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Items:         items,
		Total:         order.TotalAmount,
		ShippingInfo:  shippingInfo,
	}

	// Send email
	if err := h.emailClient.SendOrderConfirmation(emailData); err != nil {
		log.Printf("Failed to send order confirmation email: %v", err)
		return
	}

	log.Printf("Order confirmation email sent for order %s", orderID)
}

// createOrderFromSession creates a new order from a Stripe session (for cart-based checkouts)
func (h *Handler) createOrderFromSession(session *stripeLib.CheckoutSession) (*models.Order, error) {
	// Create customer ID (we'll use email as a simple identifier for now)
	customerID := uuid.New().String()

	// Extract customer email from session
	customerEmail := ""
	if session.CustomerDetails != nil {
		customerEmail = session.CustomerDetails.Email
	}

	// Calculate total from session
	totalAmount := float64(session.AmountTotal) / 100.0 // Convert from cents

	// Extract shipping details
	shippingName := ""
	shippingAddress1 := ""
	shippingAddress2 := ""
	shippingCity := ""
	shippingState := ""
	shippingZip := ""
	shippingCountry := ""

	if session.ShippingDetails != nil {
		shippingName = session.ShippingDetails.Name
		if session.ShippingDetails.Address != nil {
			shippingAddress1 = session.ShippingDetails.Address.Line1
			shippingAddress2 = session.ShippingDetails.Address.Line2
			shippingCity = session.ShippingDetails.Address.City
			shippingState = session.ShippingDetails.Address.State
			shippingZip = session.ShippingDetails.Address.PostalCode
			shippingCountry = session.ShippingDetails.Address.Country
		}
	}

	// Extract PaymentIntent ID
	paymentIntentID := ""
	if session.PaymentIntent != nil {
		paymentIntentID = session.PaymentIntent.ID
	}

	// Create order
	orderID := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO orders (
			id, customer_id, customer_email, status, total_amount, currency,
			stripe_session_id, stripe_payment_intent_id,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, orderID, customerID, customerEmail, models.OrderStatusPending, totalAmount, "usd",
		session.ID, paymentIntentID,
		shippingName, shippingAddress1, shippingAddress2,
		shippingCity, shippingState, shippingZip, shippingCountry,
		time.Now(), time.Now())

	if err != nil {
		return nil, err
	}

	// Create order items from line items
	if session.LineItems != nil && session.LineItems.Data != nil {
		for _, lineItem := range session.LineItems.Data {
			if lineItem.Price == nil {
				log.Printf("WARNING: Line item has nil Price, skipping")
				continue
			}

			itemID := uuid.New().String()
			unitPrice := float64(lineItem.Price.UnitAmount) / 100.0
			totalPrice := unitPrice * float64(lineItem.Quantity)

			_, err = h.db.Exec(`
				INSERT INTO order_items (
					id, order_id, product_id, variant_id,
					product_name, variant_name,
					quantity, unit_price, total_price, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, itemID, orderID, "", "",
				lineItem.Description, "",
				lineItem.Quantity, unitPrice, totalPrice, time.Now())

			if err != nil {
				return nil, err
			}
		}
	} else {
		log.Printf("WARNING: No line items in session %s", session.ID)
	}

	// Get the created order
	return h.orderService.GetOrder(orderID)
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
