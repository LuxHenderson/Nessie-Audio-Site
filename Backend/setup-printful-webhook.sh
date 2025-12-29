#!/bin/bash

# Printful Webhook Setup Script
# This script registers your webhook URL with Printful via their API

set -e

echo "🔧 Building webhook setup tool..."
go build -o ./bin/setup-webhook ./cmd/setup-webhook

echo ""
echo "🚀 Running webhook setup..."
./bin/setup-webhook

echo ""
echo "Done!"
