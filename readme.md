# MATE

Memories Alliances Time Engagement

A personal relationship mapping application built with:

- Go
- Neo4j
- Vanilla HTML
- Vanilla CSS
- Vanilla JavaScript

The project is intended to function as a private CRM for friends, family, organizations and shared experiences.

Configuration:

- Create a local `.mate.ini` or `mate.ini` file to override the default settings.
- The default server address is `0.0.0.0:8325`, which listens on all network interfaces.
- Neo4j must be running for the 0.7 backend API.
- Keep credentials private; `.mate.ini` is included in `.gitignore`.
- Run the server with `go run ./cmd/mate`.
