# Admin REST API

## Motivation

Blaze Mock stubs are defined in Go code, which is its core strength. However, a standalone QA mock server also needs runtime management from non-Go consumers (Python/JS test suites, Postman, CI scripts). An HTTP admin API solves this without sacrificing the Go-native experience.

## Two-tier stub model

Not all stubs need Go power. The admin API embraces this by defining two tiers:

| Tier | Created via | Capabilities | Admin API |
|------|------------|--------------|-----------|
| **Declarative** | JSON (admin API) or Go | Static responses, body from file, string matchers, templates | Full CRUD |
| **Code** | Go only | `WillRespondWith(func)`, custom `BodyMatcher`, closures | List + Delete only |

Both tiers coexist in the same registry. Matching works identically regardless of how the stub was created.

## Separate port

The admin API runs on a dedicated port (configured via `WithAdminPort`), keeping mock traffic and management traffic isolated. This avoids path conflicts with stubs and makes firewall/network rules straightforward.

```go
server := blaze.NewServer(
    blaze.WithPort(8080),
    blaze.WithAdminPort(8081),
)
```

Admin API is opt-in. If `WithAdminPort` is not set, no admin listener starts.

## Endpoints

### Stub management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/stubs` | Create a declarative stub from JSON |
| `GET` | `/stubs` | List all stubs (both tiers) |
| `GET` | `/stubs/{id}` | Get a single stub by ID |
| `DELETE` | `/stubs/{id}` | Remove a stub (any tier) |
| `DELETE` | `/stubs` | Remove all stubs |

### Request journal (future)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/requests` | List recorded requests |
| `POST` | `/requests/count` | Count requests matching a pattern |
| `DELETE` | `/requests` | Clear the journal |

## JSON stub format

Mirrors the Go builder API:

```json
{
  "id": "payment-stub",
  "request": {
    "method": "POST",
    "path": "/api/payments",
    "headers": {
      "Content-Type": {"equalTo": "application/json"}
    },
    "queryParams": {
      "version": {"prefix": "v2"}
    },
    "body": {"contains": "\"amount\""}
  },
  "response": {
    "status": 201,
    "headers": {"Content-Type": "application/json"},
    "body": "{\"id\": \"pay_123\", \"status\": \"created\"}"
  }
}
```

### String matcher JSON mapping

| Go | JSON |
|----|------|
| `EqualTo("v")` | `{"equalTo": "v"}` |
| `Prefix("v")` | `{"prefix": "v"}` |
| `Suffix("v")` | `{"suffix": "v"}` |
| `Contains("v")` | `{"contains": "v"}` |
| `MatchesRegex("v")` | `{"matches": "v"}` |

### Body matcher JSON mapping

| Go | JSON |
|----|------|
| `EqualToBody(b)` | `{"equalTo": "..."}` |
| `ContainsString("v")` | `{"contains": "v"}` |

### Response fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | int | HTTP status code |
| `headers` | map | Response headers |
| `body` | string | Inline response body |
| `bodyFile` | string | Path to a static file |

### Listing code stubs

Code stubs appear in `GET /stubs` with a `"type": "code"` marker. Their request matcher is serialized where possible, but the response shows only metadata:

```json
{
  "id": "order-confirm",
  "type": "code",
  "request": {
    "method": "POST",
    "path": "/api/orders/{id}/confirm"
  },
  "response": {
    "description": "dynamic (Go func)"
  }
}
```

## Server option

```go
func WithAdminPort(port int) ServerOption
```

## Design decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Separate port | Yes | No path conflicts with stubs, clean network isolation |
| Code stubs via API | List + Delete only | Closures can't serialize to JSON |
| Opt-in | No admin unless configured | Zero overhead when not needed |
