package neo4j

import (
	"context"
	"fmt"

	"mate/internal/config"
	"mate/internal/models"
	"mate/internal/storage"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Storage persists MATE data in Neo4j.
type Storage struct {
	driver   driver.DriverWithContext
	database string
}

// New creates a Neo4j storage implementation.
func New(cfg config.Config) (*Storage, error) {
	drv, err := driver.NewDriverWithContext(
		cfg.Neo4jURI,
		driver.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""),
	)
	if err != nil {
		return nil, err
	}

	return &Storage{
		driver:   drv,
		database: cfg.Neo4jDatabase,
	}, nil
}

// Close closes the Neo4j driver.
func (s *Storage) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// Verify verifies Neo4j connectivity.
func (s *Storage) Verify(ctx context.Context) error {
	return s.driver.VerifyConnectivity(ctx)
}

// EnsureSchema creates indexes and constraints required by MATE.
func (s *Storage) EnsureSchema(ctx context.Context) error {
	statements := []string{
		"CREATE CONSTRAINT mate_account_id IF NOT EXISTS FOR (a:Account) REQUIRE a.id IS UNIQUE",
		"CREATE CONSTRAINT mate_account_email IF NOT EXISTS FOR (a:Account) REQUIRE a.email IS UNIQUE",
		"CREATE CONSTRAINT mate_session_id IF NOT EXISTS FOR (s:Session) REQUIRE s.id IS UNIQUE",
		"CREATE CONSTRAINT mate_session_token_hash IF NOT EXISTS FOR (s:Session) REQUIRE s.token_hash IS UNIQUE",
		"CREATE CONSTRAINT mate_person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE",
		"CREATE CONSTRAINT mate_person_attribute_id IF NOT EXISTS FOR (a:PersonAttribute) REQUIRE a.id IS UNIQUE",
		"CREATE CONSTRAINT mate_organization_id IF NOT EXISTS FOR (o:Organization) REQUIRE o.id IS UNIQUE",
		"CREATE CONSTRAINT mate_organization_attribute_id IF NOT EXISTS FOR (a:OrganizationAttribute) REQUIRE a.id IS UNIQUE",
		"CREATE CONSTRAINT mate_project_id IF NOT EXISTS FOR (p:Project) REQUIRE p.id IS UNIQUE",
		"CREATE CONSTRAINT mate_diagram_node_id IF NOT EXISTS FOR (d:DiagramNode) REQUIRE d.id IS UNIQUE",
		"CREATE CONSTRAINT mate_custom_relationship_type_id IF NOT EXISTS FOR (t:CustomRelationshipType) REQUIRE t.id IS UNIQUE",
		"CREATE CONSTRAINT mate_custom_relationship_type_key IF NOT EXISTS FOR (t:CustomRelationshipType) REQUIRE (t.network_id, t.key, t.source_type, t.target_type) IS UNIQUE",
		"CREATE CONSTRAINT mate_position_key IF NOT EXISTS FOR (p:Position) REQUIRE (p.node_id, p.node_type) IS UNIQUE",
		"CREATE CONSTRAINT mate_network_id IF NOT EXISTS FOR (n:Network) REQUIRE n.id IS UNIQUE",
		"CREATE CONSTRAINT mate_network_position_key IF NOT EXISTS FOR (p:NetworkPosition) REQUIRE (p.network_id, p.node_id, p.node_type) IS UNIQUE",
	}

	for _, statement := range statements {
		if _, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
			_, err := tx.Run(ctx, statement, nil)
			return nil, err
		}); err != nil {
			return err
		}
	}

	return nil
}

// SetupStatus returns whether initial owner setup is needed.
func (s *Storage) SetupStatus(ctx context.Context) (*models.SetupStatus, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, "MATCH (a:Account) RETURN count(a) AS count", nil)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		return &models.SetupStatus{NeedsOwner: asInt(record, "count") == 0}, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.SetupStatus), nil
}

// CreateAccount creates an account.
func (s *Storage) CreateAccount(ctx context.Context, account models.Account) (*models.Account, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			CREATE (a:Account {
				id: $id,
				email: $email,
				display_name: $display_name,
				role: $role,
				disabled: false,
				password_hash: $password_hash
			})
			RETURN a.id AS id, a.email AS email, a.display_name AS display_name,
			       a.role AS role, a.disabled AS disabled, a.password_hash AS password_hash`,
			accountParams(account),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		created := accountFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Account), nil
}

// GetAccount returns one account by ID.
func (s *Storage) GetAccount(ctx context.Context, id string) (*models.Account, error) {
	return s.getAccount(ctx, "MATCH (a:Account {id: $id})", map[string]any{"id": id})
}

// GetAccountByEmail returns one account by email.
func (s *Storage) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	return s.getAccount(ctx, "MATCH (a:Account {email: $email})", map[string]any{"email": email})
}

// ListAccounts returns all accounts.
func (s *Storage) ListAccounts(ctx context.Context) ([]models.Account, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, accountReturnQuery("MATCH (a:Account)")+" ORDER BY toLower(a.email)", nil)
		if err != nil {
			return nil, err
		}
		var accounts []models.Account
		for result.Next(ctx) {
			accounts = append(accounts, accountFromRecord(result.Record()))
		}
		return accounts, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Account), nil
}

// UpdateAccountRole updates an account role.
func (s *Storage) UpdateAccountRole(ctx context.Context, id string, role models.Role) (*models.Account, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, accountReturnQuery(`
			MATCH (a:Account {id: $id})
			SET a.role = $role`),
			map[string]any{"id": id, "role": string(role)},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		account := accountFromRecord(record)
		return &account, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Account), nil
}

// DisableAccount disables an account.
func (s *Storage) DisableAccount(ctx context.Context, id string) (*models.Account, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, accountReturnQuery(`
			MATCH (a:Account {id: $id})
			SET a.disabled = true`),
			map[string]any{"id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		account := accountFromRecord(record)
		return &account, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Account), nil
}

// CreateSession stores a login session.
func (s *Storage) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (a:Account {id: $account_id})
			CREATE (a)-[:HAS_SESSION]->(s:Session {
				id: $id,
				account_id: $account_id,
				token_hash: $token_hash,
				expires_at: $expires_at
			})
			RETURN s.id AS id, s.account_id AS account_id,
			       s.token_hash AS token_hash, s.expires_at AS expires_at`,
			sessionParams(session),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := sessionFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Session), nil
}

// GetSessionByTokenHash returns one session by token hash.
func (s *Storage) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (s:Session {token_hash: $token_hash})
			RETURN s.id AS id, s.account_id AS account_id,
			       s.token_hash AS token_hash, s.expires_at AS expires_at`,
			map[string]any{"token_hash": tokenHash},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		session := sessionFromRecord(record)
		return &session, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Session), nil
}

// DeleteSession deletes a login session.
func (s *Storage) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			MATCH (s:Session {token_hash: $token_hash})
			DETACH DELETE s`,
			map[string]any{"token_hash": tokenHash},
		)
		return nil, err
	})
	return err
}

func (s *Storage) getAccount(ctx context.Context, prefix string, params map[string]any) (*models.Account, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, accountReturnQuery(prefix), params)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		account := accountFromRecord(record)
		return &account, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Account), nil
}

// CreateNetwork creates a user-owned network.
func (s *Storage) CreateNetwork(ctx context.Context, network models.Network) (*models.Network, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkReturnQuery(`
			MATCH (a:Account {id: $owner_id})
			CREATE (a)-[:OWNS_NETWORK]->(n:Network {
				id: $id,
				owner_id: $owner_id,
				name: $name,
				description: $description,
				domain: $domain,
				archived: false
			})`),
			networkParams(network),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := networkFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Network), nil
}

// GetNetwork returns one network.
func (s *Storage) GetNetwork(ctx context.Context, id string) (*models.Network, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkReturnQuery("MATCH (n:Network {id: $id})"), map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		network := networkFromRecord(record)
		return &network, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Network), nil
}

// ListNetworksForAccount returns non-archived networks owned by an account.
func (s *Storage) ListNetworksForAccount(ctx context.Context, accountID string) ([]models.Network, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkReturnQuery(`
			MATCH (:Account {id: $account_id})-[:OWNS_NETWORK]->(n:Network)
			WHERE coalesce(n.archived, false) = false`)+" ORDER BY toLower(n.name)",
			map[string]any{"account_id": accountID},
		)
		if err != nil {
			return nil, err
		}
		networks := make([]models.Network, 0)
		for result.Next(ctx) {
			networks = append(networks, networkFromRecord(result.Record()))
		}
		return networks, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Network), nil
}

