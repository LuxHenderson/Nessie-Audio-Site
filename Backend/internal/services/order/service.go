package order

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/models"
)

// Service handles order business logic
type Service struct {
	db *sql.DB
}

// NewService creates a new order service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CreateOrder creates a new pending order
func (s *Service) CreateOrder(order *models.Order, items []models.OrderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.Exec(`
		INSERT INTO orders (
			id, customer_id, status, total_amount, currency,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, order.ID, order.CustomerID, order.Status, order.TotalAmount, order.Currency,
		order.ShippingName, order.ShippingAddress1, order.ShippingAddress2,
		order.ShippingCity, order.ShippingState, order.ShippingZip, order.ShippingCountry,
		order.CreatedAt, order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// Insert order items
	for _, item := range items {
		_, err = tx.Exec(`
			INSERT INTO order_items (
				id, order_id, product_id, variant_id, quantity,
				unit_price, total_price, product_name, variant_name, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, item.ID, item.OrderID, item.ProductID, item.VariantID, item.Quantity,
			item.UnitPrice, item.TotalPrice, item.ProductName, item.VariantName, item.CreatedAt)

		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(id string) (*models.Order, error) {
	order := &models.Order{}
	err := s.db.QueryRow(`
		SELECT id, customer_id, status, total_amount, currency,
			stripe_session_id, stripe_payment_intent_id, printful_order_id,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			tracking_number, tracking_url, created_at, updated_at
		FROM orders WHERE id = ?
	`, id).Scan(
		&order.ID, &order.CustomerID, &order.Status, &order.TotalAmount, &order.Currency,
		&order.StripeSessionID, &order.StripePaymentIntentID, &order.PrintfulOrderID,
		&order.ShippingName, &order.ShippingAddress1, &order.ShippingAddress2,
		&order.ShippingCity, &order.ShippingState, &order.ShippingZip, &order.ShippingCountry,
		&order.TrackingNumber, &order.TrackingURL, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	return order, nil
}

// GetOrderItems retrieves all items for an order
func (s *Service) GetOrderItems(orderID string) ([]models.OrderItem, error) {
	rows, err := s.db.Query(`
		SELECT id, order_id, product_id, variant_id, quantity,
			unit_price, total_price, product_name, variant_name, created_at
		FROM order_items WHERE order_id = ?
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.VariantID, &item.Quantity,
			&item.UnitPrice, &item.TotalPrice, &item.ProductName, &item.VariantName, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// UpdateOrderStatus updates the order status
func (s *Service) UpdateOrderStatus(orderID, status string) error {
	_, err := s.db.Exec(`
		UPDATE orders SET status = ?, updated_at = ? WHERE id = ?
	`, status, time.Now(), orderID)

	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

// UpdateOrderWithStripeSession updates order after payment
func (s *Service) UpdateOrderWithStripeSession(order *models.Order) error {
	_, err := s.db.Exec(`
		UPDATE orders SET
			stripe_session_id = ?,
			stripe_payment_intent_id = ?,
			status = ?,
			shipping_name = ?,
			shipping_address1 = ?,
			shipping_address2 = ?,
			shipping_city = ?,
			shipping_state = ?,
			shipping_zip = ?,
			shipping_country = ?,
			updated_at = ?
		WHERE id = ?
	`, order.StripeSessionID, order.StripePaymentIntentID, order.Status,
		order.ShippingName, order.ShippingAddress1, order.ShippingAddress2,
		order.ShippingCity, order.ShippingState, order.ShippingZip, order.ShippingCountry,
		time.Now(), order.ID)

	if err != nil {
		return fmt.Errorf("update order with stripe: %w", err)
	}

	return nil
}

// UpdateOrderWithPrintful updates order with Printful details
func (s *Service) UpdateOrderWithPrintful(orderID string, printfulOrderID int64) error {
	_, err := s.db.Exec(`
		UPDATE orders SET printful_order_id = ?, updated_at = ? WHERE id = ?
	`, printfulOrderID, time.Now(), orderID)

	if err != nil {
		return fmt.Errorf("update order with printful: %w", err)
	}

	return nil
}

// UpdateOrderTracking updates tracking information
func (s *Service) UpdateOrderTracking(orderID, trackingNumber, trackingURL string) error {
	_, err := s.db.Exec(`
		UPDATE orders SET
			tracking_number = ?,
			tracking_url = ?,
			status = ?,
			updated_at = ?
		WHERE id = ?
	`, trackingNumber, trackingURL, models.OrderStatusShipped, time.Now(), orderID)

	if err != nil {
		return fmt.Errorf("update order tracking: %w", err)
	}

	return nil
}
