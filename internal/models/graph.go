package models

// Position stores a frontend graph position for a node.
type Position struct {
	NodeID   string  `json:"node_id"`
	NodeType string  `json:"node_type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// GraphResponse contains the data needed by the frontend graph.
type GraphResponse struct {
	Persons       []Person         `json:"persons"`
	Organizations []Organization   `json:"organizations"`
	Projects      []Project        `json:"projects"`
	DiagramNodes  []DiagramNode    `json:"diagram_nodes,omitempty"`
	Locations     []map[string]any `json:"locations"`
	Tags          []map[string]any `json:"tags"`
	Relationships []Relationship   `json:"relationships"`
	Positions     []Position       `json:"positions"`
}
