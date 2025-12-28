package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DatabasePath string

	// Printful
	PrintfulAPIKey    string
	PrintfulAPIURL    string
	PrintfulWebhookSecret string

	// Stripe
	StripeSecretKey      string
	StripePublishableKey string
	StripeWebhookSecret  string
	StripeSuccessURL     string
	StripeCancelURL      string

	// Production
	ProductionDomain string

	// CORS
	AllowedOrigins string

	// Logging
	LogLevel string
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getStripeSuccessURL returns the appropriate success URL based on environment
func getStripeSuccessURL(cfg *Config) string {
	// Check if explicitly set in environment
	if url := os.Getenv("STRIPE_SUCCESS_URL"); url != "" {
		return url
	}

	// Auto-detect based on environment
	if cfg.Env == "production" {
		if cfg.ProductionDomain != "" {
			return fmt.Sprintf("https://%s/cart-success.html", cfg.ProductionDomain)
		}
		// Fallback: Will need to be configured
		log.Println("WARNING: Production mode but PRODUCTION_DOMAIN not set. Set STRIPE_SUCCESS_URL or PRODUCTION_DOMAIN.")
		return "https://yoursite.com/cart-success.html"
	}

	// Development mode
	return "http://localhost:5500/cart-success.html"
}

// getStripeCancelURL returns the appropriate cancel URL based on environment
func getStripeCancelURL(cfg *Config) string {
	// Check if explicitly set in environment
	if url := os.Getenv("STRIPE_CANCEL_URL"); url != "" {
		return url
	}

	// Auto-detect based on environment
	if cfg.Env == "production" {
		if cfg.ProductionDomain != "" {
			return fmt.Sprintf("https://%s/cart-cancel.html", cfg.ProductionDomain)
		}
		// Fallback: Will need to be configured
		log.Println("WARNING: Production mode but PRODUCTION_DOMAIN not set. Set STRIPE_CANCEL_URL or PRODUCTION_DOMAIN.")
		return "https://yoursite.com/cart-cancel.html"
	}

	// Development mode
	return "http://localhost:5500/cart-cancel.html"
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		Env:                   getEnv("ENV", "development"),
		DatabasePath:          getEnv("DATABASE_PATH", "./nessie_store.db"),
		PrintfulAPIKey:        getEnv("PRINTFUL_API_KEY", ""),
		PrintfulAPIURL:        getEnv("PRINTFUL_API_URL", "https://api.printful.com"),
		PrintfulWebhookSecret: getEnv("PRINTFUL_WEBHOOK_SECRET", ""),
		StripeSecretKey:       getEnv("STRIPE_SECRET_KEY", ""),
		StripePublishableKey:  getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StripeWebhookSecret:   getEnv("STRIPE_WEBHOOK_SECRET", ""),
		ProductionDomain:      getEnv("PRODUCTION_DOMAIN", ""),
		AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
	}

	// Auto-detect Stripe redirect URLs based on environment
	cfg.StripeSuccessURL = getStripeSuccessURL(cfg)
	cfg.StripeCancelURL = getStripeCancelURL(cfg)

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	// Only check critical fields - API keys can be added later
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	// Log warnings for missing API keys
	if c.PrintfulAPIKey == "" {
		log.Println("WARNING: PRINTFUL_API_KEY not set - Printful integration will not work")
	}
	if c.StripeSecretKey == "" {
		log.Println("WARNING: STRIPE_SECRET_KEY not set - payment processing will not work")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
