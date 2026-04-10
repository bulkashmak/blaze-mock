# Go API Design

The API uses a **builder pattern** because it reads naturally in Go, gives IDE autocompletion at every step, and prevents invalid combinations at compile time.

## Usage Examples

### Static responses

```go
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
```

### Response from a static JSON file

```go
server.Stub(
    blaze.Get("/api/users").
        WillReturn(
            blaze.Response(200).
                WithHeader("Content-Type", "application/json").
                WithBodyFile("fixtures/users.json"),
        ),
)
```

### Dynamic response with Req() helper (Option A)

Full Go power — extract values from the request and build the response imperatively:

```go
server.Stub(
    blaze.Post("/api/orders/{id}/confirm").
        WillRespondWith(func(r *http.Request) blaze.Resp {
            req := blaze.Req(r)
            return blaze.Response(200).
                WithHeader("Content-Type", "application/json").
                WithBodyJSON(map[string]any{
                    "order_id": req.PathParam("id"),
                    "customer": req.JSONPath("$.customer.name"),
                    "email":    req.JSONPath("$.customer.email"),
                    "source":   req.Header("X-Source"),
                    "status":   "confirmed",
                })
        }),
)
```

### Extract + Template (Option B)

Declarative extraction with template-based response — no callback needed:

```go
server.Stub(
    blaze.Post("/api/echo").
        Extract("name", blaze.FromJSONPath("$.user.name")).
        Extract("token", blaze.FromHeader("Authorization")).
        Extract("format", blaze.FromQueryParam("format")).
        WillReturn(
            blaze.Response(200).
                WithHeader("Content-Type", "application/json").
                WithHeader("X-Auth", "{{.token}}").
                WithBodyTemplate(`{"greeting": "Hello, {{.name}}", "format": "{{.format}}"}`),
        ),
)
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
func (b *StubBuilder) Extract(name string, extractor Extractor) *StubBuilder
func (b *StubBuilder) WillReturn(resp *ResponseBuilder) *StubBuilder
func (b *StubBuilder) WillRespondWith(fn ResponseFunc) *StubBuilder

// ResponseBuilder
func Response(status int) *ResponseBuilder
func (rb *ResponseBuilder) WithHeader(k, v string) *ResponseBuilder
func (rb *ResponseBuilder) WithBody(body string) *ResponseBuilder
func (rb *ResponseBuilder) WithBodyFile(path string) *ResponseBuilder
func (rb *ResponseBuilder) WithBodyTemplate(tmpl string) *ResponseBuilder
func (rb *ResponseBuilder) WithBodyJSON(v any) *ResponseBuilder
func (rb *ResponseBuilder) WithDelay(d time.Duration) *ResponseBuilder

// StringMatcher constructors
func EqualTo(v string) StringMatcher
func Prefix(v string) StringMatcher
func Contains(v string) StringMatcher
func MatchesRegex(pattern string) StringMatcher

// Extractor constructors (for use with Extract)
func FromHeader(name string) Extractor
func FromQueryParam(name string) Extractor
func FromPathParam(name string) Extractor
func FromJSONPath(path string) Extractor
func FromBody() Extractor

// Request value helper (for use inside WillRespondWith)
func Req(r *http.Request) *RequestValue
func (rv *RequestValue) Header(name string) string
func (rv *RequestValue) QueryParam(name string) string
func (rv *RequestValue) PathParam(name string) string
func (rv *RequestValue) Body() string
func (rv *RequestValue) JSONPath(path string) string
func (rv *RequestValue) JSONPathAny(path string) any
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
