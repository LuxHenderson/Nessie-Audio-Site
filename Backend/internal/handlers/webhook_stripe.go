package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nessieaudio/ecommerce-backend/internal/logger"
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
		h.logger.Critical("Stripe webhook signature verification failed", err, map[string]interface{}{
			"endpoint": "/webhooks/stripe",
			"ip":       r.RemoteAddr,
		})
		respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// Log webhook event and check for duplicates (idempotency).
	// stripe_webhook_events.event_id has a UNIQUE constraint — if the INSERT
	// fails, this event was already processed and we should skip it.
	alreadyProcessed, err := h.logStripeWebhookEvent(event)
	if err != nil {
		log.Printf("Failed to log webhook event: %v", err)
	}
	if alreadyProcessed {
		log.Printf("Duplicate webhook event %s (type: %s), skipping", event.ID, event.Type)
		respondJSON(w, http.StatusOK, map[string]string{"status": "already_processed"})
		return
	}

	// Process based on event type
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(event)

	case "payment_intent.succeeded":
		log.Printf("PaymentIntent succeeded: %s", event.ID)

	case "payment_intent.payment_failed":
		h.handlePaymentFailed(event)

	case "payment_intent.canceled":
		h.handlePaymentCanceled(event)

	case "checkout.session.expired":
		h.handleCheckoutExpired(event)

	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// handlePaymentFailed processes payment failures and sends admin alert
func (h *Handler) handlePaymentFailed(event stripeLib.Event) {
	var paymentIntent stripeLib.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		log.Printf("Error parsing payment intent: %v", err)
		return
	}

	log.Printf("PaymentIntent failed: %s", paymentIntent.ID)

	// Extract failure details
	errorMessage := "Unknown error"
	if paymentIntent.LastPaymentError != nil {
		errorMessage = paymentIntent.LastPaymentError.Msg
	}

	// Log critical payment failure
	h.logger.CriticalWithContext(logger.ErrorContext{
		Message:  "Payment failed",
		Error:    fmt.Errorf("payment intent failed: %s", errorMessage),
		Endpoint: "/webhooks/stripe - payment_intent.payment_failed",
		Details: map[string]interface{}{
			"payment_intent_id": paymentIntent.ID,
			"amount":            float64(paymentIntent.Amount) / 100.0,
			"currency":          paymentIntent.Currency,
			"status":            paymentIntent.Status,
			"error_message":     errorMessage,
		},
	})

	// Extract customer details
	customerEmail := ""
	customerName := ""
	if paymentIntent.LatestCharge != nil && paymentIntent.LatestCharge.BillingDetails != nil {
		customerEmail = paymentIntent.LatestCharge.BillingDetails.Email
		customerName = paymentIntent.LatestCharge.BillingDetails.Name
	}

	// Build alert email
	amount := float64(paymentIntent.Amount) / 100.0
	subject := fmt.Sprintf("⚠️ Payment Failed: $%.2f", amount)

	contentHTML := fmt.Sprintf(`<p style="font-size:16px;">A payment attempt has failed on the Nessie Audio store.</p>%s%s%s`,
		email.InfoBox("Payment Details",
			email.DetailRow("Payment Intent:", paymentIntent.ID)+
				email.DetailRow("Amount:", fmt.Sprintf("$%.2f %s", amount, paymentIntent.Currency))+
				email.DetailRow("Status:", string(paymentIntent.Status))+
				email.DetailRow("Timestamp:", time.Now().Format("2006-01-02 15:04:05 MST"))),
		email.InfoBox("Customer Information",
			email.DetailRow("Email:", customerEmail)+
				email.DetailRow("Name:", customerName)),
		email.NoteBox(fmt.Sprintf("<strong>Error:</strong> %s", errorMessage), true),
	)
	htmlBody := email.EmailLayout("Payment Failed", "&#9888;&#65039;", contentHTML, true)

	// Send alert to admin
	if h.config.AdminEmail != "" {
		go func() {
			if err := h.emailClient.SendHTMLEmail(h.config.AdminEmail, subject, htmlBody); err != nil {
				log.Printf("Failed to send payment failure alert: %v", err)
			} else {
				log.Printf("Payment failure alert sent to %s", h.config.AdminEmail)
			}
		}()
	}
}

