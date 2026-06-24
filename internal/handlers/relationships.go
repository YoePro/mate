package handlers

import (
	"net/http"
	"strings"

	"mate/internal/models"
)

// Relationships handles collection requests for relationships.
func (api *API) Relationships(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := api.requireDataRead(r); err != nil {
			writeServiceError(w, err)
			return
		}
		relationships, err := api.services.Relationships.List(r.Context())
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, relationships)
	case http.MethodPost:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var relationship models.Relationship
		if err := decodeJSON(r, &relationship); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		created, err := api.services.Relationships.Create(r.Context(), relationship)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Relationship handles item requests for relationships.
func (api *API) Relationship(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/relationships/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err := api.requireDataRead(r); err != nil {
			writeServiceError(w, err)
			return
		}
		relationship, err := api.services.Relationships.Get(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, relationship)
	case http.MethodDelete:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		if err := api.services.Relationships.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
