package services

import (
	"context"
	"sort"
	"strings"

	"mate/internal/models"
	"mate/internal/storage"
)

// NetworkService contains business logic for user-owned networks.
type NetworkService struct {
	store storage.Storage
}

// Create creates a network owned by actor.
func (s *NetworkService) Create(ctx context.Context, actor *models.Account, network models.Network) (*models.Network, error) {
	if err := requireActiveActor(actor); err != nil {
		return nil, err
	}
	network.Name = normalizeSpace(network.Name)
	network.Description = normalizeSpace(network.Description)
	network.Domain = normalizeNetworkDomain(network.Domain)
	if network.Name == "" {
		return nil, ErrInvalidInput
	}
	if !validNetworkDomain(network.Domain) {
		return nil, ErrInvalidInput
	}
	if network.ID == "" {
		network.ID = newID("network")
	}
	network.OwnerID = actor.ID
	return s.store.CreateNetwork(ctx, network)
}

// List returns networks visible to actor.
func (s *NetworkService) List(ctx context.Context, actor *models.Account) ([]models.Network, error) {
	if err := requireActiveActor(actor); err != nil {
		return nil, err
	}
	return s.store.ListNetworksForAccount(ctx, actor.ID)
}

// Search returns safe metadata for discoverable networks matching query.
func (s *NetworkService) Search(ctx context.Context, actor *models.Account, query string) ([]models.NetworkSearchResult, error) {
	if err := requireActiveActor(actor); err != nil {
		return nil, err
	}
	query = normalizeSpace(query)
	if len([]rune(query)) < 2 {
		return nil, ErrInvalidInput
	}
	return s.store.SearchNetworks(ctx, actor.ID, query)
}

// Get returns one owned network.
func (s *NetworkService) Get(ctx context.Context, actor *models.Account, id string) (*models.Network, error) {
	if err := s.requireOwner(ctx, actor, id); err != nil {
		return nil, err
	}
	return s.store.GetNetwork(ctx, id)
}

// Update updates owned network metadata.
func (s *NetworkService) Update(ctx context.Context, actor *models.Account, id string, network models.Network) (*models.Network, error) {
	if err := s.requireOwner(ctx, actor, id); err != nil {
		return nil, err
	}
	network.ID = id
	network.OwnerID = actor.ID
	network.Name = normalizeSpace(network.Name)
	network.Description = normalizeSpace(network.Description)
	network.Domain = normalizeNetworkDomain(network.Domain)
	if network.Name == "" {
		return nil, ErrInvalidInput
	}
	if !validNetworkDomain(network.Domain) {
		return nil, ErrInvalidInput
	}
	return s.store.UpdateNetwork(ctx, network)
}

// Archive archives an owned network.
func (s *NetworkService) Archive(ctx context.Context, actor *models.Account, id string) error {
	if err := s.requireOwner(ctx, actor, id); err != nil {
		return err
	}
	return s.store.ArchiveNetwork(ctx, id)
}

// AddPerson links an existing or newly created global person to a network.
func (s *NetworkService) AddPerson(ctx context.Context, actor *models.Account, networkID string, person models.Person, context models.NetworkPersonContext) (*models.NetworkPerson, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	network, err := s.store.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}
	if normalizeNetworkDomain(network.Domain) != "social" {
		return nil, ErrInvalidInput
	}
	person.Name = normalizeSpace(person.Name)
	context.NetworkID = networkID
	context.Notes = normalizeSpace(context.Notes)
	context.Context = normalizeSpace(context.Context)
	if person.ID == "" {
		if person.Name == "" || !validGender(person.Gender) {
			return nil, ErrInvalidInput
		}
		person.ID = newID("person")
		created, err := s.store.CreatePerson(ctx, person)
		if err != nil {
			return nil, err
		}
		person = *created
	} else {
		existing, err := s.store.GetPerson(ctx, person.ID)
		if err != nil {
			return nil, err
		}
		person = *existing
	}
	context.PersonID = person.ID
	if _, err := s.store.AddPersonToNetwork(ctx, context); err != nil {
		return nil, err
	}
	return s.store.GetNetworkPerson(ctx, networkID, person.ID)
}

// ListPersons returns persons in an owned network.
func (s *NetworkService) ListPersons(ctx context.Context, actor *models.Account, networkID string) ([]models.NetworkPerson, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	return s.store.ListNetworkPersons(ctx, networkID)
}

