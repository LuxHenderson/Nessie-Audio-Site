package inventory

import (
	"database/sql"
	"fmt"
	"log"
)

// Service handles inventory tracking and management
type Service struct {
	db *sql.DB
}

// NewService creates a new inventory service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// StockCheck represents the result of a stock availability check
type StockCheck struct {
	VariantID       string
	Available       bool
	StockQuantity   *int // nil = unlimited
	RequestedQty    int
	TrackInventory  bool
}

// CheckStock verifies if requested quantity is available
func (s *Service) CheckStock(variantID string, requestedQty int) (*StockCheck, error) {
	var stockQty sql.NullInt64
	var trackInventory bool

	err := s.db.QueryRow(`
		SELECT stock_quantity, track_inventory
		FROM variants
		WHERE id = ?
	`, variantID).Scan(&stockQty, &trackInventory)

	if err != nil {
		return nil, fmt.Errorf("query variant stock: %w", err)
	}

	check := &StockCheck{
		VariantID:      variantID,
		RequestedQty:   requestedQty,
		TrackInventory: trackInventory,
	}

	// If inventory tracking is disabled, stock is unlimited
	if !trackInventory {
		check.Available = true
		check.StockQuantity = nil
		return check, nil
	}

	// If tracking is enabled but stock_quantity is NULL, treat as out of stock
	if !stockQty.Valid {
		check.Available = false
		check.StockQuantity = nil
		return check, nil
	}

	currentStock := int(stockQty.Int64)
	check.StockQuantity = &currentStock
	check.Available = currentStock >= requestedQty

	return check, nil
}

// DeductStock reduces stock quantity for a variant
func (s *Service) DeductStock(variantID string, quantity int) error {
	// First check if we should track inventory for this variant
	var trackInventory bool
	var stockQty sql.NullInt64
	var lowStockThreshold int
	err := s.db.QueryRow(`
		SELECT track_inventory, stock_quantity, low_stock_threshold FROM variants WHERE id = ?
	`, variantID).Scan(&trackInventory, &stockQty, &lowStockThreshold)

	if err != nil {
		return fmt.Errorf("query variant info: %w", err)
	}

	// Skip deduction if inventory tracking is disabled
	if !trackInventory {
		return nil
	}

	result, err := s.db.Exec(`
		UPDATE variants
		SET stock_quantity = stock_quantity - ?,
		    updated_at = datetime('now')
		WHERE id = ?
		AND track_inventory = 1
		AND stock_quantity >= ?
	`, quantity, variantID, quantity)

	if err != nil {
		return fmt.Errorf("deduct stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient stock for variant %s", variantID)
	}

	log.Printf("Deducted %d units from variant %s", quantity, variantID)

	// Check if stock just went below threshold (for potential alert)
	if stockQty.Valid {
		newStock := int(stockQty.Int64) - quantity
		if newStock <= lowStockThreshold && newStock >= 0 {
			log.Printf("⚠️ Low stock warning: variant %s now has %d units (threshold: %d)",
				variantID, newStock, lowStockThreshold)
		}
	}

	return nil
}

// RestoreStock adds stock back (e.g., when an order is cancelled)
func (s *Service) RestoreStock(variantID string, quantity int) error {
	// First check if we should track inventory for this variant
	var trackInventory bool
	err := s.db.QueryRow(`
		SELECT track_inventory FROM variants WHERE id = ?
	`, variantID).Scan(&trackInventory)

	if err != nil {
		return fmt.Errorf("query track_inventory: %w", err)
	}

	// Skip restoration if inventory tracking is disabled
	if !trackInventory {
		return nil
	}

	_, err = s.db.Exec(`
		UPDATE variants
		SET stock_quantity = stock_quantity + ?,
		    updated_at = datetime('now')
		WHERE id = ?
		AND track_inventory = 1
	`, quantity, variantID)

	if err != nil {
		return fmt.Errorf("restore stock: %w", err)
	}

	log.Printf("Restored %d units to variant %s", quantity, variantID)
	return nil
}

// LowStockItem represents a variant with low stock
type LowStockItem struct {
	VariantID         string
	VariantName       string
	ProductID         string
	ProductName       string
	StockQuantity     int
	LowStockThreshold int
}

// GetLowStockItems returns all variants below their low stock threshold
func (s *Service) GetLowStockItems() ([]LowStockItem, error) {
	rows, err := s.db.Query(`
		SELECT
			v.id,
			v.name,
			v.product_id,
			p.name,
			v.stock_quantity,
			v.low_stock_threshold
		FROM variants v
		JOIN products p ON v.product_id = p.id
		WHERE v.track_inventory = 1
		AND v.stock_quantity IS NOT NULL
		AND v.stock_quantity <= v.low_stock_threshold
		ORDER BY v.stock_quantity ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("query low stock items: %w", err)
	}
	defer rows.Close()

	var items []LowStockItem
	for rows.Next() {
		var item LowStockItem
		err := rows.Scan(
			&item.VariantID,
			&item.VariantName,
			&item.ProductID,
			&item.ProductName,
			&item.StockQuantity,
			&item.LowStockThreshold,
		)
		if err != nil {
			return nil, fmt.Errorf("scan low stock item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// UpdateStock updates the stock quantity for a variant
func (s *Service) UpdateStock(variantID string, newQuantity int, threshold int, trackInventory bool) error {
	_, err := s.db.Exec(`
		UPDATE variants
		SET stock_quantity = ?,
		    low_stock_threshold = ?,
		    track_inventory = ?,
		    updated_at = datetime('now')
		WHERE id = ?
	`, newQuantity, threshold, trackInventory, variantID)

	if err != nil {
		return fmt.Errorf("update stock: %w", err)
	}

	return nil
}
