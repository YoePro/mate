package storage

import (
	"context"

	"mate/internal/models"
)

// Storage defines all persistence operations used by MATE.
type Storage interface {
	Close(ctx context.Context) error

	SetupStatus(ctx context.Context) (*models.SetupStatus, error)
	CreateAccount(ctx context.Context, account models.Account) (*models.Account, error)
	GetAccount(ctx context.Context, id string) (*models.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*models.Account, error)
	ListAccounts(ctx context.Context) ([]models.Account, error)
	UpdateAccountRole(ctx context.Context, id string, role models.Role) (*models.Account, error)
	DisableAccount(ctx context.Context, id string) (*models.Account, error)
	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error

	CreatePerson(ctx context.Context, person models.Person) (*models.Person, error)
	GetPerson(ctx context.Context, id string) (*models.Person, error)
	ListPersons(ctx context.Context) ([]models.Person, error)
	UpdatePerson(ctx context.Context, person models.Person) (*models.Person, error)
	DeletePerson(ctx context.Context, id string) error
	GetPersonProfile(ctx context.Context, id string) (*models.PersonProfile, error)
	CreatePersonAttribute(ctx context.Context, attribute models.PersonAttribute) (*models.PersonAttribute, error)
	ListPersonAttributes(ctx context.Context, personID string) ([]models.PersonAttribute, error)
	UpdatePersonAttribute(ctx context.Context, attribute models.PersonAttribute) (*models.PersonAttribute, error)
	ArchivePersonAttribute(ctx context.Context, personID string, attributeID string) error

	CreateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error)
	GetOrganization(ctx context.Context, id string) (*models.Organization, error)
	ListOrganizations(ctx context.Context) ([]models.Organization, error)
	UpdateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error)
	DeleteOrganization(ctx context.Context, id string) error
	GetOrganizationProfile(ctx context.Context, id string) (*models.OrganizationProfile, error)
	CreateOrganizationAttribute(ctx context.Context, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error)
	ListOrganizationAttributes(ctx context.Context, organizationID string) ([]models.OrganizationAttribute, error)
	UpdateOrganizationAttribute(ctx context.Context, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error)
	ArchiveOrganizationAttribute(ctx context.Context, organizationID string, attributeID string) error

	CreateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error)
	GetRelationship(ctx context.Context, id string) (*models.Relationship, error)
	ListRelationships(ctx context.Context) ([]models.Relationship, error)
	DeleteRelationship(ctx context.Context, id string) error

	SavePosition(ctx context.Context, position models.Position) error
	GetGraph(ctx context.Context) (*models.GraphResponse, error)
}
