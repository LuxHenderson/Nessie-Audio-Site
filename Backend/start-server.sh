#!/bin/bash
cd "$(dirname "$0")"
export PATH="/opt/homebrew/bin:$PATH"
echo "Starting Nessie Audio Backend Server..."
go run cmd/server/main.go
