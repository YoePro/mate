package models

// RelationshipType defines the supported graph relationship types.
type RelationshipType string

const (
	RelationshipKnows       RelationshipType = "knows"
	RelationshipSpouseOf    RelationshipType = "spouse_of"
	RelationshipParentOf    RelationshipType = "parent_of"
	RelationshipWorksAt     RelationshipType = "works_at"
	RelationshipMemberOf    RelationshipType = "member_of"
	RelationshipStudiedAt   RelationshipType = "studied_at"
)

// Relationship represents a connection between two nodes.
type Relationship struct {
	ID     string           `json:"id"`
	FromID string           `json:"fromId"`
	ToID   string           `json:"toId"`
	Type   RelationshipType `json:"type"`
	Role   string           `json:"role,omitempty"`
	From   string           `json:"from,omitempty"`
	To     string           `json:"to,omitempty"`
	Notes  string           `json:"notes,omitempty"`
}
