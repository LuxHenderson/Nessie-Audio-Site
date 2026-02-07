# 🚀 Ready to Deploy - Your Merch Page Fix

## ✅ What I've Done For You

I've implemented an **automatic solution** that will fix your merch page when you deploy. Here's what was changed:

### 1. **Auto-Seeding Database** ✨
- Added automatic product seeding on server startup
- If the database is empty, it will automatically populate with all 6 products
- No manual SQL commands needed!

### 2. **Fixed Image URLs** 🖼️
- Changed all product images from `http://localhost:8080` to relative paths
- Images will now work in production

### 3. **Improved Environment Detection** 🎯
- Server now auto-detects Railway as production environment
- No need to manually set `ENV=production` (though you still can)

## 🎬 How to Deploy

You have **2 options**:

### Option A: Let Railway Auto-Deploy (Easiest)
If your Railway project is connected to a Git repository:
1. Commit these changes:
   ```bash
   git add .
   git commit -m "Fix: Auto-seed products and improve production detection"
   git push
   ```
2. Railway will automatically deploy
3. Wait ~2 minutes for deployment
4. Visit https://nessieaudio.com/merch - products should appear!

### Option B: Manual Deploy via Railway Dashboard
If not using Git auto-deploy:
1. Go to https://railway.app
2. Click your project → Backend service
3. Click **Deploy** → **Redeploy** (or trigger new deployment)
4. Wait for deployment to complete
5. Check logs to see: `✓ Seeded: [product names]`

## 📋 What Will Happen

When your server starts in production, you'll see these logs:
```
Environment auto-detected: production (Railway detected)
Database initialized
🔄 Checking for database migrations...
✅ Database schema is up to date
📦 Products table is empty - seeding with initial products...
  ✓ Seeded: Nessie Audio Eco Tote Bag
  ✓ Seeded: Hardcover bound Nessie Audio notebook
  ✓ Seeded: Nessie Audio Bubble-free stickers
  ✓ Seeded: Nessie Audio Black Glossy Mug
  ✓ Seeded: Nessie Audio Unisex Champion hoodie
  ✓ Seeded: Nessie Audio Unisex t-shirt
✅ Product seeding complete!
```

## 🧪 Verify It Works

After deployment:

1. **Check API directly:**
   - Visit: https://nessieaudio.com/api/v1/products
   - Should return JSON with 6 products

2. **Check Merch page:**
   - Visit: https://nessieaudio.com/merch
   - All 6 products should display with images

3. **Check logs in Railway:**
   - Look for "Product seeding complete!" message

## 🔧 Files Changed

- `Backend/cmd/server/main.go` - Added auto-seeding function
- `Backend/internal/config/config.go` - Improved Railway detection
- `Backend/cmd/sync-products/main.go` - Fixed image URLs
- `Backend/nessie_store.db` - Updated local database image paths

## 🆘 If Something Goes Wrong

If products still don't show after deployment:

1. Check Railway logs for errors
2. Verify environment is detected as "production" (not "development")
3. Check the API response at `/api/v1/products`
4. Look for the "Seeded:" messages in startup logs

The seeding is **idempotent** - it only runs if the database is empty, so it's safe to redeploy multiple times.

---

**That's it!** Just deploy and your merch page will work. 🎉
