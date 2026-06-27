# MATE API

## General

Base URL:

```
/api/v1
```

All requests and responses use JSON.

The web frontend is served by the same Go application.

Except for health, setup, login, logout, and static assets, API endpoints require an authenticated `mate_session` cookie. Read endpoints allow any active role. Data write endpoints require `owner`, `admin`, or `editor`. `viewer` is read-only. Account management endpoints require `owner`.

Errors use this JSON shape:

```json
{
  "error": "invalid input"
}
```

---

## Web frontend

### GET /

Returns the MATE HTML application shell. The frontend checks `GET /api/v1/auth/me` during boot and redirects unauthenticated users to `/login`.

### GET /login

Returns the login and first-owner bootstrap page. If no owner exists, the page exposes the owner creation flow. Once an owner exists, the page signs in with `POST /api/v1/auth/login`.

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

Deletes the current session and expires the session cookie. The frontend header uses this endpoint for `Sign out`.

### GET /api/v1/auth/me

Returns the authenticated account tied to the current session cookie.

---

## Accounts

Account management requires an authenticated owner session. These endpoints are API-only until the user admin UI is added.

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

## Networks

Networks are user-owned graph contexts. Users only see their own networks. Network writes require a write-capable role and ownership of the network.

### GET /api/v1/networks

Returns networks owned by the current account.

### POST /api/v1/networks

Creates a network owned by the current account.

Request:

```json
{
  "name": "Family",
  "description": "Close family and shared history"
}
```

### GET /api/v1/networks/search?q={query}

Searches discoverable networks by safe metadata. The response may include networks owned by other accounts, but search results do not grant read or edit access to the network graph.

The query must be at least two characters.

Response:

```json
[
  {
    "id": "network-123",
    "name": "Family",
    "description": "Close family and shared history",
    "owned": true,
    "can_edit": true
  },
  {
    "id": "network-456",
    "name": "Duckburg",
    "description": "Discoverable metadata only",
    "owned": false,
    "can_edit": false
  }
]
```

### GET /api/v1/networks/{id}

Returns one owned network.

### PUT /api/v1/networks/{id}

Updates owned network metadata.

### POST /api/v1/networks/{id}/archive

Archives an owned network.

### GET /api/v1/networks/{id}/graph

Returns graph data scoped to one owned network.

Response:

```json
{
  "network": {
    "id": "network-123",
    "owner_id": "acct-123",
    "name": "Family"
  },
  "persons": [
    {
      "person": {
        "id": "person-123",
        "name": "Kalle Anka",
        "gender": "m"
      },
      "context": {
        "network_id": "network-123",
        "person_id": "person-123",
        "notes": "Network-specific notes"
      }
    }
  ],
  "organizations": [],
  "projects": [],
  "relationships": [],
  "positions": [],
  "custom_relationship_types": []
}
```

### POST /api/v1/networks/{id}/positions

Stores a graph position scoped to one owned network.

Request:

```json
{
  "node_id": "person-123",
  "node_type": "person",
  "x": 120,
  "y": 180
}
```

### GET /api/v1/networks/{id}/relationship-types

Returns custom relationship types owned by this network.

Response:

```json
[
  {
    "id": "reltype-123",
    "network_id": "network-123",
    "owner_id": "acct-123",
    "key": "custom_hates",
    "label": "hates",
    "source_type": "person",
    "target_type": "person",
    "direction_behavior": "directed"
  }
]
```

### POST /api/v1/networks/{id}/relationship-types

Creates or updates a reusable custom relationship type scoped to the owned network. Non-owners cannot create relationship type metadata for another user's network.

Request:

```json
{
  "key": "custom_hates",
  "label": "hates",
  "source_type": "person",
  "target_type": "person",
  "direction_behavior": "directed"
}
```

`key` must start with `custom_` and contain only lowercase ASCII letters, digits, and underscores after that prefix.

### GET /api/v1/networks/{id}/persons

Lists people in an owned network.

### POST /api/v1/networks/{id}/persons

