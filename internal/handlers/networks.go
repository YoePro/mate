package handlers

import (
	"net/http"
	"strings"

	"mate/internal/models"
)

type networkPersonRequest struct {
	Person  models.Person               `json:"person"`
	Context models.NetworkPersonContext `json:"context"`
}

// Networks handles network collection requests.
func (api *API) Networks(w http.ResponseWriter, r *http.Request) {
	actor, err := api.currentAccount(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		networks, err := api.services.Networks.List(r.Context(), actor)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, networks)
	case http.MethodPost:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var network models.Network
		if err := decodeJSON(r, &network); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		created, err := api.services.Networks.Create(r.Context(), actor, network)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Network handles network item and nested requests.
func (api *API) Network(w http.ResponseWriter, r *http.Request) {
	actor, err := api.currentAccount(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/networks/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if parts[0] == "search" {
		api.networkSearch(w, r, actor, parts[1:])
		return
	}
	networkID := parts[0]

	if len(parts) > 1 {
		api.networkNested(w, r, actor, networkID, parts[1:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		network, err := api.services.Networks.Get(r.Context(), actor, networkID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, network)
	case http.MethodPut:
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var network models.Network
		if err := decodeJSON(r, &network); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		updated, err := api.services.Networks.Update(r.Context(), actor, networkID, network)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) networkSearch(w http.ResponseWriter, r *http.Request, actor *models.Account, parts []string) {
	if len(parts) != 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	results, err := api.services.Networks.Search(r.Context(), actor, r.URL.Query().Get("q"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (api *API) networkNested(w http.ResponseWriter, r *http.Request, actor *models.Account, networkID string, parts []string) {
	switch parts[0] {
	case "archive":
		if len(parts) != 1 || r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		if err := api.services.Networks.Archive(r.Context(), actor, networkID); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	case "graph":
		api.networkGraph(w, r, actor, networkID)
	case "positions":
		api.networkPositions(w, r, actor, networkID)
	case "persons":
		api.networkPersons(w, r, actor, networkID, parts[1:])
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (api *API) networkGraph(w http.ResponseWriter, r *http.Request, actor *models.Account, networkID string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	graph, err := api.services.Networks.Graph(r.Context(), actor, networkID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

func (api *API) networkPositions(w http.ResponseWriter, r *http.Request, actor *models.Account, networkID string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := api.requireDataWrite(r); err != nil {
		writeServiceError(w, err)
		return
	}
	var position models.Position
	if err := decodeJSON(r, &position); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := api.services.Networks.SavePosition(r.Context(), actor, networkID, position); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (api *API) networkPersons(w http.ResponseWriter, r *http.Request, actor *models.Account, networkID string, parts []string) {
	if len(parts) == 0 {
		switch r.Method {
		case http.MethodGet:
			persons, err := api.services.Networks.ListPersons(r.Context(), actor, networkID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, persons)
		case http.MethodPost:
			if err := api.requireDataWrite(r); err != nil {
				writeServiceError(w, err)
				return
			}
			var req networkPersonRequest
			if err := decodeJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			person, err := api.services.Networks.AddPerson(r.Context(), actor, networkID, req.Person, req.Context)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, person)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	personID := parts[0]
	if personID == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		person, err := api.services.Networks.GetPerson(r.Context(), actor, networkID, personID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, person)
		return
	}
	if len(parts) == 2 && parts[1] == "context" {
		if r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := api.requireDataWrite(r); err != nil {
			writeServiceError(w, err)
			return
		}
		var context models.NetworkPersonContext
		if err := decodeJSON(r, &context); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		updated, err := api.services.Networks.UpdatePersonContext(r.Context(), actor, networkID, personID, context)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
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
		if err := api.services.Networks.ArchivePerson(r.Context(), actor, networkID, personID); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}

// PersonMatches handles duplicate person suggestion requests.
func (api *API) PersonMatches(w http.ResponseWriter, r *http.Request) {
	actor, err := api.currentAccount(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req models.PersonMatchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	matches, err := api.services.Networks.MatchPersons(r.Context(), actor, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, matches)
}
