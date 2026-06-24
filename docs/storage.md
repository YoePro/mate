# Storage

MATE uses `internal/storage.Storage` as the persistence boundary.

Runtime storage is Neo4j-backed from version 0.7. Handlers call services, and services call the storage interface. Handlers should not contain database logic.

An in-memory implementation may be added later for isolated unit tests, but it is not the primary runtime storage.

The 0.7 storage interface includes account/session persistence, graph nodes, relationships, positions, and person/organization profile attributes.