// handlePaymentCanceled processes payment intent cancellations and sends admin alert
func (h *Handler) handlePaymentCanceled(event stripeLib.Event) {
	var paymentIntent stripeLib.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		log.Printf("Error parsing payment intent: %v", err)
		return
	}

	log.Printf("PaymentIntent canceled: %s (Reason: %s)", paymentIntent.ID, paymentIntent.CancellationReason)

	// Extract customer details
	customerEmail := ""
	customerName := ""
	if paymentIntent.LatestCharge != nil && paymentIntent.LatestCharge.BillingDetails != nil {
		customerEmail = paymentIntent.LatestCharge.BillingDetails.Email
		customerName = paymentIntent.LatestCharge.BillingDetails.Name
	}

	// Build alert email
	amount := float64(paymentIntent.Amount) / 100.0
	subject := fmt.Sprintf("🚫 Payment Canceled: $%.2f", amount)

	contentHTML := fmt.Sprintf(`<p style="font-size:16px;">A payment has been canceled on the Nessie Audio store.</p>%s%s%s`,
		email.InfoBox("Payment Details",
			email.DetailRow("Payment Intent:", paymentIntent.ID)+
				email.DetailRow("Amount:", fmt.Sprintf("$%.2f %s", amount, paymentIntent.Currency))+
				email.DetailRow("Status:", string(paymentIntent.Status))+
				email.DetailRow("Cancellation Reason:", string(paymentIntent.CancellationReason))+
				email.DetailRow("Timestamp:", time.Now().Format("2006-01-02 15:04:05 MST"))),
		email.InfoBox("Customer Information",
			email.DetailRow("Email:", customerEmail)+
				email.DetailRow("Name:", customerName)),
		email.NoteBox("<strong>Action:</strong> No action required unless this is unexpected. The payment was not collected.", false),
	)
	htmlBody := email.EmailLayout("Payment Canceled", "&#128683;", contentHTML, true)

	// Send alert to admin
	if h.config.AdminEmail != "" {
		go func() {
			if err := h.emailClient.SendHTMLEmail(h.config.AdminEmail, subject, htmlBody); err != nil {
				log.Printf("Failed to send payment cancellation alert: %v", err)
			} else {
				log.Printf("Payment cancellation alert sent to %s", h.config.AdminEmail)
			}
		}()
	}
}

