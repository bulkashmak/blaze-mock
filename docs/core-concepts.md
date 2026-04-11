# Core Concepts

## Stub

A **Stub** is the central domain object. It binds a **request matcher** to a **response definition**.

```go
type Stub struct {
    ID       string
    Request  RequestMatcher
    Response ResponseDef
}
```

## RequestMatcher

Describes which incoming HTTP requests a stub should handle.

```go
type RequestMatcher struct {
    Method      string
    Path        StringMatcher
    Headers     map[string]StringMatcher
    QueryParams map[string]StringMatcher
    Body        BodyMatcher
}
```

## StringMatcher

A single interface that covers exact, prefix, suffix, regex, and contains matching:

```go
type StringMatcher interface {
    Match(value string) bool
}
```

Implementations: `EqualTo(v)`, `Prefix(v)`, `Suffix(v)`, `Contains(v)`, `MatchesRegex(v)`.

## BodyMatcher

```go
type BodyMatcher interface {
    MatchBody(body []byte) bool
}
```

Implementations:

- `EqualToBody([]byte)` — exact byte-level match
- `ContainsString(string)` — substring match
- `EqualToJSON(string)` — structural JSON equality (ignores key order). Chain `.IgnoreExtraFields()` to allow the actual body to contain additional fields
- `MatchesJSONPath(path, matcher)` — extract a value at a JSONPath and check it against a `StringMatcher`
- `AllOf(matchers...)` — compose multiple `BodyMatcher`s; passes only when all pass

## ResponseDef

```go
type ResponseDef struct {
    Status   int
    Headers  map[string]string
    Body     []byte
    BodyFile string                              // path to a static file (e.g. JSON fixture)
    BodyFunc func(*http.Request) ([]byte, error) // dynamic body (imperative power)
    Delay    time.Duration
}
```

The key differentiator from WireMock: `BodyFunc`. Users can write arbitrary Go to compute a response.

Resolution order when multiple body sources are set: `BodyFunc` > `BodyFile` > `Body`.

## Request Value Extraction

Blaze provides two ways to extract values from incoming requests and use them in responses.

### Req() Helper

`Req()` wraps `*http.Request` for ergonomic extraction inside `WillRespondWith` callbacks:

```go
func Req(r *http.Request) *RequestValue

func (rv *RequestValue) Header(name string) string
func (rv *RequestValue) QueryParam(name string) string
func (rv *RequestValue) PathParam(name string) string
func (rv *RequestValue) Body() string
func (rv *RequestValue) JSONPath(path string) string    // dot-notation: "$.user.name"
func (rv *RequestValue) JSONPathAny(path string) any
```

### Extractors

`Extractor` pulls a named value from a request for use in response templates:

```go
type Extractor interface {
    Extract(r *http.Request, body []byte) string
}
```

Built-in extractors: `FromHeader(name)`, `FromQueryParam(name)`, `FromPathParam(name)`, `FromJSONPath(path)`, `FromBody()`.

Extracted values are referenced in response templates via `{{.name}}` placeholders. Templates work in both `WithBodyTemplate()` and `WithHeader()` values.
