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

// Update updates relationship properties. If the relationship type changes,
// storage replaces the underlying Neo4j edge while preserving the MATE id.
func (s *RelationshipService) Update(ctx context.Context, id string, relationship models.Relationship) (*models.Relationship, error) {
	relationship.ID = id
	relationship.CustomLabel = normalizeSpace(relationship.CustomLabel)
	relationship.Role = normalizeSpace(relationship.Role)
	relationship.StartDate = normalizeSpace(relationship.StartDate)
	relationship.EndDate = normalizeSpace(relationship.EndDate)
	relationship.Notes = normalizeSpace(relationship.Notes)
	if relationship.ID == "" {
		return nil, ErrInvalidInput
	}
	existing, err := s.store.GetRelationship(ctx, id)
	if err != nil {
		return nil, err
	}
	relationship.NetworkID = existing.NetworkID
	relationship.SourceID = existing.SourceID
	relationship.SourceType = existing.SourceType
	relationship.TargetID = existing.TargetID
	relationship.TargetType = existing.TargetType
	if relationship.Type == "" {
		relationship.Type = existing.Type
	}
	if !validRelationshipType(relationship.Type) {
		return nil, ErrInvalidInput
	}
	if validCustomRelationshipType(string(relationship.Type)) && relationship.CustomLabel == "" {
		return nil, ErrInvalidInput
	}
	return s.store.UpdateRelationship(ctx, relationship)
}

// Delete deletes a relationship.
func (s *RelationshipService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteRelationship(ctx, id)
}