// handleCheckoutExpired processes expired checkout sessions and sends admin alert
func (h *Handler) handleCheckoutExpired(event stripeLib.Event) {
	var session stripeLib.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session: %v", err)
		return
	}

	log.Printf("Checkout session expired: %s", session.ID)

	// Extract customer details
	customerEmail := ""
	customerName := ""
	if session.CustomerDetails != nil {
		customerEmail = session.CustomerDetails.Email
		customerName = session.CustomerDetails.Name
	}

	// Calculate total from session
	totalAmount := float64(session.AmountTotal) / 100.0

	// Build alert email
	subject := fmt.Sprintf("⏱️ Checkout Expired: $%.2f", totalAmount)

	contentHTML := fmt.Sprintf(`<p style="font-size:16px;">A checkout session has expired without completion on the Nessie Audio store.</p>%s%s%s`,
		email.InfoBox("Session Details",
			email.DetailRow("Session ID:", session.ID)+
				email.DetailRow("Amount:", fmt.Sprintf("$%.2f %s", totalAmount, session.Currency))+
				email.DetailRow("Status:", string(session.Status))+
				email.DetailRow("Timestamp:", time.Now().Format("2006-01-02 15:04:05 MST"))),
		email.InfoBox("Customer Information",
			email.DetailRow("Email:", customerEmail)+
				email.DetailRow("Name:", customerName)),
		email.NoteBox("<strong>Possible Reasons:</strong><br>&bull; Customer abandoned cart<br>&bull; Session timeout (24 hours)<br>&bull; Customer did not complete payment", false),
	)
	htmlBody := email.EmailLayout("Checkout Expired", "&#9200;", contentHTML, true)

	// Send alert to admin
	if h.config.AdminEmail != "" {
		go func() {
			if err := h.emailClient.SendHTMLEmail(h.config.AdminEmail, subject, htmlBody); err != nil {
				log.Printf("Failed to send checkout expiration alert: %v", err)
			} else {
				log.Printf("Checkout expiration alert sent to %s", h.config.AdminEmail)
			}
		}()
	}
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
// Implements immediate retry with exponential backoff (3 attempts: 1s, 2s, 4s)
func (h *Handler) submitOrderToPrintful(orderID string) {
	const maxImmediateRetries = 3
	backoffDurations := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 1; attempt <= maxImmediateRetries; attempt++ {
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

		// Attempt to submit to Printful
		printfulOrderID, err := h.printfulClient.CreateOrder(order, items)
		if err != nil {
			log.Printf("Printful submission attempt %d/%d failed for order %s: %v", attempt, maxImmediateRetries, orderID, err)

			// Increment retry count
			if err := h.orderService.IncrementPrintfulRetryCount(orderID); err != nil {
				log.Printf("Failed to increment retry count: %v", err)
			}

			// Record failure in audit table
			if err := h.orderService.RecordPrintfulFailure(orderID, attempt, err.Error(), ""); err != nil {
				log.Printf("Failed to record Printful failure: %v", err)
			}

			// If this wasn't the last attempt, wait and retry
			if attempt < maxImmediateRetries {
				log.Printf("Retrying in %v...", backoffDurations[attempt-1])
				time.Sleep(backoffDurations[attempt-1])
				continue
			}

			// All immediate retries exhausted
			log.Printf("All immediate retries exhausted for order %s. Will retry via background job.", orderID)
			return
		}

		// Success! Confirm the order with Printful
		if err := h.printfulClient.ConfirmOrder(printfulOrderID); err != nil {
			log.Printf("Failed to confirm Printful order: %v", err)

			// Record this as a failure too
			if err := h.orderService.IncrementPrintfulRetryCount(orderID); err != nil {
				log.Printf("Failed to increment retry count: %v", err)
			}
			if err := h.orderService.RecordPrintfulFailure(orderID, attempt, "Confirm failed: "+err.Error(), ""); err != nil {
				log.Printf("Failed to record Printful failure: %v", err)
			}

			// Retry if we have attempts left
			if attempt < maxImmediateRetries {
				log.Printf("Retrying in %v...", backoffDurations[attempt-1])
				time.Sleep(backoffDurations[attempt-1])
				continue
			}

			log.Printf("All immediate retries exhausted for order %s. Will retry via background job.", orderID)
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

		log.Printf("✅ Order %s submitted to Printful successfully (ID: %d) on attempt %d", orderID, printfulOrderID, attempt)
		return
	}
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

	// Parse cart item metadata (product/variant IDs stored during checkout)
	var cartItems []stripe.CartItemMeta
	if cartJSON, ok := session.Metadata["cart_items"]; ok && cartJSON != "" {
		if err := json.Unmarshal([]byte(cartJSON), &cartItems); err != nil {
			log.Printf("WARNING: Failed to parse cart_items metadata: %v", err)
		}
	}

	// Create order items — use cart metadata for proper product/variant IDs
	if len(cartItems) > 0 {
		// We have cart metadata — use it to insert items with correct IDs
		for _, ci := range cartItems {
			// Look up product name and variant name from the database
			var productName, variantName string
			var price float64
			err = h.db.QueryRow("SELECT name FROM products WHERE id = ?", ci.ProductID).Scan(&productName)
			if err != nil {
				log.Printf("WARNING: Could not find product %s: %v", ci.ProductID, err)
				productName = "Unknown Product"
			}
			err = h.db.QueryRow("SELECT name, price FROM variants WHERE id = ?", ci.VariantID).Scan(&variantName, &price)
			if err != nil {
				log.Printf("WARNING: Could not find variant %s: %v", ci.VariantID, err)
				variantName = ""
			}

			itemID := uuid.New().String()
			totalPrice := price * float64(ci.Quantity)

			_, err = h.db.Exec(`
				INSERT INTO order_items (
					id, order_id, product_id, variant_id,
					product_name, variant_name,
					quantity, unit_price, total_price, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, itemID, orderID, ci.ProductID, ci.VariantID,
				productName, variantName,
				ci.Quantity, price, totalPrice, time.Now())

			if err != nil {
				return nil, err
			}
		}
	} else if session.LineItems != nil && session.LineItems.Data != nil {
		// Fallback: no cart metadata (e.g. legacy sessions) — use Stripe line items
		log.Printf("WARNING: No cart_items metadata for session %s, using Stripe line items (variant IDs will be missing)", session.ID)
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

// logStripeWebhookEvent saves webhook event for audit and returns whether
// the event was already processed (duplicate). The event_id column has a
// UNIQUE constraint, so a duplicate INSERT will fail.
func (h *Handler) logStripeWebhookEvent(event stripeLib.Event) (alreadyProcessed bool, err error) {
	payload, _ := json.Marshal(event)

	_, err = h.db.Exec(`
		INSERT INTO stripe_webhook_events (id, event_type, event_id, payload, processed, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), event.Type, event.ID, string(payload), true, time.Now())

	if err != nil {
		// UNIQUE constraint violation means this event was already processed
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			return true, nil
		}
		return false, err
	}

	return false, nil
}
