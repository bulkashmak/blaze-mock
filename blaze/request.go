package blaze

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type requestBodyKey struct{}

// RequestValue provides ergonomic access to request data.
type RequestValue struct {
	r    *http.Request
	body []byte
	once sync.Once
}

// Req wraps an *http.Request for convenient value extraction.
func Req(r *http.Request) *RequestValue {
	return &RequestValue{r: r}
}

// Header returns the value of the named request header.
func (rv *RequestValue) Header(name string) string {
	return rv.r.Header.Get(name)
}

// QueryParam returns the value of the named query parameter.
func (rv *RequestValue) QueryParam(name string) string {
	return rv.r.URL.Query().Get(name)
}

// PathParam returns the value of the named path parameter.
func (rv *RequestValue) PathParam(name string) string {
	return PathParam(rv.r, name)
}

// Body returns the raw request body as a string.
func (rv *RequestValue) Body() string {
	return string(rv.loadBody())
}

// JSONPath extracts a value from the JSON request body using a simple dot-notation path.
// Supported syntax: "$.field", "$.nested.field", "$.array.0.field".
// Returns the value as a string, or "" if not found.
func (rv *RequestValue) JSONPath(path string) string {
	body := rv.loadBody()
	if len(body) == 0 {
		return ""
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	return jsonPathLookup(data, path)
}

// JSONPathAny extracts a value from the JSON request body, returning it as-is (not stringified).
func (rv *RequestValue) JSONPathAny(path string) any {
	body := rv.loadBody()
	if len(body) == 0 {
		return nil
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}

	return jsonPathLookupAny(data, path)
}

func (rv *RequestValue) loadBody() []byte {
	rv.once.Do(func() {
		// First check if body was already buffered by the handler
		if cached, ok := rv.r.Context().Value(requestBodyKey{}).([]byte); ok {
			rv.body = cached
			return
		}
		if rv.r.Body != nil {
			rv.body, _ = io.ReadAll(rv.r.Body)
		}
	})
	return rv.body
}

// jsonPathLookup navigates a JSON structure using dot-notation and returns a string.
func jsonPathLookup(data any, path string) string {
	result := jsonPathLookupAny(data, path)
	if result == nil {
		return ""
	}
	switch v := result.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// jsonPathLookupAny navigates a JSON structure using dot-notation.
// Supports: "$.field", "$.nested.field", "$.array.0.field"
func jsonPathLookupAny(data any, path string) any {
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil
			}
			current = val
		case []any:
			idx := 0
			for _, c := range part {
				if c < '0' || c > '9' {
					return nil
				}
				idx = idx*10 + int(c-'0')
			}
			if idx >= len(v) {
				return nil
			}
			current = v[idx]
		default:
			return nil
		}
	}

	return current
}
