package blaze

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type mockHandler struct {
	registry *StubRegistry
	logger   *slog.Logger
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	h.logger.Info("request received",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery,
		"headers", flatHeaders(r.Header),
		"body", string(body),
	)

	stub, params := h.registry.Match(r, body)
	if stub == nil {
		h.logger.Info("no stub matched",
			"method", r.Method,
			"path", r.URL.Path,
		)
		h.writeNoMatch(w, r, body)
		return
	}

	h.logger.Info("stub matched", "stub_id", stub.ID)

	// Store buffered body in context so Req() can access it
	r = r.WithContext(context.WithValue(r.Context(), requestBodyKey{}, body))
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
		h.logResponse(stub.ID, status, headers, respBody)
		writeResponse(w, status, headers, respBody)
		return
	}

	if resp.BodyFile != "" {
		data, err := os.ReadFile(resp.BodyFile)
		if err != nil {
			http.Error(w, "failed to read body file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logResponse(stub.ID, resp.Status, resp.Headers, data)
		writeResponse(w, resp.Status, resp.Headers, data)
		return
	}

	h.logResponse(stub.ID, resp.Status, resp.Headers, resp.Body)
	writeResponse(w, resp.Status, resp.Headers, resp.Body)
}

func (h *mockHandler) logResponse(stubID string, status int, headers map[string]string, body []byte) {
	h.logger.Info("response sent",
		"stub_id", stubID,
		"status", status,
		"headers", headers,
		"body", string(body),
	)
}

func flatHeaders(h http.Header) map[string]string {
	flat := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) == 1 {
			flat[k] = v[0]
		} else {
			flat[k] = fmt.Sprintf("%v", v)
		}
	}
	return flat
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
			Path:   fmt.Sprintf("%s", s.Request.Path),
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
