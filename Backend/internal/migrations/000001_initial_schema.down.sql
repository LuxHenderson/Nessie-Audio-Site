-- Rollback initial schema
-- This drops all tables in reverse dependency order

-- Drop indexes first
DROP INDEX IF EXISTS idx_printful_failures_order;
DROP INDEX IF EXISTS idx_variants_product;
DROP INDEX IF EXISTS idx_order_items_order;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_customer;

-- Drop tables (child tables before parent tables due to foreign keys)
DROP TABLE IF EXISTS printful_submission_failures;
DROP TABLE IF EXISTS stripe_webhook_events;
DROP TABLE IF EXISTS printful_webhook_events;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS variants;
DROP TABLE IF EXISTS products;
