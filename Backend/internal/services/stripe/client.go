package stripe

import (
	"fmt"

	"github.com/nessieaudio/ecommerce-backend/internal/models"
	stripe_lib "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

// Client wraps Stripe operations
type Client struct {
	secretKey      string
	publishableKey string
	successURL     string
	cancelURL      string
}

// NewClient creates a new Stripe client
func NewClient(secretKey, publishableKey, successURL, cancelURL string) *Client {
	stripe_lib.Key = secretKey
	return &Client{
		secretKey:      secretKey,
		publishableKey: publishableKey,
		successURL:     successURL,
		cancelURL:      cancelURL,
	}
}

// CheckoutSessionRequest represents data needed to create a checkout
type CheckoutSessionRequest struct {
	OrderID       string
	CustomerEmail string
	LineItems     []CheckoutLineItem
	ShippingAddress *ShippingAddress
}

// CheckoutLineItem represents a product in the checkout
type CheckoutLineItem struct {
	ProductName string
	VariantName string
	Quantity    int64
	UnitPrice   int64 // In cents
}

// ShippingAddress holds customer shipping details
type ShippingAddress struct {
	Name     string
	Address1 string
	Address2 string
	City     string
	State    string
	Zip      string
	Country  string
}

// CreateCheckoutSession creates a Stripe checkout session
// Returns the session ID which frontend uses to redirect to Stripe
func (c *Client) CreateCheckoutSession(req *CheckoutSessionRequest) (string, error) {
	// Build line items for Stripe
	var lineItems []*stripe_lib.CheckoutSessionLineItemParams
	for _, item := range req.LineItems {
		lineItems = append(lineItems, &stripe_lib.CheckoutSessionLineItemParams{
			PriceData: &stripe_lib.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe_lib.String("usd"),
				ProductData: &stripe_lib.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe_lib.String(fmt.Sprintf("%s - %s", item.ProductName, item.VariantName)),
				},
				UnitAmount: stripe_lib.Int64(item.UnitPrice), // Price in cents
			},
			Quantity: stripe_lib.Int64(item.Quantity),
		})
	}

	// Create Stripe checkout session
	params := &stripe_lib.CheckoutSessionParams{
		Mode:       stripe_lib.String(string(stripe_lib.CheckoutSessionModePayment)),
		SuccessURL: stripe_lib.String(c.successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe_lib.String(c.cancelURL),
		LineItems:  lineItems,
		PaymentMethodTypes: stripe_lib.StringSlice([]string{
			"card",
		}),
		Metadata: map[string]string{
			"order_id": req.OrderID, // Link back to your order
		},
		ShippingAddressCollection: &stripe_lib.CheckoutSessionShippingAddressCollectionParams{
			AllowedCountries: []*string{
				stripe_lib.String("US"),
				stripe_lib.String("CA"),
				// TODO: Add more countries as needed
			},
		},
	}

	// Only set customer email if provided, otherwise Stripe will collect it
	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe_lib.String(req.CustomerEmail)
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	return sess.ID, nil
}

// GetSession retrieves a checkout session by ID
func (c *Client) GetSession(sessionID string) (*stripe_lib.CheckoutSession, error) {
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

// ExtractShippingFromSession extracts shipping details from completed session
func ExtractShippingFromSession(sess *stripe_lib.CheckoutSession) *ShippingAddress {
	if sess.ShippingDetails == nil || sess.ShippingDetails.Address == nil {
		return nil
	}

	addr := sess.ShippingDetails.Address
	return &ShippingAddress{
		Name:     sess.ShippingDetails.Name,
		Address1: addr.Line1,
		Address2: addr.Line2,
		City:     addr.City,
		State:    addr.State,
		Zip:      addr.PostalCode,
		Country:  addr.Country,
	}
}

// UpdateOrderFromSession updates an order with Stripe session data
func UpdateOrderFromSession(order *models.Order, sess *stripe_lib.CheckoutSession) {
	order.StripeSessionID = sess.ID
	order.StripePaymentIntentID = sess.PaymentIntent.ID
	order.Status = models.OrderStatusPaid

	// Extract shipping details
	if shipping := ExtractShippingFromSession(sess); shipping != nil {
		order.ShippingName = shipping.Name
		order.ShippingAddress1 = shipping.Address1
		order.ShippingAddress2 = shipping.Address2
		order.ShippingCity = shipping.City
		order.ShippingState = shipping.State
		order.ShippingZip = shipping.Zip
		order.ShippingCountry = shipping.Country
	}
}
