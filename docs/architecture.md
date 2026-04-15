# Architecture

```
+---------------------------------------------------+
|                   blaze.Server                     |
|                                                    |
|  +-----------+                     +------------+  |
|  | Mock HTTP |                     |   Stub     |  |
|  | Handler   |-------------------->|  Registry  |  |
|  | (serves   |                     |  (in-mem)  |  |
|  |  stubs)   |                     |            |  |
|  +-----------+                     +------------+  |
|        ^                                  ^        |
+--------|----------------------------------|--------+
         |                                  |
   incoming                            blaze.Stub()
   HTTP requests                       (Go code)
```

## Package Layout

```
blaze-mock/
  blaze/              # public library package
    server.go         # Server, NewServer, Start, Shutdown
    stub.go           # Stub struct, StubBuilder
    response.go       # ResponseDef, ResponseBuilder
    matcher.go        # StringMatcher, BodyMatcher implementations
    registry.go       # StubRegistry (thread-safe in-memory store)
    handler.go        # Mock HTTP handler (matches requests to stubs)
    request.go        # Req() helper for request value extraction
    extract.go        # Extractor interface and built-in extractors
    template.go       # {{.name}} template rendering
    options.go        # ServerOption functional options
    logger.go         # LogOutput type and logger construction
  samples/            # sample projects (basic, admin-api, ...)
  docs/               # documentation
  go.mod
  README.md
```

Single public package `blaze`. No internal sub-packages for v1.

## Stub Registry

```go
type StubRegistry struct {
    mu    sync.RWMutex
    stubs []Stub // ordered by insertion order
}

func (r *StubRegistry) Add(s Stub) string
func (r *StubRegistry) Remove(id string) bool
func (r *StubRegistry) Reset()
func (r *StubRegistry) List() []Stub
func (r *StubRegistry) Match(req *http.Request) (*Stub, bool)
```

`Match` iterates stubs in insertion order, returns the first match. O(n) is fine — mock servers typically have tens of stubs, not thousands.

## Request Matching

When an HTTP request arrives at the mock handler:

1. Iterate stubs in insertion order.
2. For each stub, check all matcher fields. A stub matches only if **all** specified matchers pass. Unspecified fields are "match any".
3. Return the first fully-matching stub.
4. If no stub matches, return 404 with diagnostic info (incoming request details + all registered stubs).

### Path Parameters

`blaze.Get("/api/payments/{id}")` is internally converted to a regex `/api/payments/([^/]+)` with a named capture group. `blaze.PathParam(r, "id")` extracts the captured value from request context.

## Key Design Decisions

| Decision              | Choice                                      | Rationale                                                             |
| --------------------- | ------------------------------------------- | --------------------------------------------------------------------- |
| Stub definition style | Builder pattern                             | Reads naturally, IDE autocompletion, mirrors k6's imperative feel     |
| Dynamic responses     | `BodyFunc` via `WillRespondWith`            | Core differentiator — full Go language for response logic             |
| Matching order        | Insertion order (first registered wins)     | Simple, predictable, debuggable                                       |
| Path params           | `{name}` syntax -> regex                   | Familiar to Go developers (echo, chi patterns)                        |
| Thread safety         | `sync.RWMutex` in registry                 | Concurrent reads during matching, safe writes                         |
| Persistence           | In-memory only                              | Mock servers are ephemeral — no database, no files                    |
| 404 diagnostics       | Request details + registered stubs          | Critical for debugging why a request didn't match                     |
| Package structure     | Single `blaze` package                     | Simple import path                                                    |
| Request-to-response   | Req() helper + Extract/Template             | Req() for full Go power, Extract/Template for declarative cases       |
