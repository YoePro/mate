package models

// Person represents a human contact in MATE.
type Person struct {
	ID        string   `json:"id"`
	FullName  string   `json:"fullName"`
	NickName  string   `json:"nickName,omitempty"`
	Notes     string   `json:"notes,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}
