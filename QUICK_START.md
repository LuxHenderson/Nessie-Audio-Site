# Quick Start Guide - Nessie Audio Merch

## 🚀 Starting Your Site (Both Machines)

### 1. Start Backend
```bash
cd ~/Desktop/Coding/Nessie-Audio-Site/Backend
bash ./start-server.sh
```
✅ Server runs on port 8080

### 2. Start Frontend
- Open `merch.html` in VS Code
- Click "Go Live" or right-click → "Open with Live Server"

## 📦 What's Already Set Up

✅ Database with 6 products, 23 variants  
✅ All product descriptions (custom formatting)  
✅ Local product images (served via backend)  
✅ Price ranges (min-max display)  
✅ Variant sorting (smallest to largest)  
✅ CORS enabled (allows all origins)

## 🔄 One-Time Setup on MacBook

1. **Copy `.env` file**
   ```bash
   # From this machine:
   cd ~/Desktop/Coding/Nessie-Audio-Site/Backend
   # Copy the .env file to MacBook (USB, cloud, etc.)
   ```

2. **Sync products on MacBook**
   ```bash
   cd ~/Desktop/Coding/Nessie-Audio-Site/Backend
   go run cmd/sync-products/main.go
   ```

That's it! Everything else is in git.

## ✅ Current Status

**Database:** 6 products, 23 variants ✓  
**Backend:** Running on port 8080 ✓  
**Images:** Served from Product Photos/ ✓  
**Descriptions:** All 6 products have custom descriptions ✓  
**API:** Returns min/max price ranges ✓

## 🔍 Quick Test

```bash
# Test API
curl http://localhost:8080/api/v1/products | head -20

# Test image serving
curl -I "http://localhost:8080/Product%20Photos/Nessie%20Audio%20Unisex%20t-shirt/unisex-staple-t-shirt-black-back-6947058beaf9f.jpg"
```

Both should return 200 OK.
