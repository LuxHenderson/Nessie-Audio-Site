package printful

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/models"
)

// Client wraps the Printful API
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Printful API client
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PrintfulProduct represents a product from Printful API
type PrintfulProduct struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image"`
	Variants    []PrintfulVariant `json:"variants"`
}

// PrintfulVariant represents a variant from Printful API
type PrintfulVariant struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Size  string `json:"size"`
	Color string `json:"color"`
	Price string `json:"price"`
	Available bool `json:"in_stock"`
}

// PrintfulOrderRequest represents an order submission to Printful
type PrintfulOrderRequest struct {
	Recipient PrintfulRecipient `json:"recipient"`
	Items     []PrintfulOrderItem `json:"items"`
}

// PrintfulRecipient represents shipping details
type PrintfulRecipient struct {
	Name     string `json:"name"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2,omitempty"`
	City     string `json:"city"`
	StateCode string `json:"state_code,omitempty"`
	CountryCode string `json:"country_code"`
	Zip      string `json:"zip"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// PrintfulOrderItem represents an item in a Printful order
type PrintfulOrderItem struct {
	VariantID int64 `json:"variant_id"` // TODO: Get from your Product.PrintfulVariantID
	Quantity  int   `json:"quantity"`
}

// PrintfulOrderResponse represents Printful's order creation response
type PrintfulOrderResponse struct {
	Code   int `json:"code"`
	Result struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	} `json:"result"`
}

// GetProducts fetches products from Printful
// TODO: In production, cache this data in your database
func (c *Client) GetProducts() ([]PrintfulProduct, error) {
	// TODO: Implement actual Printful API call
	// Endpoint: GET /store/products
	// Documentation: https://developers.printful.com/docs/#tag/Store-Products-API
	
	resp, err := c.makeRequest("GET", "/store/products", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code   int `json:"code"`
		Result []PrintfulProduct `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode products: %w", err)
	}

	return result.Result, nil
}

// GetProduct fetches a single product with variants
// TODO: Store in database and refresh periodically
func (c *Client) GetProduct(productID int64) (*PrintfulProduct, error) {
	// TODO: Implement actual call
	// Endpoint: GET /store/products/{id}
	
	endpoint := fmt.Sprintf("/store/products/%d", productID)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code   int `json:"code"`
		Result PrintfulProduct `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode product: %w", err)
	}

	return &result.Result, nil
}

// CreateOrder submits an order to Printful for fulfillment
// This should ONLY be called after payment is confirmed
func (c *Client) CreateOrder(order *models.Order, items []models.OrderItem) (int64, error) {
	// Build Printful order request
	req := PrintfulOrderRequest{
		Recipient: PrintfulRecipient{
			Name:        order.ShippingName,
			Address1:    order.ShippingAddress1,
			Address2:    order.ShippingAddress2,
			City:        order.ShippingCity,
			StateCode:   order.ShippingState,
			CountryCode: order.ShippingCountry,
			Zip:         order.ShippingZip,
		},
		Items: make([]PrintfulOrderItem, len(items)),
	}

	// TODO: Map your OrderItems to Printful items
	// You need to store PrintfulVariantID in your Variant model
	for i, item := range items {
		req.Items[i] = PrintfulOrderItem{
			// TODO: Get PrintfulVariantID from your database
			VariantID: 0, // PLACEHOLDER - must be actual Printful variant ID
			Quantity:  item.Quantity,
		}
	}

	// Submit to Printful
	resp, err := c.makeRequest("POST", "/orders", req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result PrintfulOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode order response: %w", err)
	}

	if result.Code != 200 {
		return 0, fmt.Errorf("printful returned code %d", result.Code)
	}

	return result.Result.ID, nil
}

// ConfirmOrder confirms a draft order for fulfillment
func (c *Client) ConfirmOrder(printfulOrderID int64) error {
	// TODO: Implement
	// Endpoint: POST /orders/@{id}/confirm
	endpoint := fmt.Sprintf("/orders/@%d/confirm", printfulOrderID)
	resp, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// makeRequest makes an authenticated request to Printful API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Printful authentication: Bearer token
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("printful API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