// SearchNetworks returns safe network metadata matching a query.
func (s *Storage) SearchNetworks(ctx context.Context, accountID string, query string) ([]models.NetworkSearchResult, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (n:Network)
			WHERE coalesce(n.archived, false) = false
			  AND (
			    toLower(n.name) CONTAINS toLower($query)
			    OR toLower(coalesce(n.description, '')) CONTAINS toLower($query)
			    OR toLower(coalesce(n.domain, 'social')) CONTAINS toLower($query)
			  )
			RETURN n.id AS id,
			       n.name AS name,
			       n.description AS description,
			       coalesce(n.domain, 'social') AS domain,
			       n.owner_id = $account_id AS owned,
			       n.owner_id = $account_id AS can_edit
			ORDER BY owned DESC, toLower(n.name)
			LIMIT 25`,
			map[string]any{"account_id": accountID, "query": query},
		)
		if err != nil {
			return nil, err
		}
		networks := make([]models.NetworkSearchResult, 0)
		for result.Next(ctx) {
			record := result.Record()
			networks = append(networks, models.NetworkSearchResult{
				ID:          asString(record, "id"),
				Name:        asString(record, "name"),
				Description: asString(record, "description"),
				Domain:      asString(record, "domain"),
				Owned:       asBool(record, "owned"),
				CanEdit:     asBool(record, "can_edit"),
			})
		}
		return networks, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.NetworkSearchResult), nil
}

// UpdateNetwork updates network metadata.
func (s *Storage) UpdateNetwork(ctx context.Context, network models.Network) (*models.Network, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkReturnQuery(`
			MATCH (n:Network {id: $id})
			SET n.name = $name,
			    n.description = $description,
			    n.domain = $domain`),
			networkParams(network),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := networkFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Network), nil
}

// ArchiveNetwork archives a network.
func (s *Storage) ArchiveNetwork(ctx context.Context, id string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (n:Network {id: $id})
			SET n.archived = true
			RETURN count(n) AS archived`,
			map[string]any{"id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "archived") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

// AddPersonToNetwork links a person to a network with network-specific context.
func (s *Storage) AddPersonToNetwork(ctx context.Context, context models.NetworkPersonContext) (*models.NetworkPersonContext, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkPersonContextReturnQuery(`
			MATCH (n:Network {id: $network_id})
			MATCH (p:Person {id: $person_id})
			MERGE (n)-[r:CONTAINS_PERSON]->(p)
			SET r.network_id = $network_id,
			    r.person_id = $person_id,
			    r.notes = $notes,
			    r.context = $context,
			    r.archived = false`),
			networkPersonContextParams(context),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := networkPersonContextFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.NetworkPersonContext), nil
}

// GetNetworkPerson returns one person within a network.
func (s *Storage) GetNetworkPerson(ctx context.Context, networkID string, personID string) (*models.NetworkPerson, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkPersonReturnQuery(`
			MATCH (:Network {id: $network_id})-[r:CONTAINS_PERSON]->(p:Person {id: $person_id})
			WHERE coalesce(r.archived, false) = false`),
			map[string]any{"network_id": networkID, "person_id": personID},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		person := networkPersonFromRecord(record)
		return &person, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.NetworkPerson), nil
}

// ListNetworkPersons lists persons in a network.
func (s *Storage) ListNetworkPersons(ctx context.Context, networkID string) ([]models.NetworkPerson, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkPersonReturnQuery(`
			MATCH (:Network {id: $network_id})-[r:CONTAINS_PERSON]->(p:Person)
			WHERE coalesce(r.archived, false) = false`)+" ORDER BY toLower(p.name)",
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		persons := make([]models.NetworkPerson, 0)
		for result.Next(ctx) {
			persons = append(persons, networkPersonFromRecord(result.Record()))
		}
		return persons, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.NetworkPerson), nil
}

// UpdateNetworkPersonContext updates network-specific person context.
func (s *Storage) UpdateNetworkPersonContext(ctx context.Context, context models.NetworkPersonContext) (*models.NetworkPersonContext, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, networkPersonContextReturnQuery(`
			MATCH (:Network {id: $network_id})-[r:CONTAINS_PERSON]->(:Person {id: $person_id})
			WHERE coalesce(r.archived, false) = false
			SET r.notes = $notes,
			    r.context = $context`),
			networkPersonContextParams(context),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := networkPersonContextFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.NetworkPersonContext), nil
}

// ArchiveNetworkPerson archives a person's membership in a network.
func (s *Storage) ArchiveNetworkPerson(ctx context.Context, networkID string, personID string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Network {id: $network_id})-[r:CONTAINS_PERSON]->(:Person {id: $person_id})
			SET r.archived = true
			RETURN count(r) AS archived`,
			map[string]any{"network_id": networkID, "person_id": personID},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "archived") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

// SaveNetworkPosition stores a node position scoped to a network.
func (s *Storage) SaveNetworkPosition(ctx context.Context, networkID string, position models.Position) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			MATCH (n:Network {id: $network_id})
			MERGE (p:NetworkPosition {network_id: $network_id, node_id: $node_id, node_type: $node_type})
			SET p.x = $x, p.y = $y
			MERGE (n)-[:HAS_POSITION]->(p)`,
			map[string]any{
				"network_id": networkID,
				"node_id":    position.NodeID,
				"node_type":  position.NodeType,
				"x":          position.X,
				"y":          position.Y,
			},
		)
		return nil, err
	})
	return err
}

// GetNetworkGraph returns graph data scoped to one network.
func (s *Storage) GetNetworkGraph(ctx context.Context, networkID string) (*models.NetworkGraphResponse, error) {
	network, err := s.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}
	persons, err := s.ListNetworkPersons(ctx, networkID)
	if err != nil {
		return nil, err
	}
	relationships, err := s.listNetworkRelationships(ctx, networkID)
	if err != nil {
		return nil, err
	}
	organizations, err := s.listNetworkOrganizations(ctx, networkID)
	if err != nil {
		return nil, err
	}
	projects, err := s.listNetworkProjects(ctx, networkID)
	if err != nil {
		return nil, err
	}
	diagramNodes, err := s.listNetworkDiagramNodes(ctx, networkID)
	if err != nil {
		return nil, err
	}
	positions, err := s.listNetworkPositions(ctx, networkID)
	if err != nil {
		return nil, err
	}
	customRelationshipTypes, err := s.ListCustomRelationshipTypes(ctx, networkID)
	if err != nil {
		return nil, err
	}
	return &models.NetworkGraphResponse{
		Network:                 *network,
		Persons:                 persons,
		Organizations:           organizations,
		Projects:                projects,
		DiagramNodes:            diagramNodes,
		Relationships:           relationships,
		Positions:               positions,
		CustomRelationshipTypes: customRelationshipTypes,
	}, nil
}

// CreateCustomRelationshipType creates or updates a custom relationship type scoped to a network.
func (s *Storage) CreateCustomRelationshipType(ctx context.Context, relationshipType models.CustomRelationshipType) (*models.CustomRelationshipType, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (n:Network {id: $network_id})
			MERGE (n)-[:HAS_CUSTOM_RELATIONSHIP_TYPE]->(t:CustomRelationshipType {
				network_id: $network_id,
				key: $key,
				source_type: $source_type,
				target_type: $target_type
			})
			ON CREATE SET t.id = $id,
			              t.owner_id = $owner_id
			SET t.label = $label,
			    t.direction_behavior = $direction_behavior,
			    t.archived = false
			RETURN t.id AS id,
			       t.network_id AS network_id,
			       t.owner_id AS owner_id,
			       t.key AS key,
			       t.label AS label,
			       t.source_type AS source_type,
			       t.target_type AS target_type,
			       t.direction_behavior AS direction_behavior,
			       t.archived AS archived`,
			customRelationshipTypeParams(relationshipType),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := customRelationshipTypeFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.CustomRelationshipType), nil
}

// ListCustomRelationshipTypes lists custom relationship types for a network.
func (s *Storage) ListCustomRelationshipTypes(ctx context.Context, networkID string) ([]models.CustomRelationshipType, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Network {id: $network_id})-[:HAS_CUSTOM_RELATIONSHIP_TYPE]->(t:CustomRelationshipType)
			WHERE coalesce(t.archived, false) = false
			RETURN t.id AS id,
			       t.network_id AS network_id,
			       t.owner_id AS owner_id,
			       t.key AS key,
			       t.label AS label,
			       t.source_type AS source_type,
			       t.target_type AS target_type,
			       t.direction_behavior AS direction_behavior,
			       t.archived AS archived
			ORDER BY toLower(t.label)`,
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		relationshipTypes := make([]models.CustomRelationshipType, 0)
		for result.Next(ctx) {
			relationshipTypes = append(relationshipTypes, customRelationshipTypeFromRecord(result.Record()))
		}
		return relationshipTypes, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.CustomRelationshipType), nil
}

