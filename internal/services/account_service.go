package services

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"mate/internal/models"
	"mate/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 24 * time.Hour

// AccountInput contains account fields accepted by account services.
type AccountInput struct {
	Email       string
	DisplayName string
	Password    string
	Role        models.Role
}

// LoginResult contains a successful login response and session token.
type LoginResult struct {
	Account   models.Account
	Token     string
	ExpiresAt time.Time
}

// AccountService contains account and authentication logic.
type AccountService struct {
	store storage.Storage
}

// SetupStatus returns whether the first owner account still needs bootstrapping.
func (s *AccountService) SetupStatus(ctx context.Context) (*models.SetupStatus, error) {
	return s.store.SetupStatus(ctx)
}

// BootstrapOwner creates the first owner account.
func (s *AccountService) BootstrapOwner(ctx context.Context, input AccountInput) (*models.Account, error) {
	status, err := s.store.SetupStatus(ctx)
	if err != nil {
		return nil, err
	}
	if !status.NeedsOwner {
		return nil, storage.ErrConflict
	}
	input.Role = models.RoleOwner
	return s.createAccount(ctx, input)
}

// Register creates a self-service account after the first owner exists.
func (s *AccountService) Register(ctx context.Context, input AccountInput) (*models.Account, error) {
	status, err := s.store.SetupStatus(ctx)
	if err != nil {
		return nil, err
	}
	if status.NeedsOwner {
		return nil, storage.ErrConflict
	}
	input.Role = models.RoleEditor
	return s.createAccount(ctx, input)
}

// Login verifies credentials and creates a session.
func (s *AccountService) Login(ctx context.Context, email string, password string) (*LoginResult, error) {
	account, err := s.store.GetAccountByEmail(ctx, normalizeEmail(email))
	if err != nil {
		return nil, ErrUnauthorized
	}
	if account.Disabled {
		return nil, ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) != nil {
		return nil, ErrUnauthorized
	}

	token := newSessionToken()
	expiresAt := time.Now().UTC().Add(sessionDuration)
	session := models.Session{
		ID:        newID("sess"),
		AccountID: account.ID,
		TokenHash: tokenHash(token),
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}
	if _, err := s.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return &LoginResult{Account: *publicAccount(*account), Token: token, ExpiresAt: expiresAt}, nil
}

// Logout deletes a persisted session.
func (s *AccountService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, tokenHash(token))
}

// CurrentAccount returns the account tied to a session token.
func (s *AccountService) CurrentAccount(ctx context.Context, token string) (*models.Account, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}
	session, err := s.store.GetSessionByTokenHash(ctx, tokenHash(token))
	if err != nil {
		return nil, ErrUnauthorized
	}
	expiresAt, err := time.Parse(time.RFC3339, session.ExpiresAt)
	if err != nil || time.Now().UTC().After(expiresAt) {
		_ = s.store.DeleteSession(ctx, tokenHash(token))
		return nil, ErrUnauthorized
	}
	account, err := s.store.GetAccount(ctx, session.AccountID)
	if err != nil || account.Disabled {
		return nil, ErrUnauthorized
	}
	return publicAccount(*account), nil
}

// CreateAccount creates an account. Owners can create any role; admins can create non-admin accounts.
func (s *AccountService) CreateAccount(ctx context.Context, actor *models.Account, input AccountInput) (*models.Account, error) {
	if !canManageAccounts(actor) {
		return nil, ErrForbidden
	}
	if hasRole(actor, models.RoleAdmin) && (input.Role == models.RoleOwner || input.Role == models.RoleAdmin) {
		return nil, ErrForbidden
	}
	return s.createAccount(ctx, input)
}

// ListAccounts returns all accounts. Owners and admins can list accounts.
func (s *AccountService) ListAccounts(ctx context.Context, actor *models.Account) ([]models.Account, error) {
	if !canManageAccounts(actor) {
		return nil, ErrForbidden
	}
	accounts, err := s.store.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		accounts[i].PasswordHash = ""
	}
	return accounts, nil
}

