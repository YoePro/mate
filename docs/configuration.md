# Configuration

MATE reads configuration from `.mate.ini` first and then `mate.ini` if `.mate.ini` is absent.

## Server

```ini
[server]
address = 0.0.0.0:8325
```

The default address listens on all network interfaces.

## Neo4j

```ini
[neo4j]
uri = neo4j://localhost:7687
user = neo4j
password = your-local-password
database =
```

Leave `database` empty to use Neo4j's default database.

Keep `.mate.ini` private. It is ignored by git.
