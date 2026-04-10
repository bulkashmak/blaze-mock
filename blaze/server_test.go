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

// startServer creates a server on a random port, starts it in a goroutine, and returns it.
func startServer(t *testing.T) *blaze.Server {
	t.Helper()
	s := blaze.NewServer()

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			// Can't t.Fatal from a goroutine, just log
			t.Logf("server error: %v", err)
		}
	}()

	// Wait for the server to be ready
	for i := 0; i < 50; i++ {
		if s.URL() != "" {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start in time")
	return nil
}

func TestStaticResponse(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Post("/api/payments").
			WithHeader("Content-Type", blaze.EqualTo("application/json")).
			WithBodyContaining(`"amount"`).
			WillReturn(
				blaze.Response(201).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"id": "pay_123", "status": "created"}`),
			),
	)

	body := strings.NewReader(`{"amount": 100}`)
	req, _ := http.NewRequest("POST", s.URL()+"/api/payments", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(data, &result)
	if result["id"] != "pay_123" {
		t.Errorf("expected pay_123, got %s", result["id"])
	}
}

func TestDynamicResponse(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Get("/api/payments/{id}").
			WillRespondWith(func(r *http.Request) blaze.Resp {
				id := blaze.PathParam(r, "id")
				return blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyJSON(map[string]string{"id": id, "status": "created"})
			}),
	)

	resp, err := http.Get(s.URL() + "/api/payments/pay_456")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(data, &result)
	if result["id"] != "pay_456" {
		t.Errorf("expected pay_456, got %s", result["id"])
	}
}

func TestBodyFile(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "users.json")
	os.WriteFile(fixture, []byte(`[{"name": "alice"}]`), 0644)

	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Get("/api/users").
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyFile(fixture),
			),
	)

	resp, err := http.Get(s.URL() + "/api/users")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	if string(data) != `[{"name": "alice"}]` {
		t.Errorf("unexpected body: %s", data)
	}
}

func TestNoMatch404(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Get("/api/users").
			WillReturn(blaze.Response(200).WithBody("ok")),
	)

	resp, err := http.Get(s.URL() + "/api/unknown")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	var diagnostic map[string]any
	json.Unmarshal(data, &diagnostic)
	if diagnostic["message"] != "no matching stub found" {
		t.Errorf("unexpected diagnostic: %s", data)
	}
}

func TestQueryParamMatching(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Get("/api/search").
			WithQueryParam("q", blaze.EqualTo("blaze")).
			WillReturn(blaze.Response(200).WithBody("found")),
	)

	resp, err := http.Get(s.URL() + "/api/search?q=blaze")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Non-matching query
	resp2, _ := http.Get(s.URL() + "/api/search?q=other")
	defer resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp2.StatusCode)
	}
}

func TestRemoveAndResetStubs(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	id := s.Stub(
		blaze.Get("/api/a").WillReturn(blaze.Response(200).WithBody("a")),
	)
	s.Stub(
		blaze.Get("/api/b").WillReturn(blaze.Response(200).WithBody("b")),
	)

	if len(s.ListStubs()) != 2 {
		t.Fatalf("expected 2 stubs, got %d", len(s.ListStubs()))
	}

	s.RemoveStub(id)
	if len(s.ListStubs()) != 1 {
		t.Fatalf("expected 1 stub after remove, got %d", len(s.ListStubs()))
	}

	s.ResetStubs()
	if len(s.ListStubs()) != 0 {
		t.Fatalf("expected 0 stubs after reset, got %d", len(s.ListStubs()))
	}
}

func TestResponseDelay(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	s.Stub(
		blaze.Get("/api/slow").
			WillReturn(
				blaze.Response(200).
					WithBody("slow").
					WithDelay(100*time.Millisecond),
			),
	)

	start := time.Now()
	resp, err := http.Get(s.URL() + "/api/slow")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if elapsed < 100*time.Millisecond {
		t.Errorf("expected at least 100ms delay, got %v", elapsed)
	}
}

func TestInsertionOrderMatching(t *testing.T) {
	s := startServer(t)
	defer s.Shutdown()

	// Register two stubs for the same path — first one should win
	s.Stub(
		blaze.Get("/api/order").WillReturn(blaze.Response(200).WithBody("first")),
	)
	s.Stub(
		blaze.Get("/api/order").WillReturn(blaze.Response(200).WithBody("second")),
	)

	resp, err := http.Get(s.URL() + "/api/order")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if string(data) != "first" {
		t.Errorf("expected 'first' (insertion order), got '%s'", data)
	}
}
