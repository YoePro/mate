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
)

// Relationship represents a connection between two nodes.
type Relationship struct {
	ID         string           `json:"id"`
	SourceID   string           `json:"source_id"`
	SourceType string           `json:"source_type"`
	TargetID   string           `json:"target_id"`
	TargetType string           `json:"target_type"`
	Type       RelationshipType `json:"type"`
	Notes      string           `json:"notes,omitempty"`
}