// ListPersonNetworkIDs lists active network memberships for a person.
func (s *Storage) ListPersonNetworkIDs(ctx context.Context, personID string) ([]string, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (n:Network)-[r:CONTAINS_PERSON]->(:Person {id: $person_id})
			WHERE coalesce(n.archived, false) = false AND coalesce(r.archived, false) = false
			RETURN n.id AS network_id
			ORDER BY n.id`,
			map[string]any{"person_id": personID},
		)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0)
		for result.Next(ctx) {
			ids = append(ids, asString(result.Record(), "network_id"))
		}
		return ids, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]string), nil
}

// MergePersons merges removedID into survivorID and deletes the removed person.
func (s *Storage) MergePersons(ctx context.Context, survivorID string, removedID string) (*models.Person, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (survivor:Person {id: $survivor_id})
			MATCH (removed:Person {id: $removed_id})
			SET survivor.nickname = CASE WHEN coalesce(survivor.nickname, '') = '' THEN removed.nickname ELSE survivor.nickname END,
			    survivor.gender = CASE WHEN coalesce(survivor.gender, '') = '' THEN removed.gender ELSE survivor.gender END,
			    survivor.title = CASE WHEN coalesce(survivor.title, '') = '' THEN removed.title ELSE survivor.title END,
			    survivor.description = CASE WHEN coalesce(survivor.description, '') = '' THEN removed.description ELSE survivor.description END,
			    survivor.notes = CASE WHEN coalesce(survivor.notes, '') = '' THEN removed.notes ELSE survivor.notes END,
			    survivor.tags = CASE WHEN size(coalesce(survivor.tags, [])) = 0 THEN removed.tags ELSE survivor.tags END,
			    survivor.deceased = coalesce(survivor.deceased, false) OR coalesce(removed.deceased, false)
			RETURN survivor.id AS id, survivor.name AS name, survivor.nickname AS nickname, survivor.gender AS gender,
			       survivor.title AS title, survivor.description AS description, survivor.notes AS notes,
			       survivor.tags AS tags, survivor.deceased AS deceased`,
			map[string]any{"survivor_id": survivorID, "removed_id": removedID},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		if err := s.rewirePersonRelationships(ctx, tx, survivorID, removedID); err != nil {
			return nil, err
		}
		if err := s.movePersonMergeData(ctx, tx, survivorID, removedID); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx, "MATCH (removed:Person {id: $removed_id}) DETACH DELETE removed", map[string]any{"removed_id": removedID}); err != nil {
			return nil, err
		}
		survivor := personFromRecord(record)
		return &survivor, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Person), nil
}

// CreatePerson creates a person node.
func (s *Storage) CreatePerson(ctx context.Context, person models.Person) (*models.Person, error) {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			CREATE (p:Person {
				id: $id,
				name: $name,
				nickname: $nickname,
				gender: $gender,
				title: $title,
				description: $description,
				notes: $notes,
				tags: $tags,
				deceased: $deceased
			})`,
			personParams(person),
		)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return &person, nil
}

// GetPerson returns one person by ID.
func (s *Storage) GetPerson(ctx context.Context, id string) (*models.Person, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Person {id: $id})
			RETURN p.id AS id, p.name AS name, p.nickname AS nickname, p.gender AS gender,
			       p.title AS title, p.description AS description, p.notes AS notes,
			       p.tags AS tags, p.deceased AS deceased`,
			map[string]any{"id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		person := personFromRecord(record)
		return &person, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Person), nil
}

// ListPersons returns all persons.
func (s *Storage) ListPersons(ctx context.Context) ([]models.Person, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Person)
			RETURN p.id AS id, p.name AS name, p.nickname AS nickname, p.gender AS gender,
			       p.title AS title, p.description AS description, p.notes AS notes,
			       p.tags AS tags, p.deceased AS deceased
			ORDER BY toLower(p.name)`,
			nil,
		)
		if err != nil {
			return nil, err
		}
		var persons []models.Person
		for result.Next(ctx) {
			persons = append(persons, personFromRecord(result.Record()))
		}
		return persons, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Person), nil
}

// UpdatePerson updates a person node.
func (s *Storage) UpdatePerson(ctx context.Context, person models.Person) (*models.Person, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Person {id: $id})
			SET p.name = $name,
			    p.nickname = $nickname,
			    p.gender = $gender,
			    p.title = $title,
			    p.description = $description,
			    p.notes = $notes,
			    p.tags = $tags,
			    p.deceased = $deceased
			RETURN p.id AS id, p.name AS name, p.nickname AS nickname, p.gender AS gender,
			       p.title AS title, p.description AS description, p.notes AS notes,
			       p.tags AS tags, p.deceased AS deceased`,
			personParams(person),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := personFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Person), nil
}

// DeletePerson deletes a person node and its relationships.
func (s *Storage) DeletePerson(ctx context.Context, id string) error {
	return s.deleteNode(ctx, "Person", id)
}

// GetPersonProfile returns a person with attributes and relationships.
func (s *Storage) GetPersonProfile(ctx context.Context, id string) (*models.PersonProfile, error) {
	person, err := s.GetPerson(ctx, id)
	if err != nil {
		return nil, err
	}
	attributes, err := s.ListPersonAttributes(ctx, id)
	if err != nil {
		return nil, err
	}
	relationships, err := s.listRelationshipsForNode(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.PersonProfile{
		Person:        *person,
		Attributes:    attributes,
		Relationships: relationships,
	}, nil
}

// CreatePersonAttribute creates a person profile attribute.
func (s *Storage) CreatePersonAttribute(ctx context.Context, attribute models.PersonAttribute) (*models.PersonAttribute, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Person {id: $person_id})
			CREATE (p)-[:HAS_ATTRIBUTE]->(a:PersonAttribute {
				id: $id,
				person_id: $person_id,
				type: $type,
				value: $value,
				organization_id: $organization_id,
				start_date: $start_date,
				end_date: $end_date,
				current: $current,
				notes: $notes,
				archived: false
			})
			RETURN a.id AS id, a.person_id AS person_id, a.type AS type, a.value AS value,
			       a.organization_id AS organization_id, a.start_date AS start_date,
			       a.end_date AS end_date, a.current AS current, a.notes AS notes,
			       a.archived AS archived`,
			personAttributeParams(attribute),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := personAttributeFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.PersonAttribute), nil
}

// ListPersonAttributes lists non-archived attributes for a person.
func (s *Storage) ListPersonAttributes(ctx context.Context, personID string) ([]models.PersonAttribute, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Person {id: $person_id})-[:HAS_ATTRIBUTE]->(a:PersonAttribute)
			WHERE coalesce(a.archived, false) = false
			RETURN a.id AS id, a.person_id AS person_id, a.type AS type, a.value AS value,
			       a.organization_id AS organization_id, a.start_date AS start_date,
			       a.end_date AS end_date, a.current AS current, a.notes AS notes,
			       a.archived AS archived
			ORDER BY coalesce(a.start_date, ''), toLower(a.value)`,
			map[string]any{"person_id": personID},
		)
		if err != nil {
			return nil, err
		}
		attributes := make([]models.PersonAttribute, 0)
		for result.Next(ctx) {
			attributes = append(attributes, personAttributeFromRecord(result.Record()))
		}
		return attributes, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.PersonAttribute), nil
}

