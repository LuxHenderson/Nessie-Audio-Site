#!/bin/bash

# Nessie Audio Development Server Startup
# Starts both the backend server and Stripe webhook forwarding

set -e

echo "🚀 Starting Nessie Audio Development Environment..."
echo ""

# Check if stripe CLI is installed
if ! command -v stripe &> /dev/null; then
    echo "❌ Stripe CLI not found. Please install it:"
    echo "   brew install stripe/stripe-cli/stripe"
    exit 1
fi

# Check if logged into Stripe
if ! stripe config --list &> /dev/null; then
    echo "⚠️  You need to login to Stripe CLI first:"
    echo "   stripe login"
    exit 1
fi

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}✓${NC} Stripe CLI detected"
echo ""

# Trap to cleanup background processes on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down development environment..."
    kill $(jobs -p) 2>/dev/null || true
    wait 2>/dev/null || true
    echo "✓ Stopped all services"
}
trap cleanup EXIT INT TERM

# Start Stripe webhook forwarding in background
echo -e "${BLUE}[Stripe Webhooks]${NC} Starting webhook forwarding..."
stripe listen --forward-to localhost:8080/webhooks/stripe 2>&1 | sed 's/^/[Stripe] /' &
STRIPE_PID=$!

# Start Printful retry worker in background
echo -e "${BLUE}[Retry Worker]${NC} Starting Printful retry worker..."
./start-retry-worker.sh &
RETRY_PID=$!

# Wait a moment for Stripe to start
sleep 2

# Sync products and Printful IDs
echo -e "${BLUE}[Database Sync]${NC} Syncing products from Printful..."
go run cmd/sync-products/main.go 2>&1 | grep -E "✓|✅|Total|Found" | sed 's/^/[Sync] /'

echo -e "${BLUE}[Database Sync]${NC} Syncing Printful variant IDs..."
go run cmd/sync-printful-ids/main.go 2>&1 | grep -E "✓|✅|Sync complete|Updated" | sed 's/^/[Sync] /'

echo ""

# Start the Go server
echo -e "${BLUE}[Backend Server]${NC} Starting Go server..."
echo ""
go run cmd/server/main.go 2>&1 | sed 's/^/[Server] /' &
SERVER_PID=$!

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Development environment running${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Services:"
echo "  • Backend Server:    http://localhost:8080"
echo "  • Stripe Webhooks:   Forwarding to /webhooks/stripe"
echo "  • Retry Worker:      Running every 15 minutes"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Wait for all background processes
wait
