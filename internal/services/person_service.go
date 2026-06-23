package services

import (
	"context"

	"mate/internal/models"
	"mate/internal/storage"
)

// PersonService contains business logic for persons.
type PersonService struct {
	store storage.Storage
}

// Create creates a person.
func (s *PersonService) Create(ctx context.Context, person models.Person) (*models.Person, error) {
	person.Name = normalizeSpace(person.Name)
	if person.Name == "" || !validGender(person.Gender) {
		return nil, ErrInvalidInput
	}
	if person.ID == "" {
		person.ID = newID("person")
	}
	return s.store.CreatePerson(ctx, person)
}

// List returns all persons.
func (s *PersonService) List(ctx context.Context) ([]models.Person, error) {
	return s.store.ListPersons(ctx)
}

// Get returns one person.
func (s *PersonService) Get(ctx context.Context, id string) (*models.Person, error) {
	return s.store.GetPerson(ctx, id)
}

// Profile returns a person profile with attributes and relationships.
func (s *PersonService) Profile(ctx context.Context, id string) (*models.PersonProfile, error) {
	return s.store.GetPersonProfile(ctx, id)
}

// Update updates a person.
func (s *PersonService) Update(ctx context.Context, id string, person models.Person) (*models.Person, error) {
	person.ID = id
	person.Name = normalizeSpace(person.Name)
	if person.Name == "" || !validGender(person.Gender) {
		return nil, ErrInvalidInput
	}
	return s.store.UpdatePerson(ctx, person)
}

// Delete deletes a person.
func (s *PersonService) Delete(ctx context.Context, id string) error {
	return s.store.DeletePerson(ctx, id)
}

// ListAttributes returns profile attributes for a person.
func (s *PersonService) ListAttributes(ctx context.Context, personID string) ([]models.PersonAttribute, error) {
	return s.store.ListPersonAttributes(ctx, personID)
}

// CreateAttribute creates a profile attribute for a person.
func (s *PersonService) CreateAttribute(ctx context.Context, personID string, attribute models.PersonAttribute) (*models.PersonAttribute, error) {
	attribute.PersonID = personID
	attribute.Value = normalizeSpace(attribute.Value)
	if attribute.Value == "" || !validPersonAttributeType(attribute.Type) || !validAttributeDates(attribute) {
		return nil, ErrInvalidInput
	}
	if attribute.ID == "" {
		attribute.ID = newID("attr")
	}
	return s.store.CreatePersonAttribute(ctx, attribute)
}

// UpdateAttribute updates a profile attribute for a person.
func (s *PersonService) UpdateAttribute(ctx context.Context, personID, attributeID string, attribute models.PersonAttribute) (*models.PersonAttribute, error) {
	attribute.ID = attributeID
	attribute.PersonID = personID
	attribute.Value = normalizeSpace(attribute.Value)
	if attribute.Value == "" || !validPersonAttributeType(attribute.Type) || !validAttributeDates(attribute) {
		return nil, ErrInvalidInput
	}
	return s.store.UpdatePersonAttribute(ctx, attribute)
}

// ArchiveAttribute archives a profile attribute for a person.
func (s *PersonService) ArchiveAttribute(ctx context.Context, personID, attributeID string) error {
	return s.store.ArchivePersonAttribute(ctx, personID, attributeID)
}

func validGender(value string) bool {
	switch value {
	case "", "m", "f", "o":
		return true
	default:
		return false
	}
}

func validPersonAttributeType(value models.PersonAttributeType) bool {
	switch value {
	case models.PersonAttributeTitle,
		models.PersonAttributeRole,
		models.PersonAttributeEmployment,
		models.PersonAttributeEducation,
		models.PersonAttributeCertification,
		models.PersonAttributeAward,
		models.PersonAttributeBoardMembership,
		models.PersonAttributeCompetition,
		models.PersonAttributeAchievement:
		return true
	default:
		return false
	}
}

func validAttributeDates(attribute models.PersonAttribute) bool {
	return attribute.StartDate == "" || attribute.EndDate == "" || attribute.StartDate <= attribute.EndDate
}
