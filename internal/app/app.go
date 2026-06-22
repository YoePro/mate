package app

import (
	"mate/internal/config"
	"mate/internal/server"
)

// Run loads configuration and starts the application.
func Run() error {
	cfg := config.Load()

	return server.Start(cfg)
}
