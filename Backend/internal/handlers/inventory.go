package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nessieaudio/ecommerce-backend/internal/inventory"
)

// GetInventoryStatus returns inventory status for all variants
// GET /api/v1/inventory
func (h *Handler) GetInventoryStatus(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT
			v.id,
			v.name,
			v.product_id,
			p.name,
			v.stock_quantity,
			v.low_stock_threshold,
			v.track_inventory,
			v.available
		FROM variants v
		JOIN products p ON v.product_id = p.id
		WHERE v.track_inventory = 1
		ORDER BY p.name, v.name
	`)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch inventory")
		return
	}
	defer rows.Close()

	type InventoryItem struct {
		VariantID         string  `json:"variant_id"`
		VariantName       string  `json:"variant_name"`
		ProductID         string  `json:"product_id"`
		ProductName       string  `json:"product_name"`
		StockQuantity     *int    `json:"stock_quantity"`
		LowStockThreshold int     `json:"low_stock_threshold"`
		TrackInventory    bool    `json:"track_inventory"`
		Available         bool    `json:"available"`
		Status            string  `json:"status"` // "in_stock", "low_stock", "out_of_stock"
	}

	var items []InventoryItem
	for rows.Next() {
		var item InventoryItem
		var stockQty sql.NullInt64

		err := rows.Scan(
			&item.VariantID,
			&item.VariantName,
			&item.ProductID,
			&item.ProductName,
			&stockQty,
			&item.LowStockThreshold,
			&item.TrackInventory,
			&item.Available,
		)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error reading inventory")
			return
		}

		if stockQty.Valid {
			qty := int(stockQty.Int64)
			item.StockQuantity = &qty

			// Determine status
			if qty == 0 {
				item.Status = "out_of_stock"
			} else if qty <= item.LowStockThreshold {
				item.Status = "low_stock"
			} else {
				item.Status = "in_stock"
			}
		} else {
			item.Status = "unlimited"
		}

		items = append(items, item)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"inventory": items,
		"total":     len(items),
	})
}

// GetLowStockItems returns all items below their low stock threshold
// GET /api/v1/inventory/low-stock
func (h *Handler) GetLowStockItems(w http.ResponseWriter, r *http.Request) {
	inventoryService := inventory.NewService(h.db)

	items, err := inventoryService.GetLowStockItems()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch low stock items")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"low_stock_items": items,
		"count":           len(items),
	})
}

// UpdateVariantInventory updates stock quantity for a specific variant
// PUT /api/v1/inventory/{variant_id}
func (h *Handler) UpdateVariantInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	variantID := vars["variant_id"]

	type UpdateInventoryRequest struct {
		StockQuantity     *int `json:"stock_quantity"`
		LowStockThreshold int  `json:"low_stock_threshold"`
		TrackInventory    bool `json:"track_inventory"`
	}

	var req UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If tracking is enabled, stock_quantity must be provided
	if req.TrackInventory && req.StockQuantity == nil {
		respondError(w, http.StatusBadRequest, "stock_quantity required when track_inventory is true")
		return
	}

	inventoryService := inventory.NewService(h.db)

	// Update inventory
	var stockQty int
	if req.StockQuantity != nil {
		stockQty = *req.StockQuantity
	}

	if err := inventoryService.UpdateStock(variantID, stockQty, req.LowStockThreshold, req.TrackInventory); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update inventory")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":             "Inventory updated successfully",
		"variant_id":          variantID,
		"stock_quantity":      req.StockQuantity,
		"low_stock_threshold": req.LowStockThreshold,
		"track_inventory":     req.TrackInventory,
	})
}

// CheckVariantStock checks if a specific quantity is available for a variant
// GET /api/v1/inventory/{variant_id}/check?quantity=5
func (h *Handler) CheckVariantStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	variantID := vars["variant_id"]

	// Get quantity from query parameter
	quantityStr := r.URL.Query().Get("quantity")
	if quantityStr == "" {
		quantityStr = "1"
	}

	var quantity int
	if _, err := fmt.Sscanf(quantityStr, "%d", &quantity); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid quantity parameter")
		return
	}

	inventoryService := inventory.NewService(h.db)

	stockCheck, err := inventoryService.CheckStock(variantID, quantity)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check stock")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"variant_id":       stockCheck.VariantID,
		"requested_qty":    stockCheck.RequestedQty,
		"available":        stockCheck.Available,
		"stock_quantity":   stockCheck.StockQuantity,
		"track_inventory":  stockCheck.TrackInventory,
	})
}

// SendLowStockAlert manually triggers a low stock alert email
// POST /api/v1/inventory/send-alert
func (h *Handler) SendLowStockAlert(w http.ResponseWriter, r *http.Request) {
	inventoryService := inventory.NewService(h.db)

	// Get low stock items
	lowStockItems, err := inventoryService.GetLowStockItems()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get low stock items")
		return
	}

	if len(lowStockItems) == 0 {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "No low stock items found",
			"count":   0,
		})
		return
	}

	// Create alert service and send email
	alertService := inventory.NewAlertService(inventoryService, h.emailClient, h.config)

	if err := alertService.CheckAndSendLowStockAlerts(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send alert email")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Low stock alert sent successfully",
		"count":   len(lowStockItems),
		"items":   lowStockItems,
	})
}