// GetAccount returns one account. Owners and admins can inspect accounts.
func (s *AccountService) GetAccount(ctx context.Context, actor *models.Account, id string) (*models.Account, error) {
	if !canManageAccounts(actor) {
		return nil, ErrForbidden
	}
	account, err := s.store.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return publicAccount(*account), nil
}

// UpdateAccountRole updates an account role.
func (s *AccountService) UpdateAccountRole(ctx context.Context, actor *models.Account, id string, role models.Role) (*models.Account, error) {
	if !canManageAccounts(actor) {
		return nil, ErrForbidden
	}
	if !validRole(role) {
		return nil, ErrInvalidInput
	}
	target, err := s.store.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	if hasRole(actor, models.RoleAdmin) && (target.Role == models.RoleOwner || target.Role == models.RoleAdmin || role == models.RoleOwner || role == models.RoleAdmin) {
		return nil, ErrForbidden
	}
	if target.Role == models.RoleOwner && role != models.RoleOwner {
		if err := s.requireAnotherUsableOwner(ctx, id); err != nil {
			return nil, err
		}
	}
	account, err := s.store.UpdateAccountRole(ctx, id, role)
	if err != nil {
		return nil, err
	}
	return publicAccount(*account), nil
}

// DisableAccount disables an account.
func (s *AccountService) DisableAccount(ctx context.Context, actor *models.Account, id string) (*models.Account, error) {
	if !canManageAccounts(actor) {
		return nil, ErrForbidden
	}
	if actor.ID == id {
		return nil, ErrInvalidInput
	}
	target, err := s.store.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	if hasRole(actor, models.RoleAdmin) && (target.Role == models.RoleOwner || target.Role == models.RoleAdmin) {
		return nil, ErrForbidden
	}
	if target.Role == models.RoleOwner {
		if err := s.requireAnotherUsableOwner(ctx, id); err != nil {
			return nil, err
		}
	}
	account, err := s.store.DisableAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return publicAccount(*account), nil
}

func (s *AccountService) createAccount(ctx context.Context, input AccountInput) (*models.Account, error) {
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = normalizeSpace(input.DisplayName)
	if input.Role == "" {
		input.Role = models.RoleViewer
	}
	if input.Email == "" || input.DisplayName == "" || input.Password == "" || !validRole(input.Role) {
		return nil, ErrInvalidInput
	}
	if len(input.Password) < 8 {
		return nil, ErrInvalidInput
	}
	if _, err := s.store.GetAccountByEmail(ctx, input.Email); err == nil {
		return nil, storage.ErrConflict
	} else if !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	account := models.Account{
		ID:           newID("acct"),
		Email:        input.Email,
		DisplayName:  input.DisplayName,
		Role:         input.Role,
		PasswordHash: string(hash),
	}
	created, err := s.store.CreateAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	return publicAccount(*created), nil
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validRole(role models.Role) bool {
	switch role {
	case models.RoleOwner, models.RoleAdmin, models.RoleEditor, models.RoleViewer:
		return true
	default:
		return false
	}
}

func hasRole(account *models.Account, role models.Role) bool {
	return account != nil && subtle.ConstantTimeCompare([]byte(account.Role), []byte(role)) == 1
}

func canManageAccounts(account *models.Account) bool {
	return hasRole(account, models.RoleOwner) || hasRole(account, models.RoleAdmin)
}

func (s *AccountService) requireAnotherUsableOwner(ctx context.Context, targetID string) error {
	accounts, err := s.store.ListAccounts(ctx)
	if err != nil {
		return err
	}
	usableOwners := 0
	for _, account := range accounts {
		if account.ID != targetID && account.Role == models.RoleOwner && !account.Disabled {
			usableOwners++
		}
	}
	if usableOwners == 0 {
		return ErrInvalidInput
	}
	return nil
}

func publicAccount(account models.Account) *models.Account {
	account.PasswordHash = ""
	return &account
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
