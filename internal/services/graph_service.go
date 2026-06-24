package services

import (
	"context"

	"mate/internal/models"
	"mate/internal/storage"
)

// GraphService contains graph-related business logic.
type GraphService struct {
	store storage.Storage
}

// Get returns all graph data.
func (s *GraphService) Get(ctx context.Context) (*models.GraphResponse, error) {
	return s.store.GetGraph(ctx)
}

// SavePosition stores a node position.
func (s *GraphService) SavePosition(ctx context.Context, position models.Position) error {
	if position.NodeID == "" || position.NodeType == "" {
		return ErrInvalidInput
	}
	return s.store.SavePosition(ctx, position)
}
