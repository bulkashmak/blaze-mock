package blaze_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bulkashmak/blaze-mock/blaze"
)

// startAdminServer creates a server with both mock and admin ports, starts it, and returns it.
func startAdminServer(t *testing.T) *blaze.Server {
	t.Helper()
	s := blaze.NewServer(
		blaze.WithLogOutput(blaze.LogNone),
		blaze.WithAdminPort(0),
	)

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	for i := 0; i < 50; i++ {
		if s.URL() != "" && s.AdminURL() != "" {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start in time")
	return nil
}

func adminRequest(t *testing.T, method, url string, body string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func readJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON: %s (body: %s)", err, data)
	}
	return result
}

func readJSONArray(t *testing.T, resp *http.Response) []map[string]any {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON array: %s (body: %s)", err, data)
	}
	return result
}

func TestAdminCreateStub(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	stubJSON := `{
		"request": {
			"method": "GET",
			"path": "/api/hello"
		},
		"response": {
			"status": 200,
			"headers": {"Content-Type": "text/plain"},
			"body": "hello world"
		}
	}`

	resp := adminRequest(t, "POST", s.AdminURL()+"/stubs", stubJSON)
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	result := readJSON(t, resp)
	if result["id"] == "" {
		t.Fatal("expected non-empty id")
	}

	// Verify the stub is active on the mock server
	mockResp, err := http.Get(s.URL() + "/api/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer mockResp.Body.Close()

	if mockResp.StatusCode != 200 {
		t.Errorf("expected mock 200, got %d", mockResp.StatusCode)
	}
	data, _ := io.ReadAll(mockResp.Body)
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", data)
	}
}

func TestAdminCreateStubWithID(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	stubJSON := `{
		"id": "my-custom-id",
		"request": {"method": "GET", "path": "/test"},
		"response": {"status": 204}
	}`

	resp := adminRequest(t, "POST", s.AdminURL()+"/stubs", stubJSON)
	result := readJSON(t, resp)
	if result["id"] != "my-custom-id" {
		t.Errorf("expected id 'my-custom-id', got %q", result["id"])
	}
}

func TestAdminCreateStubValidation(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	tests := []struct {
		name string
		body string
	}{
		{"missing method", `{"request":{"path":"/a"},"response":{"status":200}}`},
		{"missing path", `{"request":{"method":"GET"},"response":{"status":200}}`},
		{"missing status", `{"request":{"method":"GET","path":"/a"},"response":{}}`},
		{"invalid json", `{not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := adminRequest(t, "POST", s.AdminURL()+"/stubs", tt.body)
			defer resp.Body.Close()
			if resp.StatusCode != 400 {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestAdminListStubs(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	// Create two stubs via admin API
	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"id": "stub-1",
		"request": {"method": "GET", "path": "/a"},
		"response": {"status": 200, "body": "a"}
	}`).Body.Close()

	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"id": "stub-2",
		"request": {"method": "POST", "path": "/b"},
		"response": {"status": 201, "body": "b"}
	}`).Body.Close()

	resp := adminRequest(t, "GET", s.AdminURL()+"/stubs", "")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	stubs := readJSONArray(t, resp)
	if len(stubs) != 2 {
		t.Fatalf("expected 2 stubs, got %d", len(stubs))
	}

	if stubs[0]["id"] != "stub-1" || stubs[1]["id"] != "stub-2" {
		t.Errorf("unexpected stub IDs: %v, %v", stubs[0]["id"], stubs[1]["id"])
	}
}

func TestAdminListCodeStub(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	// Register a code stub via Go API
	s.Stub(
		blaze.Get("/api/dynamic").
			WithID("code-stub").
			WillRespondWith(func(r *http.Request) blaze.Resp {
				return blaze.Response(200).WithBody("dynamic")
			}),
	)

	resp := adminRequest(t, "GET", s.AdminURL()+"/stubs", "")
	stubs := readJSONArray(t, resp)

	if len(stubs) != 1 {
		t.Fatalf("expected 1 stub, got %d", len(stubs))
	}
	if stubs[0]["type"] != "code" {
		t.Errorf("expected type='code', got %q", stubs[0]["type"])
	}

	respObj, ok := stubs[0]["response"].(map[string]any)
	if !ok {
		t.Fatal("expected response to be an object")
	}
	if respObj["description"] != "dynamic (Go func)" {
		t.Errorf("expected description='dynamic (Go func)', got %q", respObj["description"])
	}
}

func TestAdminGetStub(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"id": "get-me",
		"request": {"method": "GET", "path": "/find-me"},
		"response": {"status": 200, "body": "found"}
	}`).Body.Close()

	resp := adminRequest(t, "GET", s.AdminURL()+"/stubs/get-me", "")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	result := readJSON(t, resp)
	if result["id"] != "get-me" {
		t.Errorf("expected id 'get-me', got %q", result["id"])
	}
}

