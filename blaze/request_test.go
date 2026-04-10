package blaze

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func reqWithBody(method, url, body string) *http.Request {
	r, _ := http.NewRequest(method, url, strings.NewReader(body))
	// Simulate what the handler does: buffer body into context
	ctx := context.WithValue(r.Context(), requestBodyKey{}, []byte(body))
	return r.WithContext(ctx)
}

func TestReqHeader(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Custom", "value123")
	if got := Req(r).Header("X-Custom"); got != "value123" {
		t.Errorf("expected value123, got %s", got)
	}
}

func TestReqQueryParam(t *testing.T) {
	r, _ := http.NewRequest("GET", "/search?q=blaze&page=2", nil)
	rv := Req(r)
	if got := rv.QueryParam("q"); got != "blaze" {
		t.Errorf("expected blaze, got %s", got)
	}
	if got := rv.QueryParam("page"); got != "2" {
		t.Errorf("expected 2, got %s", got)
	}
}

func TestReqJSONPath(t *testing.T) {
	body := `{"user": {"name": "Alice", "age": 30}, "tags": ["go", "mock"]}`
	r := reqWithBody("POST", "/", body)
	rv := Req(r)

	if got := rv.JSONPath("$.user.name"); got != "Alice" {
		t.Errorf("expected Alice, got %s", got)
	}
	if got := rv.JSONPath("$.user.age"); got != "30" {
		t.Errorf("expected 30, got %s", got)
	}
	if got := rv.JSONPath("$.tags.0"); got != "go" {
		t.Errorf("expected go, got %s", got)
	}
	if got := rv.JSONPath("$.missing"); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestReqBody(t *testing.T) {
	r := reqWithBody("POST", "/", "raw body content")
	if got := Req(r).Body(); got != "raw body content" {
		t.Errorf("expected 'raw body content', got '%s'", got)
	}
}

func TestReqPathParam(t *testing.T) {
	r, _ := http.NewRequest("GET", "/api/users/42", nil)
	r = withPathParams(r, map[string]string{"id": "42"})
	if got := Req(r).PathParam("id"); got != "42" {
		t.Errorf("expected 42, got %s", got)
	}
}
