package handlers

import (
	"net/http"
	"strings"

	"mate/internal/models"
)

// Projects handles collection requests for projects.
func (api *API) Projects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := api.requireDataRead(r); err != nil {
			writeServiceError(w, err)
			return
		}
		projects, err := api.services.Projects.List(r.Context())
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, projects)
	case http.MethodPost:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var project models.Project
		if err := decodeJSON(r, &project); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		created, err := api.services.Projects.Create(r.Context(), project)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Project handles item requests for projects.
func (api *API) Project(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/projects/"), "/")
	if id == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err := api.requireDataRead(r); err != nil {
			writeServiceError(w, err)
			return
		}
		project, err := api.services.Projects.Get(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, project)
	case http.MethodPut:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var project models.Project
		if err := decodeJSON(r, &project); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		updated, err := api.services.Projects.Update(r.Context(), id, project)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		if err := api.services.Projects.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
