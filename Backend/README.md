# Nessie Audio eCommerce Backend

Production-ready Golang backend for Nessie Audio merch store with Printful fulfillment and Stripe payments.

## Overview

This backend provides a complete eCommerce API that:
- Manages products and variants from Printful
- Handles order creation and tracking
- Processes payments via Stripe
- Submits orders to Printful for fulfillment
- Receives webhooks from Stripe and Printful

**Important**: This backend uses Printful **only for fulfillment**, not as a storefront. Your frontend controls all UX and pricing.

## Quick Start

### 1. Prerequisites

- Go 1.21 or later
- Printful account (https://www.printful.com)
- Stripe account (https://stripe.com)

### 2. Setup

```bash
# Clone/navigate to backend directory
cd Backend

# Copy environment template
cp .env.example .env

# Edit .env and add your credentials (see Configuration section)
nano .env

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### 3. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Get products
curl http://localhost:8080/api/v1/products
```

## Configuration

Edit `.env` with your actual credentials:

### Required for Printful Integration

```env
PRINTFUL_API_KEY=your_printful_api_key_here
```

**Get your Printful API key:**
1. Go to https://www.printful.com/dashboard/api
2. Create a new API key
3. Copy the key to `.env`

### Required for Stripe Integration

```env
STRIPE_SECRET_KEY=sk_test_your_key_here
STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_secret_here
```

**Get your Stripe keys:**
1. Go to https://dashboard.stripe.com/apikeys
2. Copy Secret key and Publishable key
3. For webhook secret:
   - Go to https://dashboard.stripe.com/webhooks
   - Add endpoint: `https://yourdomain.com/webhooks/stripe`
   - Select events: `checkout.session.completed`
   - Copy the signing secret

### Optional Configuration

```env
PORT=8080                    # Server port
DATABASE_PATH=./store.db     # SQLite database file
ALLOWED_ORIGINS=http://localhost:3000  # CORS origins
```

## Project Structure

```
Backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── database/
│   │   └── db.go                # Database initialization
│   ├── handlers/
│   │   ├── handlers.go          # Handler setup
│   │   ├── products.go          # Product endpoints
│   │   ├── orders.go            # Order endpoints
│   │   ├── checkout.go          # Checkout endpoint
│   │   ├── webhook_stripe.go    # Stripe webhooks
│   │   └── webhook_printful.go  # Printful webhooks
│   ├── middleware/
│   │   └── middleware.go        # CORS, logging, recovery
│   ├── models/
│   │   └── models.go            # Data models
│   └── services/
│       ├── order/
│       │   └── service.go       # Order business logic
│       ├── printful/
│       │   └── client.go        # Printful API client
│       └── stripe/
│           └── client.go        # Stripe integration
├── .env.example                 # Environment template
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

## API Reference

### Products

#### List Products
```http
GET /api/v1/products
```

**Response:**
```json
{
  "products": [
    {
      "id": "uuid",
      "name": "Nessie Audio T-Shirt",
      "description": "Premium cotton tee",
      "price": 29.99,
      "currency": "USD",
      "image_url": "https://...",
      "category": "apparel"
    }
  ]
}
```

#### Get Product Details
```http
GET /api/v1/products/{id}
```

**Response:**
```json
{
  "id": "uuid",
  "name": "Nessie Audio T-Shirt",
  "price": 29.99,
  "variants": [
    {
      "id": "uuid",
      "name": "Large / Black",
      "size": "L",
      "color": "Black",
      "price": 29.99,
      "available": true
    }
  ]
}
```

### Orders

#### Create Order
```http
POST /api/v1/orders
Content-Type: application/json

{
  "customer_email": "customer@example.com",
  "items": [
    {
      "product_id": "uuid",
      "variant_id": "uuid",
      "quantity": 2
    }
  ]
}
```

**Response:**
```json
{
  "order_id": "uuid",
  "status": "pending"
}
```

#### Get Order
```http
GET /api/v1/orders/{id}
```

**Response:**
```json
{
  "order": {
    "id": "uuid",
    "customer_id": "uuid",
    "status": "paid",
    "total_amount": 59.98,
    "currency": "USD",
    "shipping_name": "John Doe",
    "tracking_number": "1Z999..."
  },
  "items": [...]
}
```

### Checkout

#### Create Checkout Session
```http
POST /api/v1/checkout
Content-Type: application/json

{
  "order_id": "uuid"
}
```

**Response:**
```json
{
  "session_id": "cs_test_..."
}
```

**Frontend usage:**
```javascript
const response = await fetch('http://localhost:8080/api/v1/checkout', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ order_id: orderId })
});

const { session_id } = await response.json();

// Redirect to Stripe Checkout
const stripe = Stripe('pk_test_...');
await stripe.redirectToCheckout({ sessionId: session_id });
```

## Frontend Integration

### Complete Checkout Flow

```javascript
// 1. Create an order
const createOrder = async (items) => {
  const response = await fetch('http://localhost:8080/api/v1/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      customer_email: 'customer@example.com',
      items: items
    })
  });
  
  const data = await response.json();
  return data.order_id;
};

// 2. Initiate checkout
const initiateCheckout = async (orderId) => {
  const response = await fetch('http://localhost:8080/api/v1/checkout', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_id: orderId })
  });
  
  const { session_id } = await response.json();
  
  // Redirect to Stripe
  const stripe = Stripe('pk_test_your_publishable_key');
  await stripe.redirectToCheckout({ sessionId: session_id });
};

