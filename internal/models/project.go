package models

// Project represents a work, initiative, or shared effort in the graph.
type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status,omitempty"`
	Description string   `json:"description,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Active      bool     `json:"active"`
	Web         string   `json:"web,omitempty"`
	Archived    bool     `json:"archived,omitempty"`
}
