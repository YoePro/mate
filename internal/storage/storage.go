package storage

import (
	"context"

	"mate/internal/models"
)

// Storage defines all persistence operations used by MATE.
type Storage interface {
	Close(ctx context.Context) error

	CreatePerson(ctx context.Context, person models.Person) (*models.Person, error)
	GetPerson(ctx context.Context, id string) (*models.Person, error)
	ListPersons(ctx context.Context) ([]models.Person, error)

	CreateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error)
	GetOrganization(ctx context.Context, id string) (*models.Organization, error)
	ListOrganizations(ctx context.Context) ([]models.Organization, error)

	CreateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error)
	ListRelationshipsForPerson(ctx context.Context, personID string) ([]models.Relationship, error)
}
