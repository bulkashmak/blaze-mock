# Blaze Mock - Sample

A working example of a Blaze Mock server demonstrating all stub types.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Static inline response |
| `GET` | `/api/users` | Response from a static JSON file (`fixtures/users.json`) |
| `GET` | `/api/users/{id}` | Extract path param + template response |
| `POST` | `/api/payments` | Static response with header and body matching |
| `POST` | `/api/orders/{id}/confirm` | Dynamic response using `Req()` helper (Option A) |
| `POST` | `/api/echo?format=...` | Declarative `Extract` + `WithBodyTemplate` (Option B) |

## Running

From the repository root:

```bash
go run ./sample/
```

The server starts on port 8080.

## Testing

Run the provided curl scripts against the running server:

```bash
./sample/requests.sh
```

Or try individual requests:

```bash
# Static file response
curl -s http://localhost:8080/api/users | jq .

# Option A: Req() helper - extract from path, body, and headers
curl -s -X POST http://localhost:8080/api/orders/ord_55/confirm \
  -H "Content-Type: application/json" \
  -H "X-Source: qa-suite" \
  -d '{"customer": {"name": "Alice", "email": "alice@example.com"}}' | jq .

# Option B: Extract + Template - declarative extraction with template placeholders
curl -s -X POST "http://localhost:8080/api/echo?format=json" \
  -H "Authorization: Bearer tok_abc" \
  -d '{"user": {"name": "Bob", "email": "bob@example.com"}}' | jq .
```
