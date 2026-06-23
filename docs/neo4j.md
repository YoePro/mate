# Neo4j

MATE 0.7 uses Neo4j as the primary runtime storage.

## Local requirements

- Neo4j must be running before MATE starts.
- The configured user must be able to create constraints and indexes.
- Set the password in `.mate.ini`.

## Current schema

MATE creates constraints for:

- `Person.id`
- `Organization.id`
- `Position.node_id` and `Position.node_type`

Current graph labels:

- `Person`
- `Organization`
- `Position`

Relationship types are created from the supported MATE relationship types, such as `knows`, `works_at`, and `member_of`.

## Unavailable database behavior

MATE verifies Neo4j connectivity during startup. If Neo4j is unavailable or credentials are wrong, startup fails with the Neo4j driver error.
