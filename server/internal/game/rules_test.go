
package game

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestCanPlayerAffordAbility(t *testing.T) {
	player := core.Player{
		Role: &core.Role{
			IsUnlocked: true,
		},
		HasUsedAbility: false,
	}

	ability := core.Ability{
		IsReady: true,
	}

	// Test that player can afford ability
	if !CanPlayerAffordAbility(player, ability) {
		t.Error("Expected player to be able to afford ability")
	}
}

func TestCanPlayerUseAbility(t *testing.T) {
	// Test case 1: Player can use ability
	player1 := core.Player{
		IsAlive: true,
		Role: &core.Role{
			IsUnlocked: true,
		},
		HasUsedAbility: false,
	}

	ability1 := core.Ability{
		IsReady: true,
	}

	if !CanPlayerUseAbility(player1, ability1) {
		t.Error("Expected player to be able to use ability")
	}

	// Test case 2: Player cannot use ability (not unlocked)
	player2 := core.Player{
		IsAlive: true,
		Role: &core.Role{
			IsUnlocked: false,
		},
	}

	if CanPlayerUseAbility(player2, ability1) {
		t.Error("Expected player to not be able to use locked ability")
	}

	// Test case 3: Player cannot use ability (already used)
	player3 := core.Player{
		IsAlive: true,
		Role: &core.Role{
			IsUnlocked: true,
		},
		HasUsedAbility: true,
	}

	if CanPlayerUseAbility(player3, ability1) {
		t.Error("Expected player to not be able to use ability twice")
	}

	// Test case 4: Player cannot use ability (system shock)
	player4 := core.Player{
		IsAlive: true,
		Role: &core.Role{
			IsUnlocked: true,
		},
		SystemShocks: []core.SystemShock{
			{
				Type:      core.ShockActionLock,
				IsActive:  true,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
		},
	}

	if CanPlayerUseAbility(player4, ability1) {
		t.Error("Expected player to not be able to use ability due to system shock")
	}
}

func TestGetNextPhase(t *testing.T) {
	testCases := []struct {
		current  core.PhaseType
		expected core.PhaseType
	}{
		{core.PhaseLobby, core.PhaseSitrep},
		{core.PhaseSitrep, core.PhasePulseCheck},
		{core.PhasePulseCheck, core.PhaseDiscussion},
		{core.PhaseDiscussion, core.PhaseExtension},
		{core.PhaseExtension, core.PhaseNomination},
		{core.PhaseNomination, core.PhaseTrial},
		{core.PhaseTrial, core.PhaseVerdict},
		{core.PhaseVerdict, core.PhaseNight},
		{core.PhaseNight, core.PhaseSitrep},
		{"UNKNOWN", core.PhaseGameOver},
	}

	for _, tc := range testCases {
		t.Run(string(tc.current), func(t *testing.T) {
			next := GetNextPhase(tc.current)
			if next != tc.expected {
				t.Errorf("Expected next phase to be %s, got %s", tc.expected, next)
			}
		})
	}
}