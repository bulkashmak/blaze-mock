#!/usr/bin/env bash
# Sample curl requests for the admin API sample.
# Start the server first: go run ./samples/admin-api/

MOCK_URL="${BLAZE_URL:-http://localhost:8080}"
ADMIN_URL="${BLAZE_ADMIN_URL:-http://localhost:8081}"

echo "=== List stubs (code stub seeded at startup) ==="
curl -s "$ADMIN_URL/stubs" | jq .
echo

echo "=== Create a stub via admin API ==="
curl -s -X POST "$ADMIN_URL/stubs" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "greeting",
    "request": {"method": "GET", "path": "/api/greeting"},
    "response": {
      "status": 200,
      "headers": {"Content-Type": "application/json"},
      "body": "{\"message\": \"hello\"}"
    }
  }' | jq .
echo

echo "=== Hit the new stub on the mock server ==="
curl -s "$MOCK_URL/api/greeting" | jq .
echo

echo "=== Update the stub ==="
curl -s -X PUT "$ADMIN_URL/stubs/greeting" \
  -H "Content-Type: application/json" \
  -d '{
    "request": {"method": "GET", "path": "/api/greeting"},
    "response": {
      "status": 200,
      "headers": {"Content-Type": "application/json"},
      "body": "{\"message\": \"hello, updated!\"}"
    }
  }' | jq .
echo

echo "=== Verify updated response ==="
curl -s "$MOCK_URL/api/greeting" | jq .
echo

echo "=== Get stub by ID ==="
curl -s "$ADMIN_URL/stubs/greeting" | jq .
echo

echo "=== Delete the stub ==="
curl -s -X DELETE "$ADMIN_URL/stubs/greeting" | jq .
echo

echo "=== Verify stub is gone (expect 404) ==="
curl -s -o /dev/null -w "HTTP %{http_code}\n" "$MOCK_URL/api/greeting"
echo

echo "=== Delete all stubs ==="
curl -s -X DELETE "$ADMIN_URL/stubs" | jq .
echo

echo "=== List stubs (expect empty) ==="
curl -s "$ADMIN_URL/stubs" | jq .
