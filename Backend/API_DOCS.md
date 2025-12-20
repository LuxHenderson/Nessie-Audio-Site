# API Documentation - Frontend Contract

This document defines the exact request/response formats for frontend integration.

## Base URL
```
http://localhost:8080/api/v1
```

Production: `https://api.nessieaudio.com/api/v1`

## Authentication
None required for public endpoints (products, orders, checkout).

## Error Responses
All errors follow this format:
```json
{
  "error": "Error message here"
}
```

Status codes:
- `400` - Bad Request (invalid input)
- `404` - Not Found
- `500` - Internal Server Error

---

## Endpoints

### 1. Get All Products

**Request:**
```http
GET /api/v1/products
```

**Response:** `200 OK`
```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Nessie Audio Classic Tee",
      "description": "Premium cotton t-shirt with iconic logo",
      "price": 29.99,
      "currency": "USD",
      "image_url": "https://printful.com/files/product.jpg",
      "thumbnail_url": "https://printful.com/files/thumb.jpg",
      "category": "apparel"
    }
  ]
}
```

**Frontend Example:**
```javascript
const getProducts = async () => {
  const response = await fetch('http://localhost:8080/api/v1/products');
  const data = await response.json();
  return data.products;
};
```

---

### 2. Get Product Details

**Request:**
```http
GET /api/v1/products/{id}
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Nessie Audio Classic Tee",
  "description": "Premium cotton t-shirt",
  "price": 29.99,
  "currency": "USD",
  "image_url": "https://...",
  "thumbnail_url": "https://...",
  "category": "apparel",
  "variants": [
    {
      "id": "variant-uuid-1",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Small / Black",
      "size": "S",
      "color": "Black",
      "price": 29.99,
      "available": true
    },
    {
      "id": "variant-uuid-2",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Medium / Black",
      "size": "M",
      "color": "Black",
      "price": 29.99,
      "available": true
    }
  ]
}
```

**Frontend Example:**
```javascript
const getProduct = async (productId) => {
  const response = await fetch(`http://localhost:8080/api/v1/products/${productId}`);
  const product = await response.json();
  return product;
};
```

---

### 3. Create Order

**Request:**
```http
POST /api/v1/orders
Content-Type: application/json

{
  "customer_email": "customer@example.com",
  "items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "variant_id": "variant-uuid-1",
      "quantity": 2
    },
    {
      "product_id": "another-product-uuid",
      "variant_id": "another-variant-uuid",
      "quantity": 1
    }
  ]
}
```

**Response:** `201 Created`
```json
{
  "order_id": "order-uuid-here",
  "status": "pending"
}
```

**Frontend Example:**
```javascript
const createOrder = async (customerEmail, cartItems) => {
  const response = await fetch('http://localhost:8080/api/v1/orders', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      customer_email: customerEmail,
      items: cartItems.map(item => ({
        product_id: item.productId,
        variant_id: item.variantId,
        quantity: item.quantity
      }))
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }

  const data = await response.json();
  return data.order_id;
};
```

---

### 4. Get Order

**Request:**
```http
GET /api/v1/orders/{order_id}
```

**Response:** `200 OK`
```json
{
  "order": {
    "id": "order-uuid",
    "customer_id": "customer-uuid",
    "status": "paid",
    "total_amount": 89.97,
    "currency": "USD",
    "stripe_session_id": "cs_test_...",
    "stripe_payment_intent_id": "pi_...",
    "printful_order_id": 123456789,
    "shipping_name": "John Doe",
    "shipping_address1": "123 Main St",
    "shipping_address2": "Apt 4",
    "shipping_city": "San Francisco",
    "shipping_state": "CA",
    "shipping_zip": "94102",
    "shipping_country": "US",
    "tracking_number": "1Z999AA10123456784",
    "tracking_url": "https://www.ups.com/track?...",
    "created_at": "2025-12-20T10:00:00Z",
    "updated_at": "2025-12-20T10:05:00Z"
  },
  "items": [
    {
      "id": "item-uuid",
      "order_id": "order-uuid",
      "product_id": "product-uuid",
      "variant_id": "variant-uuid",
      "quantity": 2,
      "unit_price": 29.99,
      "total_price": 59.98,
      "product_name": "Nessie Audio Classic Tee",
      "variant_name": "Large / Black",
      "created_at": "2025-12-20T10:00:00Z"
    }
  ]
}
```

**Order Status Values:**
- `pending` - Order created, awaiting payment
- `paid` - Payment confirmed
- `fulfilled` - Submitted to Printful
- `shipped` - Package shipped, tracking available
- `cancelled` - Order cancelled

**Frontend Example:**
```javascript
const getOrder = async (orderId) => {
  const response = await fetch(`http://localhost:8080/api/v1/orders/${orderId}`);
  const data = await response.json();
  return data;
};
```

---

### 5. Create Checkout Session

**Request:**
```http
POST /api/v1/checkout
Content-Type: application/json

