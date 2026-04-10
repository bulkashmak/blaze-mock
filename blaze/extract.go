package blaze

import (
	"encoding/json"
	"net/http"
)

// Extractor extracts a value from an incoming HTTP request.
type Extractor interface {
	Extract(r *http.Request, body []byte) string
}

// FromHeader extracts a request header value.
func FromHeader(name string) Extractor {
	return &headerExtractor{name: name}
}

type headerExtractor struct{ name string }

func (e *headerExtractor) Extract(r *http.Request, _ []byte) string {
	return r.Header.Get(e.name)
}

// FromQueryParam extracts a query parameter value.
func FromQueryParam(name string) Extractor {
	return &queryParamExtractor{name: name}
}

type queryParamExtractor struct{ name string }

func (e *queryParamExtractor) Extract(r *http.Request, _ []byte) string {
	return r.URL.Query().Get(e.name)
}

// FromPathParam extracts a path parameter value.
func FromPathParam(name string) Extractor {
	return &pathParamExtractor{name: name}
}

type pathParamExtractor struct{ name string }

func (e *pathParamExtractor) Extract(r *http.Request, _ []byte) string {
	return PathParam(r, e.name)
}

// FromJSONPath extracts a value from the JSON request body using dot-notation.
// Supported syntax: "$.field", "$.nested.field", "$.array.0.field".
func FromJSONPath(path string) Extractor {
	return &jsonPathExtractor{path: path}
}

type jsonPathExtractor struct{ path string }

func (e *jsonPathExtractor) Extract(_ *http.Request, body []byte) string {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}
	return jsonPathLookup(data, e.path)
}

// FromBody extracts the entire request body as a string.
func FromBody() Extractor {
	return &bodyExtractor{}
}

type bodyExtractor struct{}

func (e *bodyExtractor) Extract(_ *http.Request, body []byte) string {
	return string(body)
}
