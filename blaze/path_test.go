package blaze

import (
	"net/http"
	"testing"
)

func TestCompilePath_Exact(t *testing.T) {
	pm := compilePath("/api/users")
	if !pm.Match("/api/users") {
		t.Error("expected match")
	}
	if pm.Match("/api/users/123") {
		t.Error("expected no match")
	}
}

func TestCompilePath_WithParams(t *testing.T) {
	pm := compilePath("/api/users/{id}")
	if !pm.Match("/api/users/123") {
		t.Error("expected match")
	}
	if pm.Match("/api/users/") {
		t.Error("expected no match for empty param")
	}
	if pm.Match("/api/users/123/extra") {
		t.Error("expected no match for extra segments")
	}
}

func TestCompilePath_MultipleParams(t *testing.T) {
	pm := compilePath("/api/{org}/repos/{repo}")
	params := pm.extractParams("/api/acme/repos/widget")
	if params == nil {
		t.Fatal("expected match")
	}
	if params["org"] != "acme" {
		t.Errorf("expected org=acme, got %s", params["org"])
	}
	if params["repo"] != "widget" {
		t.Errorf("expected repo=widget, got %s", params["repo"])
	}
}

func TestPathParam(t *testing.T) {
	pm := compilePath("/api/items/{id}")
	params := pm.extractParams("/api/items/42")
	r, _ := http.NewRequest("GET", "/api/items/42", nil)
	r = withPathParams(r, params)

	if got := PathParam(r, "id"); got != "42" {
		t.Errorf("expected 42, got %s", got)
	}
	if got := PathParam(r, "missing"); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}