{
  "order_id": "order-uuid-from-step-3"
}
```

**Response:** `200 OK`
```json
{
  "session_id": "cs_test_a1B2c3D4e5F6g7H8i9J0"
}
```

**Frontend Example:**
```javascript
const initiateCheckout = async (orderId) => {
  // Create checkout session
  const response = await fetch('http://localhost:8080/api/v1/checkout', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ order_id: orderId })
  });

  const { session_id } = await response.json();

  // Redirect to Stripe Checkout
  const stripe = Stripe('pk_test_your_publishable_key_here');
  const { error } = await stripe.redirectToCheckout({
    sessionId: session_id
  });

  if (error) {
    console.error('Stripe error:', error);
  }
};
```

---

## Complete Checkout Flow Example

```javascript
// Step 1: User adds items to cart (frontend state)
const cart = [
  { productId: 'product-1', variantId: 'variant-1', quantity: 2 },
  { productId: 'product-2', variantId: 'variant-2', quantity: 1 }
];

// Step 2: Create order
async function checkout(customerEmail, cartItems) {
  try {
    // Create order
    const orderId = await createOrder(customerEmail, cartItems);
    console.log('Order created:', orderId);

    // Initiate Stripe checkout
    await initiateCheckout(orderId);
    
    // User is now redirected to Stripe...
    // After payment, Stripe redirects to: 
    // https://yourdomain.com/checkout/success?session_id=cs_test_...
    
  } catch (error) {
    console.error('Checkout failed:', error);
    alert('Checkout failed: ' + error.message);
  }
}

// Step 3: Handle success page
async function handleCheckoutSuccess() {
  const urlParams = new URLSearchParams(window.location.search);
  const sessionId = urlParams.get('session_id');
  
  if (!sessionId) {
    console.error('No session ID in URL');
    return;
  }

  // Order is now paid and being fulfilled
  // You can fetch order details to show confirmation
  
  console.log('Payment successful! Session:', sessionId);
  
  // Clear cart
  localStorage.removeItem('cart');
  
  // Show confirmation message
  document.getElementById('message').textContent = 
    'Thank you for your order! You will receive tracking info via email.';
}
```

---

## Frontend Integration Checklist

- [ ] Install Stripe.js in your frontend
  ```html
  <script src="https://js.stripe.com/v3/"></script>
  ```

- [ ] Set Stripe publishable key
  ```javascript
  const stripe = Stripe('pk_test_your_publishable_key');
  ```

- [ ] Update API base URL for production
  ```javascript
  const API_BASE = process.env.NODE_ENV === 'production' 
    ? 'https://api.nessieaudio.com/api/v1'
    : 'http://localhost:8080/api/v1';
  ```

- [ ] Handle CORS (backend already configured)

- [ ] Create success page at `/checkout/success`

- [ ] Create cancel page at `/checkout/cancel`

- [ ] Add error handling for all API calls

- [ ] Add loading states during checkout

- [ ] Clear cart after successful purchase

---

## CORS Configuration

Backend allows requests from origins specified in `.env`:
```env
ALLOWED_ORIGINS=http://localhost:3000,https://nessieaudio.com
```

Add your frontend URL to this list.

---

## Webhooks (Backend Only)

These endpoints are called by Stripe and Printful, not your frontend:

- `POST /webhooks/stripe` - Stripe payment events
- `POST /webhooks/printful` - Printful fulfillment events

Your frontend does not need to call these.

---

## Testing with cURL

```bash
# Get products
curl http://localhost:8080/api/v1/products

# Create order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_email": "test@example.com",
    "items": [
      {
        "product_id": "your-product-uuid",
        "variant_id": "your-variant-uuid",
        "quantity": 1
      }
    ]
  }'

# Get order
curl http://localhost:8080/api/v1/orders/{order-uuid}

# Create checkout
curl -X POST http://localhost:8080/api/v1/checkout \
  -H "Content-Type: application/json" \
  -d '{"order_id": "your-order-uuid"}'
```

---

## Rate Limiting

Currently no rate limiting. Consider adding in production.

## Caching

No caching implemented. Consider caching product listings.

## Pagination

Not implemented. All products returned in single response.
Add pagination if product count exceeds 100.
