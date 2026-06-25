package services

import (
	"context"

	"mate/internal/models"
	"mate/internal/storage"
)

// OrganizationService contains business logic for organizations.
type OrganizationService struct {
	store storage.Storage
}

// Create creates an organization.
func (s *OrganizationService) Create(ctx context.Context, organization models.Organization) (*models.Organization, error) {
	organization.Name = normalizeSpace(organization.Name)
	if organization.Name == "" || !validOrganizationType(organization.Type) {
		return nil, ErrInvalidInput
	}
	organization.Active = true
	if organization.ID == "" {
		organization.ID = newID("org")
	}
	return s.store.CreateOrganization(ctx, organization)
}

// List returns all organizations.
func (s *OrganizationService) List(ctx context.Context) ([]models.Organization, error) {
	return s.store.ListOrganizations(ctx)
}

// Get returns one organization.
func (s *OrganizationService) Get(ctx context.Context, id string) (*models.Organization, error) {
	return s.store.GetOrganization(ctx, id)
}

// Profile returns an organization profile with attributes and relationships.
func (s *OrganizationService) Profile(ctx context.Context, id string) (*models.OrganizationProfile, error) {
	return s.store.GetOrganizationProfile(ctx, id)
}

// Update updates an organization.
func (s *OrganizationService) Update(ctx context.Context, id string, organization models.Organization) (*models.Organization, error) {
	organization.ID = id
	organization.Name = normalizeSpace(organization.Name)
	if organization.Type == "" {
		existing, err := s.store.GetOrganization(ctx, id)
		if err != nil {
			return nil, err
		}
		organization.Type = existing.Type
	}
	if organization.Name == "" || !validOrganizationType(organization.Type) {
		return nil, ErrInvalidInput
	}
	organization.Active = true
	return s.store.UpdateOrganization(ctx, organization)
}

// Delete archives an organization.
func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteOrganization(ctx, id)
}

// ListAttributes returns profile attributes for an organization.
func (s *OrganizationService) ListAttributes(ctx context.Context, organizationID string) ([]models.OrganizationAttribute, error) {
	if _, err := s.store.GetOrganization(ctx, organizationID); err != nil {
		return nil, err
	}
	return s.store.ListOrganizationAttributes(ctx, organizationID)
}

// CreateAttribute creates a profile attribute for an organization.
func (s *OrganizationService) CreateAttribute(ctx context.Context, organizationID string, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error) {
	attribute.OrganizationID = organizationID
	attribute.Value = normalizeSpace(attribute.Value)
	if attribute.Value == "" || !validOrganizationAttributeType(attribute.Type) || !validOrganizationAttributeDates(attribute) {
		return nil, ErrInvalidInput
	}
	if attribute.ID == "" {
		attribute.ID = newID("attr")
	}
	return s.store.CreateOrganizationAttribute(ctx, attribute)
}

// UpdateAttribute updates a profile attribute for an organization.
func (s *OrganizationService) UpdateAttribute(ctx context.Context, organizationID, attributeID string, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error) {
	attribute.ID = attributeID
	attribute.OrganizationID = organizationID
	attribute.Value = normalizeSpace(attribute.Value)
	if attribute.Value == "" || !validOrganizationAttributeType(attribute.Type) || !validOrganizationAttributeDates(attribute) {
		return nil, ErrInvalidInput
	}
	return s.store.UpdateOrganizationAttribute(ctx, attribute)
}

// ArchiveAttribute archives a profile attribute for an organization.
func (s *OrganizationService) ArchiveAttribute(ctx context.Context, organizationID, attributeID string) error {
	return s.store.ArchiveOrganizationAttribute(ctx, organizationID, attributeID)
}

func validOrganizationAttributeType(value models.OrganizationAttributeType) bool {
	switch value {
	case models.OrganizationAttributeRole,
		models.OrganizationAttributeMembership,
		models.OrganizationAttributeBoardRole,
		models.OrganizationAttributeCertification,
		models.OrganizationAttributeAward,
		models.OrganizationAttributeMilestone:
		return true
	default:
		return false
	}
}

func validOrganizationAttributeDates(attribute models.OrganizationAttribute) bool {
	return attribute.StartDate == "" || attribute.EndDate == "" || attribute.StartDate <= attribute.EndDate
}