// GetPerson returns one network person.
func (s *NetworkService) GetPerson(ctx context.Context, actor *models.Account, networkID string, personID string) (*models.NetworkPerson, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	return s.store.GetNetworkPerson(ctx, networkID, personID)
}

// UpdatePersonContext updates network-specific person context.
func (s *NetworkService) UpdatePersonContext(ctx context.Context, actor *models.Account, networkID string, personID string, context models.NetworkPersonContext) (*models.NetworkPersonContext, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	context.NetworkID = networkID
	context.PersonID = personID
	context.Notes = normalizeSpace(context.Notes)
	context.Context = normalizeSpace(context.Context)
	return s.store.UpdateNetworkPersonContext(ctx, context)
}

// ArchivePerson archives a person membership in an owned network.
func (s *NetworkService) ArchivePerson(ctx context.Context, actor *models.Account, networkID string, personID string) error {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return err
	}
	return s.store.ArchiveNetworkPerson(ctx, networkID, personID)
}

// Graph returns graph data scoped to an owned network.
func (s *NetworkService) Graph(ctx context.Context, actor *models.Account, networkID string) (*models.NetworkGraphResponse, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	return s.store.GetNetworkGraph(ctx, networkID)
}

// SavePosition stores a graph position scoped to an owned network.
func (s *NetworkService) SavePosition(ctx context.Context, actor *models.Account, networkID string, position models.Position) error {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return err
	}
	if position.NodeID == "" || position.NodeType == "" {
		return ErrInvalidInput
	}
	return s.store.SaveNetworkPosition(ctx, networkID, position)
}

// CreateDiagramNode creates a diagram-only node scoped to an owned network.
func (s *NetworkService) CreateDiagramNode(ctx context.Context, actor *models.Account, networkID string, node models.DiagramNode) (*models.DiagramNode, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	network, err := s.store.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}
	if normalizeNetworkDomain(network.Domain) != "flowchart" {
		return nil, ErrInvalidInput
	}
	node.NetworkID = networkID
	node.Type = normalizeSpace(node.Type)
	node.Name = normalizeSpace(node.Name)
	node.Description = normalizeSpace(node.Description)
	node.Notes = normalizeSpace(node.Notes)
	if node.ID == "" {
		node.ID = newID("diag")
	}
	if !validDiagramNodeType(node.Type) || node.Name == "" {
		return nil, ErrInvalidInput
	}
	return s.store.CreateDiagramNode(ctx, node)
}

// UpdateDiagramNode updates a diagram-only node scoped to an owned network.
func (s *NetworkService) UpdateDiagramNode(ctx context.Context, actor *models.Account, networkID string, id string, node models.DiagramNode) (*models.DiagramNode, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	network, err := s.store.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}
	if normalizeNetworkDomain(network.Domain) != "flowchart" {
		return nil, ErrInvalidInput
	}
	node.ID = id
	node.NetworkID = networkID
	node.Type = normalizeSpace(node.Type)
	node.Name = normalizeSpace(node.Name)
	node.Description = normalizeSpace(node.Description)
	node.Notes = normalizeSpace(node.Notes)
	if !validDiagramNodeType(node.Type) || node.Name == "" {
		return nil, ErrInvalidInput
	}
	return s.store.UpdateDiagramNode(ctx, node)
}

// DeleteDiagramNode permanently removes a diagram-only node scoped to an owned network.
func (s *NetworkService) DeleteDiagramNode(ctx context.Context, actor *models.Account, networkID string, id string) error {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return err
	}
	return s.store.DeleteDiagramNode(ctx, networkID, id)
}

