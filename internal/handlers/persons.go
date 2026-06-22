package handlers

import "net/http"

// ListPersons returns all persons.
func ListPersons(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// CreatePerson creates a new person.
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
