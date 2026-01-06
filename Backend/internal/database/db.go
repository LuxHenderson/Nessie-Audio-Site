package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database and creates tables
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// createTables creates all necessary database tables
func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		printful_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		currency TEXT DEFAULT 'USD',
		image_url TEXT,
		thumbnail_url TEXT,
		category TEXT,
		active BOOLEAN DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS variants (
		id TEXT PRIMARY KEY,
		product_id TEXT NOT NULL,
		printful_variant_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		size TEXT,
		color TEXT,
		price REAL NOT NULL,
		available BOOLEAN DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (product_id) REFERENCES products(id)
	);

	CREATE TABLE IF NOT EXISTS customers (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		name TEXT,
		phone TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		customer_email TEXT,
		status TEXT NOT NULL,
		total_amount REAL NOT NULL,
		currency TEXT DEFAULT 'USD',
		stripe_session_id TEXT,
		stripe_payment_intent_id TEXT,
		printful_order_id INTEGER,
		shipping_name TEXT,
		shipping_address1 TEXT,
		shipping_address2 TEXT,
		shipping_city TEXT,
		shipping_state TEXT,
		shipping_zip TEXT,
		shipping_country TEXT,
		tracking_number TEXT,
		tracking_url TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (customer_id) REFERENCES customers(id)
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id TEXT PRIMARY KEY,
		order_id TEXT NOT NULL,
		product_id TEXT NOT NULL,
		variant_id TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		unit_price REAL NOT NULL,
		total_price REAL NOT NULL,
		product_name TEXT NOT NULL,
		variant_name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (order_id) REFERENCES orders(id),
		FOREIGN KEY (product_id) REFERENCES products(id),
		FOREIGN KEY (variant_id) REFERENCES variants(id)
	);

	CREATE TABLE IF NOT EXISTS printful_webhook_events (
		id TEXT PRIMARY KEY,
		event_type TEXT NOT NULL,
		order_id TEXT,
		payload TEXT NOT NULL,
		processed BOOLEAN DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS stripe_webhook_events (
		id TEXT PRIMARY KEY,
		event_type TEXT NOT NULL,
		event_id TEXT NOT NULL UNIQUE,
		payload TEXT NOT NULL,
		processed BOOLEAN DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
	CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
	CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
	CREATE INDEX IF NOT EXISTS idx_variants_product ON variants(product_id);
	`

	_, err := db.Exec(schema)
	return err
}

// runMigrations runs database migrations to add new columns
func runMigrations(db *sql.DB) error {
	// Check if customer_email column exists in orders table
	var columnExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('orders')
		WHERE name='customer_email'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("check customer_email column: %w", err)
	}

	// Add customer_email column if it doesn't exist
	if !columnExists {
		_, err := db.Exec(`ALTER TABLE orders ADD COLUMN customer_email TEXT`)
		if err != nil {
			return fmt.Errorf("add customer_email column: %w", err)
		}
	}

	// Check if printful_retry_count column exists in orders table
	var retryCountExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('orders')
		WHERE name='printful_retry_count'
	`).Scan(&retryCountExists)

	if err != nil {
		return fmt.Errorf("check printful_retry_count column: %w", err)
	}

	// Add printful_retry_count column if it doesn't exist
	if !retryCountExists {
		_, err := db.Exec(`ALTER TABLE orders ADD COLUMN printful_retry_count INTEGER DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("add printful_retry_count column: %w", err)
		}
	}

	// Create printful_submission_failures table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS printful_submission_failures (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			attempt_number INTEGER NOT NULL,
			error_message TEXT NOT NULL,
			error_details TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("create printful_submission_failures table: %w", err)
	}

	// Create index for querying failures by order
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_printful_failures_order ON printful_submission_failures(order_id)`)
	if err != nil {
		return fmt.Errorf("create printful failures index: %w", err)
	}

	// Check if stock_quantity column exists in variants table
	var stockQuantityExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('variants')
		WHERE name='stock_quantity'
	`).Scan(&stockQuantityExists)

	if err != nil {
		return fmt.Errorf("check stock_quantity column: %w", err)
	}

	// Add stock_quantity column if it doesn't exist (default to NULL for print-on-demand)
	if !stockQuantityExists {
		_, err := db.Exec(`ALTER TABLE variants ADD COLUMN stock_quantity INTEGER`)
		if err != nil {
			return fmt.Errorf("add stock_quantity column: %w", err)
		}
	}

	// Check if low_stock_threshold column exists in variants table
	var lowStockThresholdExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('variants')
		WHERE name='low_stock_threshold'
	`).Scan(&lowStockThresholdExists)

	if err != nil {
		return fmt.Errorf("check low_stock_threshold column: %w", err)
	}

	// Add low_stock_threshold column if it doesn't exist (default 5)
	if !lowStockThresholdExists {
		_, err := db.Exec(`ALTER TABLE variants ADD COLUMN low_stock_threshold INTEGER DEFAULT 5`)
		if err != nil {
			return fmt.Errorf("add low_stock_threshold column: %w", err)
		}
	}

	// Check if track_inventory column exists in variants table
	var trackInventoryExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('variants')
		WHERE name='track_inventory'
	`).Scan(&trackInventoryExists)

	if err != nil {
		return fmt.Errorf("check track_inventory column: %w", err)
	}

	// Add track_inventory column if it doesn't exist (default FALSE for print-on-demand)
	if !trackInventoryExists {
		_, err := db.Exec(`ALTER TABLE variants ADD COLUMN track_inventory BOOLEAN DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("add track_inventory column: %w", err)
		}
	}

	return nil
}
