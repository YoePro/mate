# Storage

MATE uses `internal/storage.Storage` as the persistence boundary.

Runtime storage is Neo4j-backed from version 0.7. Handlers call services, and services call the storage interface. Handlers should not contain database logic.

An in-memory implementation may be added later for isolated unit tests, but it is not the primary runtime storage.

The storage interface includes account/session persistence, graph nodes, relationships, positions, person/organization profile attributes, user-owned networks, network person context, network-scoped positions, duplicate suggestion support, guarded person merge operations, projects, and network-scoped custom relationship type metadata.

## Entity taxonomy

0.11 treats these as top-level graph entities:

- Person
- Organization
- Project

Organization subtypes stay on the `Organization.type` field instead of becoming separate top-level node labels. Supported values are `company`, `association`, `school`, `government`, `political_party`, `religious_organization`, `sports_club`, `military_unit`, `ngo`, and `community`.

Project is a top-level entity because it can be worked on, sponsored, archived, and related to both people and organizations.

Deferred taxonomy decisions:

- Event remains deferred until a concrete workflow needs event nodes instead of attached timeline/context objects.
- Family remains deferred. The UI may show a disabled family icon, but no storage model exists yet.
- Location and Tag remain legacy graph placeholders for now and should be revisited before adding richer behavior.

## Global identity versus network context

Persons are global identity records. A person can appear in multiple user-owned networks.

Network-specific information is stored on the network membership between `Network` and `Person`, not on the global person record:

- notes for this network
- context for why the person belongs in this network
- archive state for this membership

Graph positions are also scoped per network. Moving a person in one network should not move the same global person in another network.

Organizations and projects are currently global nodes. In a network graph they appear when linked to a person in that owned network. This matches the current 0.11 implementation and should be revisited if networks need explicit organization/project membership independent of person relationships.

## Relationship metadata

Relationships can carry structured context:

- `role`
- `start_date`
- `end_date`
- `current`
- `notes`
- `custom_label` for custom relationship types
- optional `network_id` for relationships created inside a selected network

The frontend uses these fields for relationship contexts such as `works_at` with a role like "Developer" or "Chairman".

Network-scoped relationship persistence is important for non-person links such as `Organization -> sponsors -> Project`, because those links cannot be inferred from a network's person memberships.

Custom relationship type definitions are network-scoped. They store the reusable key and label plus broad source/target entity types and direction behavior. The current direction behavior values are `directed` and `undirected`; 0.11 stores the value but does not yet alter graph rendering based on it.

Custom relationship type keys must be safe Neo4j relationship type names and currently use the `custom_` prefix.

## Archive behavior

Archive is the normal destructive action in the UI where supported.

- Network person archive removes the person from that network without deleting the global person.
- Organization and Project delete endpoints archive the node by setting `archived` and `active`.
- Archived organizations and projects are hidden from graph/list responses.
- Global person delete still exists for API compatibility and should be treated as an advanced destructive operation.

## Merge behavior

The storage layer can merge two global person records by moving non-conflicting properties, attributes, network memberships, and supported relationship edges to the surviving person. The service layer decides whether a merge is allowed based on network ownership and ambiguity.
