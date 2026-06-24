package server

import (
	"net/http"

	"mate/internal/config"
	"mate/internal/handlers"
	"mate/internal/services"
	"mate/internal/storage"
)

// Start starts the HTTP server.
func Start(cfg config.Config, store storage.Storage) error {
	mux := createRouter(store)

	return http.ListenAndServe(cfg.ServerAddress, mux)
}

func createRouter(store storage.Storage) *http.ServeMux {
	mux := http.NewServeMux()
	api := handlers.NewAPI(services.New(store))

	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/login", handlers.Login)

	mux.HandleFunc("/api/v1/setup/status", api.SetupStatus)
	mux.HandleFunc("/api/v1/setup/owner", api.SetupOwner)
	mux.HandleFunc("/api/v1/auth/login", api.Login)
	mux.HandleFunc("/api/v1/auth/logout", api.Logout)
	mux.HandleFunc("/api/v1/auth/me", api.Me)
	mux.HandleFunc("/api/v1/accounts", api.Accounts)
	mux.HandleFunc("/api/v1/accounts/", api.Account)
	mux.HandleFunc("/api/v1/networks", api.Networks)
	mux.HandleFunc("/api/v1/networks/", api.Network)
	mux.HandleFunc("/api/v1/person-matches", api.PersonMatches)
	mux.HandleFunc("/api/v1/persons", api.Persons)
	mux.HandleFunc("/api/v1/persons/", api.Person)
	mux.HandleFunc("/api/v1/organizations", api.Organizations)
	mux.HandleFunc("/api/v1/organizations/", api.Organization)
	mux.HandleFunc("/api/v1/relationships", api.Relationships)
	mux.HandleFunc("/api/v1/relationships/", api.Relationship)
	mux.HandleFunc("/api/v1/positions", api.Positions)
	mux.HandleFunc("/api/v1/graph", api.Graph)

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
