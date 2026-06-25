package models

// OrganizationType defines what kind of organization the node represents.
type OrganizationType string

const (
	OrganizationCompany               OrganizationType = "company"
	OrganizationAssociation           OrganizationType = "association"
	OrganizationSchool                OrganizationType = "school"
	OrganizationGovernment            OrganizationType = "government"
	OrganizationPoliticalParty        OrganizationType = "political_party"
	OrganizationReligiousOrganization OrganizationType = "religious_organization"
	OrganizationSportsClub            OrganizationType = "sports_club"
	OrganizationMilitaryUnit          OrganizationType = "military_unit"
	OrganizationNGO                   OrganizationType = "ngo"
	OrganizationCommunity             OrganizationType = "community"
)

// Organization represents an organization subtype such as company, school, or club.
type Organization struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        OrganizationType `json:"type"`
	Description string           `json:"description,omitempty"`
	Notes       string           `json:"notes,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Aliases     []string         `json:"aliases,omitempty"`
	Active      bool             `json:"active"`
	Web         string           `json:"web,omitempty"`
	Archived    bool             `json:"archived,omitempty"`
}

// OrganizationAttributeType defines time-bound organization attribute categories.
type OrganizationAttributeType string

const (
	OrganizationAttributeRole          OrganizationAttributeType = "role"
	OrganizationAttributeMembership    OrganizationAttributeType = "membership"
	OrganizationAttributeBoardRole     OrganizationAttributeType = "board_role"
	OrganizationAttributeCertification OrganizationAttributeType = "certification"
	OrganizationAttributeAward         OrganizationAttributeType = "award"
	OrganizationAttributeMilestone     OrganizationAttributeType = "milestone"
)

// OrganizationAttribute represents a profile attribute for an organization.
type OrganizationAttribute struct {
	ID             string                    `json:"id"`
	OrganizationID string                    `json:"organization_id"`
	Type           OrganizationAttributeType `json:"type"`
	Value          string                    `json:"value"`
	PersonID       string                    `json:"person_id,omitempty"`
	StartDate      string                    `json:"start_date,omitempty"`
	EndDate        string                    `json:"end_date,omitempty"`
	Current        bool                      `json:"current,omitempty"`
	Notes          string                    `json:"notes,omitempty"`
	Archived       bool                      `json:"archived,omitempty"`
}

// OrganizationProfile contains an organization profile with graph context.
type OrganizationProfile struct {
	Organization  Organization            `json:"organization"`
	Attributes    []OrganizationAttribute `json:"attributes"`
	Relationships []Relationship          `json:"relationships"`
}
