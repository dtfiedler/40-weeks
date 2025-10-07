package models

import (
	"time"
)

type VillageMember struct {
	ID                int       `json:"id" db:"id"`
	PregnancyID       int       `json:"pregnancy_id" db:"pregnancy_id"`
	Name              string    `json:"name" db:"name"`
	Email             string    `json:"email" db:"email"`
	Relationship      string    `json:"relationship" db:"relationship"`
	IsTold            bool      `json:"is_told" db:"is_told"`
	ToldDate          *time.Time `json:"told_date" db:"told_date"`
	IsSubscribed      bool      `json:"is_subscribed" db:"is_subscribed"`
	UnsubscribeToken  *string   `json:"unsubscribe_token" db:"unsubscribe_token"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// VillageStats provides summary statistics about the village
type VillageStats struct {
	TotalMembers    int `json:"total_members"`
	ToldMembers     int `json:"told_members"`
	SubscribedMembers int `json:"subscribed_members"`
	PendingMembers  int `json:"pending_members"`
}

// RelationshipGroup groups village members by relationship type
type RelationshipGroup struct {
	Relationship string          `json:"relationship"`
	Members      []VillageMember `json:"members"`
	Count        int             `json:"count"`
}

// Common relationship types
const (
	RelationshipMother     = "mother"
	RelationshipFather     = "father"
	RelationshipSister     = "sister"
	RelationshipBrother    = "brother"
	RelationshipFriend     = "friend"
	RelationshipCoworker   = "coworker"
	RelationshipGrandparent = "grandparent"
	RelationshipAunt       = "aunt"
	RelationshipUncle      = "uncle"
	RelationshipCousin     = "cousin"
	RelationshipOther      = "other"
)

// GetDisplayRelationship returns a formatted relationship string
func (vm *VillageMember) GetDisplayRelationship() string {
	switch vm.Relationship {
	case RelationshipMother:
		return "Mother"
	case RelationshipFather:
		return "Father"
	case RelationshipSister:
		return "Sister"
	case RelationshipBrother:
		return "Brother"
	case RelationshipFriend:
		return "Friend"
	case RelationshipCoworker:
		return "Coworker"
	case RelationshipGrandparent:
		return "Grandparent"
	case RelationshipAunt:
		return "Aunt"
	case RelationshipUncle:
		return "Uncle"
	case RelationshipCousin:
		return "Cousin"
	default:
		return "Family/Friend"
	}
}

// CanReceiveUpdates checks if village member can receive email updates
func (vm *VillageMember) CanReceiveUpdates() bool {
	return vm.IsTold && vm.IsSubscribed
}