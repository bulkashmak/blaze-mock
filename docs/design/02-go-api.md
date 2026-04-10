# Go API Design

The API uses a **builder pattern** because it reads naturally in Go, gives IDE autocompletion at every step, and prevents invalid combinations at compile time.

## Usage Example

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
            WithBodyContaining(`"amount"`).
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

## Builder API Surface

```go
// Entry points - one per HTTP method
func Get(path string) *StubBuilder
func Post(path string) *StubBuilder
func Put(path string) *StubBuilder
func Delete(path string) *StubBuilder
func Patch(path string) *StubBuilder
func Method(method, path string) *StubBuilder

// StubBuilder (all return *StubBuilder for chaining)
func (b *StubBuilder) WithID(id string) *StubBuilder
func (b *StubBuilder) WithHeader(name string, matcher StringMatcher) *StubBuilder
func (b *StubBuilder) WithQueryParam(name string, matcher StringMatcher) *StubBuilder
func (b *StubBuilder) WithBody(matcher BodyMatcher) *StubBuilder
func (b *StubBuilder) WithBodyContaining(substr string) *StubBuilder
func (b *StubBuilder) WillReturn(resp *ResponseBuilder) *StubBuilder
func (b *StubBuilder) WillRespondWith(fn ResponseFunc) *StubBuilder

// ResponseBuilder
func Response(status int) *ResponseBuilder
func (rb *ResponseBuilder) WithHeader(k, v string) *ResponseBuilder
func (rb *ResponseBuilder) WithBody(body string) *ResponseBuilder
func (rb *ResponseBuilder) WithBodyFile(path string) *ResponseBuilder
func (rb *ResponseBuilder) WithBodyJSON(v any) *ResponseBuilder
func (rb *ResponseBuilder) WithDelay(d time.Duration) *ResponseBuilder

// StringMatcher constructors
func EqualTo(v string) StringMatcher
func Prefix(v string) StringMatcher
func Contains(v string) StringMatcher
func MatchesRegex(pattern string) StringMatcher
```

## Server API

```go
func NewServer(opts ...ServerOption) *Server
func (s *Server) Start() error
func (s *Server) Shutdown() error
func (s *Server) URL() string

// Stub management
func (s *Server) Stub(b *StubBuilder) string
func (s *Server) RemoveStub(id string) bool
func (s *Server) ResetStubs()
func (s *Server) ListStubs() []Stub

// Server options
func WithPort(port int) ServerOption
```

`NewServer` creates the server and registers stubs. `Start` begins listening and blocks until the server is shut down.