// UpdatePersonAttribute updates a person profile attribute.
func (s *Storage) UpdatePersonAttribute(ctx context.Context, attribute models.PersonAttribute) (*models.PersonAttribute, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Person {id: $person_id})-[:HAS_ATTRIBUTE]->(a:PersonAttribute {id: $id})
			SET a.type = $type,
			    a.value = $value,
			    a.organization_id = $organization_id,
			    a.start_date = $start_date,
			    a.end_date = $end_date,
			    a.current = $current,
			    a.notes = $notes
			RETURN a.id AS id, a.person_id AS person_id, a.type AS type, a.value AS value,
			       a.organization_id AS organization_id, a.start_date AS start_date,
			       a.end_date AS end_date, a.current AS current, a.notes AS notes,
			       a.archived AS archived`,
			personAttributeParams(attribute),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := personAttributeFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.PersonAttribute), nil
}

// ArchivePersonAttribute archives a person profile attribute.
func (s *Storage) ArchivePersonAttribute(ctx context.Context, personID string, attributeID string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Person {id: $person_id})-[:HAS_ATTRIBUTE]->(a:PersonAttribute {id: $id})
			SET a.archived = true
			RETURN count(a) AS archived`,
			map[string]any{"person_id": personID, "id": attributeID},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "archived") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

// CreateOrganization creates an organization node.
func (s *Storage) CreateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error) {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			CREATE (o:Organization {
				id: $id,
				name: $name,
				type: $type,
				description: $description,
				notes: $notes,
				tags: $tags,
				aliases: $aliases,
				active: $active,
				web: $web,
				archived: false
			})`,
			organizationParams(organization),
		)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return &organization, nil
}

// GetOrganization returns one organization by ID.
func (s *Storage) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, organizationReturnQuery("MATCH (o:Organization {id: $id})"), map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		org := organizationFromRecord(record)
		return &org, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Organization), nil
}

// ListOrganizations returns all organizations.
func (s *Storage) ListOrganizations(ctx context.Context) ([]models.Organization, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, organizationReturnQuery("MATCH (o:Organization) WHERE coalesce(o.archived, false) = false")+" ORDER BY toLower(o.name)", nil)
		if err != nil {
			return nil, err
		}
		var organizations []models.Organization
		for result.Next(ctx) {
			organizations = append(organizations, organizationFromRecord(result.Record()))
		}
		return organizations, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Organization), nil
}

// UpdateOrganization updates an organization node.
func (s *Storage) UpdateOrganization(ctx context.Context, organization models.Organization) (*models.Organization, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, organizationReturnQuery(`
			MATCH (o:Organization {id: $id})
			SET o.name = $name,
			    o.type = $type,
			    o.description = $description,
			    o.notes = $notes,
			    o.tags = $tags,
			    o.aliases = $aliases,
			    o.active = $active,
			    o.web = $web`),
			organizationParams(organization),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		org := organizationFromRecord(record)
		return &org, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Organization), nil
}

// DeleteOrganization deletes an organization node and its relationships.
func (s *Storage) DeleteOrganization(ctx context.Context, id string) error {
	return s.archiveNode(ctx, "Organization", id)
}

// CreateProject creates a project node.
func (s *Storage) CreateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			CREATE (p:Project {
				id: $id,
				name: $name,
				status: $status,
				description: $description,
				notes: $notes,
				tags: $tags,
				aliases: $aliases,
				active: $active,
				web: $web,
				archived: false
			})`,
			projectParams(project),
		)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProject returns one project by ID.
func (s *Storage) GetProject(ctx context.Context, id string) (*models.Project, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, projectReturnQuery("MATCH (p:Project {id: $id})"), map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		project := projectFromRecord(record)
		return &project, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Project), nil
}

// ListProjects returns all projects.
func (s *Storage) ListProjects(ctx context.Context) ([]models.Project, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, projectReturnQuery("MATCH (p:Project) WHERE coalesce(p.archived, false) = false")+" ORDER BY toLower(p.name)", nil)
		if err != nil {
			return nil, err
		}
		var projects []models.Project
		for result.Next(ctx) {
			projects = append(projects, projectFromRecord(result.Record()))
		}
		return projects, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Project), nil
}

// UpdateProject updates a project node.
func (s *Storage) UpdateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, projectReturnQuery(`
			MATCH (p:Project {id: $id})
			SET p.name = $name,
			    p.status = $status,
			    p.description = $description,
			    p.notes = $notes,
			    p.tags = $tags,
			    p.aliases = $aliases,
			    p.active = $active,
			    p.web = $web`),
			projectParams(project),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := projectFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Project), nil
}

// DeleteProject deletes a project node and its relationships.
func (s *Storage) DeleteProject(ctx context.Context, id string) error {
	return s.archiveNode(ctx, "Project", id)
}

// CreateDiagramNode creates a network-scoped diagram-only node.
func (s *Storage) CreateDiagramNode(ctx context.Context, node models.DiagramNode) (*models.DiagramNode, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, diagramNodeReturnQuery(`
			MATCH (n:Network {id: $network_id})
			CREATE (n)-[:CONTAINS_DIAGRAM_NODE]->(d:DiagramNode {
				id: $id,
				network_id: $network_id,
				type: $type,
				name: $name,
				description: $description,
				notes: $notes,
				archived: false
			})`),
			diagramNodeParams(node),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := diagramNodeFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.DiagramNode), nil
}

// GetDiagramNode returns one network-scoped diagram node.
func (s *Storage) GetDiagramNode(ctx context.Context, networkID string, id string) (*models.DiagramNode, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, diagramNodeReturnQuery(`
			MATCH (:Network {id: $network_id})-[:CONTAINS_DIAGRAM_NODE]->(d:DiagramNode {id: $id})
			WHERE coalesce(d.archived, false) = false`),
			map[string]any{"network_id": networkID, "id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		node := diagramNodeFromRecord(record)
		return &node, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.DiagramNode), nil
}

// UpdateDiagramNode updates a network-scoped diagram node.
func (s *Storage) UpdateDiagramNode(ctx context.Context, node models.DiagramNode) (*models.DiagramNode, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, diagramNodeReturnQuery(`
			MATCH (:Network {id: $network_id})-[:CONTAINS_DIAGRAM_NODE]->(d:DiagramNode {id: $id})
			WHERE coalesce(d.archived, false) = false
			SET d.type = $type,
			    d.name = $name,
			    d.description = $description,
			    d.notes = $notes`),
			diagramNodeParams(node),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := diagramNodeFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.DiagramNode), nil
}

// DeleteDiagramNode permanently removes a network-scoped diagram node.
func (s *Storage) DeleteDiagramNode(ctx context.Context, networkID string, id string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		params := map[string]any{"network_id": networkID, "id": id}
		if _, err := tx.Run(ctx, `
			MATCH (:Network {id: $network_id})-[:CONTAINS_DIAGRAM_NODE]->(d:DiagramNode {id: $id})-[r]-()
			WHERE r.network_id = $network_id
			DELETE r`,
			params,
		); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx, `
			MATCH (:Network {id: $network_id})-[positionRel:HAS_POSITION]->(position:NetworkPosition {node_id: $id})
			DELETE positionRel, position`,
			params,
		); err != nil {
			return nil, err
		}
		_, err := tx.Run(ctx, `
			MATCH (:Network {id: $network_id})-[:CONTAINS_DIAGRAM_NODE]->(d:DiagramNode {id: $id})
			DETACH DELETE d`,
			params,
		)
		return nil, err
	})
	return err
}

