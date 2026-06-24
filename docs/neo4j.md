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
- `Position`
- `Network`
- `NetworkPosition`

Relationship types are created from the supported MATE relationship types, such as `knows`, `works_at`, and `member_of`.

Network relationships:

- `(:Account)-[:OWNS_NETWORK]->(:Network)`
- `(:Network)-[:CONTAINS_PERSON]->(:Person)`
- `(:Network)-[:HAS_POSITION]->(:NetworkPosition)`

`CONTAINS_PERSON` stores network-specific `notes`, `context`, and `archived` properties.

## Unavailable database behavior

MATE verifies Neo4j connectivity during startup. If Neo4j is unavailable or credentials are wrong, startup fails with the Neo4j driver error.

## Authentication data

Accounts are stored as `Account` nodes. Passwords are stored as bcrypt hashes only. Login sessions are stored as `Session` nodes related to accounts and are looked up by a SHA-256 hash of the browser cookie token.