func TestAdminGetStubNotFound(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	resp := adminRequest(t, "GET", s.AdminURL()+"/stubs/nonexistent", "")
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAdminUpdateStub(t *testing.T) {
	// init
	s := startAdminServer(t)
	defer s.Shutdown()

	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"id": "update-me",
		"request": {"method": "GET", "path": "/hi"},
		"response": {"status": 200}
	}`).Body.Close()

	// when
	resp := adminRequest(t, "PUT", s.AdminURL()+"/stubs/update-me", `{
		"request": {"method": "POST", "path": "/hi"},
		"response": {"status": 200}
	}`)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// then
	respVerify := adminRequest(t, "GET", s.AdminURL()+"/stubs/update-me", "")
	if respVerify.StatusCode != 200 {
		t.Fatalf("expected 200 after update, got %d", respVerify.StatusCode)
	}
	respVerify.Body.Close()
}

func TestAdminDeleteStub(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"id": "delete-me",
		"request": {"method": "GET", "path": "/bye"},
		"response": {"status": 200}
	}`).Body.Close()

	resp := adminRequest(t, "DELETE", s.AdminURL()+"/stubs/delete-me", "")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify it's gone
	resp2 := adminRequest(t, "GET", s.AdminURL()+"/stubs/delete-me", "")
	if resp2.StatusCode != 404 {
		t.Errorf("expected 404 after delete, got %d", resp2.StatusCode)
	}
	resp2.Body.Close()
}

func TestAdminDeleteStubNotFound(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	resp := adminRequest(t, "DELETE", s.AdminURL()+"/stubs/nonexistent", "")
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAdminDeleteAllStubs(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"request": {"method": "GET", "path": "/a"},
		"response": {"status": 200}
	}`).Body.Close()
	adminRequest(t, "POST", s.AdminURL()+"/stubs", `{
		"request": {"method": "GET", "path": "/b"},
		"response": {"status": 200}
	}`).Body.Close()

	resp := adminRequest(t, "DELETE", s.AdminURL()+"/stubs", "")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	listResp := adminRequest(t, "GET", s.AdminURL()+"/stubs", "")
	stubs := readJSONArray(t, listResp)
	if len(stubs) != 0 {
		t.Errorf("expected 0 stubs after delete all, got %d", len(stubs))
	}
}

func TestAdminCreateStubWithMatchers(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	stubJSON := `{
		"request": {
			"method": "POST",
			"path": "/api/payments",
			"headers": {
				"Content-Type": {"equalTo": "application/json"},
				"X-Request-ID": {"prefix": "req-"}
			},
			"queryParams": {
				"version": {"contains": "v2"}
			},
			"body": {"contains": "\"amount\""}
		},
		"response": {
			"status": 201,
			"headers": {"Content-Type": "application/json"},
			"body": "{\"status\": \"created\"}"
		}
	}`

	resp := adminRequest(t, "POST", s.AdminURL()+"/stubs", stubJSON)
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Matching request
	body := strings.NewReader(`{"amount": 100}`)
	req, _ := http.NewRequest("POST", s.URL()+"/api/payments?version=v2.1", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "req-abc123")

	mockResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer mockResp.Body.Close()

	if mockResp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", mockResp.StatusCode)
	}

	// Non-matching request (wrong Content-Type)
	body2 := strings.NewReader(`{"amount": 100}`)
	req2, _ := http.NewRequest("POST", s.URL()+"/api/payments?version=v2.1", body2)
	req2.Header.Set("Content-Type", "text/plain")
	req2.Header.Set("X-Request-ID", "req-abc123")

	mockResp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer mockResp2.Body.Close()

	if mockResp2.StatusCode != 404 {
		t.Errorf("expected 404 for non-matching request, got %d", mockResp2.StatusCode)
	}
}

func TestAdminCreateStubWithBodyFile(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "response.json")
	os.WriteFile(fixture, []byte(`{"users": ["alice", "bob"]}`), 0644)

	s := startAdminServer(t)
	defer s.Shutdown()

	stubJSON := `{
		"request": {"method": "GET", "path": "/api/users"},
		"response": {"status": 200, "bodyFile": "` + fixture + `"}
	}`

	resp := adminRequest(t, "POST", s.AdminURL()+"/stubs", stubJSON)
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	mockResp, err := http.Get(s.URL() + "/api/users")
	if err != nil {
		t.Fatal(err)
	}
	defer mockResp.Body.Close()

	data, _ := io.ReadAll(mockResp.Body)
	if string(data) != `{"users": ["alice", "bob"]}` {
		t.Errorf("unexpected body: %s", data)
	}
}

func TestAdminNoPortNoAdmin(t *testing.T) {
	s := blaze.NewServer(blaze.WithLogOutput(blaze.LogNone))

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	for i := 0; i < 50; i++ {
		if s.URL() != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer s.Shutdown()

	if s.AdminURL() != "" {
		t.Errorf("expected empty AdminURL when no admin port configured, got %q", s.AdminURL())
	}
}

func TestAdminCreateStubWithPathParams(t *testing.T) {
	s := startAdminServer(t)
	defer s.Shutdown()

	stubJSON := `{
		"request": {"method": "GET", "path": "/api/users/{id}"},
		"response": {"status": 200, "body": "user found"}
	}`

	resp := adminRequest(t, "POST", s.AdminURL()+"/stubs", stubJSON)
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	mockResp, err := http.Get(s.URL() + "/api/users/123")
	if err != nil {
		t.Fatal(err)
	}
	defer mockResp.Body.Close()

	if mockResp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", mockResp.StatusCode)
	}
}
