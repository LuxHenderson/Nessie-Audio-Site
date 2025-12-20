<!-- Workspace instructions for Nessie Audio eCommerce Backend -->

## Project Overview
Production-ready Golang eCommerce backend with Printful fulfillment and Stripe payments.

## Quick Start

```bash
# Install Go (if not installed)
brew install go  # macOS
# or download from https://go.dev/dl/

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env with your API keys
nano .env

# Run the server
go run cmd/server/main.go
```

## Required Credentials

Before running:
1. **Printful API Key**: Get from https://www.printful.com/dashboard/api
2. **Stripe Keys**: Get from https://dashboard.stripe.com/apikeys
3. **Copy `.env.example` to `.env`**: `cp .env.example .env`
4. **Add your real keys to `.env`** (never commit this file!)

**Security Note**: The `.env` file contains real secrets and is gitignored. Only `.env.example` with placeholder values should be committed to the repository.

## Project Completed

All components have been scaffolded:
- ✅ Project structure
- ✅ Configuration management
- ✅ Data models
- ✅ Printful API client
- ✅ Stripe integration
- ✅ Order management service
- ✅ Database layer (SQLite)
- ✅ API handlers (products, orders, checkout)
- ✅ Webhook handlers (Stripe & Printful)
- ✅ Middleware (CORS, logging, recovery)
- ✅ Main server entry point
- ✅ Comprehensive documentation

## Next Steps

1. **Install Go** (if needed)
2. **Add API credentials** to `.env`
3. **Add products** to database (see README)
4. **Run server**: `go run cmd/server/main.go`
5. **Test API**: `curl http://localhost:8080/health`
6. **Integrate with frontend** (see API_DOCS.md)

## Documentation

- `README.md` - Complete setup and usage guide
- `API_DOCS.md` - Frontend integration contract
- `.env.example` - Environment configuration template

## Architecture Highlights

- **Clean separation**: handlers → services → database
- **Webhook security**: Signature verification for Stripe & Printful
- **Payment flow**: Only submits to Printful after confirmed payment
- **Audit logs**: All webhook events stored for debugging
- **Production-ready**: Graceful shutdown, error handling, CORS

## Frontend Integration

The backend exposes REST APIs for:
- Listing products and variants
- Creating orders
- Initiating Stripe checkout
- Retrieving order status

See `API_DOCS.md` for complete request/response examples.
