package blaze_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bulkashmak/blaze-mock/blaze"
)

func TestExtractWithTemplate(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Post("/api/echo").
			Extract("name", blaze.FromJSONPath("$.user.name")).
			Extract("token", blaze.FromHeader("Authorization")).
			Extract("format", blaze.FromQueryParam("format")).
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithHeader("X-Token", "{{.token}}").
					WithBodyTemplate(`{"greeting": "Hello, {{.name}}", "format": "{{.format}}"}`),
			),
	)

	body := strings.NewReader(`{"user": {"name": "Alice"}}`)
	req, _ := http.NewRequest("POST", s.URL()+"/api/echo?format=json", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer abc123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if got := resp.Header.Get("X-Token"); got != "Bearer abc123" {
		t.Errorf("expected 'Bearer abc123' in X-Token header, got '%s'", got)
	}

	data, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(data, &result)

	if result["greeting"] != "Hello, Alice" {
		t.Errorf("expected 'Hello, Alice', got '%s'", result["greeting"])
	}
	if result["format"] != "json" {
		t.Errorf("expected 'json', got '%s'", result["format"])
	}
}

func TestExtractPathParam(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Get("/api/users/{id}").
			Extract("user_id", blaze.FromPathParam("id")).
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyTemplate(`{"id": "{{.user_id}}", "found": true}`),
			),
	)

	resp, err := http.Get(s.URL() + "/api/users/usr_99")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(data, &result)

	if result["id"] != "usr_99" {
		t.Errorf("expected 'usr_99', got '%v'", result["id"])
	}
}

func TestReqHelperInWillRespondWith(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Post("/api/transform/{id}").
			WillRespondWith(func(r *http.Request) blaze.Resp {
				req := blaze.Req(r)
				return blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyJSON(map[string]any{
						"id":     req.PathParam("id"),
						"name":   req.JSONPath("$.name"),
						"source": req.Header("X-Source"),
						"page":   req.QueryParam("page"),
					})
			}),
	)

	body := strings.NewReader(`{"name": "Widget"}`)
	req, _ := http.NewRequest("POST", s.URL()+"/api/transform/item_7?page=3", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Source", "test-suite")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(data, &result)

	if result["id"] != "item_7" {
		t.Errorf("expected 'item_7', got '%s'", result["id"])
	}
	if result["name"] != "Widget" {
		t.Errorf("expected 'Widget', got '%s'", result["name"])
	}
	if result["source"] != "test-suite" {
		t.Errorf("expected 'test-suite', got '%s'", result["source"])
	}
	if result["page"] != "3" {
		t.Errorf("expected '3', got '%s'", result["page"])
	}
}
