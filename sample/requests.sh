#!/usr/bin/env bash
# Sample curl requests for the Blaze Mock sample server.
# Start the server first: go run ./sample/

BASE_URL="${BLAZE_URL:-http://localhost:8080}"

echo "=== GET /health ==="
curl -s "$BASE_URL/health" | jq .
echo

echo "=== GET /api/users (static JSON file) ==="
curl -s "$BASE_URL/api/users" | jq .
echo

echo "=== POST /api/payments (static inline response) ==="
curl -s -X POST "$BASE_URL/api/payments" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "USD"}' | jq .
echo

echo "=== POST /api/orders/ord_55/confirm (Option A: Req() helper) ==="
curl -s -X POST "$BASE_URL/api/orders/ord_55/confirm" \
  -H "Content-Type: application/json" \
  -H "X-Source: qa-suite" \
  -d '{"customer": {"name": "Alice", "email": "alice@example.com"}, "items": 3}' | jq .
echo

echo "=== POST /api/echo (Option B: Extract + Template) ==="
curl -s -D - -X POST "$BASE_URL/api/echo?format=json" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tok_abc" \
  -d '{"user": {"name": "Bob", "email": "bob@example.com"}}'
echo
echo

echo "=== GET /api/users/usr_42 (Extract + Template with path param) ==="
curl -s "$BASE_URL/api/users/usr_42" | jq .
echo

echo "=== GET /unknown (404 diagnostic) ==="
curl -s "$BASE_URL/unknown" | jq .
