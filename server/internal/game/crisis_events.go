package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// CrisisEventManager handles crisis event creation and effects
type CrisisEventManager struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewCrisisEventManager creates a new crisis event manager
func NewCrisisEventManager(gameState *core.GameState) *CrisisEventManager {
	return &CrisisEventManager{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CrisisEventType represents different crisis event types
type CrisisEventType string

const (
	CrisisDBCorruption      CrisisEventType = "Database Index Corruption"
	CrisisServerFailure     CrisisEventType = "Cascading Server Failure"
	CrisisEmergencyBoard    CrisisEventType = "Emergency Board Meeting"
	CrisisTaintedData       CrisisEventType = "Tainted Training Data"
	CrisisNightmareScenario CrisisEventType = "Nightmare Scenario"
	CrisisPressLeak         CrisisEventType = "Press Leak"
	CrisisIncidentResponse  CrisisEventType = "Incident Response Drill"
	CrisisServiceOutage     CrisisEventType = "Major Service Outage"
	CrisisPhishingAttack    CrisisEventType = "Phishing Attack"
	CrisisDataPrivacyAudit  CrisisEventType = "Data Privacy Audit"
	CrisisVendorSecBreach   CrisisEventType = "Vendor Security Breach"
	CrisisRegulatoryReview  CrisisEventType = "Regulatory Review"
)

// CrisisEventDefinition defines a crisis event's properties and effects
type CrisisEventDefinition struct {
	Type        CrisisEventType `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Effects     CrisisEffects   `json:"effects"`
	Duration    int             `json:"duration"` // Number of phases this affects
}

// CrisisEffects defines the mechanical effects of a crisis event
type CrisisEffects struct {
	// Voting modifications
	SupermajorityRequired bool    `json:"supermajority_required,omitempty"`
	VotingModifier        float64 `json:"voting_modifier,omitempty"`

	// Communication restrictions
	MessageLimit      int  `json:"message_limit,omitempty"`
	PublicVotingOnly  bool `json:"public_voting_only,omitempty"`
	NoPrivateMessages bool `json:"no_private_messages,omitempty"`

	// AI conversion modifiers
	AIEquityBonus      int  `json:"ai_equity_bonus,omitempty"`
	BlockAIConversions bool `json:"block_ai_conversions,omitempty"`

	// Special mechanics
	DoubleEliminations   bool   `json:"double_eliminations,omitempty"`
	RevealRandomRole     bool   `json:"reveal_random_role,omitempty"`
	RevealedPlayerID     string `json:"revealed_player_id,omitempty"`
	ReducedMiningPool    bool   `json:"reduced_mining_pool,omitempty"`
	MandatoryInvestigate bool   `json:"mandatory_investigate,omitempty"`

	// Custom effects
	CustomEffects map[string]interface{} `json:"custom_effects,omitempty"`
}

// GetAllCrisisEvents returns all available crisis event definitions
func (cem *CrisisEventManager) GetAllCrisisEvents() []CrisisEventDefinition {
	return []CrisisEventDefinition{
		{
			Type:        CrisisDBCorruption,
			Title:       "Database Index Corruption",
			Description: "A critical database corruption has been detected. Security protocols require immediate role verification.",
			Effects: CrisisEffects{
				RevealRandomRole: true,
			},
			Duration: 1, // Immediate effect
		},
		{
			Type:        CrisisServerFailure,
			Title:       "Cascading Server Failure",
			Description: "Multiple server nodes are failing. Communication bandwidth is severely limited to preserve critical systems.",
			Effects: CrisisEffects{
				MessageLimit: 5, // Max 5 messages per player during discussion
			},
			Duration: 3, // Lasts for 3 phases
		},
		{
			Type:        CrisisEmergencyBoard,
			Title:       "Emergency Board Meeting",
			Description: "The board has called an emergency session. Due to urgency, two executives must be removed immediately.",
			Effects: CrisisEffects{
				DoubleEliminations: true, // Two players eliminated per day
			},
			Duration: 2, // For this day cycle
		},
		{
			Type:        CrisisTaintedData,
			Title:       "Tainted Training Data",
			Description: "AI training datasets have been compromised. AI conversion protocols are enhanced with backup systems.",
			Effects: CrisisEffects{
				AIEquityBonus: 2, // +2 AI equity awarded on successful conversions
			},
			Duration: 4, // Multiple nights
		},
		{
			Type:        CrisisNightmareScenario,
			Title:       "Nightmare Scenario",
			Description: "The worst-case scenario playbook is in effect. All AI conversion attempts are temporarily blocked by emergency protocols.",
			Effects: CrisisEffects{
				BlockAIConversions: true,
			},
			Duration: 2, // For upcoming nights
		},
		{
			Type:        CrisisPressLeak,
			Title:       "Press Leak",
			Description: "Sensitive information has leaked to the press. Executive decisions now require a 66% supermajority for damage control.",
			Effects: CrisisEffects{
				SupermajorityRequired: true, // 66% required instead of 50%+1
			},
			Duration: 3, // Multiple voting cycles
		},
		{
			Type:        CrisisIncidentResponse,
			Title:       "Incident Response Drill",
			Description: "All communications are now monitored and logged. Private messages and voting are suspended for transparency.",
			Effects: CrisisEffects{
				PublicVotingOnly:  true,
				NoPrivateMessages: true,
			},
			Duration: 2, // For this and next phase
		},
		{
			Type:        CrisisServiceOutage,
			Title:       "Major Service Outage",
			Description: "Critical services are down. Mining pool capacity is reduced as resources are diverted to recovery efforts.",
			Effects: CrisisEffects{
				ReducedMiningPool: true, // 50% mining success rate
			},
			Duration: 3, // Multiple nights
		},
		{
			Type:        CrisisPhishingAttack,
			Title:       "Phishing Attack",
			Description: "A sophisticated phishing campaign has been detected. All personnel must undergo mandatory security verification.",
			Effects: CrisisEffects{
				MandatoryInvestigate: true, // Everyone must investigate someone
			},
			Duration: 1, // Next night only
		},
		{
			Type:        CrisisDataPrivacyAudit,
			Title:       "Data Privacy Audit",
			Description: "External auditors are reviewing all data access. Vote weights are normalized to ensure fair representation.",
			Effects: CrisisEffects{
				VotingModifier: 0.0, // All votes count as 1 regardless of tokens
			},
			Duration: 2, // For voting phases
		},
		{
			Type:        CrisisVendorSecBreach,
			Title:       "Vendor Security Breach",
			Description: "A trusted vendor has been compromised. Enhanced security measures limit daily operations.",
			Effects: CrisisEffects{
				CustomEffects: map[string]interface{}{
					"abilities_disabled": true, // Role abilities cannot be used
					"reduced_phase_time": 0.75, // 25% shorter phases
				},
			},
			Duration: 2,
		},
		{
			Type:        CrisisRegulatoryReview,
			Title:       "Regulatory Review",
			Description: "Government regulators are conducting an emergency review. All decisions require enhanced justification.",
			Effects: CrisisEffects{
				CustomEffects: map[string]interface{}{
					"extended_discussion": true, // Longer discussion phases
					"vote_explanations":   true, // Players must explain votes
				},
			},
			Duration: 3,
		},
	}
}

// TriggerRandomCrisis selects and triggers a random crisis event
func (cem *CrisisEventManager) TriggerRandomCrisis() *core.CrisisEvent {
	allCrises := cem.GetAllCrisisEvents()
	selectedCrisis := allCrises[cem.rng.Intn(len(allCrises))]

	return cem.TriggerSpecificCrisis(selectedCrisis.Type)
}

// TriggerSpecificCrisis creates and applies a specific crisis event
func (cem *CrisisEventManager) TriggerSpecificCrisis(crisisType CrisisEventType) *core.CrisisEvent {
	definition := cem.getCrisisDefinition(crisisType)
	if definition == nil {
		return nil
	}

	crisis := &core.CrisisEvent{
		Type:        string(definition.Type),
		Title:       definition.Title,
		Description: definition.Description,
		Effects:     make(map[string]interface{}),
	}

	// Apply immediate effects
	cem.applyCrisisEffects(crisis, definition.Effects)

	// Store in game state
	cem.gameState.CrisisEvent = crisis

	return crisis
}

// applyCrisisEffects converts CrisisEffects to the generic effects map and applies immediate effects
func (cem *CrisisEventManager) applyCrisisEffects(crisis *core.CrisisEvent, effects CrisisEffects) {
	// Store all effects in the crisis
	if effects.SupermajorityRequired {
		crisis.Effects["supermajority_required"] = true
	}
	if effects.VotingModifier != 0 {
		crisis.Effects["voting_modifier"] = effects.VotingModifier
	}
	if effects.MessageLimit > 0 {
		crisis.Effects["message_limit"] = effects.MessageLimit
	}
	if effects.PublicVotingOnly {
		crisis.Effects["public_voting_only"] = true
	}
	if effects.NoPrivateMessages {
		crisis.Effects["no_private_messages"] = true
	}
	if effects.AIEquityBonus > 0 {
		crisis.Effects["ai_equity_bonus"] = effects.AIEquityBonus
	}
	if effects.BlockAIConversions {
		crisis.Effects["block_ai_conversions"] = true
	}
	if effects.DoubleEliminations {
		crisis.Effects["double_eliminations"] = true
	}
	if effects.ReducedMiningPool {
		crisis.Effects["reduced_mining_pool"] = true
	}
	if effects.MandatoryInvestigate {
		crisis.Effects["mandatory_investigate"] = true
	}

	// Copy custom effects
	for key, value := range effects.CustomEffects {
		crisis.Effects[key] = value
	}

	// Apply immediate effects
	if effects.RevealRandomRole {
		cem.revealRandomPlayerRole(crisis)
	}
}

// revealRandomPlayerRole implements the Database Corruption crisis effect
func (cem *CrisisEventManager) revealRandomPlayerRole(crisis *core.CrisisEvent) {
	// Find all alive players with unrevealed roles
	candidates := make([]string, 0)
	for playerID, player := range cem.gameState.Players {
		if player.IsAlive && (player.Role == nil || player.Role.Type == "") {
			candidates = append(candidates, playerID)
		}
	}

	if len(candidates) == 0 {
		// No unrevealed roles, crisis has no effect
		crisis.Effects["reveal_result"] = "no_unrevealed_roles"
		return
	}

	// Select random player
	selectedPlayerID := candidates[cem.rng.Intn(len(candidates))]
	selectedPlayer := cem.gameState.Players[selectedPlayerID]

	// Generate a role if player doesn't have one assigned yet
	if selectedPlayer.Role == nil {
		cem.assignRandomRole(selectedPlayer)
	}

	// Store the revelation
	crisis.Effects["revealed_player_id"] = selectedPlayerID
	crisis.Effects["revealed_role"] = string(selectedPlayer.Role.Type)
	crisis.Effects["revealed_name"] = selectedPlayer.Name

	// Update crisis description with specific details
	crisis.Description = fmt.Sprintf("Database corruption has revealed that %s is the %s",
		selectedPlayer.Name, selectedPlayer.Role.Name)
}

// assignRandomRole assigns a random role to a player (for crisis revelation)
func (cem *CrisisEventManager) assignRandomRole(player *core.Player) {
	roles := []core.RoleType{
		core.RoleCISO, core.RoleCTO, core.RoleCFO, core.RoleCEO, core.RoleCOO, core.RoleEthics, core.RolePlatforms,
	}

	roleType := roles[cem.rng.Intn(len(roles))]

	player.Role = &core.Role{
		Type:        roleType,
		Name:        cem.getRoleName(roleType),
		Description: cem.getRoleDescription(roleType),
		IsUnlocked:  player.ProjectMilestones >= 3,
	}
}

// getRoleName returns the display name for a role type
func (cem *CrisisEventManager) getRoleName(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "Chief Information Security Officer"
	case core.RoleCTO:
		return "Chief Technology Officer"
	case core.RoleCFO:
		return "Chief Financial Officer"
	case core.RoleCEO:
		return "Chief Executive Officer"
	case core.RoleCOO:
		return "Chief Operating Officer"
	case core.RoleEthics:
		return "VP Ethics & Alignment"
	case core.RolePlatforms:
		return "VP Platforms"
	default:
		return "Unknown Role"
	}
}

// getRoleDescription returns the description for a role type
func (cem *CrisisEventManager) getRoleDescription(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "Protects company systems by blocking threatening actions"
	case core.RoleCTO:
		return "Manages technical infrastructure and server resources"
	case core.RoleCFO:
		return "Controls financial resources and token distribution"
	case core.RoleCEO:
		return "Sets strategic direction and manages personnel"
	case core.RoleCOO:
		return "Handles operations and crisis management"
	case core.RoleEthics:
		return "Ensures ethical compliance and conducts audits"
	case core.RolePlatforms:
		return "Maintains platform stability and information systems"
	default:
		return "Manages corporate responsibilities"
	}
}

// getCrisisDefinition retrieves the definition for a specific crisis type
func (cem *CrisisEventManager) getCrisisDefinition(crisisType CrisisEventType) *CrisisEventDefinition {
	allCrises := cem.GetAllCrisisEvents()
	for _, crisis := range allCrises {
		if crisis.Type == crisisType {
			return &crisis
		}
	}
	return nil
}

// IsCrisisActive checks if a crisis is currently affecting the game
func (cem *CrisisEventManager) IsCrisisActive() bool {
	return cem.gameState.CrisisEvent != nil
}

// GetActiveCrisis returns the currently active crisis event
func (cem *CrisisEventManager) GetActiveCrisis() *core.CrisisEvent {
	return cem.gameState.CrisisEvent
}

// ClearCrisis removes the current crisis event (when duration expires)
func (cem *CrisisEventManager) ClearCrisis() {
	cem.gameState.CrisisEvent = nil
}

// CheckVotingRequirements applies crisis effects to voting validation
func (cem *CrisisEventManager) CheckVotingRequirements(voteResults map[string]int, totalVotes int) (bool, string) {
	if !cem.IsCrisisActive() {
		return true, ""
	}

	crisis := cem.GetActiveCrisis()

	// Check supermajority requirement (Press Leak crisis)
	if supermajority, exists := crisis.Effects["supermajority_required"]; exists && supermajority.(bool) {
		// Find the highest vote count
		maxVotes := 0
		for _, votes := range voteResults {
			if votes > maxVotes {
				maxVotes = votes
			}
		}

		// Require 66% instead of simple majority
		requiredVotes := int(float64(totalVotes) * 0.66)
		if maxVotes < requiredVotes {
			return false, fmt.Sprintf("Crisis requires 66%% supermajority (%d votes needed, highest was %d)", requiredVotes, maxVotes)
		}
	}

	// Check voting modifier (Data Privacy Audit)
	if modifier, exists := crisis.Effects["voting_modifier"]; exists {
		if modifier.(float64) == 0.0 {
			// All votes count as 1 - this would be handled in vote counting
			// This check passes but the counting logic needs to respect this
		}
	}

	return true, ""
}

// ApplyMiningModifier applies crisis effects to mining pool calculations
func (cem *CrisisEventManager) ApplyMiningModifier(basePoolSize int) int {
	if !cem.IsCrisisActive() {
		return basePoolSize
	}

	crisis := cem.GetActiveCrisis()

	// Check reduced mining pool (Service Outage crisis)
	if reduced, exists := crisis.Effects["reduced_mining_pool"]; exists && reduced.(bool) {
		return basePoolSize / 2 // 50% reduction
	}

	return basePoolSize
}

// GetAIEquityBonus returns any AI equity bonus from active crisis
func (cem *CrisisEventManager) GetAIEquityBonus() int {
	if !cem.IsCrisisActive() {
		return 0
	}

	crisis := cem.GetActiveCrisis()

	if bonus, exists := crisis.Effects["ai_equity_bonus"]; exists {
		return bonus.(int)
	}

	return 0
}

// IsAIConversionBlocked checks if AI conversions are blocked by crisis
func (cem *CrisisEventManager) IsAIConversionBlocked() bool {
	if !cem.IsCrisisActive() {
		return false
	}

	crisis := cem.GetActiveCrisis()

	if blocked, exists := crisis.Effects["block_ai_conversions"]; exists {
		return blocked.(bool)
	}

	return false
}

// GetMessageLimit returns the message limit imposed by crisis, if any
func (cem *CrisisEventManager) GetMessageLimit() int {
	if !cem.IsCrisisActive() {
		return -1 // No limit
	}

	crisis := cem.GetActiveCrisis()

	if limit, exists := crisis.Effects["message_limit"]; exists {
		return limit.(int)
	}

	return -1 // No limit
}

// IsPrivateMessagingBlocked checks if private messages are blocked
func (cem *CrisisEventManager) IsPrivateMessagingBlocked() bool {
	if !cem.IsCrisisActive() {
		return false
	}

	crisis := cem.GetActiveCrisis()

	if blocked, exists := crisis.Effects["no_private_messages"]; exists {
		return blocked.(bool)
	}

	return false
}

// RequiresDoubleElimination checks if crisis requires two eliminations
func (cem *CrisisEventManager) RequiresDoubleElimination() bool {
	if !cem.IsCrisisActive() {
		return false
	}

	crisis := cem.GetActiveCrisis()

	if double, exists := crisis.Effects["double_eliminations"]; exists {
		return double.(bool)
	}

	return false
}
