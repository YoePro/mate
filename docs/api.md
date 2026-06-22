# MATE API

## General

Base URL:

```
/api/v1
```

All requests and responses use JSON.

The web frontend is served by the same Go application.

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
  "fullName": "John Doe"
}
```

### PUT /api/v1/persons/{id}

Updates a person.

### DELETE /api/v1/persons/{id}

Deletes a person.

---

## Organizations

### GET /api/v1/organizations

Returns all organizations.

### GET /api/v1/organizations/{id}

Returns a specific organization.

### POST /api/v1/organizations

Creates a new organization.

### PUT /api/v1/organizations/{id}

Updates an organization.

### DELETE /api/v1/organizations/{id}

Deletes an organization.

---

## Relationships

### GET /api/v1/relationships

Returns relationships.

### POST /api/v1/relationships

Creates a relationship.

### DELETE /api/v1/relationships/{id}

Deletes a relationship.

---

## Graph

### GET /api/v1/graph

Returns graph data for visualization.

Response:

```json
{
  "nodes": [],
  "links": []
}
```