// GetOrganizationProfile returns an organization with attributes and relationships.
func (s *Storage) GetOrganizationProfile(ctx context.Context, id string) (*models.OrganizationProfile, error) {
	organization, err := s.GetOrganization(ctx, id)
	if err != nil {
		return nil, err
	}
	attributes, err := s.ListOrganizationAttributes(ctx, id)
	if err != nil {
		return nil, err
	}
	relationships, err := s.listRelationshipsForNode(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.OrganizationProfile{
		Organization:  *organization,
		Attributes:    attributes,
		Relationships: relationships,
	}, nil
}

// CreateOrganizationAttribute creates an organization profile attribute.
func (s *Storage) CreateOrganizationAttribute(ctx context.Context, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (o:Organization {id: $organization_id})
			CREATE (o)-[:HAS_ATTRIBUTE]->(a:OrganizationAttribute {
				id: $id,
				organization_id: $organization_id,
				type: $type,
				value: $value,
				person_id: $person_id,
				start_date: $start_date,
				end_date: $end_date,
				current: $current,
				notes: $notes,
				archived: false
			})
			RETURN a.id AS id, a.organization_id AS organization_id, a.type AS type, a.value AS value,
			       a.person_id AS person_id, a.start_date AS start_date,
			       a.end_date AS end_date, a.current AS current, a.notes AS notes,
			       a.archived AS archived`,
			organizationAttributeParams(attribute),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := organizationAttributeFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.OrganizationAttribute), nil
}

// ListOrganizationAttributes lists non-archived attributes for an organization.
func (s *Storage) ListOrganizationAttributes(ctx context.Context, organizationID string) ([]models.OrganizationAttribute, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Organization {id: $organization_id})-[:HAS_ATTRIBUTE]->(a:OrganizationAttribute)
			WHERE coalesce(a.archived, false) = false
			RETURN a.id AS id, a.organization_id AS organization_id, a.type AS type, a.value AS value,
			       a.person_id AS person_id, a.start_date AS start_date,
			       a.end_date AS end_date, a.current AS current, a.notes AS notes,
			       a.archived AS archived
			ORDER BY coalesce(a.start_date, ''), toLower(a.value)`,
			map[string]any{"organization_id": organizationID},
		)
		if err != nil {
			return nil, err
		}
		attributes := make([]models.OrganizationAttribute, 0)
		for result.Next(ctx) {
			attributes = append(attributes, organizationAttributeFromRecord(result.Record()))
		}
		return attributes, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.OrganizationAttribute), nil
}

// UpdateOrganizationAttribute updates an organization profile attribute.
func (s *Storage) UpdateOrganizationAttribute(ctx context.Context, attribute models.OrganizationAttribute) (*models.OrganizationAttribute, error) {
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Organization {id: $organization_id})-[:HAS_ATTRIBUTE]->(a:OrganizationAttribute {id: $id})
			SET a.type = $type,
			    a.value = $value,
			    a.person_id = $person_id,
			    a.start_date = $start_date,
			    a.end_date = $end_date,
			    a.current = $current,
			    a.notes = $notes
			RETURN a.id AS id, a.organization_id AS organization_id, a.type AS type, a.value AS value,
			       a.person_id AS person_id, a.start_date AS start_date,
			       a.end_date AS end_date, a.current AS current, a.notes AS notes,
			       a.archived AS archived`,
			organizationAttributeParams(attribute),
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := organizationAttributeFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.OrganizationAttribute), nil
}

// ArchiveOrganizationAttribute archives an organization profile attribute.
func (s *Storage) ArchiveOrganizationAttribute(ctx context.Context, organizationID string, attributeID string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Organization {id: $organization_id})-[:HAS_ATTRIBUTE]->(a:OrganizationAttribute {id: $id})
			SET a.archived = true
			RETURN count(a) AS archived`,
			map[string]any{"organization_id": organizationID, "id": attributeID},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "archived") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

// CreateRelationship creates a relationship edge.
func (s *Storage) CreateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error) {
	relType := string(relationship.Type)
	query := fmt.Sprintf(`
		MATCH (source {id: $source_id})
		MATCH (target {id: $target_id})
		CREATE (source)-[r:%s {
			id: $id,
			network_id: $network_id,
			source_type: $source_type,
			target_type: $target_type,
			custom_label: $custom_label,
			role: $role,
			start_date: $start_date,
			end_date: $end_date,
			current: $current,
			notes: $notes
		}]->(target)
		RETURN r.id AS id,
		       r.network_id AS network_id,
		       type(r) AS type,
		       source.id AS source_id,
		       coalesce(r.source_type, source.type, CASE WHEN source:Project THEN 'project' WHEN source:DiagramNode THEN source.type ELSE 'person' END) AS source_type,
		       target.id AS target_id,
		       coalesce(r.target_type, target.type, CASE WHEN target:Project THEN 'project' WHEN target:DiagramNode THEN target.type ELSE 'person' END) AS target_type,
		       r.custom_label AS custom_label,
		       r.role AS role,
		       r.start_date AS start_date,
		       r.end_date AS end_date,
		       r.current AS current,
		       r.notes AS notes`, relType)
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, relationshipParams(relationship))
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		created := relationshipFromRecord(record)
		return &created, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Relationship), nil
}

// GetRelationship returns one relationship by ID.
func (s *Storage) GetRelationship(ctx context.Context, id string) (*models.Relationship, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, relationshipReturnQuery("MATCH (source)-[r]->(target) WHERE r.id = $id"), map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		relationship := relationshipFromRecord(record)
		return &relationship, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Relationship), nil
}

// UpdateRelationship updates editable relationship properties.
func (s *Storage) UpdateRelationship(ctx context.Context, relationship models.Relationship) (*models.Relationship, error) {
	relType := string(relationship.Type)
	query := fmt.Sprintf(`
		MATCH (source)-[old]->(target)
		WHERE old.id = $id
		DELETE old
		CREATE (source)-[r:%s {
			id: $id,
			network_id: $network_id,
			source_type: $source_type,
			target_type: $target_type,
			custom_label: $custom_label,
			role: $role,
			start_date: $start_date,
			end_date: $end_date,
			current: $current,
			notes: $notes
		}]->(target)
		RETURN r.id AS id,
		       r.network_id AS network_id,
		       type(r) AS type,
		       source.id AS source_id,
		       coalesce(r.source_type, source.type, CASE WHEN source:Project THEN 'project' WHEN source:DiagramNode THEN source.type ELSE 'person' END) AS source_type,
		       target.id AS target_id,
		       coalesce(r.target_type, target.type, CASE WHEN target:Project THEN 'project' WHEN target:DiagramNode THEN target.type ELSE 'person' END) AS target_type,
		       r.custom_label AS custom_label,
		       r.role AS role,
		       r.start_date AS start_date,
		       r.end_date AS end_date,
		       r.current AS current,
		       r.notes AS notes`, relType)
	value, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, relationshipParams(relationship))
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, storage.ErrNotFound
		}
		updated := relationshipFromRecord(record)
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*models.Relationship), nil
}

// ListRelationships returns all relationships.
func (s *Storage) ListRelationships(ctx context.Context) ([]models.Relationship, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, relationshipReturnQuery("MATCH (source)-[r]->(target) WHERE r.id IS NOT NULL"), nil)
		if err != nil {
			return nil, err
		}
		var relationships []models.Relationship
		for result.Next(ctx) {
			relationships = append(relationships, relationshipFromRecord(result.Record()))
		}
		return relationships, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Relationship), nil
}

func (s *Storage) listRelationshipsForNode(ctx context.Context, nodeID string) ([]models.Relationship, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, relationshipReturnQuery("MATCH (source)-[r]->(target) WHERE r.id IS NOT NULL AND (source.id = $id OR target.id = $id)"), map[string]any{"id": nodeID})
		if err != nil {
			return nil, err
		}
		var relationships []models.Relationship
		for result.Next(ctx) {
			relationships = append(relationships, relationshipFromRecord(result.Record()))
		}
		return relationships, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Relationship), nil
}

// DeleteRelationship deletes a relationship edge.
func (s *Storage) DeleteRelationship(ctx context.Context, id string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH ()-[r]->()
			WHERE r.id = $id
			WITH r, count(r) AS deleted
			DELETE r
			RETURN deleted`,
			map[string]any{"id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "deleted") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

// SavePosition stores a node position.
func (s *Storage) SavePosition(ctx context.Context, position models.Position) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			MERGE (p:Position {node_id: $node_id, node_type: $node_type})
			SET p.x = $x, p.y = $y`,
			map[string]any{
				"node_id":   position.NodeID,
				"node_type": position.NodeType,
				"x":         position.X,
				"y":         position.Y,
			},
		)
		return nil, err
	})
	return err
}

// GetGraph returns all graph data needed by the frontend.
func (s *Storage) GetGraph(ctx context.Context) (*models.GraphResponse, error) {
	persons, err := s.ListPersons(ctx)
	if err != nil {
		return nil, err
	}
	organizations, err := s.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := s.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	relationships, err := s.ListRelationships(ctx)
	if err != nil {
		return nil, err
	}
	positions, err := s.listPositions(ctx)
	if err != nil {
		return nil, err
	}

	return &models.GraphResponse{
		Persons:       persons,
		Organizations: organizations,
		Projects:      projects,
		Locations:     []map[string]any{},
		Tags:          []map[string]any{},
		Relationships: relationships,
		Positions:     positions,
	}, nil
}

