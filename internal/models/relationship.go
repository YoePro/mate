package models

// RelationshipType defines the supported graph relationship types.
type RelationshipType string

const (
	RelationshipKnows     RelationshipType = "knows"
	RelationshipSpouseOf  RelationshipType = "spouse_of"
	RelationshipParentOf  RelationshipType = "parent_of"
	RelationshipSiblingOf RelationshipType = "sibling_of"
	RelationshipWorksAt   RelationshipType = "works_at"
	RelationshipMemberOf  RelationshipType = "member_of"
	RelationshipStudiedAt RelationshipType = "studied_at"
	RelationshipLivesIn   RelationshipType = "lives_in"
	RelationshipHasTag    RelationshipType = "has_tag"
	RelationshipWorksOn   RelationshipType = "works_on"
	RelationshipSponsors  RelationshipType = "sponsors"
	RelationshipPartnerOf RelationshipType = "partner_of"
	RelationshipOwns      RelationshipType = "owns"
)

// Relationship represents a connection between two nodes.
type Relationship struct {
	ID          string           `json:"id"`
	NetworkID   string           `json:"network_id,omitempty"`
	SourceID    string           `json:"source_id"`
	SourceType  string           `json:"source_type"`
	TargetID    string           `json:"target_id"`
	TargetType  string           `json:"target_type"`
	Type        RelationshipType `json:"type"`
	CustomLabel string           `json:"custom_label,omitempty"`
	Role        string           `json:"role,omitempty"`
	StartDate   string           `json:"start_date,omitempty"`
	EndDate     string           `json:"end_date,omitempty"`
	Current     bool             `json:"current,omitempty"`
	Notes       string           `json:"notes,omitempty"`
}
