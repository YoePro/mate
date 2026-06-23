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
		"CREATE CONSTRAINT mate_position_key IF NOT EXISTS FOR (p:Position) REQUIRE (p.node_id, p.node_type) IS UNIQUE",
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
		var attributes []models.PersonAttribute
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
				web: $web
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
		result, err := tx.Run(ctx, organizationReturnQuery("MATCH (o:Organization)")+" ORDER BY toLower(o.name)", nil)
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
	return s.deleteNode(ctx, "Organization", id)
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
		var attributes []models.OrganizationAttribute
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
			source_type: $source_type,
			target_type: $target_type,
			notes: $notes
		}]->(target)
		RETURN r.id AS id,
		       type(r) AS type,
		       source.id AS source_id,
		       coalesce(source.type, 'person') AS source_type,
		       target.id AS target_id,
		       coalesce(target.type, 'person') AS target_type,
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
	}
}

func relationshipParams(relationship models.Relationship) map[string]any {
	return map[string]any{
		"id":          relationship.ID,
		"source_id":   relationship.SourceID,
		"source_type": relationship.SourceType,
		"target_id":   relationship.TargetID,
		"target_type": relationship.TargetType,
		"type":        string(relationship.Type),
		"notes":       relationship.Notes,
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
	}
}

func relationshipFromRecord(record *driver.Record) models.Relationship {
	return models.Relationship{
		ID:         asString(record, "id"),
		SourceID:   asString(record, "source_id"),
		SourceType: asString(record, "source_type"),
		TargetID:   asString(record, "target_id"),
		TargetType: asString(record, "target_type"),
		Type:       models.RelationshipType(asString(record, "type")),
		Notes:      asString(record, "notes"),
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

func organizationReturnQuery(prefix string) string {
	return prefix + `
		RETURN o.id AS id, o.name AS name, o.type AS type, o.description AS description,
		       o.notes AS notes, o.tags AS tags, o.aliases AS aliases, o.active AS active,
		       o.web AS web`
}

func relationshipReturnQuery(prefix string) string {
	return prefix + `
		RETURN r.id AS id,
		       type(r) AS type,
		       source.id AS source_id,
		       coalesce(source.type, 'person') AS source_type,
		       target.id AS target_id,
		       coalesce(target.type, 'person') AS target_type,
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
