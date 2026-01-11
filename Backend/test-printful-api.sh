#!/bin/bash

# Test Printful API directly
API_KEY="V6QIxMywq21i0z2CYpwiqCbwLaeUiLNlBOJzAb5Y"
API_URL="https://api.printful.com"

echo "=== Testing Printful API ==="
echo ""

# First, let's list store products to see what we have
echo "1. Fetching store products..."
curl -s -X GET \
  -H "Authorization: Bearer $API_KEY" \
  "$API_URL/store/products" | jq '.result[0:3] | .[] | {id, name, variants}'

echo ""
echo "2. Getting details for Eco Tote Bag (ID: should be in the list above)..."
echo "Enter the sync_product ID from above (or press enter to skip): "
read PRODUCT_ID

if [ -n "$PRODUCT_ID" ]; then
  curl -s -X GET \
    -H "Authorization: Bearer $API_KEY" \
    "$API_URL/store/products/$PRODUCT_ID" | jq '.result.sync_variants[0] | {sync_variant_id, name, retail_price, files}'
fi

echo ""
echo "3. Creating a test draft order..."
echo ""

# Create order with sync_variant_id
RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  "$API_URL/orders" \
  -d '{
    "recipient": {
      "name": "John Doe",
      "address1": "123 Main Street",
      "city": "Los Angeles",
      "state_code": "CA",
      "country_code": "US",
      "zip": "90001"
    },
    "items": [
      {
        "sync_variant_id": 10457,
        "quantity": 1
      }
    ]
  }')

echo "$RESPONSE" | jq '.'

# Check if successful
if echo "$RESPONSE" | jq -e '.result.id' > /dev/null 2>&1; then
  ORDER_ID=$(echo "$RESPONSE" | jq -r '.result.id')
  echo ""
  echo "✓ Draft order created successfully!"
  echo "Order ID: $ORDER_ID"
  echo ""
  echo "View in dashboard: https://www.printful.com/dashboard/default/orders"
else
  echo ""
  echo "❌ Failed to create order"
  echo "Error: $(echo "$RESPONSE" | jq -r '.error.message // .result')"
fi
