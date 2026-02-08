package handlers

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"
)

// URL represents a single URL entry in the sitemap
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// URLSet represents the root element of the sitemap
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// GetSitemap generates and serves the sitemap.xml
func (h *Handler) GetSitemap(w http.ResponseWriter, r *http.Request) {
	// Determine base URL based on environment
	baseURL := h.getBaseURL()

	// Create sitemap structure
	sitemap := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []URL{},
	}

	// Current timestamp for lastmod
	now := time.Now().Format("2006-01-02")

	// Add static pages
	staticPages := []struct {
		path       string
		changefreq string
		priority   float64
	}{
		{"/home", "monthly", 1.0},          // Homepage - highest priority
		{"/portfolio", "weekly", 0.9},      // Portfolio - high priority
		{"/merch", "daily", 0.9},           // Merch/shop - high priority, changes often
		{"/tour", "weekly", 0.8},           // Tour dates - medium-high priority
		{"/about", "monthly", 0.7},         // About page
		{"/gallery", "monthly", 0.7},       // Gallery
		{"/nessie-digital", "monthly", 0.7}, // Nessie Digital
		{"/contact", "monthly", 0.6},       // Contact page
		{"/cart", "yearly", 0.3},           // Cart - low priority, functional page
	}

	for _, page := range staticPages {
		sitemap.URLs = append(sitemap.URLs, URL{
			Loc:        baseURL + page.path,
			LastMod:    now,
			ChangeFreq: page.changefreq,
			Priority:   page.priority,
		})
	}

	// Add dynamic product pages from database
	products, err := h.getProductsForSitemap()
	if err != nil {
		log.Printf("Error fetching products for sitemap: %v", err)
		// Continue anyway with static pages
	} else {
		for _, productID := range products {
			sitemap.URLs = append(sitemap.URLs, URL{
				Loc:        fmt.Sprintf("%s/product-detail?id=%s", baseURL, productID),
				LastMod:    now,
				ChangeFreq: "weekly",
				Priority:   0.8, // Product pages are important for e-commerce
			})
		}
	}

	// Set headers for XML response
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Write XML declaration
	w.Write([]byte(xml.Header))

	// Encode and write sitemap
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(sitemap); err != nil {
		log.Printf("Error encoding sitemap: %v", err)
		http.Error(w, "Failed to generate sitemap", http.StatusInternalServerError)
		return
	}
}

// getBaseURL returns the appropriate base URL based on environment
func (h *Handler) getBaseURL() string {
	switch h.config.Env {
	case "production":
		if h.config.ProductionDomain != "" {
			return "https://" + h.config.ProductionDomain
		}
		// Fallback if production domain not set
		return "https://nessieaudio.com"
	case "staging":
		return "https://staging.nessieaudio.com" // Update with your staging domain
	default: // development
		return "http://localhost:5500"
	}
}

// getProductsForSitemap fetches all product IDs from the database
func (h *Handler) getProductsForSitemap() ([]string, error) {
	rows, err := h.db.Query("SELECT id FROM products WHERE active = 1")
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var productIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan product id: %w", err)
		}
		productIDs = append(productIDs, id)
	}

	return productIDs, nil
}
