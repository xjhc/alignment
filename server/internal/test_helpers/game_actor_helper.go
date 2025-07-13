package test_helpers

import (
	"context"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/actors"
)

// CreateTestGameActor is a helper function for creating a GameActor for integration tests
func CreateTestGameActor(t *testing.T, numPlayers int) *actors.GameActor {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	
	players := make(map[string]*core.Player)
	for i := 0; i < numPlayers; i++ {
		playerID := string(rune('A' + i))
		players[playerID] = &core.Player{
			ID:           playerID,
			Name:         "Player " + playerID,
			IsAlive:      true,
			ControlType:  "HUMAN",
			Alignment:    core.AlignmentHuman,
			Tokens:       1,
		}
	}
	
	actor := actors.NewGameActor(ctx, cancel, "test-game", players, nil)
	require.NotNil(t, actor, "Failed to create game actor")
	
	actor.Start()
	t.Cleanup(actor.Stop)
	
	return actor
}