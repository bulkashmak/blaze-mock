# Blaze Mock

Blaze Mock is a lightweight, standalone HTTP mock server for QA testing. Define stubs imperatively in Go code - no JSON, no YAML, no separate config files.

Inspired by [Grafana k6](https://k6.io/)'s imperative approach to load testing, but applied to HTTP mocking.

## Why Blaze?

- **Imperative over declarative** - stubs are Go code with full language power (variables, loops, conditionals, type safety)
- **Dynamic responses** - compute responses with arbitrary Go logic via `WillRespondWith`
- **Standalone mock server** - deploy as a single container for QA environments
- **Lightweight** - single Go binary, starts in milliseconds, no JVM

## Quick Start

```go
package main

import (
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

    // Dynamic response
    server.Stub(
        blaze.Get("/api/payments/{id}").
            WillRespondWith(func(r *http.Request) blaze.Resp {
                id := blaze.PathParam(r, "id")
                return blaze.Response(200).
                    WithBodyJSON(map[string]string{"id": id, "status": "created"})
            }),
    )

    server.Start()
}
```

## Documentation

See [docs/design/](docs/design/README.md) for the full design document.
