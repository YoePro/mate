package models

// Role defines an account permission level.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Account represents a user account.
type Account struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	Role         Role   `json:"role"`
	Disabled     bool   `json:"disabled"`
	PasswordHash string `json:"-"`
}

// Session represents a persisted login session.
type Session struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	TokenHash string `json:"-"`
	ExpiresAt string `json:"expires_at"`
}

// SetupStatus describes whether initial account setup is needed.
type SetupStatus struct {
	NeedsOwner bool `json:"needs_owner"`
}
