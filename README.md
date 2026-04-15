# Blaze Mock

Blaze Mock is a lightweight, standalone HTTP mock server for QA testing. Define stubs imperatively in Go code - no JSON, no YAML, no separate config files.

Inspired by [Grafana k6](https://k6.io/)'s imperative approach to load testing, but applied to HTTP mocking.

## Why Blaze?

- **Imperative over declarative** - stubs are Go code with full language power (variables, loops, conditionals, type safety)
- **Dynamic responses** - compute responses with arbitrary Go logic via `WillRespondWith`
- **Request-to-response mapping** - extract values from requests and inject them into responses
- **Standalone mock server** - deploy as a single container for QA environments
- **Lightweight** - single Go binary, starts in milliseconds, no JVM

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    "github.com/bulkashmak/blaze-mock/blaze"
)

func main() {
    server := blaze.NewServer(blaze.WithPort(8080))

    // Static response
    server.Stub(
        blaze.Post("/api/payments").
            WithHeader("Content-Type", blaze.EqualTo("application/json")).
            WillReturn(
                blaze.Response(201).
                    WithHeader("Content-Type", "application/json").
                    WithBody(`{"id": "pay_123", "status": "created"}`),
            ),
    )

    // Response from a static JSON file
    server.Stub(
        blaze.Get("/api/users").
            WillReturn(
                blaze.Response(200).
                    WithHeader("Content-Type", "application/json").
                    WithBodyFile("fixtures/users.json"),
            ),
    )

    // JSON body matching — structural equality (ignores key order)
    server.Stub(
        blaze.Post("/api/invoices").
            WithBody(blaze.EqualToJSON(`{"amount": 500, "currency": "EUR"}`)).
            WillReturn(
                blaze.Response(201).
                    WithHeader("Content-Type", "application/json").
                    WithBody(`{"id": "inv_001", "status": "created"}`),
            ),
    )

    // JSON body matching — match a specific field by JSONPath
    server.Stub(
        blaze.Post("/api/refunds").
            WithBody(blaze.MatchesJSONPath("$.reason", blaze.Contains("defective"))).
            WillReturn(blaze.Response(202)),
    )

    // Dynamic response with Req() helper (Option A)
    server.Stub(
        blaze.Post("/api/orders/{id}/confirm").
            WillRespondWith(func(r *http.Request) blaze.Resp {
                req := blaze.Req(r)
                return blaze.Response(200).
                    WithHeader("Content-Type", "application/json").
                    WithBodyJSON(map[string]any{
                        "order_id": req.PathParam("id"),
                        "customer": req.JSONPath("$.customer.name"),
                        "source":   req.Header("X-Source"),
                        "status":   "confirmed",
                    })
            }),
    )

    // Extract + Template (Option B)
    server.Stub(
        blaze.Post("/api/echo").
            Extract("name", blaze.FromJSONPath("$.user.name")).
            Extract("token", blaze.FromHeader("Authorization")).
            WillReturn(
                blaze.Response(200).
                    WithHeader("Content-Type", "application/json").
                    WithHeader("X-Auth", "{{.token}}").
                    WithBodyTemplate(`{"greeting": "Hello, {{.name}}"}`),
            ),
    )

    log.Fatal(server.Start())
}
```

## Samples

| Sample | Description |
|--------|-------------|
| [`samples/basic/`](samples/basic/) | All stub types: static, file, JSON matching, dynamic, templates |
| [`samples/admin-api/`](samples/admin-api/) | HTTP admin API for runtime stub CRUD |

Run any sample from the repository root:

```bash
go run ./samples/basic/
go run ./samples/admin-api/
```

## Documentation

See [docs/](docs/README.md) for the full documentation.