func (s *Storage) listPositions(ctx context.Context) ([]models.Position, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Position)
			RETURN p.node_id AS node_id, p.node_type AS node_type, p.x AS x, p.y AS y`,
			nil,
		)
		if err != nil {
			return nil, err
		}
		var positions []models.Position
		for result.Next(ctx) {
			record := result.Record()
			positions = append(positions, models.Position{
				NodeID:   asString(record, "node_id"),
				NodeType: asString(record, "node_type"),
				X:        asFloat(record, "x"),
				Y:        asFloat(record, "y"),
			})
		}
		return positions, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Position), nil
}

func (s *Storage) listNetworkRelationships(ctx context.Context, networkID string) ([]models.Relationship, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, relationshipReturnQuery(`
			MATCH (source)-[r]->(target)
			WHERE r.id IS NOT NULL
			  AND (
			    r.network_id = $network_id
			    OR EXISTS {
			      MATCH (:Network {id: $network_id})-[membership:CONTAINS_PERSON]->(person:Person)
			      WHERE coalesce(membership.archived, false) = false
			        AND (source.id = person.id OR target.id = person.id)
			    }
			    OR (
			      EXISTS {
			        MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: source.id})
			      }
			      AND EXISTS {
			        MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: target.id})
			      }
			    )
			  )`),
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		relationships := make([]models.Relationship, 0)
		for result.Next(ctx) {
			relationship := relationshipFromRecord(result.Record())
			if seen[relationship.ID] {
				continue
			}
			seen[relationship.ID] = true
			relationships = append(relationships, relationship)
		}
		return relationships, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Relationship), nil
}

func (s *Storage) listNetworkOrganizations(ctx context.Context, networkID string) ([]models.Organization, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, organizationReturnQuery(`
			MATCH (o:Organization)
			WHERE coalesce(o.archived, false) = false
			  AND (
			    EXISTS {
			      MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: o.id})
			    }
			    OR EXISTS {
			      MATCH (source)-[r]->(target)
			      WHERE r.id IS NOT NULL
			        AND (source.id = o.id OR target.id = o.id)
			        AND (
			          r.network_id = $network_id
			          OR EXISTS {
			            MATCH (:Network {id: $network_id})-[membership:CONTAINS_PERSON]->(person:Person)
			            WHERE coalesce(membership.archived, false) = false
			              AND (source.id = person.id OR target.id = person.id)
			          }
			          OR (
			            EXISTS {
			              MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: source.id})
			            }
			            AND EXISTS {
			              MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: target.id})
			            }
			          )
			        )
			    }
			  )`)+" ORDER BY toLower(o.name)",
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		organizations := make([]models.Organization, 0)
		for result.Next(ctx) {
			organization := organizationFromRecord(result.Record())
			if seen[organization.ID] {
				continue
			}
			seen[organization.ID] = true
			organizations = append(organizations, organization)
		}
		return organizations, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Organization), nil
}

func (s *Storage) listNetworkProjects(ctx context.Context, networkID string) ([]models.Project, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, projectReturnQuery(`
			MATCH (p:Project)
			WHERE coalesce(p.archived, false) = false
			  AND (
			    EXISTS {
			      MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: p.id})
			    }
			    OR EXISTS {
			      MATCH (source)-[r]->(target)
			      WHERE r.id IS NOT NULL
			        AND (source.id = p.id OR target.id = p.id)
			        AND (
			          r.network_id = $network_id
			          OR EXISTS {
			            MATCH (:Network {id: $network_id})-[membership:CONTAINS_PERSON]->(person:Person)
			            WHERE coalesce(membership.archived, false) = false
			              AND (source.id = person.id OR target.id = person.id)
			          }
			          OR (
			            EXISTS {
			              MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: source.id})
			            }
			            AND EXISTS {
			              MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(:NetworkPosition {node_id: target.id})
			            }
			          )
			        )
			    }
			  )`)+" ORDER BY toLower(p.name)",
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		projects := make([]models.Project, 0)
		for result.Next(ctx) {
			project := projectFromRecord(result.Record())
			if seen[project.ID] {
				continue
			}
			seen[project.ID] = true
			projects = append(projects, project)
		}
		return projects, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Project), nil
}

func (s *Storage) listNetworkDiagramNodes(ctx context.Context, networkID string) ([]models.DiagramNode, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, diagramNodeReturnQuery(`
			MATCH (:Network {id: $network_id})-[:CONTAINS_DIAGRAM_NODE]->(d:DiagramNode)
			WHERE coalesce(d.archived, false) = false`)+" ORDER BY toLower(d.name)",
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		nodes := make([]models.DiagramNode, 0)
		for result.Next(ctx) {
			nodes = append(nodes, diagramNodeFromRecord(result.Record()))
		}
		return nodes, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.DiagramNode), nil
}

func (s *Storage) listNetworkPositions(ctx context.Context, networkID string) ([]models.Position, error) {
	value, err := s.read(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (:Network {id: $network_id})-[:HAS_POSITION]->(p:NetworkPosition)
			RETURN p.node_id AS node_id, p.node_type AS node_type, p.x AS x, p.y AS y`,
			map[string]any{"network_id": networkID},
		)
		if err != nil {
			return nil, err
		}
		positions := make([]models.Position, 0)
		for result.Next(ctx) {
			record := result.Record()
			positions = append(positions, models.Position{
				NodeID:   asString(record, "node_id"),
				NodeType: asString(record, "node_type"),
				X:        asFloat(record, "x"),
				Y:        asFloat(record, "y"),
			})
		}
		return positions, result.Err()
	})
	if err != nil {
		return nil, err
	}
	return value.([]models.Position), nil
}

