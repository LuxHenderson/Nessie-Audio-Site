# Inventory Tracking & Low-Stock Alerts

## Overview

The Nessie Audio backend includes a comprehensive inventory tracking system with automatic low-stock alerts. This system is designed to work seamlessly with both print-on-demand products and physical inventory.

## Features

- **Flexible Inventory Tracking**: Choose which variants to track (print-on-demand items don't need tracking)
- **Automatic Stock Deduction**: Stock automatically decreases when orders are placed
- **Low-Stock Alerts**: Email notifications when inventory drops below threshold
- **Stock Availability Checks**: Prevents overselling by checking stock before order creation
- **REST API**: Full API for inventory management

## How It Works

### Print-on-Demand vs Physical Inventory

**Print-on-Demand (Default):**
- `track_inventory = false`
- `stock_quantity = NULL`
- No stock limitations
- Orders always allowed

**Physical Inventory:**
- `track_inventory = true`
- `stock_quantity = [number]`
- Low stock threshold monitoring
- Orders blocked when out of stock

### Automatic Stock Management

1. **Order Creation**: When a customer places an order, the system:
   - Checks if requested quantity is available
   - Deducts stock from all ordered variants
   - Logs a warning if stock drops below threshold

2. **Low Stock Detection**: Stock is considered "low" when:
   - `stock_quantity <= low_stock_threshold`
   - Threshold defaults to 5 units

3. **Email Alerts**: When stock is low:
   - Alert email sent to `ADMIN_EMAIL`
   - Includes all low-stock items
   - Shows current stock vs threshold

## Database Schema

### Variants Table (New Columns)

```sql
stock_quantity       INTEGER   -- NULL = unlimited (print-on-demand)
low_stock_threshold  INTEGER   -- Alert when stock <= this (default: 5)
track_inventory      BOOLEAN   -- Enable/disable tracking (default: false)
```

## API Endpoints

All endpoints are under `/api/v1/inventory` with rate limiting (60 requests/min).

### GET /api/v1/inventory
Get inventory status for all tracked variants.

**Response:**
```json
{
  "inventory": [
    {
      "variant_id": "abc-123",
      "variant_name": "Large / Black",
      "product_id": "prod-456",
      "product_name": "T-Shirt",
      "stock_quantity": 15,
      "low_stock_threshold": 5,
      "track_inventory": true,
      "available": true,
      "status": "in_stock"  // "in_stock", "low_stock", "out_of_stock", "unlimited"
    }
  ],
  "total": 1
}
```

### GET /api/v1/inventory/low-stock
Get only items below their low stock threshold.

**Response:**
```json
{
  "low_stock_items": [
    {
      "variant_id": "abc-123",
      "variant_name": "Medium / White",
      "product_id": "prod-789",
      "product_name": "Hoodie",
      "stock_quantity": 3,
      "low_stock_threshold": 5
    }
  ],
  "count": 1
}
```

### PUT /api/v1/inventory/{variant_id}
Update inventory settings for a variant.

**Request Body:**
```json
{
  "stock_quantity": 50,
  "low_stock_threshold": 10,
  "track_inventory": true
}
```

**Response:**
```json
{
  "message": "Inventory updated successfully",
  "variant_id": "abc-123",
  "stock_quantity": 50,
  "low_stock_threshold": 10,
  "track_inventory": true
}
```

### GET /api/v1/inventory/{variant_id}/check?quantity=5
Check if a specific quantity is available.

**Response:**
```json
{
  "variant_id": "abc-123",
  "requested_qty": 5,
  "available": true,
  "stock_quantity": 15,
  "track_inventory": true
}
```

### POST /api/v1/inventory/send-alert
Manually trigger a low-stock alert email (for testing or manual checks).

**Response:**
```json
{
  "message": "Low stock alert sent successfully",
  "count": 3,
  "items": [...]
}
```

## Usage Examples

### Enable Inventory Tracking for a Variant

```bash
curl -X PUT http://localhost:8080/api/v1/inventory/variant-123 \
  -H "Content-Type: application/json" \
  -d '{
    "stock_quantity": 100,
    "low_stock_threshold": 10,
    "track_inventory": true
  }'
```

### Check Current Inventory

```bash
curl http://localhost:8080/api/v1/inventory
```

### Check Low Stock Items

```bash
curl http://localhost:8080/api/v1/inventory/low-stock
```

### Check if Stock is Available

```bash
curl "http://localhost:8080/api/v1/inventory/variant-123/check?quantity=25"
```

### Send Low Stock Alert

```bash
curl -X POST http://localhost:8080/api/v1/inventory/send-alert
```

## Testing

Run the comprehensive test suite:

```bash
go run cmd/test-inventory/main.go
```

**Tests include:**
- Database schema verification
- Stock availability checking
- Stock deduction
- Low stock detection
- Stock restoration
- Email alert system

## Configuration

Set these environment variables in your `.env.*` files:

```env
# Email address for low stock alerts
ADMIN_EMAIL=your-email@example.com

# SMTP configuration (for sending alerts)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_FROM_NAME=Nessie Audio Inventory
```

## Low Stock Alert Email

When stock drops below threshold, you'll receive an HTML email with:
- List of all low-stock items
- Current stock levels
- Threshold values
- Status badges (Critical/Low/Out of Stock)
- Recommended actions

## Implementation Details

### Order Creation Flow

1. User submits order
2. System checks stock for all items:
   - If `track_inventory = false`: Allow order (print-on-demand)
   - If `track_inventory = true`:
     - Check `stock_quantity >= requested_quantity`
     - If insufficient, reject order with error message
3. If stock check passes:
   - Create order in database
   - Deduct stock from variants
   - Log warnings for low stock items
4. Commit transaction

### Services Architecture

**inventory.Service** (`internal/inventory/inventory.go`):
- `CheckStock()`: Verify availability
- `DeductStock()`: Reduce stock quantity
- `RestoreStock()`: Add stock back (for cancelled orders)
- `GetLowStockItems()`: Find items below threshold
- `UpdateStock()`: Modify inventory settings

**inventory.AlertService** (`internal/inventory/alerts.go`):
- `CheckAndSendLowStockAlerts()`: Check and email alerts
- `SendImmediateLowStockAlert()`: Send alert for specific item

**order.Service** (`internal/services/order/service.go`):
- Integrates inventory checks into order creation
- Automatic stock deduction during checkout

## Best Practices

1. **Set Realistic Thresholds**: Base `low_stock_threshold` on your reorder time and sales velocity
   - Fast-selling items: Higher threshold (e.g., 20)
   - Slow-selling items: Lower threshold (e.g., 5)

2. **Monitor Alerts**: Check your admin email regularly for low stock notifications

3. **Disable for Print-on-Demand**: Keep `track_inventory = false` for Printful items (default)

4. **Enable for Physical Products**: Set `track_inventory = true` for items you stock yourself

5. **Test Before Production**: Use the test script to verify everything works

## Troubleshooting

**Stock not deducting?**
- Check if `track_inventory = true` for the variant
- Verify `stock_quantity` is not NULL

**Not receiving alerts?**
- Verify `ADMIN_EMAIL` is set in environment variables
- Check SMTP credentials are correct
- Test with: `go run cmd/test-inventory/main.go`

**Order rejected for insufficient stock?**
- Check current stock: `GET /api/v1/inventory`
- Update stock if needed: `PUT /api/v1/inventory/{variant_id}`

**Want to disable tracking?**
```bash
curl -X PUT http://localhost:8080/api/v1/inventory/variant-123 \
  -H "Content-Type: application/json" \
  -d '{"track_inventory": false}'
```

## Future Enhancements

Potential features for future versions:
- Automatic reorder suggestions
- Stock movement history/audit log
- Multi-location inventory
- Reserved stock (for incomplete orders)
- Bulk inventory updates
- CSV import/export
- Scheduled alert digests (daily summary instead of immediate)

## Files

- `internal/inventory/inventory.go` - Core inventory service
- `internal/inventory/alerts.go` - Low stock alert system
- `internal/handlers/inventory.go` - REST API endpoints
- `internal/services/order/service.go` - Order integration
- `internal/models/models.go` - Variant model with inventory fields
- `internal/database/db.go` - Database migrations
- `cmd/test-inventory/main.go` - Test suite
