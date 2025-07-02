package party

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// Party represents a group of friends who want to play together
type Party struct {
	ID      string                                         `json:"id"`
	Leader  string                                         `json:"leader"`   // Player ID of the party leader
	Members map[string]interfaces.PlayerActorInterface    `json:"members"`  // Map of playerID -> PlayerActor
	Invites map[string]*interfaces.PartyInvite                       `json:"invites"`  // Map of playerID -> invite
	CreatedAt time.Time                                   `json:"createdAt"`
	mutex   sync.RWMutex
}


// PartyManager manages party creation, invites, and coordination
type PartyManager struct {
	parties map[string]*Party        // Map of partyID -> Party
	invites map[string]*interfaces.PartyInvite  // Map of inviteID -> interfaces.PartyInvite
	playerParty map[string]string    // Map of playerID -> partyID
	mutex   sync.RWMutex
}

// NewPartyManager creates a new party manager
func NewPartyManager() *PartyManager {
	pm := &PartyManager{
		parties:     make(map[string]*Party),
		invites:     make(map[string]*interfaces.PartyInvite),
		playerParty: make(map[string]string),
	}

	// Start cleanup routine for expired invites
	go pm.startCleanupRoutine()

	return pm
}

// startCleanupRoutine periodically removes expired invites
func (pm *PartyManager) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		pm.cleanupExpiredInvites()
	}
}

// cleanupExpiredInvites removes invites that have expired
func (pm *PartyManager) cleanupExpiredInvites() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	now := time.Now()
	var expiredInvites []string

	for inviteID, invite := range pm.invites {
		if now.After(invite.ExpiresAt) && invite.Status == "pending" {
			expiredInvites = append(expiredInvites, inviteID)
		}
	}

	// Remove expired invites
	for _, inviteID := range expiredInvites {
		invite := pm.invites[inviteID]
		if party, exists := pm.parties[invite.PartyID]; exists {
			party.mutex.Lock()
			delete(party.Invites, invite.InviteeID)
			party.mutex.Unlock()
		}
		delete(pm.invites, inviteID)
	}
}

// CreateParty creates a new party with the given player as leader
func (pm *PartyManager) CreateParty(leader interfaces.PlayerActorInterface) (string, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	leaderID := leader.GetPlayerID()

	// Check if player is already in a party
	if _, exists := pm.playerParty[leaderID]; exists {
		return "", errors.New("player is already in a party")
	}

	partyID := uuid.New().String()
	party := &Party{
		ID:        partyID,
		Leader:    leaderID,
		Members:   make(map[string]interfaces.PlayerActorInterface),
		Invites:   make(map[string]*interfaces.PartyInvite),
		CreatedAt: time.Now(),
	}

	party.Members[leaderID] = leader
	pm.parties[partyID] = party
	pm.playerParty[leaderID] = partyID

	return partyID, nil
}

// InviteToPartyByInviter creates an invitation for a player to join the inviter's party
func (pm *PartyManager) InviteToPartyByInviter(inviterID, inviteeID string) (*interfaces.PartyInvite, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Find the inviter's party
	partyID, exists := pm.playerParty[inviterID]
	if !exists {
		return nil, errors.New("inviter is not in a party")
	}

	party, exists := pm.parties[partyID]
	if !exists {
		return nil, errors.New("party not found")
	}

	// Validate inviter is party leader
	if party.Leader != inviterID {
		return nil, errors.New("only party leader can send invites")
	}

	// Check if invitee is already in a party
	if _, exists := pm.playerParty[inviteeID]; exists {
		return nil, errors.New("player is already in a party")
	}

	// Check if there's already a pending invite for this player
	party.mutex.RLock()
	if _, exists := party.Invites[inviteeID]; exists {
		party.mutex.RUnlock()
		return nil, errors.New("invite already pending for this player")
	}
	party.mutex.RUnlock()

	// Create the invite
	inviteID := uuid.New().String()
	invite := &interfaces.PartyInvite{
		ID:        inviteID,
		PartyID:   partyID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    "pending",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute), // Invites expire after 5 minutes
	}

	pm.invites[inviteID] = invite
	
	party.mutex.Lock()
	party.Invites[inviteeID] = invite
	party.mutex.Unlock()

	return invite, nil
}

// InviteToParty creates an invitation for a player to join a party
func (pm *PartyManager) InviteToParty(partyID, inviterID, inviteeID string) (*interfaces.PartyInvite, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Validate party exists
	party, exists := pm.parties[partyID]
	if !exists {
		return nil, errors.New("party not found")
	}

	// Validate inviter is party leader
	if party.Leader != inviterID {
		return nil, errors.New("only party leader can send invites")
	}

	// Check if invitee is already in a party
	if _, exists := pm.playerParty[inviteeID]; exists {
		return nil, errors.New("player is already in a party")
	}

	// Check if there's already a pending invite for this player
	party.mutex.RLock()
	if _, exists := party.Invites[inviteeID]; exists {
		party.mutex.RUnlock()
		return nil, errors.New("invite already pending for this player")
	}
	party.mutex.RUnlock()

	// Create the invite
	inviteID := uuid.New().String()
	invite := &interfaces.PartyInvite{
		ID:        inviteID,
		PartyID:   partyID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    "pending",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute), // Invites expire after 5 minutes
	}

	pm.invites[inviteID] = invite
	
	party.mutex.Lock()
	party.Invites[inviteeID] = invite
	party.mutex.Unlock()

	return invite, nil
}