func (s *Storage) movePersonMergeData(ctx context.Context, tx driver.ManagedTransaction, survivorID string, removedID string) error {
	statements := []string{
		`
		MATCH (removed:Person {id: $removed_id})-[:HAS_ATTRIBUTE]->(a:PersonAttribute)
		MATCH (survivor:Person {id: $survivor_id})
		MERGE (survivor)-[:HAS_ATTRIBUTE]->(a)
		SET a.person_id = $survivor_id`,
		`
		MATCH (n:Network)-[r:CONTAINS_PERSON]->(:Person {id: $removed_id})
		MATCH (survivor:Person {id: $survivor_id})
		MERGE (n)-[existing:CONTAINS_PERSON]->(survivor)
		SET existing.network_id = n.id,
		    existing.person_id = $survivor_id,
		    existing.notes = CASE WHEN coalesce(existing.notes, '') = '' THEN r.notes ELSE existing.notes END,
		    existing.context = CASE WHEN coalesce(existing.context, '') = '' THEN r.context ELSE existing.context END,
		    existing.archived = coalesce(existing.archived, r.archived, false)`,
	}
	for _, statement := range statements {
		if _, err := tx.Run(ctx, statement, map[string]any{"survivor_id": survivorID, "removed_id": removedID}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) rewirePersonRelationships(ctx context.Context, tx driver.ManagedTransaction, survivorID string, removedID string) error {
	types := []models.RelationshipType{
		models.RelationshipKnows,
		models.RelationshipSpouseOf,
		models.RelationshipParentOf,
		models.RelationshipSiblingOf,
		models.RelationshipWorksAt,
		models.RelationshipMemberOf,
		models.RelationshipStudiedAt,
		models.RelationshipLivesIn,
		models.RelationshipHasTag,
	}
	for _, relType := range types {
		if err := s.rewireOutgoingPersonRelationships(ctx, tx, relType, survivorID, removedID); err != nil {
			return err
		}
		if err := s.rewireIncomingPersonRelationships(ctx, tx, relType, survivorID, removedID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) rewireOutgoingPersonRelationships(ctx context.Context, tx driver.ManagedTransaction, relType models.RelationshipType, survivorID string, removedID string) error {
	query := fmt.Sprintf(`
		MATCH (removed:Person {id: $removed_id})-[r:%s]->(target)
		WHERE r.id IS NOT NULL AND target.id <> $survivor_id
		MATCH (survivor:Person {id: $survivor_id})
		CREATE (survivor)-[copy:%s]->(target)
		SET copy = properties(r),
		    copy.source_id = $survivor_id
		DELETE r`, relType, relType)
	_, err := tx.Run(ctx, query, map[string]any{"survivor_id": survivorID, "removed_id": removedID})
	return err
}

func (s *Storage) rewireIncomingPersonRelationships(ctx context.Context, tx driver.ManagedTransaction, relType models.RelationshipType, survivorID string, removedID string) error {
	query := fmt.Sprintf(`
		MATCH (source)-[r:%s]->(removed:Person {id: $removed_id})
		WHERE r.id IS NOT NULL AND source.id <> $survivor_id
		MATCH (survivor:Person {id: $survivor_id})
		CREATE (source)-[copy:%s]->(survivor)
		SET copy = properties(r),
		    copy.target_id = $survivor_id
		DELETE r`, relType, relType)
	_, err := tx.Run(ctx, query, map[string]any{"survivor_id": survivorID, "removed_id": removedID})
	return err
}

func (s *Storage) deleteNode(ctx context.Context, label, id string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, fmt.Sprintf(`
			MATCH (n:%s {id: $id})
			WITH n, count(n) AS deleted
			DETACH DELETE n
			RETURN deleted`, label),
			map[string]any{"id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "deleted") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

func (s *Storage) archiveNode(ctx context.Context, label, id string) error {
	_, err := s.write(ctx, func(tx driver.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, fmt.Sprintf(`
			MATCH (n:%s {id: $id})
			SET n.archived = true,
			    n.active = false
			RETURN count(n) AS archived`, label),
			map[string]any{"id": id},
		)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		if asInt(record, "archived") == 0 {
			return nil, storage.ErrNotFound
		}
		return nil, nil
	})
	return err
}

func (s *Storage) read(ctx context.Context, work func(driver.ManagedTransaction) (any, error)) (any, error) {
	session := s.driver.NewSession(ctx, driver.SessionConfig{DatabaseName: s.database})
	defer session.Close(ctx)
	return session.ExecuteRead(ctx, work)
}

func (s *Storage) write(ctx context.Context, work func(driver.ManagedTransaction) (any, error)) (any, error) {
	session := s.driver.NewSession(ctx, driver.SessionConfig{DatabaseName: s.database})
	defer session.Close(ctx)
	return session.ExecuteWrite(ctx, work)
}

func personParams(person models.Person) map[string]any {
	return map[string]any{
		"id":          person.ID,
		"name":        person.Name,
		"nickname":    person.Nickname,
		"gender":      person.Gender,
		"title":       person.Title,
		"description": person.Description,
		"notes":       person.Notes,
		"tags":        person.Tags,
		"deceased":    person.Deceased,
	}
}

func accountParams(account models.Account) map[string]any {
	return map[string]any{
		"id":            account.ID,
		"email":         account.Email,
		"display_name":  account.DisplayName,
		"role":          string(account.Role),
		"disabled":      account.Disabled,
		"password_hash": account.PasswordHash,
	}
}

func sessionParams(session models.Session) map[string]any {
	return map[string]any{
		"id":         session.ID,
		"account_id": session.AccountID,
		"token_hash": session.TokenHash,
		"expires_at": session.ExpiresAt,
	}
}

func networkParams(network models.Network) map[string]any {
	return map[string]any{
		"id":          network.ID,
		"owner_id":    network.OwnerID,
		"name":        network.Name,
		"description": network.Description,
		"domain":      network.Domain,
		"archived":    network.Archived,
	}
}

func networkPersonContextParams(context models.NetworkPersonContext) map[string]any {
	return map[string]any{
		"network_id": context.NetworkID,
		"person_id":  context.PersonID,
		"notes":      context.Notes,
		"context":    context.Context,
		"archived":   context.Archived,
	}
}

func organizationParams(organization models.Organization) map[string]any {
	return map[string]any{
		"id":          organization.ID,
		"name":        organization.Name,
		"type":        string(organization.Type),
		"description": organization.Description,
		"notes":       organization.Notes,
		"tags":        organization.Tags,
		"aliases":     organization.Aliases,
		"active":      organization.Active,
		"web":         organization.Web,
		"archived":    organization.Archived,
	}
}

func projectParams(project models.Project) map[string]any {
	return map[string]any{
		"id":          project.ID,
		"name":        project.Name,
		"status":      project.Status,
		"description": project.Description,
		"notes":       project.Notes,
		"tags":        project.Tags,
		"aliases":     project.Aliases,
		"active":      project.Active,
		"web":         project.Web,
		"archived":    project.Archived,
	}
}

func diagramNodeParams(node models.DiagramNode) map[string]any {
	return map[string]any{
		"id":          node.ID,
		"network_id":  node.NetworkID,
		"type":        node.Type,
		"name":        node.Name,
		"description": node.Description,
		"notes":       node.Notes,
		"archived":    node.Archived,
	}
}

func customRelationshipTypeParams(relationshipType models.CustomRelationshipType) map[string]any {
	return map[string]any{
		"id":                 relationshipType.ID,
		"network_id":         relationshipType.NetworkID,
		"owner_id":           relationshipType.OwnerID,
		"key":                relationshipType.Key,
		"label":              relationshipType.Label,
		"source_type":        relationshipType.SourceType,
		"target_type":        relationshipType.TargetType,
		"direction_behavior": relationshipType.DirectionBehavior,
		"archived":           relationshipType.Archived,
	}
}

func relationshipParams(relationship models.Relationship) map[string]any {
	return map[string]any{
		"id":           relationship.ID,
		"network_id":   relationship.NetworkID,
		"source_id":    relationship.SourceID,
		"source_type":  relationship.SourceType,
		"target_id":    relationship.TargetID,
		"target_type":  relationship.TargetType,
		"type":         string(relationship.Type),
		"custom_label": relationship.CustomLabel,
		"role":         relationship.Role,
		"start_date":   relationship.StartDate,
		"end_date":     relationship.EndDate,
		"current":      relationship.Current,
		"notes":        relationship.Notes,
	}
}

func personAttributeParams(attribute models.PersonAttribute) map[string]any {
	return map[string]any{
		"id":              attribute.ID,
		"person_id":       attribute.PersonID,
		"type":            string(attribute.Type),
		"value":           attribute.Value,
		"organization_id": attribute.OrganizationID,
		"start_date":      attribute.StartDate,
		"end_date":        attribute.EndDate,
		"current":         attribute.Current,
		"notes":           attribute.Notes,
	}
}

func organizationAttributeParams(attribute models.OrganizationAttribute) map[string]any {
	return map[string]any{
		"id":              attribute.ID,
		"organization_id": attribute.OrganizationID,
		"type":            string(attribute.Type),
		"value":           attribute.Value,
		"person_id":       attribute.PersonID,
		"start_date":      attribute.StartDate,
		"end_date":        attribute.EndDate,
		"current":         attribute.Current,
		"notes":           attribute.Notes,
	}
}

func accountFromRecord(record *driver.Record) models.Account {
	return models.Account{
		ID:           asString(record, "id"),
		Email:        asString(record, "email"),
		DisplayName:  asString(record, "display_name"),
		Role:         models.Role(asString(record, "role")),
		Disabled:     asBool(record, "disabled"),
		PasswordHash: asString(record, "password_hash"),
	}
}

func sessionFromRecord(record *driver.Record) models.Session {
	return models.Session{
		ID:        asString(record, "id"),
		AccountID: asString(record, "account_id"),
		TokenHash: asString(record, "token_hash"),
		ExpiresAt: asString(record, "expires_at"),
	}
}

func networkFromRecord(record *driver.Record) models.Network {
	return models.Network{
		ID:          asString(record, "id"),
		OwnerID:     asString(record, "owner_id"),
		Name:        asString(record, "name"),
		Description: asString(record, "description"),
		Domain:      asString(record, "domain"),
		Archived:    asBool(record, "archived"),
	}
}

func personFromRecord(record *driver.Record) models.Person {
	return models.Person{
		ID:          asString(record, "id"),
		Name:        asString(record, "name"),
		Nickname:    asString(record, "nickname"),
		Gender:      asString(record, "gender"),
		Title:       asString(record, "title"),
		Description: asString(record, "description"),
		Notes:       asString(record, "notes"),
		Tags:        asStringSlice(record, "tags"),
		Deceased:    asBool(record, "deceased"),
	}
}

func networkPersonContextFromRecord(record *driver.Record) models.NetworkPersonContext {
	return models.NetworkPersonContext{
		NetworkID: asString(record, "network_id"),
		PersonID:  asString(record, "person_id"),
		Notes:     asString(record, "context_notes"),
		Context:   asString(record, "context"),
		Archived:  asBool(record, "context_archived"),
	}
}

func networkPersonFromRecord(record *driver.Record) models.NetworkPerson {
	return models.NetworkPerson{
		Person:  personFromRecord(record),
		Context: networkPersonContextFromRecord(record),
	}
}

func organizationFromRecord(record *driver.Record) models.Organization {
	return models.Organization{
		ID:          asString(record, "id"),
		Name:        asString(record, "name"),
		Type:        models.OrganizationType(asString(record, "type")),
		Description: asString(record, "description"),
		Notes:       asString(record, "notes"),
		Tags:        asStringSlice(record, "tags"),
		Aliases:     asStringSlice(record, "aliases"),
		Active:      asBool(record, "active"),
		Web:         asString(record, "web"),
		Archived:    asBool(record, "archived"),
	}
}

func projectFromRecord(record *driver.Record) models.Project {
	return models.Project{
		ID:          asString(record, "id"),
		Name:        asString(record, "name"),
		Status:      asString(record, "status"),
		Description: asString(record, "description"),
		Notes:       asString(record, "notes"),
		Tags:        asStringSlice(record, "tags"),
		Aliases:     asStringSlice(record, "aliases"),
		Active:      asBool(record, "active"),
		Web:         asString(record, "web"),
		Archived:    asBool(record, "archived"),
	}
}

func diagramNodeFromRecord(record *driver.Record) models.DiagramNode {
	return models.DiagramNode{
		ID:          asString(record, "id"),
		NetworkID:   asString(record, "network_id"),
		Type:        asString(record, "type"),
		Name:        asString(record, "name"),
		Description: asString(record, "description"),
		Notes:       asString(record, "notes"),
		Archived:    asBool(record, "archived"),
	}
}

func customRelationshipTypeFromRecord(record *driver.Record) models.CustomRelationshipType {
	return models.CustomRelationshipType{
		ID:                asString(record, "id"),
		NetworkID:         asString(record, "network_id"),
		OwnerID:           asString(record, "owner_id"),
		Key:               asString(record, "key"),
		Label:             asString(record, "label"),
		SourceType:        asString(record, "source_type"),
		TargetType:        asString(record, "target_type"),
		DirectionBehavior: asString(record, "direction_behavior"),
		Archived:          asBool(record, "archived"),
	}
}

func relationshipFromRecord(record *driver.Record) models.Relationship {
	return models.Relationship{
		ID:          asString(record, "id"),
		NetworkID:   asString(record, "network_id"),
		SourceID:    asString(record, "source_id"),
		SourceType:  asString(record, "source_type"),
		TargetID:    asString(record, "target_id"),
		TargetType:  asString(record, "target_type"),
		Type:        models.RelationshipType(asString(record, "type")),
		CustomLabel: asString(record, "custom_label"),
		Role:        asString(record, "role"),
		StartDate:   asString(record, "start_date"),
		EndDate:     asString(record, "end_date"),
		Current:     asBool(record, "current"),
		Notes:       asString(record, "notes"),
	}
}

func personAttributeFromRecord(record *driver.Record) models.PersonAttribute {
	return models.PersonAttribute{
		ID:             asString(record, "id"),
		PersonID:       asString(record, "person_id"),
		Type:           models.PersonAttributeType(asString(record, "type")),
		Value:          asString(record, "value"),
		OrganizationID: asString(record, "organization_id"),
		StartDate:      asString(record, "start_date"),
		EndDate:        asString(record, "end_date"),
		Current:        asBool(record, "current"),
		Notes:          asString(record, "notes"),
		Archived:       asBool(record, "archived"),
	}
}

func organizationAttributeFromRecord(record *driver.Record) models.OrganizationAttribute {
	return models.OrganizationAttribute{
		ID:             asString(record, "id"),
		OrganizationID: asString(record, "organization_id"),
		Type:           models.OrganizationAttributeType(asString(record, "type")),
		Value:          asString(record, "value"),
		PersonID:       asString(record, "person_id"),
		StartDate:      asString(record, "start_date"),
		EndDate:        asString(record, "end_date"),
		Current:        asBool(record, "current"),
		Notes:          asString(record, "notes"),
		Archived:       asBool(record, "archived"),
	}
}

func accountReturnQuery(prefix string) string {
	return prefix + `
		RETURN a.id AS id, a.email AS email, a.display_name AS display_name,
		       a.role AS role, a.disabled AS disabled, a.password_hash AS password_hash`
}

func networkReturnQuery(prefix string) string {
	return prefix + `
		RETURN n.id AS id, n.owner_id AS owner_id, n.name AS name,
		       n.description AS description, coalesce(n.domain, 'social') AS domain,
		       n.archived AS archived`
}

func networkPersonContextReturnQuery(prefix string) string {
	return prefix + `
		RETURN r.network_id AS network_id, r.person_id AS person_id,
		       r.notes AS context_notes, r.context AS context,
		       r.archived AS context_archived`
}

func networkPersonReturnQuery(prefix string) string {
	return prefix + `
		RETURN p.id AS id, p.name AS name, p.nickname AS nickname, p.gender AS gender,
		       p.title AS title, p.description AS description, p.notes AS notes,
		       p.tags AS tags, p.deceased AS deceased,
		       r.network_id AS network_id, r.person_id AS person_id,
		       r.notes AS context_notes, r.context AS context,
		       r.archived AS context_archived`
}

func organizationReturnQuery(prefix string) string {
	return prefix + `
		RETURN o.id AS id, o.name AS name, o.type AS type, o.description AS description,
		       o.notes AS notes, o.tags AS tags, o.aliases AS aliases, o.active AS active,
		       o.web AS web, o.archived AS archived`
}

func projectReturnQuery(prefix string) string {
	return prefix + `
		RETURN p.id AS id, p.name AS name, p.status AS status, p.description AS description,
		       p.notes AS notes, p.tags AS tags, p.aliases AS aliases, p.active AS active,
		       p.web AS web, p.archived AS archived`
}

func diagramNodeReturnQuery(prefix string) string {
	return prefix + `
		RETURN d.id AS id, d.network_id AS network_id, d.type AS type,
		       d.name AS name, d.description AS description, d.notes AS notes,
		       d.archived AS archived`
}

func relationshipReturnQuery(prefix string) string {
	return prefix + `
		RETURN r.id AS id,
		       r.network_id AS network_id,
		       type(r) AS type,
		       source.id AS source_id,
		       coalesce(r.source_type, source.type, CASE WHEN source:Project THEN 'project' WHEN source:DiagramNode THEN source.type ELSE 'person' END) AS source_type,
		       target.id AS target_id,
		       coalesce(r.target_type, target.type, CASE WHEN target:Project THEN 'project' WHEN target:DiagramNode THEN target.type ELSE 'person' END) AS target_type,
		       r.custom_label AS custom_label,
		       r.role AS role,
		       r.start_date AS start_date,
		       r.end_date AS end_date,
		       r.current AS current,
		       r.notes AS notes`
}

func asString(record *driver.Record, key string) string {
	value, ok := record.Get(key)
	if !ok || value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprint(value)
}

func asBool(record *driver.Record, key string) bool {
	value, ok := record.Get(key)
	if !ok || value == nil {
		return false
	}
	b, _ := value.(bool)
	return b
}

func asFloat(record *driver.Record, key string) float64 {
	value, ok := record.Get(key)
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}

func asInt(record *driver.Record, key string) int64 {
	value, ok := record.Get(key)
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func asStringSlice(record *driver.Record, key string) []string {
	value, ok := record.Get(key)
	if !ok || value == nil {
		return nil
	}
	items, ok := value.([]any)
	if !ok {
		if strings, ok := value.([]string); ok {
			return strings
		}
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if str, ok := item.(string); ok {
			out = append(out, str)
		}
	}
	return out
}
