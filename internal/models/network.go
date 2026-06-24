package models

// Network represents a user-owned graph context.
type Network struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
}

// NetworkSearchResult is safe metadata returned by network discovery.
type NetworkSearchResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Owned       bool   `json:"owned"`
	CanEdit     bool   `json:"can_edit"`
}

// NetworkPersonContext stores network-specific context for a global person.
type NetworkPersonContext struct {
	NetworkID string `json:"network_id"`
	PersonID  string `json:"person_id"`
	Notes     string `json:"notes,omitempty"`
	Context   string `json:"context,omitempty"`
	Archived  bool   `json:"archived,omitempty"`
}

// NetworkPerson contains a global person plus network-specific context.
type NetworkPerson struct {
	Person  Person               `json:"person"`
	Context NetworkPersonContext `json:"context"`
}

// NetworkGraphResponse contains graph data scoped to one network.
type NetworkGraphResponse struct {
	Network       Network         `json:"network"`
	Persons       []NetworkPerson `json:"persons"`
	Organizations []Organization  `json:"organizations"`
	Relationships []Relationship  `json:"relationships"`
	Positions     []Position      `json:"positions"`
}

// PersonMatchRequest describes person duplicate search input.
type PersonMatchRequest struct {
	Name          string   `json:"name"`
	Nickname      string   `json:"nickname,omitempty"`
	Organization  string   `json:"organization,omitempty"`
	School        string   `json:"school,omitempty"`
	Location      string   `json:"location,omitempty"`
	Relationship  string   `json:"relationship,omitempty"`
	Relationships []string `json:"relationships,omitempty"`
}

// PersonMatchSuggestion describes a possible duplicate person.
type PersonMatchSuggestion struct {
	Person     Person   `json:"person"`
	Confidence float64  `json:"confidence"`
	Reasons    []string `json:"reasons"`
}

// PersonMergeResult describes the result of a completed merge.
type PersonMergeResult struct {
	Survivor      Person `json:"survivor"`
	RemovedPerson string `json:"removed_person_id"`
}
