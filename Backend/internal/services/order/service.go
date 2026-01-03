package order

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/inventory"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
)

// Service handles order business logic
type Service struct {
	db                *sql.DB
	inventoryService  *inventory.Service
}

// NewService creates a new order service
func NewService(db *sql.DB) *Service {
	return &Service{
		db:               db,
		inventoryService: inventory.NewService(db),
	}
}

// CreateOrder creates a new pending order
func (s *Service) CreateOrder(order *models.Order, items []models.OrderItem) error {
	// First, check stock availability for all items
	for _, item := range items {
		stockCheck, err := s.inventoryService.CheckStock(item.VariantID, item.Quantity)
		if err != nil {
			return fmt.Errorf("check stock for variant %s: %w", item.VariantID, err)
		}

		if !stockCheck.Available {
			if stockCheck.StockQuantity != nil {
				return fmt.Errorf("insufficient stock for %s: requested %d, available %d",
					item.VariantName, item.Quantity, *stockCheck.StockQuantity)
			}
			return fmt.Errorf("item %s is out of stock", item.VariantName)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.Exec(`
		INSERT INTO orders (
			id, customer_id, customer_email, status, total_amount, currency,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, order.ID, order.CustomerID, order.CustomerEmail, order.Status, order.TotalAmount, order.Currency,
		order.ShippingName, order.ShippingAddress1, order.ShippingAddress2,
		order.ShippingCity, order.ShippingState, order.ShippingZip, order.ShippingCountry,
		order.CreatedAt, order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// Insert order items and deduct stock
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

		// Deduct stock for this item
		if err := s.inventoryService.DeductStock(item.VariantID, item.Quantity); err != nil {
			return fmt.Errorf("deduct stock for variant %s: %w", item.VariantID, err)
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
	var printfulOrderID sql.NullInt64
	var printfulRetryCount sql.NullInt64
	var trackingNumber sql.NullString
	var trackingURL sql.NullString

	err := s.db.QueryRow(`
		SELECT id, customer_id, customer_email, status, total_amount, currency,
			stripe_session_id, stripe_payment_intent_id, printful_order_id,
			COALESCE(printful_retry_count, 0) as printful_retry_count,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			tracking_number, tracking_url, created_at, updated_at
		FROM orders WHERE id = ?
	`, id).Scan(
		&order.ID, &order.CustomerID, &order.CustomerEmail, &order.Status, &order.TotalAmount, &order.Currency,
		&order.StripeSessionID, &order.StripePaymentIntentID, &printfulOrderID,
		&printfulRetryCount,
		&order.ShippingName, &order.ShippingAddress1, &order.ShippingAddress2,
		&order.ShippingCity, &order.ShippingState, &order.ShippingZip, &order.ShippingCountry,
		&trackingNumber, &trackingURL, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	// Convert nullable fields
	if printfulOrderID.Valid {
		order.PrintfulOrderID = printfulOrderID.Int64
	}
	if printfulRetryCount.Valid {
		order.PrintfulRetryCount = int(printfulRetryCount.Int64)
	}
	if trackingNumber.Valid {
		order.TrackingNumber = trackingNumber.String
	}
	if trackingURL.Valid {
		order.TrackingURL = trackingURL.String
	}

	return order, nil
}

// GetOrderItems retrieves all items for an order
func (s *Service) GetOrderItems(orderID string) ([]models.OrderItem, error) {
	rows, err := s.db.Query(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.variant_id,
			COALESCE(v.printful_variant_id, 0) as printful_variant_id,
			oi.quantity, oi.unit_price, oi.total_price,
			oi.product_name, oi.variant_name, oi.created_at
		FROM order_items oi
		LEFT JOIN variants v ON oi.variant_id = v.id
		WHERE oi.order_id = ?
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.VariantID,
			&item.PrintfulVariantID,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice,
			&item.ProductName, &item.VariantName, &item.CreatedAt,
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

// IncrementPrintfulRetryCount increments the retry count for an order
func (s *Service) IncrementPrintfulRetryCount(orderID string) error {
	_, err := s.db.Exec(`
		UPDATE orders SET
			printful_retry_count = printful_retry_count + 1,
			updated_at = ?
		WHERE id = ?
	`, time.Now(), orderID)

	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}

	return nil
}

// RecordPrintfulFailure logs a Printful submission failure
func (s *Service) RecordPrintfulFailure(orderID string, attemptNumber int, errorMsg, errorDetails string) error {
	_, err := s.db.Exec(`
		INSERT INTO printful_submission_failures (
			id, order_id, attempt_number, error_message, error_details, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`, fmt.Sprintf("%s-%d", orderID, attemptNumber), orderID, attemptNumber, errorMsg, errorDetails, time.Now())

	if err != nil {
		return fmt.Errorf("record printful failure: %w", err)
	}

	return nil
}

// GetFailedPrintfulOrders returns orders that failed Printful submission and are eligible for retry
func (s *Service) GetFailedPrintfulOrders() ([]models.Order, error) {
	// Find orders that:
	// 1. Have status = 'paid' (payment confirmed)
	// 2. Have NULL printful_order_id (not yet submitted)
	// 3. Were created less than 24 hours ago
	// 4. Have at least one retry attempt (failed at least once)
	rows, err := s.db.Query(`
		SELECT id, customer_id, customer_email, status, total_amount, currency,
			stripe_session_id, stripe_payment_intent_id, printful_order_id,
			printful_retry_count,
			shipping_name, shipping_address1, shipping_address2,
			shipping_city, shipping_state, shipping_zip, shipping_country,
			tracking_number, tracking_url, created_at, updated_at
		FROM orders
		WHERE status = ?
			AND printful_order_id IS NULL
			AND created_at > datetime('now', '-24 hours')
			AND printful_retry_count > 0
		ORDER BY created_at ASC
	`, models.OrderStatusPaid)

	if err != nil {
		return nil, fmt.Errorf("query failed orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var printfulOrderID sql.NullInt64
		var trackingNumber sql.NullString
		var trackingURL sql.NullString
		var shippingAddress2 sql.NullString
		var stripeSessionID sql.NullString
		var stripePaymentIntentID sql.NullString

		if err := rows.Scan(
			&order.ID, &order.CustomerID, &order.CustomerEmail, &order.Status, &order.TotalAmount, &order.Currency,
			&stripeSessionID, &stripePaymentIntentID, &printfulOrderID,
			&order.PrintfulRetryCount,
			&order.ShippingName, &order.ShippingAddress1, &shippingAddress2,
			&order.ShippingCity, &order.ShippingState, &order.ShippingZip, &order.ShippingCountry,
			&trackingNumber, &trackingURL, &order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		// Convert nullable fields
		if printfulOrderID.Valid {
			order.PrintfulOrderID = printfulOrderID.Int64
		}
		if trackingNumber.Valid {
			order.TrackingNumber = trackingNumber.String
		}
		if trackingURL.Valid {
			order.TrackingURL = trackingURL.String
		}
		if shippingAddress2.Valid {
			order.ShippingAddress2 = shippingAddress2.String
		}
		if stripeSessionID.Valid {
			order.StripeSessionID = stripeSessionID.String
		}
		if stripePaymentIntentID.Valid {
			order.StripePaymentIntentID = stripePaymentIntentID.String
		}

		orders = append(orders, order)
	}

	return orders, nil
}
