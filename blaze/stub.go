package blaze

import (
	"net/http"

	"github.com/google/uuid"
)

// Stub binds a request matcher to a response definition.
type Stub struct {
	ID       string
	Request  RequestMatcher
	Response ResponseDef
}

// RequestMatcher describes which incoming HTTP requests a stub should handle.
type RequestMatcher struct {
	Method      string
	Path        StringMatcher
	Headers     map[string]StringMatcher
	QueryParams map[string]StringMatcher
	Body        BodyMatcher
}

// StubBuilder constructs a Stub using a fluent API.
type StubBuilder struct {
	id           string
	method       string
	path         string
	headers      map[string]StringMatcher
	queryParams  map[string]StringMatcher
	body         BodyMatcher
	extractors   map[string]Extractor
	responseDef  ResponseDef
	responseFunc ResponseFunc
}

func newStubBuilder(method, path string) *StubBuilder {
	return &StubBuilder{
		method:      method,
		path:        path,
		headers:     make(map[string]StringMatcher),
		queryParams: make(map[string]StringMatcher),
	}
}

// Get creates a StubBuilder for GET requests.
func Get(path string) *StubBuilder { return newStubBuilder(http.MethodGet, path) }

// Post creates a StubBuilder for POST requests.
func Post(path string) *StubBuilder { return newStubBuilder(http.MethodPost, path) }

// Put creates a StubBuilder for PUT requests.
func Put(path string) *StubBuilder { return newStubBuilder(http.MethodPut, path) }

// Delete creates a StubBuilder for DELETE requests.
func Delete(path string) *StubBuilder { return newStubBuilder(http.MethodDelete, path) }

// Patch creates a StubBuilder for PATCH requests.
func Patch(path string) *StubBuilder { return newStubBuilder(http.MethodPatch, path) }

// Method creates a StubBuilder for an arbitrary HTTP method.
func Method(method, path string) *StubBuilder { return newStubBuilder(method, path) }

func (b *StubBuilder) WithID(id string) *StubBuilder {
	b.id = id
	return b
}

func (b *StubBuilder) WithHeader(name string, matcher StringMatcher) *StubBuilder {
	b.headers[name] = matcher
	return b
}

func (b *StubBuilder) WithQueryParam(name string, matcher StringMatcher) *StubBuilder {
	b.queryParams[name] = matcher
	return b
}

func (b *StubBuilder) WithBody(matcher BodyMatcher) *StubBuilder {
	b.body = matcher
	return b
}

func (b *StubBuilder) WithBodyContaining(substr string) *StubBuilder {
	b.body = ContainsString(substr)
	return b
}

// Extract defines a named extractor that pulls a value from the incoming request.
// Extracted values are available in response templates via {{.name}}.
func (b *StubBuilder) Extract(name string, extractor Extractor) *StubBuilder {
	if b.extractors == nil {
		b.extractors = make(map[string]Extractor)
	}
	b.extractors[name] = extractor
	return b
}

func (b *StubBuilder) WillReturn(resp *ResponseBuilder) *StubBuilder {
	b.responseDef = resp.build()
	return b
}

func (b *StubBuilder) WillRespondWith(fn ResponseFunc) *StubBuilder {
	b.responseFunc = fn
	return b
}

func (b *StubBuilder) build() Stub {
	id := b.id
	if id == "" {
		id = uuid.New().String()
	}

	pathMatcher := compilePath(b.path)

	resp := b.responseDef
	if b.responseFunc != nil {
		fn := b.responseFunc
		resp.BodyFunc = func(r *http.Request) (int, map[string]string, []byte, error) {
			rb := fn(r)
			built := rb.build()
			return built.Status, built.Headers, built.Body, nil
		}
	} else if len(b.extractors) > 0 {
		extractors := b.extractors
		base := resp
		resp.BodyFunc = func(r *http.Request) (int, map[string]string, []byte, error) {
			// Get buffered body from context
			var body []byte
			if cached, ok := r.Context().Value(requestBodyKey{}).([]byte); ok {
				body = cached
			}

			// Run all extractors
			values := make(map[string]string, len(extractors))
			for name, ext := range extractors {
				values[name] = ext.Extract(r, body)
			}

			// Apply template substitution to response body
			respBody := base.Body
			if base.BodyTemplate != "" {
				respBody = []byte(renderTemplate(base.BodyTemplate, values))
			}

			// Apply template substitution to response headers
			headers := make(map[string]string, len(base.Headers))
			for k, v := range base.Headers {
				headers[k] = renderTemplate(v, values)
			}

			return base.Status, headers, respBody, nil
		}
	}

	return Stub{
		ID: id,
		Request: RequestMatcher{
			Method:      b.method,
			Path:        pathMatcher,
			Headers:     b.headers,
			QueryParams: b.queryParams,
			Body:        b.body,
		},
		Response: resp,
	}
}
