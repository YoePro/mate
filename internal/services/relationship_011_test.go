package services

import (
	"context"
	"errors"
	"testing"

	"mate/internal/models"
	"mate/internal/storage"
)

func TestValidRelationshipType_AllowsBuiltInsAndCustomTypes(t *testing.T) {
	valid := []models.RelationshipType{
		models.RelationshipKnows,
		models.RelationshipWorksAt,
		models.RelationshipWorksOn,
		models.RelationshipSponsors,
		models.RelationshipPartnerOf,
		models.RelationshipOwns,
		models.RelationshipType("custom_hates"),
		models.RelationshipType("custom_advises_2026"),
	}

	for _, typ := range valid {
		if !validRelationshipType(typ) {
			t.Fatalf("expected %q to be valid", typ)
		}
	}

	invalid := []models.RelationshipType{
		"",
		"custom_",
		"custom_1bad",
		"custom_BAD",
		"custom_bad-name",
		"unknown",
	}

	for _, typ := range invalid {
		if validRelationshipType(typ) {
			t.Fatalf("expected %q to be invalid", typ)
		}
	}
}

func TestRelationshipService_CreateTrimsAndPersistsStructuredAttributes(t *testing.T) {
	store := &serviceTestStore{}
	service := &RelationshipService{store: store}

	created, err := service.Create(context.Background(), models.Relationship{
		SourceID:    "person-1",
		SourceType:  "person",
		TargetID:    "org-1",
		TargetType:  "company",
		Type:        models.RelationshipWorksAt,
		CustomLabel: " ",
		Role:        " Developer ",
		StartDate:   " 2024 ",
		EndDate:     " ",
		Current:     true,
		Notes:       " Current role ",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected service to assign relationship id")
	}
	if store.createdRelationship.Role != "Developer" {
		t.Fatalf("expected trimmed role, got %q", store.createdRelationship.Role)
	}
	if store.createdRelationship.StartDate != "2024" {
		t.Fatalf("expected trimmed start date, got %q", store.createdRelationship.StartDate)
	}
	if store.createdRelationship.EndDate != "" {
		t.Fatalf("expected empty trimmed end date, got %q", store.createdRelationship.EndDate)
	}
	if store.createdRelationship.Notes != "Current role" {
		t.Fatalf("expected trimmed notes, got %q", store.createdRelationship.Notes)
	}
	if !store.createdRelationship.Current {
		t.Fatal("expected current flag to persist")
	}
}

func TestRelationshipService_CreateRequiresCustomLabelForCustomTypes(t *testing.T) {
	store := &serviceTestStore{}
	service := &RelationshipService{store: store}

	_, err := service.Create(context.Background(), models.Relationship{
		SourceID:   "person-1",
		SourceType: "person",
		TargetID:   "person-2",
		TargetType: "person",
		Type:       models.RelationshipType("custom_hates"),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRelationshipService_UpdatePreservesEndpointsAndAllowsTypeChange(t *testing.T) {
	store := &serviceTestStore{
		createdRelationship: models.Relationship{
			ID:         "rel-1",
			NetworkID:  "network-1",
			SourceID:   "person-1",
			SourceType: "person",
			TargetID:   "org-1",
			TargetType: "company",
			Type:       models.RelationshipWorksAt,
		},
	}
	service := &RelationshipService{store: store}

	updated, err := service.Update(context.Background(), "rel-1", models.Relationship{
		SourceID:   "other-source",
		SourceType: "project",
		TargetID:   "other-target",
		TargetType: "project",
		Type:       models.RelationshipSponsors,
		Role:       " Developer ",
		StartDate:  " 2024 ",
		Current:    true,
		Notes:      " Active role ",
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}
	if updated.SourceID != "person-1" || updated.TargetID != "org-1" {
		t.Fatalf("expected endpoints to be preserved, got %#v", updated)
	}
	if updated.Type != models.RelationshipSponsors {
		t.Fatalf("expected type to be updated, got %#v", updated)
	}
	if updated.Role != "Developer" || updated.StartDate != "2024" || updated.Notes != "Active role" {
		t.Fatalf("expected normalized editable fields, got %#v", updated)
	}
}

func TestNetworkService_CreateCustomRelationshipTypeRequiresNetworkOwner(t *testing.T) {
	store := &serviceTestStore{
		network: models.Network{ID: "network-1", OwnerID: "owner-1", Name: "Owned"},
	}
	service := &NetworkService{store: store}

	created, err := service.CreateCustomRelationshipType(context.Background(),
		&models.Account{ID: "owner-1", Role: models.RoleOwner},
		"network-1",
		models.CustomRelationshipType{
			Key:        " custom_advises ",
			Label:      " advises ",
			SourceType: " person ",
			TargetType: " person ",
		},
	)
	if err != nil {
		t.Fatalf("CreateCustomRelationshipType returned error: %v", err)
	}
	if created.NetworkID != "network-1" {
		t.Fatalf("expected network id to be set, got %q", created.NetworkID)
	}
	if created.OwnerID != "owner-1" {
		t.Fatalf("expected owner id to be set, got %q", created.OwnerID)
	}
	if created.DirectionBehavior != "directed" {
		t.Fatalf("expected default directed behavior, got %q", created.DirectionBehavior)
	}
	if created.Label != "advises" || created.Key != "custom_advises" {
		t.Fatalf("expected trimmed key and label, got %q / %q", created.Key, created.Label)
	}

	_, err = service.CreateCustomRelationshipType(context.Background(),
		&models.Account{ID: "other-1", Role: models.RoleOwner},
		"network-1",
		models.CustomRelationshipType{Key: "custom_hates", Label: "hates", SourceType: "person", TargetType: "person"},
	)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-owner, got %v", err)
	}
}

type serviceTestStore struct {
	needsOwner                bool
	network                   models.Network
	accounts                  []models.Account
	createdRelationship       models.Relationship
	createdCustomRelationship models.CustomRelationshipType
}

func (s *serviceTestStore) Close(ctx context.Context) error { return nil }
func (s *serviceTestStore) SetupStatus(ctx context.Context) (*models.SetupStatus, error) {
	return &models.SetupStatus{NeedsOwner: s.needsOwner}, nil
}
func (s *serviceTestStore) CreateAccount(ctx context.Context, account models.Account) (*models.Account, error) {
	s.accounts = append(s.accounts, account)
	return &account, nil
}
func (s *serviceTestStore) GetAccount(ctx context.Context, id string) (*models.Account, error) {
	for _, account := range s.accounts {
		if account.ID == id {
			return &account, nil
		}
	}
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	for _, account := range s.accounts {
		if account.Email == email {
			return &account, nil
		}
	}
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListAccounts(ctx context.Context) ([]models.Account, error) {
	return append([]models.Account(nil), s.accounts...), nil
}
func (s *serviceTestStore) UpdateAccountRole(ctx context.Context, id string, role models.Role) (*models.Account, error) {
	for i := range s.accounts {
		if s.accounts[i].ID == id {
			s.accounts[i].Role = role
			return &s.accounts[i], nil
		}
	}
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) DisableAccount(ctx context.Context, id string) (*models.Account, error) {
	for i := range s.accounts {
		if s.accounts[i].ID == id {
			s.accounts[i].Disabled = true
			return &s.accounts[i], nil
		}
	}
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) DeleteSession(ctx context.Context, tokenHash string) error { return nil }
func (s *serviceTestStore) CreateNetwork(ctx context.Context, network models.Network) (*models.Network, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetNetwork(ctx context.Context, id string) (*models.Network, error) {
	if s.network.ID == id {
		return &s.network, nil
	}
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListNetworksForAccount(ctx context.Context, accountID string) ([]models.Network, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) SearchNetworks(ctx context.Context, accountID string, query string) ([]models.NetworkSearchResult, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdateNetwork(ctx context.Context, network models.Network) (*models.Network, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ArchiveNetwork(ctx context.Context, id string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) AddPersonToNetwork(ctx context.Context, context models.NetworkPersonContext) (*models.NetworkPersonContext, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetNetworkPerson(ctx context.Context, networkID string, personID string) (*models.NetworkPerson, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListNetworkPersons(ctx context.Context, networkID string) ([]models.NetworkPerson, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdateNetworkPersonContext(ctx context.Context, context models.NetworkPersonContext) (*models.NetworkPersonContext, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ArchiveNetworkPerson(ctx context.Context, networkID string, personID string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) SaveNetworkPosition(ctx context.Context, networkID string, position models.Position) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) GetNetworkGraph(ctx context.Context, networkID string) (*models.NetworkGraphResponse, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) CreateCustomRelationshipType(ctx context.Context, relationshipType models.CustomRelationshipType) (*models.CustomRelationshipType, error) {
	s.createdCustomRelationship = relationshipType
	return &s.createdCustomRelationship, nil
}
func (s *serviceTestStore) ListCustomRelationshipTypes(ctx context.Context, networkID string) ([]models.CustomRelationshipType, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListPersonNetworkIDs(ctx context.Context, personID string) ([]string, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) MergePersons(ctx context.Context, survivorID string, removedID string) (*models.Person, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) CreatePerson(ctx context.Context, person models.Person) (*models.Person, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetPerson(ctx context.Context, id string) (*models.Person, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListPersons(ctx context.Context) ([]models.Person, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdatePerson(ctx context.Context, person models.Person) (*models.Person, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) DeletePerson(ctx context.Context, id string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) GetPersonProfile(ctx context.Context, id string) (*models.PersonProfile, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) CreatePersonAttribute(ctx context.Context, attribute models.PersonAttribute) (*models.PersonAttribute, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListPersonAttributes(ctx context.Context, personID string) ([]models.PersonAttribute, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdatePersonAttribute(ctx context.Context, attribute models.PersonAttribute) (*models.PersonAttribute, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ArchivePersonAttribute(ctx context.Context, personID string, attributeID string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) CreateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListOrganizations(ctx context.Context) ([]models.Organization, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) DeleteOrganization(ctx context.Context, id string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) GetOrganizationProfile(ctx context.Context, id string) (*models.OrganizationProfile, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) CreateOrganizationAttribute(ctx context.Context, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListOrganizationAttributes(ctx context.Context, organizationID string) ([]models.OrganizationAttribute, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdateOrganizationAttribute(ctx context.Context, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ArchiveOrganizationAttribute(ctx context.Context, organizationID string, attributeID string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) CreateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) GetProject(ctx context.Context, id string) (*models.Project, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListProjects(ctx context.Context) ([]models.Project, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) DeleteProject(ctx context.Context, id string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) CreateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error) {
	s.createdRelationship = relationship
	return &s.createdRelationship, nil
}
func (s *serviceTestStore) GetRelationship(ctx context.Context, id string) (*models.Relationship, error) {
	if s.createdRelationship.ID == id {
		return &s.createdRelationship, nil
	}
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) ListRelationships(ctx context.Context) ([]models.Relationship, error) {
	return nil, storage.ErrNotFound
}
func (s *serviceTestStore) UpdateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error) {
	s.createdRelationship = relationship
	return &s.createdRelationship, nil
}
func (s *serviceTestStore) DeleteRelationship(ctx context.Context, id string) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) SavePosition(ctx context.Context, position models.Position) error {
	return storage.ErrNotFound
}
func (s *serviceTestStore) GetGraph(ctx context.Context) (*models.GraphResponse, error) {
	return nil, storage.ErrNotFound
}
