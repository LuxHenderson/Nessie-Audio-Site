package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

// GetProductsResponse represents the products API response
type GetProductsResponse struct {
	Products []ProductResponse `json:"products"`
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Price        float64           `json:"price"`
	MinPrice     float64           `json:"min_price,omitempty"`
	MaxPrice     float64           `json:"max_price,omitempty"`
	Currency     string            `json:"currency"`
	ImageURL     string            `json:"image_url"`
	ThumbnailURL string            `json:"thumbnail_url"`
	Category     string            `json:"category"`
	Variants     []VariantResponse `json:"variants,omitempty"`
}

// VariantResponse represents a product variant
type VariantResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Size      string  `json:"size"`
	Color     string  `json:"color"`
	Price     float64 `json:"price"`
	Available bool    `json:"available"`
}

// GetProducts returns all active products
// GET /api/v1/products
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Query products from database
	rows, err := h.db.Query(`
		SELECT id, name, description, price, currency, image_url, thumbnail_url, category
		FROM products WHERE active = 1
		ORDER BY created_at DESC
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}
	defer rows.Close()

	var products []ProductResponse
	for rows.Next() {
		var p ProductResponse
		var description, imageURL, thumbnailURL, category sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &description, &p.Price, &p.Currency,
			&imageURL, &thumbnailURL, &category); err != nil {
			respondError(w, http.StatusInternalServerError, "Error reading products")
			return
		}

		p.Description = description.String
		p.ImageURL = imageURL.String
		p.ThumbnailURL = thumbnailURL.String
		p.Category = category.String

		// Get min and max prices from variants
		var minPrice, maxPrice sql.NullFloat64
		err := h.db.QueryRow(`
			SELECT MIN(price), MAX(price)
			FROM variants
			WHERE product_id = ? AND available = 1
		`, p.ID).Scan(&minPrice, &maxPrice)
		
		if err == nil && minPrice.Valid && maxPrice.Valid {
			p.MinPrice = minPrice.Float64
			p.MaxPrice = maxPrice.Float64
		}

		products = append(products, p)
	}

	respondJSON(w, http.StatusOK, GetProductsResponse{Products: products})
}

// GetProduct returns a single product with variants
// GET /api/v1/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	// Get product
	var product ProductResponse
	var description, imageURL, thumbnailURL, category sql.NullString
	err := h.db.QueryRow(`
		SELECT id, name, description, price, currency, image_url, thumbnail_url, category
		FROM products WHERE id = ? AND active = 1
	`, productID).Scan(&product.ID, &product.Name, &description, &product.Price,
		&product.Currency, &imageURL, &thumbnailURL, &category)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch product")
		return
	}

	product.Description = description.String
	product.ImageURL = imageURL.String
	product.ThumbnailURL = thumbnailURL.String
	product.Category = category.String

	// Get variants
	rows, err := h.db.Query(`
		SELECT id, product_id, name, COALESCE(size, ''), COALESCE(color, ''), price, available
		FROM variants WHERE product_id = ? AND available = 1
		ORDER BY name
	`, productID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch variants")
		return
	}
	defer rows.Close()

	var variants []VariantResponse
	for rows.Next() {
		var v VariantResponse
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Name, &v.Size, &v.Color, &v.Price, &v.Available); err != nil {
			respondError(w, http.StatusInternalServerError, "Error reading variants")
			return
		}
		variants = append(variants, v)
	}

	product.Variants = variants

	respondJSON(w, http.StatusOK, product)
}
