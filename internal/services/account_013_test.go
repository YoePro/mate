package services

import (
	"context"
	"errors"
	"testing"

	"mate/internal/models"
)

func TestAccountService_AdminCanCreateNonAdminAccount(t *testing.T) {
	store := &serviceTestStore{}
	service := &AccountService{store: store}
	actor := &models.Account{ID: "admin-1", Role: models.RoleAdmin}

	account, err := service.CreateAccount(context.Background(), actor, AccountInput{
		Email:       "editor@example.com",
		DisplayName: "Editor",
		Password:    "password123",
		Role:        models.RoleEditor,
	})
	if err != nil {
		t.Fatalf("expected admin to create editor account, got %v", err)
	}
	if account.PasswordHash != "" {
		t.Fatal("expected public account response to hide password hash")
	}
	if len(store.accounts) != 1 || store.accounts[0].Role != models.RoleEditor {
		t.Fatalf("expected stored editor account, got %#v", store.accounts)
	}
}

func TestAccountService_RegisterCreatesEditorAfterOwnerBootstrap(t *testing.T) {
	store := &serviceTestStore{needsOwner: false}
	service := &AccountService{store: store}

	account, err := service.Register(context.Background(), AccountInput{
		Email:       "self@example.com",
		DisplayName: "Self",
		Password:    "password123",
		Role:        models.RoleOwner,
	})
	if err != nil {
		t.Fatalf("expected self registration to succeed, got %v", err)
	}
	if account.Role != models.RoleEditor {
		t.Fatalf("expected self-registered account to be editor, got %q", account.Role)
	}
}

func TestAccountService_RegisterBlockedBeforeOwnerBootstrap(t *testing.T) {
	store := &serviceTestStore{needsOwner: true}
	service := &AccountService{store: store}

	_, err := service.Register(context.Background(), AccountInput{
		Email:       "self@example.com",
		DisplayName: "Self",
		Password:    "password123",
	})
	if err == nil {
		t.Fatal("expected self registration to fail before owner bootstrap")
	}
}

func TestAccountService_AdminCannotCreateAdminOrOwnerAccount(t *testing.T) {
	store := &serviceTestStore{}
	service := &AccountService{store: store}
	actor := &models.Account{ID: "admin-1", Role: models.RoleAdmin}

	for _, role := range []models.Role{models.RoleAdmin, models.RoleOwner} {
		_, err := service.CreateAccount(context.Background(), actor, AccountInput{
			Email:       string(role) + "@example.com",
			DisplayName: string(role),
			Password:    "password123",
			Role:        role,
		})
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden for role %s, got %v", role, err)
		}
	}
}

func TestAccountService_BlocksDemotingLastUsableOwner(t *testing.T) {
	store := &serviceTestStore{
		accounts: []models.Account{
			{ID: "owner-1", Email: "owner@example.com", Role: models.RoleOwner},
			{ID: "admin-1", Email: "admin@example.com", Role: models.RoleAdmin},
		},
	}
	service := &AccountService{store: store}
	actor := &models.Account{ID: "owner-1", Role: models.RoleOwner}

	_, err := service.UpdateAccountRole(context.Background(), actor, "owner-1", models.RoleAdmin)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for last owner demotion, got %v", err)
	}
}

func TestAccountService_BlocksDisablingLastUsableOwner(t *testing.T) {
	store := &serviceTestStore{
		accounts: []models.Account{
			{ID: "owner-1", Email: "owner@example.com", Role: models.RoleOwner},
			{ID: "owner-2", Email: "disabled-owner@example.com", Role: models.RoleOwner, Disabled: true},
		},
	}
	service := &AccountService{store: store}
	actor := &models.Account{ID: "admin-1", Role: models.RoleAdmin}

	_, err := service.DisableAccount(context.Background(), actor, "owner-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected admin to be forbidden from disabling owner, got %v", err)
	}

	actor = &models.Account{ID: "owner-2", Role: models.RoleOwner}
	_, err = service.DisableAccount(context.Background(), actor, "owner-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for last owner disable, got %v", err)
	}
}
