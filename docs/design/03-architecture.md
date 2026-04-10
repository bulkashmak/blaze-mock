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
  sample/             # sample project
    main.go
    fixtures/
    requests.sh       # curl requests for testing
  docs/
    design/           # design documents
  go.mod
  README.md
```

Single public package `blaze`. No internal sub-packages for v1.

## StubRegistry

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

`Match` iterates stubs in insertion order, returns the first match. O(n) is fine for v1 - mock servers typically have tens of stubs, not thousands.
