# MATE API

## General

Base URL:

```
/api/v1
```

All requests and responses use JSON.

The web frontend is served by the same Go application.

Errors use this JSON shape:

```json
{
  "error": "invalid input"
}
```

---

## Web frontend

### GET /

Returns the MATE HTML application shell.

### GET /static/

Serves frontend assets from `web/static`, including CSS, JavaScript, images, fonts, and icons.

---

## Health

### GET /health

Returns server health information.

Response:

```json
{
  "status": "ok"
}
```

---

## Setup And Auth

### GET /api/v1/setup/status

Returns whether the first owner account needs to be created.

Response:

```json
{
  "needs_owner": true
}
```

### POST /api/v1/setup/owner

Creates the first owner account. This returns `409` once an account already exists.

Request:

```json
{
  "email": "owner@example.com",
  "display_name": "Owner",
  "password": "minimum-8-chars"
}
```

### POST /api/v1/auth/login

Logs in with email and password. On success, the server sets a HTTP-only `mate_session` cookie.

Request:

```json
{
  "email": "owner@example.com",
  "password": "minimum-8-chars"
}
```

Response:

```json
{
  "account": {
    "id": "acct-123",
    "email": "owner@example.com",
    "display_name": "Owner",
    "role": "owner",
    "disabled": false
  },
  "expires_at": "2026-06-24T12:00:00Z"
}
```

### POST /api/v1/auth/logout

Deletes the current session and expires the session cookie.

### GET /api/v1/auth/me

Returns the authenticated account tied to the current session cookie.

---

## Accounts

Account management requires an authenticated owner session.

Supported roles are `owner`, `admin`, `editor`, and `viewer`.

### GET /api/v1/accounts

Returns all accounts.

### POST /api/v1/accounts

Creates an account.

Request:

```json
{
  "email": "viewer@example.com",
  "display_name": "Viewer",
  "password": "minimum-8-chars",
  "role": "viewer"
}
```

### GET /api/v1/accounts/{id}

Returns one account.

### PATCH /api/v1/accounts/{id}

Updates an account role.

Request:

```json
{
  "role": "editor"
}
```

### POST /api/v1/accounts/{id}/disable

Disables an account.

---

## Persons

### GET /api/v1/persons

Returns all persons.

### GET /api/v1/persons/{id}

Returns a specific person.

### POST /api/v1/persons

Creates a new person.

Request:

```json
{
  "name": "John Doe",
  "nickname": "Johnny",
  "gender": "m",
  "title": "Board member",
  "notes": "Met at school"
}
```

Supported gender values are `m`, `f`, and `o`. Leave the field empty if it is unknown or not set.

### PUT /api/v1/persons/{id}

Updates a person.

### DELETE /api/v1/persons/{id}

Deletes a person.

### GET /api/v1/persons/{id}/profile

Returns a person profile with attributes and relationships.

### GET /api/v1/persons/{id}/attributes

Returns active profile attributes for a person.

### POST /api/v1/persons/{id}/attributes

Creates a person profile attribute.

Request:

```json
{
  "type": "education",
  "value": "University of Duckburg",
  "organization_id": "org-123",
  "start_date": "1998",
  "end_date": "2002",
  "current": false,
  "notes": "Optional notes"
}
```

Supported person attribute types are `title`, `role`, `employment`, `education`, `certification`, `award`, `board_membership`, `competition`, and `achievement`.

### PUT /api/v1/persons/{id}/attributes/{attributeId}

Updates a person profile attribute.

### POST /api/v1/persons/{id}/attributes/{attributeId}/archive

Archives a person profile attribute.

---

## Positions

### POST /api/v1/positions

Stores a graph position for a node.

Request:

```json
{
  "node_id": "person-123",
  "node_type": "person",
  "x": 120,
  "y": 180
}
```

---

## Organizations

### GET /api/v1/organizations

Returns all organizations.

### GET /api/v1/organizations/{id}

Returns a specific organization.

### POST /api/v1/organizations

Creates a new organization.

Request:

```json
{
  "name": "Example Company",
  "type": "company",
  "description": "A useful organization profile.",
  "web": "https://example.com",
  "notes": "Optional notes"
}
```

Supported organization types are `company`, `association`, and `school`.

### PUT /api/v1/organizations/{id}

Updates an organization.

### DELETE /api/v1/organizations/{id}

Deletes an organization.

### GET /api/v1/organizations/{id}/profile

Returns an organization profile with attributes and relationships.

### GET /api/v1/organizations/{id}/attributes

Returns active profile attributes for an organization.

### POST /api/v1/organizations/{id}/attributes

Creates an organization profile attribute.

Request:

```json
{
  "type": "milestone",
  "value": "Founded",
  "person_id": "person-123",
  "start_date": "1998",
  "end_date": "",
  "current": false,
  "notes": "Optional notes"
}
```

Supported organization attribute types are `role`, `membership`, `board_role`, `certification`, `award`, and `milestone`.

### PUT /api/v1/organizations/{id}/attributes/{attributeId}

Updates an organization profile attribute.

### POST /api/v1/organizations/{id}/attributes/{attributeId}/archive

Archives an organization profile attribute.

---

## Relationships

Supported relationship types are `knows`, `spouse_of`, `parent_of`, `sibling_of`, `works_at`, `member_of`, `studied_at`, `lives_in`, and `has_tag`.

### GET /api/v1/relationships

Returns relationships.

### POST /api/v1/relationships

Creates a relationship.

Request:

```json
{
  "source_id": "person-123",
  "source_type": "person",
  "target_id": "org-456",
  "target_type": "company",
  "type": "works_at",
  "notes": "Current role"
}
```

### GET /api/v1/relationships/{id}

Returns a relationship.

### DELETE /api/v1/relationships/{id}

Deletes a relationship.

---

## Graph

### GET /api/v1/graph

Returns graph data for visualization.

Response:

```json
{
  "persons": [],
  "organizations": [],
  "locations": [],
  "tags": [],
  "relationships": [],
  "positions": []
}
```
