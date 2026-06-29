# Neo4j

MATE 0.7 uses Neo4j as the primary runtime storage.

## Local requirements

- Neo4j must be running before MATE starts.
- The configured user must be able to create constraints and indexes.
- Set the password in `.mate.ini`.

## Current schema

MATE creates constraints for:

- `Account.id`
- `Account.email`
- `Session.id`
- `Session.token_hash`
- `Person.id`
- `PersonAttribute.id`
- `Organization.id`
- `OrganizationAttribute.id`
- `Project.id`
- `CustomRelationshipType.id`
- `CustomRelationshipType.network_id`, `CustomRelationshipType.key`, `CustomRelationshipType.source_type`, and `CustomRelationshipType.target_type`
- `Position.node_id` and `Position.node_type`
- `Network.id`
- `NetworkPosition.network_id`, `NetworkPosition.node_id`, and `NetworkPosition.node_type`

Current graph labels:

- `Account`
- `Session`
- `Person`
- `PersonAttribute`
- `Organization`
- `OrganizationAttribute`
- `Project`
- `DiagramNode`
- `CustomRelationshipType`
- `Position`
- `Network`
- `NetworkPosition`

Relationship types are created from the supported MATE relationship types, such as `knows`, `works_at`, `member_of`, `works_on`, and `sponsors`.

Custom relationship edges use validated dynamic relationship types with a `custom_` prefix, for example `custom_hates`. Reusable metadata for these types is stored as `CustomRelationshipType` nodes scoped to a network.

Network relationships:

- `(:Account)-[:OWNS_NETWORK]->(:Network)`
- `(:Network)-[:CONTAINS_PERSON]->(:Person)`
- `(:Network)-[:HAS_POSITION]->(:NetworkPosition)`
- `(:Network)-[:CONTAINS_DIAGRAM_NODE]->(:DiagramNode)`
- `(:Network)-[:HAS_CUSTOM_RELATIONSHIP_TYPE]->(:CustomRelationshipType)`

`CONTAINS_PERSON` stores network-specific `notes`, `context`, and `archived` properties.

`CONTAINS_DIAGRAM_NODE` scopes diagram-only nodes, such as Flowchart nodes, to one network.

Deleting a `DiagramNode` physically removes the node, its network-scoped relationships, and its saved network position.

`Network` stores:

- `id`
- `owner_id`
- `name`
- `description`
- `domain`
- `archived`

`CustomRelationshipType` stores:

- `network_id`
- `owner_id`
- `key`
- `label`
- `source_type`
- `target_type`
- `direction_behavior`
- `archived`

Relationship edges may store:

- `source_type`
- `target_type`
- `custom_label`
- `role`
- `start_date`
- `end_date`
- `current`
- `notes`

## 0.11 taxonomy decisions

Project is implemented as a first-class `Project` node.

Organization subtypes are stored as `Organization.type`; they are not separate labels.

Event and Family are not implemented as graph nodes in 0.11. Event remains an open workflow decision. Family remains an open model decision and may become a group/entity, network concept, or relationship cluster later.

Location and Tag are still legacy graph placeholders and should be revisited before deeper 1.0 behavior is built on top of them.

## Unavailable database behavior

MATE verifies Neo4j connectivity during startup. If Neo4j is unavailable or credentials are wrong, startup fails with the Neo4j driver error.

## Authentication data

Accounts are stored as `Account` nodes. Passwords are stored as bcrypt hashes only. Login sessions are stored as `Session` nodes related to accounts and are looked up by a SHA-256 hash of the browser cookie token.
