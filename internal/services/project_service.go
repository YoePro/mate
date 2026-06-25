package services

import (
	"context"

	"mate/internal/models"
	"mate/internal/storage"
)

// ProjectService contains business logic for projects.
type ProjectService struct {
	store storage.Storage
}

// Create creates a project.
func (s *ProjectService) Create(ctx context.Context, project models.Project) (*models.Project, error) {
	project.Name = normalizeSpace(project.Name)
	if project.Name == "" {
		return nil, ErrInvalidInput
	}
	project.Active = true
	if project.ID == "" {
		project.ID = newID("prj")
	}
	return s.store.CreateProject(ctx, project)
}

// List returns all projects.
func (s *ProjectService) List(ctx context.Context) ([]models.Project, error) {
	return s.store.ListProjects(ctx)
}

// Get returns one project.
func (s *ProjectService) Get(ctx context.Context, id string) (*models.Project, error) {
	return s.store.GetProject(ctx, id)
}

// Update updates a project.
func (s *ProjectService) Update(ctx context.Context, id string, project models.Project) (*models.Project, error) {
	project.ID = id
	project.Name = normalizeSpace(project.Name)
	if project.Name == "" {
		return nil, ErrInvalidInput
	}
	project.Active = true
	return s.store.UpdateProject(ctx, project)
}

// Delete archives a project.
func (s *ProjectService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteProject(ctx, id)
}
