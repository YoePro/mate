package server

import (
	"net/http"

	"mate/internal/config"
	"mate/internal/handlers"
)

// Start starts the HTTP server.
func Start(cfg config.Config) error {
	mux := createRouter()

	return http.ListenAndServe(cfg.ServerAddress, mux)
}

func createRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.Health)

	mux.HandleFunc("/api/v1/persons", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListPersons(w, r)
		case http.MethodPost:
			handlers.CreatePerson(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/organizations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListOrganizations(w, r)
		case http.MethodPost:
			handlers.CreateOrganization(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/graph", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.GetGraph(w, r)
	})

	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("web/static")),
		),
	)

	mux.HandleFunc("/", handlers.Index)

	return mux
}
