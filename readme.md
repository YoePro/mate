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

## Login

MATE 0.7 includes persisted accounts and sessions. Open `/login` after starting the server.

- If no owner exists, `/login` shows the first-owner creation flow.
- After owner bootstrap, sign in with the account name and password.
- The main app at `/` requires a valid session and redirects to `/login` otherwise.
- The header shows the current account and provides `Sign out`.
- Owner account management endpoints exist in the API; a full account-management UI is planned later.
