package handlers

import (
	"net/http"
	"strings"
	"time"

	"mate/internal/models"
	"mate/internal/services"
)

const sessionCookieName = "mate_session"

type accountRequest struct {
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name"`
	Password    string      `json:"password"`
	Role        models.Role `json:"role"`
}

type roleRequest struct {
	Role models.Role `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Account   models.Account `json:"account"`
	ExpiresAt string         `json:"expires_at"`
}

// SetupStatus returns whether owner bootstrap is needed.
func (api *API) SetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status, err := api.services.Accounts.SetupStatus(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

// SetupOwner bootstraps the first owner account.
func (api *API) SetupOwner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req accountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	account, err := api.services.Accounts.BootstrapOwner(r.Context(), accountInput(req))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

// Register creates a self-service account after owner bootstrap.
func (api *API) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req accountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	account, err := api.services.Accounts.Register(r.Context(), accountInput(req))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

// Login creates a persisted session and sets the session cookie.
func (api *API) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := api.services.Accounts.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	setSessionCookie(w, result.Token, result.ExpiresAt)
	writeJSON(w, http.StatusOK, loginResponse{
		Account:   result.Account,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	})
}

// Logout removes the current session and expires the session cookie.
func (api *API) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := api.services.Accounts.Logout(r.Context(), sessionToken(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	clearSessionCookie(w)
	writeJSON(w, http.StatusNoContent, nil)
}

// Me returns the authenticated account.
func (api *API) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	account, err := api.currentAccount(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, account)
}

// Accounts handles account collection requests.
func (api *API) Accounts(w http.ResponseWriter, r *http.Request) {
	actor, err := api.currentAccount(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		accounts, err := api.services.Accounts.ListAccounts(r.Context(), actor)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, accounts)
	case http.MethodPost:
		var req accountRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		account, err := api.services.Accounts.CreateAccount(r.Context(), actor, accountInput(req))
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, account)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Account handles account item requests.
func (api *API) Account(w http.ResponseWriter, r *http.Request) {
	actor, err := api.currentAccount(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]

	if len(parts) == 2 && parts[1] == "disable" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		account, err := api.services.Accounts.DisableAccount(r.Context(), actor, id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, account)
		return
	}
	if len(parts) > 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		account, err := api.services.Accounts.GetAccount(r.Context(), actor, id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, account)
	case http.MethodPatch:
		var req roleRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		account, err := api.services.Accounts.UpdateAccountRole(r.Context(), actor, id, req.Role)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, account)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) currentAccount(r *http.Request) (*models.Account, error) {
	return api.services.Accounts.CurrentAccount(r.Context(), sessionToken(r))
}

func accountInput(req accountRequest) services.AccountInput {
	return services.AccountInput{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		Role:        req.Role,
	}
}

func sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
