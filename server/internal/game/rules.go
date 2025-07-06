
package game

import (
	"time"

	"github.com/xjhc/alignment/core"
)

// CanPlayerAffordAbility checks if a player can afford to use their role ability
func CanPlayerAffordAbility(player core.Player, ability core.Ability) bool {
	// For now, abilities don't have costs, so they're always affordable if unlocked
	return true
}

// CanPlayerUseAbility checks if a player can use their role ability
func CanPlayerUseAbility(player core.Player, ability core.Ability) bool {
	if !player.IsAlive {
		return false
	}

	if player.Role == nil || !player.Role.IsUnlocked {
		return false
	}

	if player.HasUsedAbility {
		return false
	}

	// Check for system shock
	for _, shock := range player.SystemShocks {
		if shock.Type == core.ShockActionLock && shock.IsActive && time.Now().Before(shock.ExpiresAt) {
			return false
		}
	}

	// Check for crisis effects that disable abilities
	// This would require access to the game state or passing crisis effects here
	// For now, we'll assume no crisis effects

	return true
}

// GetNextPhase returns the next phase after the current one
func GetNextPhase(currentPhase core.PhaseType) core.PhaseType {
	switch currentPhase {
	case core.PhaseLobby:
		return core.PhaseSitrep
	case core.PhaseSitrep:
		return core.PhasePulseCheck
	case core.PhasePulseCheck:
		return core.PhaseDiscussion
	case core.PhaseDiscussion:
		return core.PhaseExtension
	case core.PhaseExtension:
		return core.PhaseNomination
	case core.PhaseNomination:
		return core.PhaseTrial
	case core.PhaseTrial:
		return core.PhaseVerdict
	case core.PhaseVerdict:
		return core.PhaseNight
	case core.PhaseNight:
		return core.PhaseSitrep
	default:
		return core.PhaseGameOver
	}
}

// GetPhaseDuration returns the duration for a specific phase
func GetPhaseDuration(phase core.PhaseType, settings core.GameSettings) time.Duration {
	switch phase {
	case core.PhaseSitrep:
		return settings.SitrepDuration
	case core.PhasePulseCheck:
		return settings.PulseCheckDuration
	case core.PhaseDiscussion:
		return settings.DiscussionDuration
	case core.PhaseExtension:
		return settings.ExtensionDuration
	case core.PhaseNomination:
		return settings.NominationDuration
	case core.PhaseTrial:
		return settings.TrialDuration
	case core.PhaseVerdict:
		return settings.VerdictDuration
	case core.PhaseNight:
		return settings.NightDuration
	default:
		return 0
	}
}