package models

// DiagramNode represents a network-scoped diagram-only node.
type DiagramNode struct {
	ID          string `json:"id"`
	NetworkID   string `json:"network_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
}
