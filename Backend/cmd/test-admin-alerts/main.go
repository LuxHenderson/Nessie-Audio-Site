package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	stripeLib "github.com/stripe/stripe-go/v76"
)

func main() {
	log.Println("🧪 Testing Admin Alert System")
	log.Println("==============================")

	// Load config to get webhook secret
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	baseURL := "http://localhost:8080"

	log.Printf("\n📧 Admin alerts will be sent to: %s\n", cfg.AdminEmail)

	// Test 1: Payment Failed
	log.Println("\n1️⃣ Testing Payment Failure Alert...")
	testPaymentFailed(baseURL, cfg.StripeWebhookSecret)
	time.Sleep(2 * time.Second)

	// Test 2: Payment Canceled
	log.Println("\n2️⃣ Testing Payment Cancellation Alert...")
	testPaymentCanceled(baseURL, cfg.StripeWebhookSecret)
	time.Sleep(2 * time.Second)

	// Test 3: Checkout Expired
	log.Println("\n3️⃣ Testing Checkout Expiration Alert...")
	testCheckoutExpired(baseURL, cfg.StripeWebhookSecret)

	log.Println("\n✅ All test events sent!")
	log.Println("\n📬 Check your email inbox (nessieaudio@gmail.com) for 3 admin alerts:")
	log.Println("   - ⚠️ Payment Failed")
	log.Println("   - 🚫 Payment Canceled")
	log.Println("   - ⏱️ Checkout Expired")
}

func testPaymentFailed(baseURL, webhookSecret string) {
	// Create a mock payment_intent.payment_failed event
	event := stripeLib.Event{
		ID:      "evt_test_payment_failed_" + fmt.Sprint(time.Now().Unix()),
		Type:    "payment_intent.payment_failed",
		Created: time.Now().Unix(),
	}

	// Create mock PaymentIntent data
	paymentIntent := map[string]interface{}{
		"id":       "pi_test_failed_12345",
		"object":   "payment_intent",
		"amount":   4999, // $49.99
		"currency": "usd",
		"status":   "requires_payment_method",
		"last_payment_error": map[string]interface{}{
			"message": "Your card was declined. Please try another payment method.",
			"code":    "card_declined",
			"type":    "card_error",
		},
		"latest_charge": map[string]interface{}{
			"id": "ch_test_12345",
			"billing_details": map[string]interface{}{
				"email": "test.customer@example.com",
				"name":  "Test Customer",
			},
		},
	}

	rawData, _ := json.Marshal(paymentIntent)
	event.Data = &stripeLib.EventData{
		Raw: rawData,
	}

	sendWebhook(baseURL, event, webhookSecret)
}

func testPaymentCanceled(baseURL, webhookSecret string) {
	// Create a mock payment_intent.canceled event
	event := stripeLib.Event{
		ID:      "evt_test_payment_canceled_" + fmt.Sprint(time.Now().Unix()),
		Type:    "payment_intent.canceled",
		Created: time.Now().Unix(),
	}

	// Create mock PaymentIntent data
	paymentIntent := map[string]interface{}{
		"id":                  "pi_test_canceled_12345",
		"object":              "payment_intent",
		"amount":              2999, // $29.99
		"currency":            "usd",
		"status":              "canceled",
		"cancellation_reason": "abandoned",
		"latest_charge": map[string]interface{}{
			"id": "ch_test_67890",
			"billing_details": map[string]interface{}{
				"email": "abandoned.cart@example.com",
				"name":  "Abandoned User",
			},
		},
	}

	rawData, _ := json.Marshal(paymentIntent)
	event.Data = &stripeLib.EventData{
		Raw: rawData,
	}

	sendWebhook(baseURL, event, webhookSecret)
}

func testCheckoutExpired(baseURL, webhookSecret string) {
	// Create a mock checkout.session.expired event
	event := stripeLib.Event{
		ID:      "evt_test_checkout_expired_" + fmt.Sprint(time.Now().Unix()),
		Type:    "checkout.session.expired",
		Created: time.Now().Unix(),
	}

	// Create mock CheckoutSession data
	session := map[string]interface{}{
		"id":           "cs_test_expired_12345",
		"object":       "checkout.session",
		"amount_total": 7998, // $79.98
		"currency":     "usd",
		"status":       "expired",
		"customer_details": map[string]interface{}{
			"email": "timeout.customer@example.com",
			"name":  "Timeout Customer",
		},
	}

	rawData, _ := json.Marshal(session)
	event.Data = &stripeLib.EventData{
		Raw: rawData,
	}

	sendWebhook(baseURL, event, webhookSecret)
}

func sendWebhook(baseURL string, event stripeLib.Event, webhookSecret string) {
	// Serialize event
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	// Generate Stripe signature
	timestamp := time.Now().Unix()
	signature := generateStripeSignature(payload, timestamp, webhookSecret)

	// Send webhook request
	url := baseURL + "/webhooks/stripe"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", signature)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ Webhook sent successfully (Event: %s)", event.Type)
	} else {
		log.Printf("❌ Webhook failed: %s - %s", resp.Status, string(body))
	}
}

func generateStripeSignature(payload []byte, timestamp int64, secret string) string {
	// Generate proper Stripe signature using HMAC-SHA256
	// Format: t=timestamp,v1=signature
	signedPayload := fmt.Sprintf("%d.%s", timestamp, payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("t=%d,v1=%s", timestamp, signature)
}
