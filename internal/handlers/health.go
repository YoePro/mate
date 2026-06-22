package handlers

import "net/http"

// Health returns a simple server health response.
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("MATE is running"))
}
