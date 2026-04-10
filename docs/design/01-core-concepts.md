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

Describes which incoming HTTP requests this stub should handle.

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

Concrete implementations: `EqualTo(v)`, `Prefix(v)`, `Suffix(v)`, `Contains(v)`, `MatchesRegex(v)`.

## BodyMatcher

```go
type BodyMatcher interface {
    MatchBody(body []byte) bool
}
```

Implementations for v1: `EqualToBody([]byte)`, `ContainsString(string)`, `MatchesJSONPath(expr, expected)`.

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
