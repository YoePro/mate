package handlers

import (
	"net/http"
	"strings"

	"mate/internal/models"
)

// ListPersons returns all persons.
func ListPersons(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// CreatePerson creates a new person.
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Persons handles collection requests for persons.
func (api *API) Persons(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		persons, err := api.services.Persons.List(r.Context())
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, persons)
	case http.MethodPost:
		var person models.Person
		if err := decodeJSON(r, &person); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		created, err := api.services.Persons.Create(r.Context(), person)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Person handles item requests for persons.
func (api *API) Person(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/persons/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]

	if len(parts) > 1 {
		api.personNested(w, r, id, parts[1:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		person, err := api.services.Persons.Get(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, person)
	case http.MethodPut:
		var person models.Person
		if err := decodeJSON(r, &person); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		updated, err := api.services.Persons.Update(r.Context(), id, person)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := api.services.Persons.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) personNested(w http.ResponseWriter, r *http.Request, personID string, parts []string) {
	switch parts[0] {
	case "profile":
		api.personProfile(w, r, personID)
	case "attributes":
		api.personAttributes(w, r, personID, parts[1:])
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (api *API) personProfile(w http.ResponseWriter, r *http.Request, personID string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	profile, err := api.services.Persons.Profile(r.Context(), personID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (api *API) personAttributes(w http.ResponseWriter, r *http.Request, personID string, parts []string) {
	if len(parts) == 0 {
		switch r.Method {
		case http.MethodGet:
			attributes, err := api.services.Persons.ListAttributes(r.Context(), personID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, attributes)
		case http.MethodPost:
			var attribute models.PersonAttribute
			if err := decodeJSON(r, &attribute); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			created, err := api.services.Persons.CreateAttribute(r.Context(), personID, attribute)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, created)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	attributeID := parts[0]
	if attributeID == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if len(parts) == 2 && parts[1] == "archive" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := api.services.Persons.ArchiveAttribute(r.Context(), personID, attributeID); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
		return
	}

	if len(parts) > 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var attribute models.PersonAttribute
	if err := decodeJSON(r, &attribute); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	updated, err := api.services.Persons.UpdateAttribute(r.Context(), personID, attributeID, attribute)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
