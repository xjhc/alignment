package game

import (
	"math/rand"

	"github.com/xjhc/alignment/core"
)

// ActionHandler defines the interface for processing a specific game action.
type ActionHandler interface {
	Handle(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error)
}

// ActionAcker provides a way for handlers to access GameActor-specific utilities
// without creating a circular dependency.
type ActionAcker interface {
	GetPlayer(id string) (*core.Player, bool)
	GetRandom() *rand.Rand
	GetGameID() string
	ValidateActionPayloadSize(action core.Action) error
}