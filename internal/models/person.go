package models

// Person represents a human contact in MATE.
type Person struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Nickname    string   `json:"nickname,omitempty"`
	Gender      string   `json:"gender,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Deceased    bool     `json:"deceased,omitempty"`
}

// PersonAttributeType defines time-bound person attribute categories.
type PersonAttributeType string

const (
	PersonAttributeTitle           PersonAttributeType = "title"
	PersonAttributeRole            PersonAttributeType = "role"
	PersonAttributeEmployment      PersonAttributeType = "employment"
	PersonAttributeEducation       PersonAttributeType = "education"
	PersonAttributeCertification   PersonAttributeType = "certification"
	PersonAttributeAward           PersonAttributeType = "award"
	PersonAttributeBoardMembership PersonAttributeType = "board_membership"
	PersonAttributeCompetition     PersonAttributeType = "competition"
	PersonAttributeAchievement     PersonAttributeType = "achievement"
)

// PersonAttribute represents a profile attribute for a person.
type PersonAttribute struct {
	ID             string              `json:"id"`
	PersonID       string              `json:"person_id"`
	Type           PersonAttributeType `json:"type"`
	Value          string              `json:"value"`
	OrganizationID string              `json:"organization_id,omitempty"`
	StartDate      string              `json:"start_date,omitempty"`
	EndDate        string              `json:"end_date,omitempty"`
	Current        bool                `json:"current,omitempty"`
	Notes          string              `json:"notes,omitempty"`
	Archived       bool                `json:"archived,omitempty"`
}

// PersonProfile contains a person profile with graph context.
type PersonProfile struct {
	Person        Person            `json:"person"`
	Attributes    []PersonAttribute `json:"attributes"`
	Relationships []Relationship    `json:"relationships"`
}
