package models

// CustomRelationshipType stores a reusable relationship type scoped to a network.
type CustomRelationshipType struct {
	ID                string `json:"id"`
	NetworkID         string `json:"network_id"`
	OwnerID           string `json:"owner_id"`
	Key               string `json:"key"`
	Label             string `json:"label"`
	SourceType        string `json:"source_type"`
	TargetType        string `json:"target_type"`
	DirectionBehavior string `json:"direction_behavior"`
	Archived          bool   `json:"archived,omitempty"`
}
