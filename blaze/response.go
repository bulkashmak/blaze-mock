package blaze

import (
	"encoding/json"
	"net/http"
	"time"
)

// Resp is a convenience alias used in WillRespondWith callbacks.
type Resp = *ResponseBuilder

// ResponseFunc is the signature for dynamic response handlers.
type ResponseFunc func(*http.Request) *ResponseBuilder

// ResponseDef holds the response configuration for a stub.
type ResponseDef struct {
	Status       int
	Headers      map[string]string
	Body         []byte
	BodyFile     string
	BodyTemplate string
	BodyFunc     func(*http.Request) (int, map[string]string, []byte, error)
	Delay        time.Duration
}

// ResponseBuilder constructs a ResponseDef using a fluent API.
type ResponseBuilder struct {
	status       int
	headers      map[string]string
	body         []byte
	bodyFile     string
	bodyTemplate string
	bodyJSON     any
	delay        time.Duration
}

// Response creates a new ResponseBuilder with the given HTTP status code.
func Response(status int) *ResponseBuilder {
	return &ResponseBuilder{
		status:  status,
		headers: make(map[string]string),
	}
}

func (rb *ResponseBuilder) WithHeader(k, v string) *ResponseBuilder {
	rb.headers[k] = v
	return rb
}

func (rb *ResponseBuilder) WithBody(body string) *ResponseBuilder {
	rb.body = []byte(body)
	return rb
}

func (rb *ResponseBuilder) WithBodyFile(path string) *ResponseBuilder {
	rb.bodyFile = path
	return rb
}

// WithBodyTemplate sets a response body template with {{.name}} placeholders
// that are resolved from values defined via Extract() on the stub.
func (rb *ResponseBuilder) WithBodyTemplate(tmpl string) *ResponseBuilder {
	rb.bodyTemplate = tmpl
	return rb
}

func (rb *ResponseBuilder) WithBodyJSON(v any) *ResponseBuilder {
	rb.bodyJSON = v
	return rb
}

func (rb *ResponseBuilder) WithDelay(d time.Duration) *ResponseBuilder {
	rb.delay = d
	return rb
}

func (rb *ResponseBuilder) build() ResponseDef {
	body := rb.body
	if rb.bodyJSON != nil {
		data, err := json.Marshal(rb.bodyJSON)
		if err != nil {
			panic("blaze: failed to marshal JSON body: " + err.Error())
		}
		body = data
	}
	return ResponseDef{
		Status:       rb.status,
		Headers:      rb.headers,
		Body:         body,
		BodyFile:     rb.bodyFile,
		BodyTemplate: rb.bodyTemplate,
		Delay:        rb.delay,
	}
}
