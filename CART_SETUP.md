# Shopping Cart & Checkout Setup Guide

## ✅ Completed Implementation

The shopping cart and Stripe checkout system has been fully implemented! Here's what's been added:

### Frontend Components

1. **Cart Management (`cart.js`)**
   - ShoppingCart class with localStorage persistence
   - Add, remove, update quantity functions
   - Cart count badge in navigation
   - "Added to cart" notification popup

2. **Cart Page (`cart.html` + `cart-page.js`)**
   - Display cart items with images, names, variants, prices
   - Quantity controls (increment/decrement)
   - Remove item buttons
   - Order summary sidebar (subtotal, shipping, total)
   - "Proceed to Checkout" button

3. **Success/Cancel Pages**
   - `cart-success.html` - Displayed after successful payment
   - `cart-cancel.html` - Displayed if user cancels checkout
   - Cart is cleared on success page

4. **Product Integration**
   - "Add to Cart" button on product detail pages
   - "Buy Now" button adds to cart and redirects to cart page
   - Selected variant and quantity are added to cart

### Backend Components

1. **Cart Checkout Endpoint**
   - `POST /api/v1/cart/checkout`
   - Accepts cart items and email
   - Validates products and variants
   - Creates Stripe Checkout Session
   - Returns session_id for redirect

2. **Config Endpoint**
   - `GET /api/v1/config`
   - Returns Stripe publishable key for frontend

### Styling

- Cart page fully styled to match existing site design
- Responsive layout for mobile and desktop
- Success/cancel pages styled with branded colors
- Cart notification popups with animations

## 🔧 Required Configuration

Before the checkout system will work, you need to set your **Stripe Publishable Key** in the backend `.env` file:

### Step 1: Get Your Stripe Keys

1. Go to https://dashboard.stripe.com/apikeys
2. Copy your **Publishable key** (starts with `pk_test_` or `pk_live_`)
3. Your **Secret key** is already configured (`sk_live_...`)

### Step 2: Update .env File

Edit `Backend/.env` and replace the placeholder:

```env
STRIPE_PUBLISHABLE_KEY=pk_test_your_actual_key_here
```

**Important:** The secret key should use `sk_test_` for testing (not `sk_live_`).

### Step 3: Restart Backend Server

```bash
cd Backend
bash ./start-server.sh
```

## 🧪 Testing the Cart System

### Test Flow:

1. **Start the backend server** (if not already running):
   ```bash
   cd Backend
   bash ./start-server.sh
   ```

2. **Open the site** with Live Server on port 5500:
   - Open `merch.html` or any page in VS Code
   - Right-click → "Open with Live Server"

3. **Add items to cart**:
   - Click on any product
   - Select a size/variant
   - Click "Add to Cart" or "Buy Now"
   - Check the cart icon - count should increase

4. **View cart**:
   - Click the cart icon (🛒) in navigation
   - See all items with correct prices and variants
   - Test quantity controls
   - Test remove button

5. **Test checkout** (requires valid Stripe keys):
   - Click "Proceed to Checkout"
   - Enter your email when prompted
   - Should redirect to Stripe Checkout page
   - Complete or cancel the payment

6. **Success/Cancel flows**:
   - Success: Redirects to `cart-success.html`, cart is cleared
   - Cancel: Redirects to `cart-cancel.html`, items remain in cart

## 📋 Cart Features

- ✅ Persistent cart (localStorage)
- ✅ Cart count badge in navigation
- ✅ Add to cart notifications
- ✅ Quantity controls
- ✅ Remove items
- ✅ Price calculations
- ✅ Variant selection
- ✅ Stripe Checkout integration
- ✅ Success/cancel handling
- ✅ Responsive design
- ✅ Mobile-friendly

## 🔍 Troubleshooting

### Cart not working?
- Check browser console for errors
- Verify `cart.js` is loaded before `cart-page.js` and `product-detail.js`
- Check localStorage: Open DevTools → Application → Local Storage → `nessie_audio_cart`

### Checkout button not working?
- Verify Stripe publishable key is set in `Backend/.env`
- Check backend console for errors
- Verify backend is running on port 8080
- Check browser console for Stripe initialization errors

### Items not adding to cart?
- Check product-detail.js is loaded after cart.js
- Verify variant is selected
- Check browser console for errors

### Wrong URLs after checkout?
- Verify these settings in `Backend/.env`:
  ```
  STRIPE_SUCCESS_URL=http://localhost:5500/cart-success.html
  STRIPE_CANCEL_URL=http://localhost:5500/cart-cancel.html
  ```

## 📝 Files Modified/Created

### New Files:
- `cart.js` - Cart state management
- `cart.html` - Cart page HTML
- `cart-page.js` - Cart page functionality
- `cart-success.html` - Success page
- `cart-cancel.html` - Cancel page
- `Backend/internal/handlers/checkout.go` - Added `CreateCartCheckout` handler
- `CART_SETUP.md` - This file

### Modified Files:
- `merch.html` - Added cart icon to navigation
- `product-detail.html` - Added cart icon, cart.js script
- `product-detail.js` - Updated `handleAddToCart()` and `handleBuyNow()`
- `style.css` - Added cart styles, success/cancel styles
- `Backend/internal/handlers/handlers.go` - Added `/cart/checkout` and `/config` routes
- `Backend/cmd/server/main.go` - Added endpoint logging
- `Backend/.env` - Updated success/cancel URLs

## 🚀 Next Steps

1. Set your Stripe publishable key in `.env`
2. Test the full checkout flow with Stripe test mode
3. When ready for production:
   - Switch to live Stripe keys
   - Update URLs to production domain
   - Test thoroughly with real payment methods

## 💡 Notes

- Cart data is stored in browser localStorage (persists across sessions)
- Cart is cleared after successful checkout
- Shipping is set to "TBD" (can be updated later)
- Email is collected at checkout time
- Stripe handles payment processing securely
- No credit card data is stored on your server
