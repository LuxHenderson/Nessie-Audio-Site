# Nessie Audio Merch Page Documentation

## Overview
This document provides comprehensive guidance for the Nessie Audio Merch page framework, including structure, styling, and backend integration points for Golang eCommerce functionality.

## File Structure

```
/Nessie-Audio-Site/
├── merch.html          # Main merch page markup
├── style.css           # Global styles (includes merch-specific styles at bottom)
├── merch.js            # JavaScript for dynamic product loading and cart functionality
└── MERCH_README.md     # This documentation file
```

## Page Architecture

### HTML Structure (merch.html)

The page follows a semantic HTML5 structure:

```html
<main>
  <section class="merch-page">
    <h1>- Merch</h1>
    <p class="merch-intro">Intro blurb</p>
    
    <div class="merch-grid">
      <article class="merch-item">
        <div class="merch-image-container">
          <img class="merch-image" />
        </div>
        <div class="merch-details">
          <h3 class="merch-title">Product Name</h3>
          <p class="merch-description">Description</p>
          <p class="merch-price">$XX.XX</p>
          <button class="merch-buy-btn" data-product-id="XXX">Buy Now</button>
        </div>
      </article>
    </div>
  </section>
</main>
```

### Key HTML Classes

| Class | Purpose |
|-------|---------|
| `.merch-page` | Main container for merch section |
| `.merch-intro` | Intro blurb styling |
| `.merch-grid` | CSS Grid container for product cards |
| `.merch-item` | Individual product card |
| `.merch-image-container` | Fixed-height image wrapper |
| `.merch-image` | Product image with hover effects |
| `.merch-details` | Product information container |
| `.merch-title` | Product name/title |
| `.merch-description` | Product description text |
| `.merch-price` | Product price display |
| `.merch-buy-btn` | Call-to-action button |

## CSS Styling (style.css)

All merch-specific styles are located at the bottom of `style.css` under the "MERCH PAGE STYLES" section.

### Responsive Grid System

The product grid uses CSS Grid with `auto-fill` to automatically adjust columns:

```css
.merch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 2rem;
}
```

**Breakpoints:**
- **Desktop (>900px):** Multi-column grid (300px min column width)
- **Tablet (600px-900px):** 2-3 columns (250px min column width)
- **Mobile (<600px):** Single column layout

### Interactive Features

1. **Product Card Hover:**
   - Card lifts up (`translateY(-8px)`)
   - Shadow increases
   - Image zooms slightly (`scale(1.05)`)

2. **Buy Button Hover:**
   - Background brightens
   - Border becomes more visible
   - Button scales up (`scale(1.05)`)
   - Glowing shadow effect

3. **Buy Button Click Feedback:**
   - Scale down effect (`scale(0.98)`)
   - Text changes to "Added!"
   - Background color changes temporarily

## JavaScript Functionality (merch.js)

### Configuration

Update these constants for your backend:

```javascript
const API_BASE_URL = 'http://localhost:8080/api';
const PRODUCTS_ENDPOINT = `${API_BASE_URL}/products`;
const CART_ENDPOINT = `${API_BASE_URL}/cart`;
```

### Key Functions

#### 1. `fetchProducts()`
Fetches product data from Golang backend API.

**Expected Response Format:**
```json
[
  {
    "id": "001",
    "name": "Product Name",
    "description": "Product description",
    "price": 29.99,
    "imageUrl": "https://example.com/image.jpg",
    "category": "apparel",
    "stock": 50,
    "featured": false
  }
]
```

#### 2. `renderProducts(products)`
Dynamically generates and renders product cards to the DOM.

#### 3. `addToCart(productId)`
Handles adding products to shopping cart via backend API.

**Expected Request Format:**
```json
{
  "productId": "001",
  "quantity": 1
}
```

#### 4. `showNotification(message, type)`
Displays success/error notifications to users.

### Current Implementation

The page currently uses **static HTML** with placeholder products. To switch to **dynamic loading**:

1. Uncomment the API call in `fetchProducts()`
2. Uncomment the dynamic rendering code in `initMerchPage()`
3. Comment out or remove static HTML products in `merch.html`

## Backend Integration Guide

### Step 1: Golang API Endpoints

Create the following REST API endpoints:

#### GET `/api/products`
Returns all available products.

**Response:**
```json
{
  "products": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "price": float64,
      "imageUrl": "string",
      "category": "string",
      "stock": int,
      "featured": bool
    }
  ]
}
```

#### POST `/api/cart`
Adds a product to the shopping cart.

**Request Body:**
```json
{
  "productId": "string",
  "quantity": int
}
```

**Response:**
```json
{
  "success": true,
  "cartId": "string",
  "message": "Product added to cart"
}
```

#### GET `/api/cart/:cartId`
Retrieves cart contents.

### Step 2: Database Schema

Suggested PostgreSQL schema:

```sql
CREATE TABLE products (
  id VARCHAR(50) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10, 2) NOT NULL,
  image_url VARCHAR(500),
  category VARCHAR(100),
  stock INT DEFAULT 0,
  featured BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE carts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cart_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cart_id UUID REFERENCES carts(id),
  product_id VARCHAR(50) REFERENCES products(id),
  quantity INT DEFAULT 1,
  added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Step 3: CORS Configuration

Enable CORS in your Golang server:

```go
import "github.com/rs/cors"

