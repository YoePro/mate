package handlers

import "net/http"

// Index serves the application entry point.
func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/index.html")
}
