#!/bin/bash

# Printful Retry Worker - runs every 15 minutes
# Retries failed Printful order submissions

while true; do
    echo "[Retry Worker] Running Printful retry job..."
    go run cmd/retry-printful/main.go 2>&1 | sed 's/^/[Retry] /'
    echo "[Retry Worker] Next run in 15 minutes..."
    sleep 900  # 15 minutes
done
