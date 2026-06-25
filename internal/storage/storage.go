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

	CreateNetwork(ctx context.Context, network models.Network) (*models.Network, error)
	GetNetwork(ctx context.Context, id string) (*models.Network, error)
	ListNetworksForAccount(ctx context.Context, accountID string) ([]models.Network, error)
	SearchNetworks(ctx context.Context, accountID string, query string) ([]models.NetworkSearchResult, error)
	UpdateNetwork(ctx context.Context, network models.Network) (*models.Network, error)
	ArchiveNetwork(ctx context.Context, id string) error
	AddPersonToNetwork(ctx context.Context, context models.NetworkPersonContext) (*models.NetworkPersonContext, error)
	GetNetworkPerson(ctx context.Context, networkID string, personID string) (*models.NetworkPerson, error)
	ListNetworkPersons(ctx context.Context, networkID string) ([]models.NetworkPerson, error)
	UpdateNetworkPersonContext(ctx context.Context, context models.NetworkPersonContext) (*models.NetworkPersonContext, error)
	ArchiveNetworkPerson(ctx context.Context, networkID string, personID string) error
	SaveNetworkPosition(ctx context.Context, networkID string, position models.Position) error
	GetNetworkGraph(ctx context.Context, networkID string) (*models.NetworkGraphResponse, error)
	CreateCustomRelationshipType(ctx context.Context, relationshipType models.CustomRelationshipType) (*models.CustomRelationshipType, error)
	ListCustomRelationshipTypes(ctx context.Context, networkID string) ([]models.CustomRelationshipType, error)
	ListPersonNetworkIDs(ctx context.Context, personID string) ([]string, error)
	MergePersons(ctx context.Context, survivorID string, removedID string) (*models.Person, error)

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

	CreateProject(ctx context.Context, project models.Project) (*models.Project, error)
	GetProject(ctx context.Context, id string) (*models.Project, error)
	ListProjects(ctx context.Context) ([]models.Project, error)
	UpdateProject(ctx context.Context, project models.Project) (*models.Project, error)
	DeleteProject(ctx context.Context, id string) error

	CreateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error)
	GetRelationship(ctx context.Context, id string) (*models.Relationship, error)
	ListRelationships(ctx context.Context) ([]models.Relationship, error)
	DeleteRelationship(ctx context.Context, id string) error

	SavePosition(ctx context.Context, position models.Position) error
	GetGraph(ctx context.Context) (*models.GraphResponse, error)
}