Adds an existing global person to a network or creates a new global person and adds it to the network.

Request for a new person:

```json
{
  "person": {
    "name": "Kalle Anka",
    "gender": "m"
  },
  "context": {
    "notes": "Seen in this network",
    "context": "Family branch"
  }
}
```

Request for an existing person:

```json
{
  "person": {
    "id": "person-123"
  },
  "context": {
    "notes": "Different notes in this network"
  }
}
```

### GET /api/v1/networks/{id}/persons/{personId}

Returns one network person plus network-specific context.

### PUT /api/v1/networks/{id}/persons/{personId}/context

Updates network-specific person context.

### POST /api/v1/networks/{id}/persons/{personId}/archive

Archives the person membership in that network. It does not delete the global person.

---

## Duplicate Suggestions And Merge

### POST /api/v1/person-matches

Returns possible duplicate global persons. Suggestions include confidence and reasons.

Request:

```json
{
  "name": "Kalle Anka",
  "nickname": "Kalle",
  "organization": "Duckburg",
  "school": "",
  "location": "",
  "relationships": []
}
```

Response:

```json
[
  {
    "person": {
      "id": "person-123",
      "name": "Kalle Anka",
      "gender": "m"
    },
    "confidence": 0.9,
    "reasons": ["same name", "same nickname"]
  }
]
```

### POST /api/v1/persons/{id}/merge

Merges a duplicate person into `{id}` when the operation is unambiguous and the current user owns all affected networks.

Request:

```json
{
  "removed_person_id": "person-duplicate"
}
```

The merge preserves the survivor's existing conflicting fields. Empty survivor fields may be filled from the removed person. Network memberships, attributes, and supported relationship edges are moved to the survivor.

Ambiguous cross-network merges return `403` until a reviewed merge policy is implemented.

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

Supported organization types are `company`, `association`, `school`, `government`, `political_party`, `religious_organization`, `sports_club`, `military_unit`, `ngo`, and `community`.

### PUT /api/v1/organizations/{id}

Updates an organization.

### DELETE /api/v1/organizations/{id}

Archives an organization. Archived organizations are hidden from graph/list responses.

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

## Projects

Project is a first-class top-level entity for initiatives that people can work on and organizations can sponsor.

### GET /api/v1/projects

Returns all active projects.

### GET /api/v1/projects/{id}

Returns a specific project.

### POST /api/v1/projects

Creates a new project.

Request:

```json
{
  "name": "MATE",
  "status": "active",
  "description": "Relationship mapping project",
  "web": "https://example.com",
  "notes": "Optional notes"
}
```

### PUT /api/v1/projects/{id}

Updates a project.

### DELETE /api/v1/projects/{id}

Archives a project. Archived projects are hidden from graph/list responses.

---

## Relationships

Supported built-in relationship types are `knows`, `spouse_of`, `parent_of`, `sibling_of`, `works_at`, `member_of`, `studied_at`, `lives_in`, `has_tag`, `works_on`, `sponsors`, `partner_of`, and `owns`.

Custom relationship types are supported when the type key matches `custom_*` validation and the relationship includes `custom_label`. The frontend stores reusable custom relationship type metadata through the network relationship-type endpoints before creating the relationship.

Relationship type choices in the frontend are filtered by broad source/target entity type, such as person-to-person, person-to-organization, organization-to-organization, person-to-project, and organization-to-project.

### GET /api/v1/relationships

Returns relationships.

### POST /api/v1/relationships

Creates a relationship.

Request:

```json
{
  "network_id": "network-123",
  "source_id": "person-123",
  "source_type": "person",
  "target_id": "org-456",
  "target_type": "company",
  "type": "works_at",
  "role": "Developer",
  "start_date": "2024",
  "end_date": "",
  "current": true,
  "notes": "Current role"
}
```

`network_id` is optional for legacy/global calls. When it is present, the current account must own that network and the relationship is returned by that network's graph even when neither endpoint is a person.

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
  "projects": [],
  "locations": [],
  "tags": [],
  "relationships": [],
  "positions": []
}
```