// JoinParty allows a player to join a party via invite
func (pm *PartyManager) JoinParty(playerID, inviteID string, playerActor interfaces.PlayerActorInterface) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Validate invite exists
	invite, exists := pm.invites[inviteID]
	if !exists {
		return errors.New("invite not found")
	}

	// Validate player matches invite
	if invite.InviteeID != playerID {
		return errors.New("invite not for this player")
	}

	// Validate invite is still pending
	if invite.Status != "pending" {
		return errors.New("invite is no longer pending")
	}

	// Validate invite hasn't expired
	if time.Now().After(invite.ExpiresAt) {
		return errors.New("invite has expired")
	}

	// Check if player is already in a party
	if _, exists := pm.playerParty[playerID]; exists {
		return errors.New("player is already in a party")
	}

	// Validate party still exists
	party, exists := pm.parties[invite.PartyID]
	if !exists {
		return errors.New("party no longer exists")
	}

	// Add player to party
	party.mutex.Lock()
	party.Members[playerID] = playerActor
	delete(party.Invites, playerID) // Remove the invite
	party.mutex.Unlock()

	pm.playerParty[playerID] = invite.PartyID
	invite.Status = "accepted"

	return nil
}

// LeaveParty removes a player from their party
func (pm *PartyManager) LeaveParty(playerID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Check if player is in a party
	partyID, exists := pm.playerParty[playerID]
	if !exists {
		return errors.New("player is not in a party")
	}

	party, exists := pm.parties[partyID]
	if !exists {
		// Clean up orphaned player reference
		delete(pm.playerParty, playerID)
		return errors.New("party not found")
	}

	// Remove player from party
	party.mutex.Lock()
	delete(party.Members, playerID)
	isLeader := party.Leader == playerID
	memberCount := len(party.Members)
	party.mutex.Unlock()

	delete(pm.playerParty, playerID)

	// If this was the leader and there are other members, transfer leadership
	if isLeader && memberCount > 0 {
		pm.transferLeadership(party)
	}

	// If party is now empty, dissolve it
	if memberCount == 0 {
		pm.dissolveParty(partyID)
	}

	return nil
}

// transferLeadership transfers party leadership to another member
func (pm *PartyManager) transferLeadership(party *Party) {
	party.mutex.Lock()
	defer party.mutex.Unlock()

	// Find any member to be the new leader
	for memberID := range party.Members {
		party.Leader = memberID
		break
	}
}

// dissolveParty removes a party completely
func (pm *PartyManager) dissolveParty(partyID string) {
	// Cancel all pending invites for this party
	var invitesToCancel []string
	for inviteID, invite := range pm.invites {
		if invite.PartyID == partyID && invite.Status == "pending" {
			invitesToCancel = append(invitesToCancel, inviteID)
		}
	}

	for _, inviteID := range invitesToCancel {
		pm.invites[inviteID].Status = "cancelled"
		delete(pm.invites, inviteID)
	}

	// Remove the party
	delete(pm.parties, partyID)
}

// GetParty returns party information for a player
func (pm *PartyManager) GetParty(playerID string) (*Party, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	partyID, exists := pm.playerParty[playerID]
	if !exists {
		return nil, errors.New("player is not in a party")
	}

	party, exists := pm.parties[partyID]
	if !exists {
		// Clean up orphaned reference
		delete(pm.playerParty, playerID)
		return nil, errors.New("party not found")
	}

	return party, nil
}

// GetPendingInvites returns all pending invites for a player
func (pm *PartyManager) GetPendingInvites(playerID string) []*interfaces.PartyInvite {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var invites []*interfaces.PartyInvite
	for _, invite := range pm.invites {
		if invite.InviteeID == playerID && invite.Status == "pending" && time.Now().Before(invite.ExpiresAt) {
			invites = append(invites, invite)
		}
	}

	return invites
}

// IsInParty checks if a player is currently in a party
func (pm *PartyManager) IsInParty(playerID string) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	_, exists := pm.playerParty[playerID]
	return exists
}

// GetPartyMembers returns a slice of member IDs for a party
func (pm *PartyManager) GetPartyMembers(partyID string) []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	party, exists := pm.parties[partyID]
	if !exists {
		return nil
	}

	party.mutex.RLock()
	defer party.mutex.RUnlock()

	var members []string
	for memberID := range party.Members {
		members = append(members, memberID)
	}

	return members
}