package models

// OrganizationType defines what kind of organization the node represents.
type OrganizationType string

const (
	OrganizationCompany     OrganizationType = "company"
	OrganizationAssociation OrganizationType = "association"
	OrganizationSchool      OrganizationType = "school"
)

// Organization represents a company, association, or school.
type Organization struct {
	ID    string           `json:"id"`
	Name  string           `json:"name"`
	Type  OrganizationType `json:"type"`
	Notes string           `json:"notes,omitempty"`
}
