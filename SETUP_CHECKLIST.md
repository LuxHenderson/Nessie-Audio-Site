# Nessie Audio Site - Setup Checklist

This checklist ensures your merch page works consistently across both machines.

## ✅ Files That Must Be Present on Both Machines

### Backend Files
1. **`.env` file** in `Backend/` directory
   - Location: `/Backend/.env`
   - Contains: API keys for Printful and Stripe
   - Status: ✅ Present (hidden file, not in git)
   - **Action Required**: Copy this file to your MacBook's Backend folder

2. **`nessie_store.db`** in `Backend/` directory
   - Location: `/Backend/nessie_store.db`
   - Contains: Product data, descriptions, and variants
   - Status: ✅ Present
   - **Action Required**: Either sync products or copy database to MacBook

3. **Product Photos** directory
   - Location: `/Product Photos/` (root of project)
   - Contains: All product images
   - Status: ✅ Present
   - **Action Required**: Ensure this folder exists on MacBook

### Frontend Files
All frontend files (HTML, CSS, JS) are tracked in git and should be consistent.

## 🚀 Startup Process (Same on Both Machines)

### Step 1: Start Backend Server
```bash
cd /path/to/Nessie-Audio-Site/Backend
bash ./start-server.sh
```

**Expected Output:**
```
Starting Nessie Audio Backend Server...
Server listening on port 8080
Serving product images from local Product Photos directory
```

### Step 2: Start Frontend (Live Server in VS Code)
1. Open `merch.html` in VS Code
2. Right-click and select "Open with Live Server"
3. Or use the "Go Live" button in VS Code status bar

**Expected Result:** Page opens at `http://127.0.0.1:5500/merch.html` (or similar)

## 🔄 Initial Setup on New Machine (MacBook)

Run these commands once on the MacBook:

### 1. Copy Environment Variables
```bash
# Copy .env file from this machine to MacBook
# (You'll need to manually transfer this file)
```

### 2. Sync Products from Printful
```bash
cd Backend
go run cmd/sync-products/main.go
```

**Expected Output:**
```
Found 6 products from Printful
✓ Added product: Nessie Audio Unisex t-shirt
✓ Added product: Nessie Audio Unisex Champion hoodie
✓ Added product: Nessie Audio Black Glossy Mug
✓ Added product: Hardcover bound Nessie Audio notebook
✓ Added product: Nessie Audio Eco Tote Bag
✓ Added product: Nessie Audio Bubble-free stickers
✅ Sync complete!
```

## 🧪 Verification Checklist

Run these tests on both machines to ensure consistency:

### Backend Tests
```bash
# 1. Check server is running
curl http://localhost:8080/health

# 2. Check products API
curl http://localhost:8080/api/v1/products | python3 -m json.tool | head -20

# 3. Check images are served
curl -I "http://localhost:8080/Product%20Photos/Nessie%20Audio%20Unisex%20t-shirt/unisex-staple-t-shirt-black-back-6947058beaf9f.jpg"
```

**Expected Results:**
1. Health endpoint returns 200 OK
2. Products API returns 6 products with descriptions
3. Image returns 200 OK with Content-Type: image/jpeg or image/png

### Frontend Tests
Open in browser and verify:
- [ ] All 6 products display on merch page
- [ ] Product images load correctly
- [ ] Prices show ranges (e.g., "$15.00 - $18.00" for mug)
- [ ] Clicking product opens detail page
- [ ] Product descriptions appear in "About this product" box
- [ ] Bullet points are properly formatted
- [ ] Size dropdown is in correct order (smallest to largest)
- [ ] No descriptions appear on merch listing cards

## 📋 What's Permanent (Won't Need Re-setup)

✅ **Product Descriptions** - Stored in sync script and database  
✅ **Product Images** - Served from local Product Photos directory  
✅ **Price Ranges** - Automatically calculated from variants  
✅ **Size Sorting** - Built into frontend JavaScript  
✅ **Static File Serving** - Backend configured to serve Product Photos  

## ⚠️ Known Configuration Details

### Hardcoded Localhost References
These are intentional and correct for local development:
- `merch.js`: `const API_BASE_URL = 'http://localhost:8080/api/v1'`
- `product-detail.js`: `const API_BASE_URL = 'http://localhost:8080/api/v1'`
- Product images: `http://localhost:8080/Product Photos/...`

### CORS Configuration
Backend allows all origins (`ALLOWED_ORIGINS=*` in .env)

## 🔧 Troubleshooting

### Issue: "No products available"
**Solution:** Run `go run cmd/sync-products/main.go` in Backend directory

### Issue: Images not loading
**Solutions:**
1. Check Product Photos directory exists
2. Restart backend server
3. Verify server shows "Serving product images from local Product Photos directory"

### Issue: Port 8080 already in use
**Solution:** `lsof -ti:8080 | xargs kill -9`

### Issue: Missing descriptions
**Solution:** Product descriptions are in sync script. Re-sync products.

### Issue: .env file missing
**Solution:** Copy .env file from other machine (it's gitignored for security)

## 📝 Daily Development Workflow

**Every time you start working:**
1. `cd Backend && bash ./start-server.sh`
2. Open merch.html with Live Server in VS Code
3. That's it! Everything else is persistent.

**No need to:**
- ❌ Re-sync products (unless adding new ones to Printful)
- ❌ Update descriptions (they're permanent)
- ❌ Fix image paths (they're correct)
- ❌ Reorder variants (handled automatically)
