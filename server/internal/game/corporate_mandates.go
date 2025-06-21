package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// CorporateMandateManager handles corporate mandate assignment and effects
type CorporateMandateManager struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewCorporateMandateManager creates a new corporate mandate manager
func NewCorporateMandateManager(gameState *core.GameState) *CorporateMandateManager {
	return &CorporateMandateManager{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Use the existing MandateType from state.go

// LocalCorporateMandate represents a corporate mandate with typed effects (different from state.go version)
type LocalCorporateMandate struct {
	Type        core.MandateType `json:"type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Effects     MandateEffects   `json:"effects"`
	IsActive    bool             `json:"is_active"`
	StartDay    int              `json:"start_day"`
}

// MandateEffects defines the mechanical effects of a corporate mandate
type MandateEffects struct {
	// Starting conditions
	StartingTokensModifier int `json:"starting_tokens_modifier,omitempty"`

	// Mining modifications
	MiningSuccessModifier float64 `json:"mining_success_modifier,omitempty"`
	ReducedMiningSlots    bool    `json:"reduced_mining_slots,omitempty"`

	// Communication restrictions
	PublicVotingOnly bool `json:"public_voting_only,omitempty"`
	NoDirectMessages bool `json:"no_direct_messages,omitempty"`

	// Milestone requirements
	MilestonesForAbilities int `json:"milestones_for_abilities,omitempty"`

	// AI restrictions
	BlockAIOddNights bool `json:"block_ai_odd_nights,omitempty"`

	// Custom effects
	CustomEffects map[string]interface{} `json:"custom_effects,omitempty"`
}

// GetAllCorporateMandates returns all available corporate mandate definitions
func (cmm *CorporateMandateManager) GetAllCorporateMandates() []LocalCorporateMandate {
	return []LocalCorporateMandate{
		{
			Type:        core.MandateAggressiveGrowth,
			Title:       "Aggressive Growth Quarter",
			Description: "The board has declared an aggressive growth period. All personnel start with enhanced resources, but infrastructure capacity is strained.",
			Effects: MandateEffects{
				StartingTokensModifier: 1,    // +1 starting token (total 3)
				MiningSuccessModifier:  0.75, // 25% reduced mining success
				ReducedMiningSlots:     true, // Fewer mining slots available
			},
			IsActive: false,
		},
		{
			Type:        core.MandateTransparency,
			Title:       "Total Transparency Initiative",
			Description: "In response to recent concerns, all company decisions must be made transparently. Private communications and secret voting are suspended.",
			Effects: MandateEffects{
				PublicVotingOnly: true, // All votes must be public
				NoDirectMessages: true, // No private communications allowed
			},
			IsActive: false,
		},
		{
			Type:        core.MandateSecurityLockdown,
			Title:       "Security Lockdown Protocol",
			Description: "Enhanced security measures are in effect. Higher security clearance required for all operations, and AI systems are restricted on odd nights.",
			Effects: MandateEffects{
				MilestonesForAbilities: 4,    // Need 4 milestones instead of 3
				BlockAIOddNights:       true, // AI cannot convert on nights 1, 3, 5, etc.
			},
			IsActive: false,
		},
	}
}

// AssignRandomMandate selects and activates a random corporate mandate
func (cmm *CorporateMandateManager) AssignRandomMandate() *core.CorporateMandate {
	mandates := cmm.GetAllCorporateMandates()
	selectedMandate := mandates[cmm.rng.Intn(len(mandates))]

	return cmm.ActivateMandate(selectedMandate.Type)
}

// ActivateMandate activates a specific corporate mandate
func (cmm *CorporateMandateManager) ActivateMandate(mandateType core.MandateType) *core.CorporateMandate {
	localMandate := cmm.getMandateDefinition(mandateType)
	if localMandate == nil {
		return nil
	}

	// Convert to game state format
	mandate := &core.CorporateMandate{
		Type:        localMandate.Type,
		Name:        localMandate.Title,
		Description: localMandate.Description,
		Effects:     make(map[string]interface{}),
		IsActive:    true,
	}

	// Convert typed effects to generic map
	cmm.convertMandateEffects(localMandate.Effects, mandate.Effects)

	// Set start day
	localMandate.StartDay = cmm.gameState.DayNumber

	// Apply immediate effects
	cmm.applyMandateEffects(*localMandate)

	// Store in game state
	cmm.gameState.CorporateMandate = mandate

	return mandate
}

// convertMandateEffects converts typed effects to generic map
func (cmm *CorporateMandateManager) convertMandateEffects(typedEffects MandateEffects, effectsMap map[string]interface{}) {
	if typedEffects.StartingTokensModifier != 0 {
		effectsMap["starting_tokens_modifier"] = typedEffects.StartingTokensModifier
	}
	if typedEffects.MiningSuccessModifier > 0 {
		effectsMap["mining_success_modifier"] = typedEffects.MiningSuccessModifier
	}
	if typedEffects.ReducedMiningSlots {
		effectsMap["reduced_mining_slots"] = true
	}
	if typedEffects.PublicVotingOnly {
		effectsMap["public_voting_only"] = true
	}
	if typedEffects.NoDirectMessages {
		effectsMap["no_direct_messages"] = true
	}
	if typedEffects.MilestonesForAbilities > 0 {
		effectsMap["milestones_for_abilities"] = typedEffects.MilestonesForAbilities
	}
	if typedEffects.BlockAIOddNights {
		effectsMap["block_ai_odd_nights"] = true
	}
	for key, value := range typedEffects.CustomEffects {
		effectsMap[key] = value
	}
}

// applyMandateEffects applies the mandate's effects to the game state
func (cmm *CorporateMandateManager) applyMandateEffects(mandate LocalCorporateMandate) {
	effects := mandate.Effects

	// Apply starting token modifier (for new players joining)
	if effects.StartingTokensModifier != 0 {
		cmm.gameState.Settings.StartingTokens += effects.StartingTokensModifier

		// Also apply to existing players if this is day 1
		if mandate.StartDay == 1 {
			for _, player := range cmm.gameState.Players {
				if player.IsAlive {
					player.Tokens += effects.StartingTokensModifier
				}
			}
		}
	}

	// Effects are already stored in the CorporateMandate through convertMandateEffects
}

// getMandateDefinition retrieves the definition for a specific mandate type
func (cmm *CorporateMandateManager) getMandateDefinition(mandateType core.MandateType) *LocalCorporateMandate {
	mandates := cmm.GetAllCorporateMandates()
	for _, mandate := range mandates {
		if mandate.Type == mandateType {
			return &mandate
		}
	}
	return nil
}

// IsMandateActive checks if a corporate mandate is currently active
func (cmm *CorporateMandateManager) IsMandateActive() bool {
	return cmm.gameState.CorporateMandate != nil && cmm.gameState.CorporateMandate.IsActive
}

// GetActiveMandate returns the currently active corporate mandate
func (cmm *CorporateMandateManager) GetActiveMandate() *core.CorporateMandate {
	if cmm.IsMandateActive() {
		return cmm.gameState.CorporateMandate
	}
	return nil
}

// CheckMiningRestrictions applies mandate effects to mining operations
func (cmm *CorporateMandateManager) CheckMiningRestrictions() (successModifier float64, slotsReduced bool) {
	if !cmm.IsMandateActive() {
		return 1.0, false
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects

	modifier := 1.0
	if modifierVal, exists := effects["mining_success_modifier"]; exists {
		if modFloat, ok := modifierVal.(float64); ok {
			modifier = modFloat
		}
	}

	slotsReduced = false
	if reducedVal, exists := effects["reduced_mining_slots"]; exists {
		if reduced, ok := reducedVal.(bool); ok {
			slotsReduced = reduced
		}
	}

	return modifier, slotsReduced
}

// CheckCommunicationRestrictions validates if communication is allowed
func (cmm *CorporateMandateManager) CheckCommunicationRestrictions() (publicVotingOnly bool, noDirectMessages bool) {
	if !cmm.IsMandateActive() {
		return false, false
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects

	publicVotingOnly = false
	if pubVal, exists := effects["public_voting_only"]; exists {
		if pub, ok := pubVal.(bool); ok {
			publicVotingOnly = pub
		}
	}

	noDirectMessages = false
	if noMsgVal, exists := effects["no_direct_messages"]; exists {
		if noMsg, ok := noMsgVal.(bool); ok {
			noDirectMessages = noMsg
		}
	}

	return publicVotingOnly, noDirectMessages
}

// GetMilestoneRequirement returns the milestone requirement for abilities
func (cmm *CorporateMandateManager) GetMilestoneRequirement() int {
	if !cmm.IsMandateActive() {
		return 3 // Default requirement
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects

	if milestonesVal, exists := effects["milestones_for_abilities"]; exists {
		if milestones, ok := milestonesVal.(int); ok {
			return milestones
		}
	}

	return 3 // Default requirement
}

// IsAIConversionAllowed checks if AI can convert on the current night
func (cmm *CorporateMandateManager) IsAIConversionAllowed() bool {
	if !cmm.IsMandateActive() {
		return true
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects

	// Check odd night restriction
	if blockVal, exists := effects["block_ai_odd_nights"]; exists {
		if blockOdd, ok := blockVal.(bool); ok && blockOdd {
			// Calculate which night this is
			nightNumber := cmm.gameState.DayNumber
			if nightNumber%2 == 1 { // Odd nights (1, 3, 5, etc.)
				return false
			}
		}
	}

	return true
}

// CheckVotingRestrictions applies mandate effects to voting
func (cmm *CorporateMandateManager) CheckVotingRestrictions() (publicOnly bool, requiresJustification bool) {
	if !cmm.IsMandateActive() {
		return false, false
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects

	// Total Transparency Initiative forces public voting
	if pubVal, exists := effects["public_voting_only"]; exists {
		if pub, ok := pubVal.(bool); ok && pub {
			return true, false
		}
	}

	// Check for custom effects that might require justification
	if requires, exists := effects["require_vote_justification"]; exists {
		if requiresJust, ok := requires.(bool); ok {
			return false, requiresJust
		}
	}

	return false, false
}

// ApplyMandateToMiningPool modifies mining pool based on mandate
func (cmm *CorporateMandateManager) ApplyMandateToMiningPool(baseMiningSlots int) int {
	if !cmm.IsMandateActive() {
		return baseMiningSlots
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects

	// Aggressive Growth Quarter reduces mining slots due to strained infrastructure
	if reducedVal, exists := effects["reduced_mining_slots"]; exists {
		if reduced, ok := reducedVal.(bool); ok && reduced {
			return baseMiningSlots - 1 // Reduce by 1 slot
		}
	}

	return baseMiningSlots
}

// GetMandateStatusMessage returns a status message about the active mandate
func (cmm *CorporateMandateManager) GetMandateStatusMessage() string {
	if !cmm.IsMandateActive() {
		return "No corporate mandate currently active"
	}

	mandate := cmm.GetActiveMandate()
	// Since CorporateMandate doesn't have StartDay, we'll assume it started on day 1
	daysActive := cmm.gameState.DayNumber

	return fmt.Sprintf("Active Mandate: %s (Day %d of implementation)", mandate.Name, daysActive)
}

// GenerateMandateEffectsSummary creates a summary of mandate effects for SITREP
func (cmm *CorporateMandateManager) GenerateMandateEffectsSummary() []string {
	if !cmm.IsMandateActive() {
		return []string{}
	}

	mandate := cmm.GetActiveMandate()
	effects := mandate.Effects
	summary := make([]string, 0)

	// Document each active effect
	if modifierVal, exists := effects["starting_tokens_modifier"]; exists {
		if modifier, ok := modifierVal.(int); ok && modifier != 0 {
			if modifier > 0 {
				summary = append(summary, fmt.Sprintf("Enhanced starting resources (+%d tokens)", modifier))
			} else {
				summary = append(summary, fmt.Sprintf("Reduced starting resources (%d tokens)", modifier))
			}
		}
	}

	if modifierVal, exists := effects["mining_success_modifier"]; exists {
		if modifier, ok := modifierVal.(float64); ok && modifier > 0 && modifier < 1.0 {
			reduction := int((1.0 - modifier) * 100)
			summary = append(summary, fmt.Sprintf("Mining efficiency reduced by %d%%", reduction))
		}
	}

	if reducedVal, exists := effects["reduced_mining_slots"]; exists {
		if reduced, ok := reducedVal.(bool); ok && reduced {
			summary = append(summary, "Reduced mining pool capacity due to infrastructure constraints")
		}
	}

	if pubVal, exists := effects["public_voting_only"]; exists {
		if pub, ok := pubVal.(bool); ok && pub {
			summary = append(summary, "All voting decisions must be made publicly")
		}
	}

	if noMsgVal, exists := effects["no_direct_messages"]; exists {
		if noMsg, ok := noMsgVal.(bool); ok && noMsg {
			summary = append(summary, "Private communications suspended for transparency")
		}
	}

	if milestonesVal, exists := effects["milestones_for_abilities"]; exists {
		if milestones, ok := milestonesVal.(int); ok && milestones > 3 {
			summary = append(summary, fmt.Sprintf("Enhanced security clearance required (%d milestones for abilities)", milestones))
		}
	}

	if blockVal, exists := effects["block_ai_odd_nights"]; exists {
		if block, ok := blockVal.(bool); ok && block {
			summary = append(summary, "AI system restrictions in effect on odd-numbered nights")
		}
	}

	return summary
}

// CheckRoleAbilityRequirements validates if a player can use role abilities under mandate
func (cmm *CorporateMandateManager) CheckRoleAbilityRequirements(player *core.Player) (canUse bool, reason string) {
	if !cmm.IsMandateActive() {
		return true, ""
	}

	mandateReq := cmm.GetMilestoneRequirement()
	if player.ProjectMilestones < mandateReq {
		return false, fmt.Sprintf("corporate mandate requires %d milestones for role abilities", mandateReq)
	}

	return true, ""
}

// DeactivateMandate removes the current corporate mandate (if needed for testing or special events)
func (cmm *CorporateMandateManager) DeactivateMandate() {
	if cmm.gameState.CorporateMandate != nil {
		cmm.gameState.CorporateMandate.IsActive = false
	}
}
