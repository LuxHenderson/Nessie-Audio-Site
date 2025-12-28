# Environment Configuration Guide

This document explains how environment detection works in your Nessie Audio website.

## Overview

Your site now automatically detects whether it's running in **development** or **production** mode and configures itself accordingly.

---

## Frontend (JavaScript)

### How It Works

The [config.js](config.js) file automatically detects the environment:

- **Development**: Uses `http://localhost:8080/api/v1`
- **Production**: Uses your live domain's API endpoint

### Detection Logic

```javascript
if (window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1') {
  // Development mode
} else {
  // Production mode
}
```

### What You Need to Do

**Nothing!** It works automatically.

When you deploy to production, the frontend will automatically use your production domain.

---

## Backend (Go Server)

### How It Works

The backend uses the `ENV` variable in your `.env` file:

- `ENV=development` → Development mode
- `ENV=production` → Production mode

### Current Configuration (Development)

**File: `Backend/.env`**

```env
ENV=development

# Stripe Test Keys (Safe for development)
STRIPE_SECRET_KEY=sk_test_YOUR_STRIPE_TEST_SECRET_KEY
STRIPE_PUBLISHABLE_KEY=pk_test_YOUR_STRIPE_TEST_PUBLISHABLE_KEY

# URLs are auto-detected (leave blank)
STRIPE_SUCCESS_URL=
STRIPE_CANCEL_URL=

# Production domain (optional, for when ENV=production)
PRODUCTION_DOMAIN=
```

### Auto-Detection Logic

**Development Mode (`ENV=development`):**
- Stripe Success URL: `http://localhost:5500/cart-success.html`
- Stripe Cancel URL: `http://localhost:5500/cart-cancel.html`

**Production Mode (`ENV=production`):**
- Stripe Success URL: `https://[PRODUCTION_DOMAIN]/cart-success.html`
- Stripe Cancel URL: `https://[PRODUCTION_DOMAIN]/cart-cancel.html`

---

## When You Deploy to Production

### Step 1: Update Backend .env File

```env
ENV=production
PRODUCTION_DOMAIN=nessieaudio.com

# Switch to LIVE Stripe keys
STRIPE_SECRET_KEY=sk_live_YOUR_STRIPE_LIVE_SECRET_KEY
STRIPE_PUBLISHABLE_KEY=pk_live_YOUR_STRIPE_LIVE_PUBLISHABLE_KEY
```

### Step 2: Deploy Your Code

No code changes needed! The configuration automatically adapts.

### Step 3: Verify

Check your server logs for:
```
Stripe Success URL: https://nessieaudio.com/cart-success.html
Stripe Cancel URL: https://nessieaudio.com/cart-cancel.html
```

---

## Stripe Keys Reference

### Test Keys (Development)
- **Publishable**: `pk_test_YOUR_STRIPE_TEST_PUBLISHABLE_KEY`
- **Secret**: `sk_test_YOUR_STRIPE_TEST_SECRET_KEY`
- **Use these for**: Local development, testing checkout flow
- **No real charges**: All payments are simulated

### Live Keys (Production)
- **Publishable**: `pk_live_YOUR_STRIPE_LIVE_PUBLISHABLE_KEY`
- **Secret**: `sk_live_YOUR_STRIPE_LIVE_SECRET_KEY`
- **Use these for**: Production deployment only
- **Real charges**: Customers will be charged real money

**Note**: Keep your actual API keys in a secure private location. They are stored in your `.env` file which is protected by `.gitignore`.

---

## Testing Locally

### Frontend Testing
1. Open your site at `http://localhost:5500` (or whatever port you use)
2. Open browser console (F12)
3. Look for: `API Configuration loaded: http://localhost:8080/api/v1`

### Backend Testing
1. Start the server: `cd Backend && ./start-server.sh`
2. Check logs for:
   ```
   Stripe Success URL: http://localhost:5500/cart-success.html
   Stripe Cancel URL: http://localhost:5500/cart-cancel.html
   ```

---

## Files Modified

### Frontend
- ✅ [config.js](config.js) - New centralized config
- ✅ [merch.js](merch.js) - Uses API_CONFIG
- ✅ [cart-page.js](cart-page.js) - Uses API_CONFIG
- ✅ [product-detail.js](product-detail.js) - Uses API_CONFIG
- ✅ [merch.html](merch.html) - Loads config.js
- ✅ [cart.html](cart.html) - Loads config.js
- ✅ [product-detail.html](product-detail.html) - Loads config.js

### Backend
- ✅ [Backend/internal/config/config.go](Backend/internal/config/config.go) - Auto-detection logic
- ✅ [Backend/.env](Backend/.env) - Updated with test keys

---

## Troubleshooting

### Issue: Frontend shows CORS errors
**Solution**: Make sure backend `ALLOWED_ORIGINS` includes your frontend URL

### Issue: Stripe checkout fails
**Solution**: Verify you're using the correct keys for your environment

### Issue: Wrong redirect URL
**Solution**: Check `ENV` variable in `.env` file matches your environment

---

## Security Notes

- **Never commit** `.env` file to git (it's in `.gitignore`)
- **Test keys** are safe to use in development
- **Live keys** should only be used in production
- Always use HTTPS in production
