package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"mate/internal/models"
	"mate/internal/storage"
)

// Services groups application services.
type Services struct {
	Persons       *PersonService
	Organizations *OrganizationService
	Projects      *ProjectService
	Relationships *RelationshipService
	Graph         *GraphService
	Accounts      *AccountService
	Networks      *NetworkService
}

// New creates all application services.
func New(store storage.Storage) *Services {
	return &Services{
		Persons:       &PersonService{store: store},
		Organizations: &OrganizationService{store: store},
		Projects:      &ProjectService{store: store},
		Relationships: &RelationshipService{store: store},
		Graph:         &GraphService{store: store},
		Accounts:      &AccountService{store: store},
		Networks:      &NetworkService{store: store},
	}
}

// ErrInvalidInput indicates that input failed service validation.
var ErrInvalidInput = errors.New("invalid input")

// ErrUnauthorized indicates that a valid authenticated account is required.
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden indicates that the authenticated account lacks permission.
var ErrForbidden = errors.New("forbidden")

func newID(prefix string) string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(err)
	}
	return prefix + "-" + hex.EncodeToString(bytes[:])
}

func newSessionToken() string {
	return newID("tok")
}

func normalizeSpace(value string) string {
	return strings.TrimSpace(value)
}

func validOrganizationType(value models.OrganizationType) bool {
	switch value {
	case models.OrganizationCompany,
		models.OrganizationAssociation,
		models.OrganizationSchool,
		models.OrganizationGovernment,
		models.OrganizationPoliticalParty,
		models.OrganizationReligiousOrganization,
		models.OrganizationSportsClub,
		models.OrganizationMilitaryUnit,
		models.OrganizationNGO,
		models.OrganizationCommunity:
		return true
	default:
		return false
	}
}

func validRelationshipType(value models.RelationshipType) bool {
	switch value {
	case models.RelationshipKnows,
		models.RelationshipSpouseOf,
		models.RelationshipParentOf,
		models.RelationshipSiblingOf,
		models.RelationshipWorksAt,
		models.RelationshipMemberOf,
		models.RelationshipStudiedAt,
		models.RelationshipLivesIn,
		models.RelationshipHasTag,
		models.RelationshipWorksOn,
		models.RelationshipSponsors,
		models.RelationshipPartnerOf,
		models.RelationshipOwns:
		return true
	default:
		return validCustomRelationshipType(string(value))
	}
}

var customRelationshipTypePattern = regexp.MustCompile(`^custom_[a-z][a-z0-9_]{0,63}$`)

func validCustomRelationshipType(value string) bool {
	return customRelationshipTypePattern.MatchString(value)
}

func validCustomRelationshipDirection(value string) bool {
	switch value {
	case "directed", "undirected":
		return true
	default:
		return false
	}
}
