-- Initial database schema for Nessie Audio eCommerce
-- This migration captures the existing schema at the time migrations were introduced

-- Products table
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

-- Variants table
CREATE TABLE IF NOT EXISTS variants (
	id TEXT PRIMARY KEY,
	product_id TEXT NOT NULL,
	printful_variant_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	size TEXT,
	color TEXT,
	price REAL NOT NULL,
	available BOOLEAN DEFAULT 1,
	stock_quantity INTEGER,
	low_stock_threshold INTEGER DEFAULT 5,
	track_inventory BOOLEAN DEFAULT 0,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Customers table
CREATE TABLE IF NOT EXISTS customers (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	name TEXT,
	phone TEXT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

-- Orders table
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
	printful_retry_count INTEGER DEFAULT 0,
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

-- Order items table
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

-- Printful webhook events table
CREATE TABLE IF NOT EXISTS printful_webhook_events (
	id TEXT PRIMARY KEY,
	event_type TEXT NOT NULL,
	order_id TEXT,
	payload TEXT NOT NULL,
	processed BOOLEAN DEFAULT 0,
	created_at DATETIME NOT NULL
);

-- Stripe webhook events table
CREATE TABLE IF NOT EXISTS stripe_webhook_events (
	id TEXT PRIMARY KEY,
	event_type TEXT NOT NULL,
	event_id TEXT NOT NULL UNIQUE,
	payload TEXT NOT NULL,
	processed BOOLEAN DEFAULT 0,
	created_at DATETIME NOT NULL
);

-- Printful submission failures table
CREATE TABLE IF NOT EXISTS printful_submission_failures (
	id TEXT PRIMARY KEY,
	order_id TEXT NOT NULL,
	attempt_number INTEGER NOT NULL,
	error_message TEXT NOT NULL,
	error_details TEXT,
	created_at DATETIME NOT NULL,
	FOREIGN KEY (order_id) REFERENCES orders(id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_variants_product ON variants(product_id);
CREATE INDEX IF NOT EXISTS idx_printful_failures_order ON printful_submission_failures(order_id);
