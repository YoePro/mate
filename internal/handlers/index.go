package handlers

import "net/http"

// Index serves the application entry point.
func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "web/html/index.html")
}

// Login serves the login and owner bootstrap page.
func Login(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "web/html/login.html")
}
