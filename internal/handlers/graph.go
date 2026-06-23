package handlers

import (
	"encoding/json"
	"net/http"

	"mate/internal/models"
)

// GetGraph returns graph data.
func GetGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"nodes": []any{},
		"links": []any{},
	})
}

// Graph handles graph data requests.
func (api *API) Graph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	graph, err := api.services.Graph.Get(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

// Positions handles graph position updates.
func (api *API) Positions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var position models.Position
	if err := decodeJSON(r, &position); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := api.services.Graph.SavePosition(r.Context(), position); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
