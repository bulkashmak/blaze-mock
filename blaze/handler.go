package blaze

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type mockHandler struct {
	registry *StubRegistry
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	stub, params := h.registry.Match(r, body)
	if stub == nil {
		h.writeNoMatch(w, r, body)
		return
	}

	r = withPathParams(r, params)
	resp := &stub.Response

	if resp.Delay > 0 {
		time.Sleep(resp.Delay)
	}

	// Resolution order: BodyFunc > BodyFile > Body
	if resp.BodyFunc != nil {
		status, headers, respBody, err := resp.BodyFunc(r)
		if err != nil {
			http.Error(w, "dynamic response error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeResponse(w, status, headers, respBody)
		return
	}

	if resp.BodyFile != "" {
		data, err := os.ReadFile(resp.BodyFile)
		if err != nil {
			http.Error(w, "failed to read body file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeResponse(w, resp.Status, resp.Headers, data)
		return
	}

	writeResponse(w, resp.Status, resp.Headers, resp.Body)
}

func writeResponse(w http.ResponseWriter, status int, headers map[string]string, body []byte) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	if len(body) > 0 {
		w.Write(body)
	}
}

func (h *mockHandler) writeNoMatch(w http.ResponseWriter, r *http.Request, body []byte) {
	stubs := h.registry.List()

	type stubInfo struct {
		ID     string `json:"id"`
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	registered := make([]stubInfo, len(stubs))
	for i, s := range stubs {
		registered[i] = stubInfo{
			ID:     s.ID,
			Method: s.Request.Method,
			Path:   fmt.Sprintf("%v", s.Request.Path),
		}
	}

	diagnostic := map[string]any{
		"message": "no matching stub found",
		"request": map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		},
		"registered_stubs": registered,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(diagnostic)
}
