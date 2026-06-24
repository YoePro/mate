# Storage

MATE uses `internal/storage.Storage` as the persistence boundary.

Runtime storage is Neo4j-backed from version 0.7. Handlers call services, and services call the storage interface. Handlers should not contain database logic.

An in-memory implementation may be added later for isolated unit tests, but it is not the primary runtime storage.

The storage interface includes account/session persistence, graph nodes, relationships, positions, person/organization profile attributes, user-owned networks, network person context, network-scoped positions, duplicate suggestion support, and guarded person merge operations.

## Global identity versus network context

Persons are global identity records. A person can appear in multiple user-owned networks.

Network-specific information is stored on the network membership between `Network` and `Person`, not on the global person record:

- notes for this network
- context for why the person belongs in this network
- archive state for this membership

Graph positions are also scoped per network. Moving a person in one network should not move the same global person in another network.

## Merge behavior

The storage layer can merge two global person records by moving non-conflicting properties, attributes, network memberships, and supported relationship edges to the surviving person. The service layer decides whether a merge is allowed based on network ownership and ambiguity.
