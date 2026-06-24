package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"mate/internal/models"
	"mate/internal/services"
	"mate/internal/storage"
)

// API contains HTTP handlers for the JSON API.
type API struct {
	services *services.Services
}

// NewAPI creates API handlers.
func NewAPI(services *services.Services) *API {
	return &API{services: services}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func (api *API) requireDataRead(r *http.Request) error {
	_, err := api.currentAccount(r)
	return err
}

func (api *API) requireDataWrite(r *http.Request) error {
	account, err := api.currentAccount(r)
	if err != nil {
		return err
	}
	switch account.Role {
	case models.RoleOwner, models.RoleAdmin, models.RoleEditor:
		return nil
	default:
		return services.ErrForbidden
	}
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, services.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, services.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, storage.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, storage.ErrConflict):
		writeError(w, http.StatusConflict, "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
