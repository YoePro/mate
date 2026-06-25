package services

import (
	"context"

	"mate/internal/models"
	"mate/internal/storage"
)

// RelationshipService contains relationship business logic.
type RelationshipService struct {
	store storage.Storage
}

// Create creates a relationship.
func (s *RelationshipService) Create(ctx context.Context, relationship models.Relationship) (*models.Relationship, error) {
	relationship.CustomLabel = normalizeSpace(relationship.CustomLabel)
	relationship.Role = normalizeSpace(relationship.Role)
	relationship.StartDate = normalizeSpace(relationship.StartDate)
	relationship.EndDate = normalizeSpace(relationship.EndDate)
	relationship.Notes = normalizeSpace(relationship.Notes)
	if relationship.SourceID == "" || relationship.TargetID == "" || !validRelationshipType(relationship.Type) {
		return nil, ErrInvalidInput
	}
	if validCustomRelationshipType(string(relationship.Type)) && relationship.CustomLabel == "" {
		return nil, ErrInvalidInput
	}
	if relationship.ID == "" {
		relationship.ID = newID("rel")
	}
	return s.store.CreateRelationship(ctx, relationship)
}

// List returns all relationships.
func (s *RelationshipService) List(ctx context.Context) ([]models.Relationship, error) {
	return s.store.ListRelationships(ctx)
}

// Get returns one relationship.
func (s *RelationshipService) Get(ctx context.Context, id string) (*models.Relationship, error) {
	return s.store.GetRelationship(ctx, id)
}

// Delete deletes a relationship.
func (s *RelationshipService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteRelationship(ctx, id)
}
