package mcp

import "github.com/xjhc/alignment/core"

// PublicGameState is a filtered view of the game state, safe to expose to an AI.
type PublicGameState struct {
	YourPlayerID string                `json:"your_player_id"`
	GameID       string                `json:"game_id"`
	CurrentPhase string                `json:"current_phase"`
	DayNumber    int                   `json:"day_number"`
	Players      []PublicPlayer        `json:"players"`
	ChatLog      []core.ChatMessage    `json:"chat_log"`
	CrisisEvent  *core.CrisisEvent     `json:"crisis_event,omitempty"`
	VoteState    *core.VoteState       `json:"vote_state,omitempty"`
}

// PublicPlayer contains only the public information about a player.
type PublicPlayer struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	IsAlive           bool   `json:"is_alive"`
	Tokens            int    `json:"tokens"`
	ProjectMilestones int    `json:"project_milestones"`
	StatusMessage     string `json:"status_message"`
}

// FilterGameStateForAI creates a secure, public-only view of the game state.
// It strips all secret information (alignments, roles, etc.) for all players except the AI itself.
func FilterGameStateForAI(state *core.GameState, aiPlayerID string) PublicGameState {
	publicPlayers := make([]PublicPlayer, 0, len(state.Players))
	for _, p := range state.Players {
		publicPlayers = append(publicPlayers, PublicPlayer{
			ID:                p.ID,
			Name:              p.Name,
			IsAlive:           p.IsAlive,
			Tokens:            p.Tokens,
			ProjectMilestones: p.ProjectMilestones,
			StatusMessage:     p.StatusMessage,
		})
	}

	return PublicGameState{
		YourPlayerID: aiPlayerID,
		GameID:       state.ID,
		CurrentPhase: string(state.Phase.Type),
		DayNumber:    state.DayNumber,
		Players:      publicPlayers,
		ChatLog:      state.ChatMessages, // Chat is public
		CrisisEvent:  state.CrisisEvent,
		VoteState:    state.VoteState,
	}
}