func main() {
    // ... your router setup ...
    
    // CORS middleware
    c := cors.New(cors.Options{
        AllowedOrigins: []string{"http://localhost:3000", "https://nessieaudio.com"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type", "Authorization"},
    })
    
    handler := c.Handler(yourRouter)
    // ... start server ...
}
```

### Step 4: Image Hosting

Options for product images:
1. **AWS S3:** Reliable, scalable object storage
2. **Cloudinary:** Image CDN with transformation capabilities
3. **Self-hosted:** Store in `/static/images/products/` directory

Update `imageUrl` in product data accordingly.

### Step 5: Payment Integration

Consider integrating:
- **Stripe:** `stripe-go` library
- **PayPal:** PayPal SDK
- **Square:** Square Go SDK

Add checkout endpoint:
```
POST /api/checkout
```

## Adding New Products

### Method 1: Static HTML (Current)

1. Open `merch.html`
2. Locate the `.merch-grid` div
3. Copy an existing `<article class="merch-item">` block
4. Update:
   - `img src` with new image URL
   - `alt` text
   - `.merch-title` text
   - `.merch-description` text
   - `.merch-price` value
   - `data-product-id` attribute
5. Save file

### Method 2: Dynamic Loading (After Backend Integration)

1. Add product to database via admin panel or SQL:
   ```sql
   INSERT INTO products (id, name, description, price, image_url, category, stock)
   VALUES ('007', 'New Product', 'Description', 39.99, 'url', 'apparel', 100);
   ```
2. Product will automatically appear on page load

## Responsive Design

The page is optimized for desktop viewing with mobile compatibility:

- **Desktop (>900px):** 3-4 column grid, full product details
- **Tablet (600-900px):** 2-3 column grid, condensed spacing
- **Mobile (<600px):** Single column, optimized touch targets

### Testing Responsive Layout

1. Open `merch.html` in browser
2. Open DevTools (F12)
3. Toggle Device Toolbar (Ctrl+Shift+M)
4. Test various screen sizes

## Accessibility Features

- Semantic HTML5 elements (`<article>`, `<section>`)
- ARIA labels on buttons
- Focus states for keyboard navigation
- Alt text on all images
- Sufficient color contrast ratios

## Performance Optimization

### Current Optimizations
- CSS Grid (no JS required for layout)
- Lazy loading ready (add `loading="lazy"` to images)
- Minimal JavaScript execution on page load
- CSS transitions (GPU-accelerated)

### Future Optimizations
- Image CDN implementation
- Product data caching
- Lazy load images below the fold
- Service Worker for offline support

## Browser Compatibility

Tested and compatible with:
- Chrome/Edge (v90+)
- Firefox (v88+)
- Safari (v14+)
- Mobile Safari (iOS 14+)
- Chrome Mobile (Android 10+)

## Customization Guide

### Change Grid Layout

Adjust in `style.css`:
```css
.merch-grid {
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); /* Change 350px */
  gap: 3rem; /* Change spacing */
}
```

### Adjust Color Scheme

Colors are controlled by CSS variables in `style.css`:
```css
:root {
  --bg: #ffffff25;
  --panel: #ffffff25;
  --text: #ffffff;
  --accent: #efefef25;
}
```

### Modify Hover Effects

Edit transitions in `.merch-item:hover` and `.merch-buy-btn:hover`.

## Testing Checklist

- [ ] All product images load correctly
- [ ] Hover effects work on desktop
- [ ] Buy buttons are clickable
- [ ] Responsive layout works on mobile
- [ ] Notifications appear when clicking Buy
- [ ] Page is accessible via keyboard
- [ ] No console errors
- [ ] Integrates with existing site navigation
- [ ] Fog effect renders properly

## Troubleshooting

### Products not displaying
- Check console for JavaScript errors
- Verify `.merch-grid` exists in HTML
- Check if `merch.js` is loading

### Buy buttons not responding
- Verify `data-product-id` attributes are set
- Check JavaScript console for errors
- Ensure `merch.js` is loaded after DOM

### Styling issues
- Clear browser cache
- Check if `style.css` contains merch styles
- Verify class names match between HTML and CSS

## Next Steps

1. **Set up Golang backend:**
   - Create REST API endpoints
   - Set up database
   - Implement CORS

2. **Integrate payment processing:**
   - Choose payment provider (Stripe recommended)
   - Add checkout flow
   - Implement order confirmation

3. **Add admin panel:**
   - Product management interface
   - Inventory tracking
   - Order management

4. **Enhance features:**
   - Product search/filtering
   - Shopping cart page
   - Wishlist functionality
   - Product reviews

5. **Deploy:**
   - Set up production server
   - Configure environment variables
   - Enable HTTPS
   - Set up CDN for images

## Support & Questions

For questions or issues, contact the development team or refer to:
- Golang documentation: https://golang.org/doc/
- Stripe API docs: https://stripe.com/docs/api
- MDN Web Docs: https://developer.mozilla.org/

---

**Last Updated:** December 19, 2025  
**Version:** 1.0.0  
**Maintainer:** Nessie Audio Development Team
