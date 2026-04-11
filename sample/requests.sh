#!/usr/bin/env bash
# Sample curl requests for the Blaze Mock sample server.
# Start the server first: go run ./sample/

BASE_URL="${BLAZE_URL:-http://localhost:8080}"

echo "--- GET /health"
curl -s "$BASE_URL/health" | jq .
echo

echo "--- GET /api/users"
curl -s "$BASE_URL/api/users" | jq .
echo

echo "--- POST /api/payments"
curl -s -X POST "$BASE_URL/api/payments" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "USD"}' | jq .
echo

echo "--- POST /api/invoices (EqualToJSON)"
curl -s -X POST "$BASE_URL/api/invoices" \
  -H "Content-Type: application/json" \
  -d '{"currency": "EUR", "amount": 500}' | jq .
echo

echo "--- POST /api/refunds (MatchesJSONPath)"
curl -s -X POST "$BASE_URL/api/refunds" \
  -H "Content-Type: application/json" \
  -d '{"order_id": "ord_99", "reason": "item defective on arrival"}' | jq .
echo

echo "--- POST /api/orders/ord_55/confirm"
curl -s -X POST "$BASE_URL/api/orders/ord_55/confirm" \
  -H "Content-Type: application/json" \
  -H "X-Source: qa-suite" \
  -d '{"customer": {"name": "Alice", "email": "alice@example.com"}, "items": 3}' | jq .
echo

echo "--- POST /api/echo"
curl -s -D - -X POST "$BASE_URL/api/echo?format=json" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tok_abc" \
  -d '{"user": {"name": "Bob", "email": "bob@example.com"}}'
echo
echo

echo "--- GET /api/users/usr_42"
curl -s "$BASE_URL/api/users/usr_42" | jq .
echo

echo "--- GET /unknown"
curl -s "$BASE_URL/unknown" | jq .