// ValidateRelationshipForNetwork enforces domain-specific relationship rules for an owned network.
func (s *NetworkService) ValidateRelationshipForNetwork(ctx context.Context, actor *models.Account, relationship models.Relationship) error {
	if relationship.NetworkID == "" {
		return nil
	}
	network, err := s.Get(ctx, actor, relationship.NetworkID)
	if err != nil {
		return err
	}
	domain := normalizeNetworkDomain(network.Domain)
	switch domain {
	case "flowchart":
		if !isDiagramNodeType(relationship.SourceType) ||
			!isDiagramNodeType(relationship.TargetType) ||
			!validFlowchartRelationshipType(relationship.Type) {
			return ErrInvalidInput
		}
		if _, err := s.store.GetDiagramNode(ctx, relationship.NetworkID, relationship.SourceID); err != nil {
			return err
		}
		if _, err := s.store.GetDiagramNode(ctx, relationship.NetworkID, relationship.TargetID); err != nil {
			return err
		}
	case "social":
		if isDiagramNodeType(relationship.SourceType) ||
			isDiagramNodeType(relationship.TargetType) ||
			validFlowchartRelationshipType(relationship.Type) {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}
	return nil
}

// CreateCustomRelationshipType creates or updates a reusable custom relationship type for an owned network.
func (s *NetworkService) CreateCustomRelationshipType(ctx context.Context, actor *models.Account, networkID string, relationshipType models.CustomRelationshipType) (*models.CustomRelationshipType, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	relationshipType.NetworkID = networkID
	relationshipType.OwnerID = actor.ID
	relationshipType.Key = normalizeSpace(relationshipType.Key)
	relationshipType.Label = normalizeSpace(relationshipType.Label)
	relationshipType.SourceType = normalizeSpace(relationshipType.SourceType)
	relationshipType.TargetType = normalizeSpace(relationshipType.TargetType)
	relationshipType.DirectionBehavior = normalizeSpace(relationshipType.DirectionBehavior)
	if relationshipType.DirectionBehavior == "" {
		relationshipType.DirectionBehavior = "directed"
	}
	if relationshipType.ID == "" {
		relationshipType.ID = newID("reltype")
	}
	if relationshipType.Label == "" ||
		relationshipType.SourceType == "" ||
		relationshipType.TargetType == "" ||
		!validCustomRelationshipType(relationshipType.Key) ||
		!validCustomRelationshipDirection(relationshipType.DirectionBehavior) {
		return nil, ErrInvalidInput
	}
	return s.store.CreateCustomRelationshipType(ctx, relationshipType)
}

// ListCustomRelationshipTypes lists reusable custom relationship types for an owned network.
func (s *NetworkService) ListCustomRelationshipTypes(ctx context.Context, actor *models.Account, networkID string) ([]models.CustomRelationshipType, error) {
	if err := s.requireOwner(ctx, actor, networkID); err != nil {
		return nil, err
	}
	return s.store.ListCustomRelationshipTypes(ctx, networkID)
}

// MatchPersons returns duplicate suggestions for global persons.
func (s *NetworkService) MatchPersons(ctx context.Context, actor *models.Account, req models.PersonMatchRequest) ([]models.PersonMatchSuggestion, error) {
	if err := requireActiveActor(actor); err != nil {
		return nil, err
	}
	req.Name = normalizeSpace(req.Name)
	req.Nickname = normalizeSpace(req.Nickname)
	req.Organization = normalizeSpace(req.Organization)
	req.School = normalizeSpace(req.School)
	req.Location = normalizeSpace(req.Location)
	req.Relationship = normalizeSpace(req.Relationship)
	if req.Name == "" && req.Nickname == "" {
		return nil, ErrInvalidInput
	}
	persons, err := s.store.ListPersons(ctx)
	if err != nil {
		return nil, err
	}
	suggestions := make([]models.PersonMatchSuggestion, 0)
	for _, person := range persons {
		confidence, reasons := scorePersonMatch(req, person)
		if confidence >= 0.35 {
			suggestions = append(suggestions, models.PersonMatchSuggestion{
				Person:     person,
				Confidence: confidence,
				Reasons:    reasons,
			})
		}
	}
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Confidence == suggestions[j].Confidence {
			return strings.ToLower(suggestions[i].Person.Name) < strings.ToLower(suggestions[j].Person.Name)
		}
		return suggestions[i].Confidence > suggestions[j].Confidence
	})
	if len(suggestions) > 8 {
		suggestions = suggestions[:8]
	}
	return suggestions, nil
}

