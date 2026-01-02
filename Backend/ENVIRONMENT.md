# Environment Configuration Guide

## Overview

The Nessie Audio backend uses **smart environment auto-detection** with environment-specific configuration files. No manual switching required!

## How It Works

The system automatically detects which environment you're in based on:

1. **ENV environment variable** (highest priority)
2. **Server hostname** (production/staging servers)
3. **Marker files** (`.production`, `.staging`)
4. **Default** (local development)

## Environment Files

### `.env.development` - Local Development
- Auto-loaded on your local machines
- Test Stripe keys (no real charges)
- Localhost URLs
- Verbose logging (`debug` level)
- HTTP allowed (no HTTPS redirect)

### `.env.staging` - Staging Server
- Auto-loaded on staging servers
- Test Stripe keys (safe testing)
- Staging domain URLs
- Full logging (`info` level)
- HTTPS enforced

### `.env.production` - Production Server
- Auto-loaded on production servers
- **LIVE Stripe keys** (real charges!)
- Production domain URLs
- Error-only logging (`error` level)
- HTTPS enforced
- Stricter security

### `.env.example` - Template
- Committed to Git (no secrets)
- Use as reference when setting up new machines

## Auto-Detection Logic

```
1. Check ENV variable → Use if set
2. Check hostname:
   - Contains "nessieaudio.com" or "production" → production
   - Contains "staging" or "stage-" → staging
3. Check for marker files:
   - .production file exists → production
   - .staging file exists → staging
4. Default → development (your local machine)
```

## Setup on Your Machines

### Initial Setup (One Time Per Machine)

1. Clone the repository
2. Copy environment files from your password manager or backup
3. Place them in the Backend directory:
   ```bash
   cd Backend/
   # Add your .env.development, .env.staging, .env.production files here
   ```

**That's it!** The system will automatically use `.env.development` on your local machines.

### Moving Between Machines

**Option 1: Keep env files in a separate folder**
```bash
# On Machine 1
cp .env.* ~/secure-config/nessie-backend/

# On Machine 2
cp ~/secure-config/nessie-backend/.env.* ./Backend/
```

**Option 2: Use a password manager (recommended)**
- Store all `.env.*` files as secure notes in 1Password/Bitwarden
- Copy/paste when setting up a new machine

**Option 3: Cloud sync (be careful!)**
- Store in encrypted cloud folder (not regular Dropbox!)
- Sync across machines
- Make sure it's encrypted at rest

## Testing Environment Detection

Run the test utility:
```bash
go run cmd/test-env/main.go
```

Expected output:
```
Environment auto-detected: development (default/local)
Loaded configuration from .env.development
Environment: development
Stripe Mode: TEST (safe)
```

## Manual Override (If Needed)

Force a specific environment by setting ENV:
```bash
# Force staging
export ENV=staging
./server

# Force production (be careful!)
export ENV=production
./server
```

## Hostname Detection Examples

These hostnames auto-detect as **production**:
- `nessieaudio.com`
- `www.nessieaudio.com`
- `prod-server-01`
- `production-backend`

These hostnames auto-detect as **staging**:
- `staging.nessieaudio.com`
- `stage-api-server`
- `stg-backend-01`

Everything else defaults to **development**.

## Security Best Practices

### ✅ DO:
- Keep `.env.development`, `.env.staging`, `.env.production` in `.gitignore`
- Store real credentials in a password manager
- Use different Printful webhook secrets per environment
- Review which Stripe keys are in which file

### ❌ DON'T:
- Commit `.env.*` files to Git (they're gitignored for a reason)
- Share production credentials in Slack/email
- Use production Stripe keys in development
- Store unencrypted env files in public cloud storage

## Environment Comparison

| Feature | Development | Staging | Production |
|---------|------------|---------|------------|
| Stripe Keys | Test | Test | **Live** |
| Database | `nessie_store.db` | `nessie_store_staging.db` | `nessie_store_production.db` |
| URLs | localhost | staging.nessieaudio.com | nessieaudio.com |
| HTTPS Redirect | ❌ Disabled | ✅ Enabled | ✅ Enabled |
| Logging | Verbose (debug) | Full (info) | Errors only (error) |
| Auto-detected on | Your local machines | Staging servers | Production servers |

## Troubleshooting

**Wrong environment detected?**
```bash
# Check what's being detected
go run cmd/test-env/main.go

# Force the correct environment
export ENV=development  # or staging, or production
```

**Can't find .env file?**
```bash
# Check which file it's trying to load
ls -la .env.*

# Create from example
cp .env.example .env.development
# Then edit with your real values
```

**Accidentally using production keys in development?**
```bash
# Check your Stripe keys
grep STRIPE_SECRET_KEY .env.development

# Should start with sk_test_ (not sk_live_!)
```

## Files

- **`.env.development`** - Development config (gitignored)
- **`.env.staging`** - Staging config (gitignored)
- **`.env.production`** - Production config (gitignored)
- **`.env.example`** - Template (committed to Git)
- **`internal/config/config.go`** - Auto-detection logic
- **`cmd/test-env/main.go`** - Environment detection test

## Questions?

- Check which environment is active: `go run cmd/test-env/main.go`
- Force an environment: `export ENV=development`
- See all config values: Check server startup logs
