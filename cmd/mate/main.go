package main

import (
	"log"

	"mate/internal/app"
)

// main is the application entry point.
func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