// MergePersons merges duplicate global persons when network ownership makes it unambiguous.
func (s *NetworkService) MergePersons(ctx context.Context, actor *models.Account, survivorID string, removedID string) (*models.PersonMergeResult, error) {
	if err := requireActiveActor(actor); err != nil {
		return nil, err
	}
	if survivorID == "" || removedID == "" || survivorID == removedID {
		return nil, ErrInvalidInput
	}
	survivorNetworks, err := s.store.ListPersonNetworkIDs(ctx, survivorID)
	if err != nil {
		return nil, err
	}
	removedNetworks, err := s.store.ListPersonNetworkIDs(ctx, removedID)
	if err != nil {
		return nil, err
	}
	affected := unionStrings(survivorNetworks, removedNetworks)
	if len(affected) == 0 {
		return nil, ErrForbidden
	}
	for _, networkID := range affected {
		if err := s.requireOwner(ctx, actor, networkID); err != nil {
			return nil, err
		}
	}
	if hasDisjointReferences(survivorNetworks, removedNetworks) {
		return nil, ErrForbidden
	}
	survivor, err := s.store.MergePersons(ctx, survivorID, removedID)
	if err != nil {
		return nil, err
	}
	return &models.PersonMergeResult{Survivor: *survivor, RemovedPerson: removedID}, nil
}

func (s *NetworkService) requireOwner(ctx context.Context, actor *models.Account, networkID string) error {
	if err := requireActiveActor(actor); err != nil {
		return err
	}
	network, err := s.store.GetNetwork(ctx, networkID)
	if err != nil {
		return err
	}
	if network.Archived {
		return storage.ErrNotFound
	}
	if network.OwnerID != actor.ID {
		return ErrForbidden
	}
	return nil
}

func requireActiveActor(actor *models.Account) error {
	if actor == nil || actor.Disabled {
		return ErrUnauthorized
	}
	return nil
}

func normalizeNetworkDomain(domain string) string {
	domain = normalizeSpace(domain)
	if domain == "" {
		return "social"
	}
	return strings.ToLower(domain)
}

func validNetworkDomain(domain string) bool {
	switch domain {
	case "social", "flowchart":
		return true
	default:
		return false
	}
}

func validDiagramNodeType(value string) bool {
	switch value {
	case "flow_start", "flow_stop", "flow_process", "flow_decision", "flow_input", "flow_output", "flow_merge", "flow_delay":
		return true
	default:
		return false
	}
}

func isDiagramNodeType(value string) bool {
	return strings.HasPrefix(value, "flow_")
}

func validFlowchartRelationshipType(value models.RelationshipType) bool {
	switch value {
	case models.RelationshipNext, models.RelationshipYes, models.RelationshipNo, models.RelationshipLoop, models.RelationshipError:
		return true
	default:
		return false
	}
}

func scorePersonMatch(req models.PersonMatchRequest, person models.Person) (float64, []string) {
	var score float64
	reasons := make([]string, 0)
	reqName := strings.ToLower(req.Name)
	personName := strings.ToLower(person.Name)
	reqNick := strings.ToLower(req.Nickname)
	personNick := strings.ToLower(person.Nickname)
	if reqName != "" && personName != "" {
		switch {
		case reqName == personName:
			score += 0.7
			reasons = append(reasons, "same name")
		case strings.Contains(personName, reqName) || strings.Contains(reqName, personName):
			score += 0.45
			reasons = append(reasons, "similar name")
		}
	}
	if reqNick != "" && personNick != "" {
		if reqNick == personNick {
			score += 0.2
			reasons = append(reasons, "same nickname")
		} else if strings.Contains(personNick, reqNick) || strings.Contains(reqNick, personNick) {
			score += 0.1
			reasons = append(reasons, "similar nickname")
		}
	}
	for _, tag := range person.Tags {
		lowerTag := strings.ToLower(tag)
		if req.Organization != "" && lowerTag == strings.ToLower(req.Organization) {
			score += 0.1
			reasons = append(reasons, "same organization tag")
		}
		if req.School != "" && lowerTag == strings.ToLower(req.School) {
			score += 0.1
			reasons = append(reasons, "same school tag")
		}
		if req.Location != "" && lowerTag == strings.ToLower(req.Location) {
			score += 0.1
			reasons = append(reasons, "same location tag")
		}
	}
	if score > 1 {
		score = 1
	}
	return score, reasons
}

func unionStrings(a []string, b []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	for _, value := range append(a, b...) {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func hasDisjointReferences(a []string, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, value := range a {
		seen[value] = true
	}
	for _, value := range b {
		if seen[value] {
			return false
		}
	}
	return true
}
