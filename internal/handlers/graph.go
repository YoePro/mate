package handlers

import (
	"encoding/json"
	"net/http"
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
