package app

import (
	"context"

	"mate/internal/config"
	"mate/internal/server"
	neo4jstorage "mate/internal/storage/neo4j"
)

// Run loads configuration and starts the application.
func Run() error {
	cfg := config.Load()
	ctx := context.Background()

	store, err := neo4jstorage.New(cfg)
	if err != nil {
		return err
	}
	defer store.Close(ctx)

	if err := store.Verify(ctx); err != nil {
		return err
	}

	if err := store.EnsureSchema(ctx); err != nil {
		return err
	}

	return server.Start(cfg, store)
}
