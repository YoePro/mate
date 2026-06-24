package handlers

import (
	"net/http"
	"strings"

	"mate/internal/models"
)

// ListOrganizations returns all organizations.
func ListOrganizations(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// CreateOrganization creates a new organization.
func CreateOrganization(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Organizations handles collection requests for organizations.
func (api *API) Organizations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := api.requireDataRead(r); err != nil {
			writeServiceError(w, err)
			return
		}
		organizations, err := api.services.Organizations.List(r.Context())
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, organizations)
	case http.MethodPost:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var organization models.Organization
		if err := decodeJSON(r, &organization); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		created, err := api.services.Organizations.Create(r.Context(), organization)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Organization handles item requests for organizations.
func (api *API) Organization(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]

	if len(parts) > 1 {
		api.organizationNested(w, r, id, parts[1:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err := api.requireDataRead(r); err != nil {
			writeServiceError(w, err)
			return
		}
		organization, err := api.services.Organizations.Get(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, organization)
	case http.MethodPut:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var organization models.Organization
		if err := decodeJSON(r, &organization); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		updated, err := api.services.Organizations.Update(r.Context(), id, organization)
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
		if err := api.services.Organizations.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) organizationNested(w http.ResponseWriter, r *http.Request, organizationID string, parts []string) {
	switch parts[0] {
	case "profile":
		api.organizationProfile(w, r, organizationID)
	case "attributes":
		api.organizationAttributes(w, r, organizationID, parts[1:])
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (api *API) organizationProfile(w http.ResponseWriter, r *http.Request, organizationID string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := api.requireDataRead(r); err != nil {
		writeServiceError(w, err)
		return
	}
	profile, err := api.services.Organizations.Profile(r.Context(), organizationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (api *API) organizationAttributes(w http.ResponseWriter, r *http.Request, organizationID string, parts []string) {
	if len(parts) == 0 {
		switch r.Method {
		case http.MethodGet:
			if err := api.requireDataRead(r); err != nil {
				writeServiceError(w, err)
				return
			}
			attributes, err := api.services.Organizations.ListAttributes(r.Context(), organizationID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, attributes)
		case http.MethodPost:
			if err := api.requireDataWrite(r); err != nil {
				writeServiceError(w, err)
				return
			}
			var attribute models.OrganizationAttribute
			if err := decodeJSON(r, &attribute); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			created, err := api.services.Organizations.CreateAttribute(r.Context(), organizationID, attribute)
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
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		if err := api.services.Organizations.ArchiveAttribute(r.Context(), organizationID, attributeID); err != nil {
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
	var attribute models.OrganizationAttribute
	if err := decodeJSON(r, &attribute); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	updated, err := api.services.Organizations.UpdateAttribute(r.Context(), organizationID, attributeID, attribute)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
