# Features

This document summarizes implemented product features at the start of 0.12 stabilization.

## Accounts and sessions

- First-owner bootstrap through `/login`.
- Login and logout with HTTP-only session cookie.
- Roles: `owner`, `admin`, `editor`, and `viewer`.
- Owner-only account management API.

## Networks

- Users can create and rename owned networks.
- Users can list owned networks.
- Users can search discoverable network metadata without receiving edit access.
- Network graph data is scoped to the selected owned network.
- Person membership, notes, context, and positions are network-scoped.
- Non-owners cannot edit another user's network.

## Graph entities

- Person.
- Organization with subtypes:
  - company
  - association
  - school
  - government
  - political_party
  - religious_organization
  - sports_club
  - military_unit
  - ngo
  - community
- Project.
- Location and Tag remain legacy placeholders pending a later taxonomy decision.
- Family, Event, and Interest are not implemented yet.

## Relationship behavior

- Built-in relationship types include person, organization, and project-oriented links such as `works_at`, `member_of`, `works_on`, and `sponsors`.
- Relationship type options are filtered in the UI by broad source and target entity type.
- Custom relationship types are network-owned metadata and can be reused.
- Relationship records can store `role`, `start_date`, `end_date`, `current`, and `notes`.

## Graph UI

- Grouped toolbox areas for Navigation, Selection, Graph, and Create.
- Icon-based Create tools for person genders, organization subtypes, and project.
- Person node colors reflect gender:
  - `m`: blue
  - `f`: pink
  - `o` or unknown: neutral
- Ctrl/Cmd-click multi-select.
- Select connected nodes.
- Fit all, fit selection, zoom in, zoom out, and reset zoom.
- Hide selected, show hidden, and auto layout.
- Archive is the normal destructive UI action where supported.

## 0.12 stabilization focus

- Service tests for custom relationship validation, relationship attributes, and network-owner checks.
- Manual smoke testing for grouped tools, relationship type filtering, custom relationship reuse, Project behavior, gender colors, archive behavior, and network permission boundaries.