// 3. Handle success (on your success page)
const urlParams = new URLSearchParams(window.location.search);
const sessionId = urlParams.get('session_id');
// Display confirmation, order is being fulfilled
```

## Webhooks

### Stripe Webhook

**Endpoint:** `POST /webhooks/stripe`

**Events handled:**
- `checkout.session.completed` - Payment confirmed, order submitted to Printful

**Setup:**
1. Go to Stripe Dashboard > Webhooks
2. Add endpoint: `https://yourdomain.com/webhooks/stripe`
3. Select event: `checkout.session.completed`
4. Copy signing secret to `.env`

### Printful Webhook

**Endpoint:** `POST /webhooks/printful`

**Events handled:**
- `order_updated` - Order status changed
- `shipment_created` - Tracking info available
- `order_failed` - Fulfillment failed

**Setup:**
1. Go to Printful Dashboard > Settings > Webhooks
2. Add webhook URL: `https://yourdomain.com/webhooks/printful`
3. Copy webhook secret to `.env`

## Adding Products

Products must be added to the database before they can be sold. Currently, this is a manual process:

### Option 1: From Printful API (Future)

TODO: Create a script to sync products from Printful

### Option 2: Manual SQL Insert

```sql
-- Add a product
INSERT INTO products (
  id, printful_id, name, description, price, currency,
  image_url, thumbnail_url, category, active, created_at, updated_at
) VALUES (
  'uuid-here',
  1234567,  -- Printful product ID
  'Nessie Audio T-Shirt',
  'Premium cotton t-shirt with logo',
  29.99,
  'USD',
  'https://printful.com/path/to/image.jpg',
  'https://printful.com/path/to/thumb.jpg',
  'apparel',
  1,
  datetime('now'),
  datetime('now')
);

-- Add variants
INSERT INTO variants (
  id, product_id, printful_variant_id, name,
  size, color, price, available, created_at, updated_at
) VALUES (
  'variant-uuid',
  'product-uuid',
  9876543,  -- Printful variant ID
  'Large / Black',
  'L',
  'Black',
  29.99,
  1,
  datetime('now'),
  datetime('now')
);
```

**Finding Printful IDs:**
1. Go to your Printful store
2. Select a product
3. Get product ID from URL or API
4. Get variant IDs from product API endpoint

## Order Fulfillment Flow

1. **Customer creates order** → Order status: `pending`
2. **Customer pays via Stripe** → Stripe webhook fires
3. **Backend confirms payment** → Order status: `paid`
4. **Backend submits to Printful** → Order status: `fulfilled`
5. **Printful ships order** → Printful webhook fires
6. **Backend updates tracking** → Order status: `shipped`

## Database

SQLite database with the following tables:

- `products` - Product catalog
- `variants` - Product variants (size/color)
- `customers` - Customer records
- `orders` - Order records
- `order_items` - Line items
- `printful_webhook_events` - Audit log
- `stripe_webhook_events` - Audit log

**Database file:** `nessie_store.db` (created automatically)

## Security Considerations

✅ Webhook signature verification (Stripe & Printful)  
✅ CORS protection  
✅ Environment-based secrets  
✅ SQL injection protection (parameterized queries)  
✅ Request timeout limits  

🔒 **Production checklist:**
- Use HTTPS (required for webhooks)
- Set strong `PRINTFUL_WEBHOOK_SECRET`
- Use Stripe production keys
- Restrict `ALLOWED_ORIGINS`
- Enable rate limiting
- Add authentication for admin endpoints

## Development

```bash
# Run server with auto-reload (install air)
go install github.com/cosmtrek/air@latest
air

# Run tests
go test ./...

# Build production binary
go build -o server cmd/server/main.go

# Run production binary
./server
```

## Deployment

### Option 1: Digital Ocean / AWS / GCP

```bash
# Build binary
GOOS=linux GOARCH=amd64 go build -o server cmd/server/main.go

# Upload to server
scp server user@yourserver:/opt/nessie-backend/
scp .env user@yourserver:/opt/nessie-backend/

# Run with systemd or supervisord
```

### Option 2: Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
COPY .env .
EXPOSE 8080
CMD ["./server"]
```

```bash
docker build -t nessie-backend .
docker run -p 8080:8080 nessie-backend
```

## Troubleshooting

### "PRINTFUL_API_KEY not set"
- Copy `.env.example` to `.env`
- Add your Printful API key

### "Failed to create checkout session"
- Verify `STRIPE_SECRET_KEY` is set
- Check Stripe Dashboard for errors

### Webhook not receiving events
- Ensure webhook URL is publicly accessible (use ngrok for local testing)
- Verify webhook secret matches
- Check server logs for errors

### Orders not submitting to Printful
- Verify `PrintfulVariantID` is set in variants table
- Check Printful Dashboard for rejected orders
- Review server logs

## TODO / Future Enhancements

- [ ] Product sync script from Printful API
- [ ] Admin API for product management
- [ ] Email notifications (order confirmation, shipping updates)
- [ ] Inventory tracking
- [ ] Discount codes
- [ ] Multiple currencies
- [ ] PostgreSQL support
- [ ] Redis caching
- [ ] Prometheus metrics

## Support

For issues related to:
- **Printful**: https://www.printful.com/api/docs
- **Stripe**: https://stripe.com/docs
- **This backend**: Check server logs or GitHub issues

## License

Proprietary - Nessie Audio © 2025
