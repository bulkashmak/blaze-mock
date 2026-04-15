package blaze

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type adminHandler struct {
	registry *StubRegistry
	logger   *slog.Logger
}

func newAdminMux(registry *StubRegistry, logger *slog.Logger) http.Handler {
	h := &adminHandler{registry: registry, logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /stubs", h.createStub)
	mux.HandleFunc("GET /stubs", h.listStubs)
	mux.HandleFunc("GET /stubs/{id}", h.getStub)
	mux.HandleFunc("PUT /stubs/{id}", h.updateStub)
	mux.HandleFunc("DELETE /stubs/{id}", h.deleteStub)
	mux.HandleFunc("DELETE /stubs", h.deleteAllStubs)

	return mux
}

func (h *adminHandler) createStub(w http.ResponseWriter, r *http.Request) {
	var dto stubDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	stub, err := dtoToStub(dto)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id := h.registry.Add(stub)
	h.logger.Info("stub created via admin API", "stub_id", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *adminHandler) listStubs(w http.ResponseWriter, r *http.Request) {
	stubs := h.registry.List()
	dtos := make([]stubDTO, len(stubs))
	for i, s := range stubs {
		dtos[i] = stubToDTO(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtos)
}

func (h *adminHandler) getStub(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stub := h.registry.Get(id)
	if stub == nil {
		h.writeError(w, http.StatusNotFound, "stub not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stubToDTO(*stub))
}

func (h *adminHandler) updateStub(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var dto stubDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	stub, err := dtoToStub(dto)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if updated := h.registry.Update(id, stub); updated == "" {
		h.writeError(w, http.StatusNotFound, "stub not found")
		return
	}

	h.logger.Info("stub updated via admin API", "stub_id", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *adminHandler) deleteStub(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.registry.Remove(id) {
		h.writeError(w, http.StatusNotFound, "stub not found")
		return
	}

	h.logger.Info("stub deleted via admin API", "stub_id", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *adminHandler) deleteAllStubs(w http.ResponseWriter, r *http.Request) {
	h.registry.Reset()
	h.logger.Info("all stubs deleted via admin API")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *adminHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
