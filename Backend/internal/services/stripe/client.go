package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/circuitbreaker"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
	stripe_lib "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/client"
)

// CartItemMeta is the compact representation stored in Stripe session metadata
// so we can recover product/variant IDs when the webhook fires.
type CartItemMeta struct {
	ProductID string `json:"p"`
	VariantID string `json:"v"`
	Quantity  int64  `json:"q"`
}

// Client wraps Stripe operations
type Client struct {
	secretKey      string
	publishableKey string
	successURL     string
	cancelURL      string
	circuitBreaker *circuitbreaker.CircuitBreaker
	api            *client.API
}

// NewClient creates a new Stripe client
func NewClient(secretKey, publishableKey, successURL, cancelURL string) *Client {
	stripe_lib.Key = secretKey

	// Configure HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create Stripe API client with custom HTTP client
	stripeAPI := &client.API{}
	stripeAPI.Init(secretKey, &stripe_lib.Backends{
		API: stripe_lib.GetBackendWithConfig(
			stripe_lib.APIBackend,
			&stripe_lib.BackendConfig{
				HTTPClient: httpClient,
			},
		),
	})

	return &Client{
		secretKey:      secretKey,
		publishableKey: publishableKey,
		successURL:     successURL,
		cancelURL:      cancelURL,
		api:            stripeAPI,
		circuitBreaker: circuitbreaker.New(circuitbreaker.Config{
			Name:            "stripe",
			MaxFailures:     5,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 1,
		}),
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
	ProductID   string // Database product UUID (for cart checkouts)
	VariantID   string // Database variant UUID (for cart checkouts)
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

// CreateCheckoutSession creates a Stripe checkout session with circuit breaker protection
// Returns the session ID which frontend uses to redirect to Stripe
func (c *Client) CreateCheckoutSession(req *CheckoutSessionRequest) (string, error) {
	var sessionID string

	err := c.circuitBreaker.Execute(func() error {
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

		// Build metadata — always include order_id
		metadata := map[string]string{
			"order_id": req.OrderID,
		}

		// For cart checkouts, store product/variant IDs so the webhook
		// can create order_items with the correct foreign keys.
		var cartMetas []CartItemMeta
		for _, item := range req.LineItems {
			if item.ProductID != "" && item.VariantID != "" {
				cartMetas = append(cartMetas, CartItemMeta{
					ProductID: item.ProductID,
					VariantID: item.VariantID,
					Quantity:  item.Quantity,
				})
			}
		}
		if len(cartMetas) > 0 {
			cartJSON, err := json.Marshal(cartMetas)
			if err == nil {
				metadata["cart_items"] = string(cartJSON)
			}
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
			Metadata: metadata,
			ShippingAddressCollection: &stripe_lib.CheckoutSessionShippingAddressCollectionParams{
				AllowedCountries: stripe_lib.StringSlice(worldwideCountries),
			},
		}

		// Set customer email if provided, otherwise tell Stripe to collect it
		if req.CustomerEmail != "" {
			params.CustomerEmail = stripe_lib.String(req.CustomerEmail)
		} else {
			// Force Stripe to collect email during checkout
			params.CustomerCreation = stripe_lib.String("always")
		}

		sess, err := session.New(params)
		if err != nil {
			return fmt.Errorf("create checkout session: %w", err)
		}

		sessionID = sess.ID
		return nil
	})

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetSession retrieves a checkout session by ID with line items expanded with circuit breaker protection
func (c *Client) GetSession(sessionID string) (*stripe_lib.CheckoutSession, error) {
	var sess *stripe_lib.CheckoutSession

	err := c.circuitBreaker.Execute(func() error {
		params := &stripe_lib.CheckoutSessionParams{}
		params.AddExpand("line_items")

		s, err := session.Get(sessionID, params)
		if err != nil {
			return fmt.Errorf("get session: %w", err)
		}
		sess = s
		return nil
	})

	if err != nil {
		return nil, err
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

// worldwideCountries lists ISO 3166-1 alpha-2 country codes accepted by
// Stripe shipping address collection (stripe-go v76).
// Excluded per Stripe docs: AS, CX, CC, CU, HM, IR, KP, MH, FM, NF, MP, PW, SD, SY, UM, VI.
// Also excluded: AC, XK (not valid ISO 3166-1 alpha-2 codes).
var worldwideCountries = []string{
	"AD", "AE", "AF", "AG", "AI", "AL", "AM", "AO", "AR", "AT", "AU",
	"AW", "AZ", "BA", "BB", "BD", "BE", "BF", "BG", "BH", "BI", "BJ", "BL",
	"BM", "BN", "BO", "BQ", "BR", "BS", "BT", "BW", "BY", "BZ", "CA", "CD",
	"CF", "CG", "CH", "CI", "CK", "CL", "CM", "CN", "CO", "CR", "CV", "CW",
	"CY", "CZ", "DE", "DJ", "DK", "DM", "DO", "DZ", "EC", "EE", "EG", "ER",
	"ES", "ET", "FI", "FJ", "FK", "FO", "FR", "GA", "GB", "GD", "GE",
	"GF", "GG", "GH", "GI", "GL", "GM", "GN", "GP", "GQ", "GR", "GT", "GU",
	"GW", "GY", "HK", "HN", "HR", "HT", "HU", "ID", "IE", "IL", "IM", "IN",
	"IO", "IQ", "IS", "IT", "JE", "JM", "JO", "JP", "KE", "KG", "KH", "KI",
	"KM", "KN", "KR", "KW", "KY", "KZ", "LA", "LB", "LC", "LI", "LK", "LR",
	"LS", "LT", "LU", "LV", "LY", "MA", "MC", "MD", "ME", "MF", "MG",
	"MK", "ML", "MM", "MN", "MO", "MQ", "MR", "MS", "MT", "MU", "MV",
	"MW", "MX", "MY", "MZ", "NA", "NC", "NE", "NG", "NI", "NL", "NO",
	"NP", "NR", "NU", "NZ", "OM", "PA", "PE", "PF", "PG", "PH", "PK", "PL",
	"PM", "PN", "PR", "PS", "PT", "PY", "QA", "RE", "RO", "RS", "RU",
	"RW", "SA", "SB", "SC", "SE", "SG", "SH", "SI", "SJ", "SK", "SL",
	"SM", "SN", "SO", "SR", "SS", "ST", "SV", "SX", "SZ", "TC", "TD", "TF",
	"TG", "TH", "TJ", "TK", "TL", "TM", "TN", "TO", "TR", "TT", "TV", "TW",
	"TZ", "UA", "UG", "US", "UY", "UZ", "VA", "VC", "VE", "VG", "VN",
	"VU", "WF", "WS", "YE", "YT", "ZA", "ZM", "ZW",
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
