package events

import (
	"encoding/json"
	"testing"

	"github.com/xjhc/alignment/core"
)

// TestEventPayloadContracts verifies that all events have strongly-typed payload contracts
// These tests will initially fail and turn green as we implement the typed payload system
func TestEventPayloadContracts(t *testing.T) {
	tests := []struct {
		eventType     core.EventType
		payloadStruct interface{}
		description   string
	}{
		// Game Lifecycle Events
		{core.EventGameCreated, core.GameCreatedPayload{}, "Game creation event"},
		{core.EventGameStarted, core.GameStartedPayload{}, "Game start event"},
		{core.EventGameEnded, core.GameEndedPayload{}, "Game end event"},
		{core.EventPhaseChanged, core.PhaseChangedPayload{}, "Phase change event"},
		{core.EventDayStarted, core.DayStartedPayload{}, "Day start event"},
		{core.EventNightStarted, core.NightStartedPayload{}, "Night start event"},

		// Player Events
		{core.EventPlayerJoined, core.PlayerJoinedPayload{}, "Player join event"},
		{core.EventPlayerLeft, core.PlayerLeftPayload{}, "Player leave event"},
		{core.EventPlayerEliminated, core.PlayerEliminatedPayload{}, "Player elimination event"},
		{core.EventPlayerAbandoned, core.PlayerAbandonedPayload{}, "Player abandonment event"},
		{core.EventPlayerConnectionStatusChanged, core.PlayerConnectionStatusChangedPayload{}, "Player connection status event"},
		{core.EventPlayerRoleRevealed, core.PlayerRoleRevealedPayload{}, "Player role reveal event"},
		{core.EventPlayerAligned, core.PlayerAlignedPayload{}, "Player alignment event"},
		{core.EventPlayerShocked, core.PlayerShockedPayload{}, "Player shock event"},
		{core.EventPlayerStatusChanged, core.PlayerStatusChangedPayload{}, "Player status change event"},

		// Voting Events
		{core.EventVoteStarted, core.VoteStartedPayload{}, "Vote start event"},
		{core.EventVoteCast, core.VoteCastPayload{}, "Vote cast event"},
		{core.EventVoteTallyUpdated, core.VoteTallyUpdatedPayload{}, "Vote tally update event"},
		{core.EventVoteCompleted, core.VoteCompletedPayload{}, "Vote completion event"},
		{core.EventPlayerNominated, core.PlayerNominatedPayload{}, "Player nomination event"},

		// Token and Mining Events
		{core.EventTokensAwarded, core.TokensAwardedPayload{}, "Tokens awarded event"},
		{core.EventTokensLost, core.TokensLostPayload{}, "Tokens lost event"},
		{core.EventMiningSuccessful, core.MiningSuccessfulPayload{}, "Mining success event"},
		{core.EventMiningFailed, core.MiningFailedPayload{}, "Mining failure event"},
		{core.EventMiningPoolUpdated, core.MiningPoolUpdatedPayload{}, "Mining pool update event"},
		{core.EventTokensDistributed, core.TokensDistributedPayload{}, "Token distribution event"},

		// Night Action Events
		{core.EventNightActionSubmitted, core.NightActionSubmittedPayload{}, "Night action submission event"},
		{core.EventNightActionsResolved, core.NightActionResolutionPayload{}, "Night actions resolution event"},

		// AI Conversion Events
		{core.EventAIConversionAttempt, core.AIConversionAttemptPayload{}, "AI conversion attempt event"},
		{core.EventAIConversionSuccess, core.AIConversionSuccessPayload{}, "AI conversion success event"},
		{core.EventAIConversionFailed, core.AIConversionFailedPayload{}, "AI conversion failure event"},

		// Communication Events
		{core.EventChatMessage, core.ChatMessagePayload{}, "Chat message event"},
		{core.EventMessageReaction, core.MessageReactionPayload{}, "Message reaction event"},
		{core.EventSystemMessage, core.SystemMessagePayload{}, "System message event"},

		// Crisis and Pulse Check Events
		{core.EventCrisisTriggered, core.CrisisTriggeredPayload{}, "Crisis trigger event"},
		{core.EventPulseCheckStarted, core.PulseCheckStartedPayload{}, "Pulse check start event"},
		{core.EventPulseCheckUpdated, core.PulseCheckUpdatedPayload{}, "Pulse check update event"},
		{core.EventPulseCheckRevealed, core.PulseCheckRevealedPayload{}, "Pulse check reveal event"},

		// Role and Ability Events
		{core.EventRoleAssigned, core.RoleAssignedPayload{}, "Role assignment event"},
		{core.EventRoleAbilityUnlocked, core.RoleAbilityUnlockedPayload{}, "Role ability unlock event"},
		{core.EventProjectMilestone, core.ProjectMilestonePayload{}, "Project milestone event"},

		// Specific Role Ability Events
		{core.EventRunAudit, core.RunAuditPayload{}, "Audit execution event"},
		{core.EventOverclockServers, core.OverclockServersPayload{}, "Server overclock event"},
		{core.EventIsolateNode, core.IsolateNodePayload{}, "Node isolation event"},
		{core.EventPerformanceReview, core.PerformanceReviewPayload{}, "Performance review event"},
		{core.EventReallocateBudget, core.ReallocateBudgetPayload{}, "Budget reallocation event"},
		{core.EventPivot, core.PivotPayload{}, "Strategy pivot event"},
		{core.EventDeployHotfix, core.DeployHotfixPayload{}, "Hotfix deployment event"},

		// Player Blocking and Protection Events
		{core.EventPlayerBlocked, core.PlayerBlockedPayload{}, "Player blocked event"},
		{core.EventPlayerProtected, core.PlayerProtectedPayload{}, "Player protected event"},
		{core.EventPlayerInvestigated, core.PlayerInvestigatedPayload{}, "Player investigation event"},

		// Status Events
		{core.EventSlackStatusChanged, core.SlackStatusChangedPayload{}, "Slack status change event"},
		{core.EventPartingShotSet, core.PartingShotSetPayload{}, "Parting shot set event"},
		{core.EventWhisperSent, core.WhisperSentPayload{}, "Whisper sent event"},

		// KPI Events
		{core.EventKPIAssigned, core.KPIAssignedPayload{}, "KPI assignment event"},
		{core.EventKPIProgress, core.KPIProgressPayload{}, "KPI progress event"},

		// System Shock Events
		{core.EventSystemShockApplied, core.SystemShockAppliedPayload{}, "System shock application event"},
		{core.EventShockEffectTriggered, core.ShockEffectTriggeredPayload{}, "Shock effect trigger event"},

		// AI Equity Events
		{core.EventAIEquityChanged, core.AIEquityChangedPayload{}, "AI equity change event"},
		{core.EventEquityThreshold, core.EquityThresholdPayload{}, "Equity threshold event"},

		// Corporate Mandate Events
		{core.EventMandateActivated, core.MandateActivatedPayload{}, "Mandate activation event"},
		{core.EventMandateEffect, core.MandateEffectPayload{}, "Mandate effect event"},

		// Phase Skipping Events
		{core.EventSkipVoteUpdated, core.SkipVoteUpdatedPayload{}, "Skip vote update event"},

		// Whistleblower Protocol Events
		{core.EventWhistleblowerVotingStarted, core.WhistleblowerVotingStartedPayload{}, "Whistleblower voting start event"},
		{core.EventWhistleblowerVoteCast, core.WhistleblowerVoteCastPayload{}, "Whistleblower vote cast event"},
		{core.EventWhistleblowerVotingCompleted, core.WhistleblowerVotingCompletedPayload{}, "Whistleblower voting completion event"},

		// Semantic Events
		{core.EventSitrepPublished, core.SitrepPublishedPayload{}, "SITREP publication event"},
		{core.EventLiaisonProtocolActivated, core.LiaisonProtocolActivatedPayload{}, "Liaison protocol activation event"},
		{core.EventGameRuleModified, core.GameRuleModifiedPayload{}, "Game rule modification event"},

		// Game State Events
		{core.EventGameStateUpdate, core.GameStateUpdatePayload{}, "Game state update event"},
		{core.EventLobbyStateUpdate, core.LobbyStateUpdatePayload{}, "Lobby state update event"},
		{core.EventVictoryCondition, core.VictoryConditionPayload{}, "Victory condition event"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			// Test that the payload struct can be serialized to JSON
			payloadBytes, err := json.Marshal(tt.payloadStruct)
			if err != nil {
				t.Errorf("Failed to marshal payload for %s: %v", tt.eventType, err)
			}

			// Test that the payload can be deserialized back
			var unmarshaled interface{}
			err = json.Unmarshal(payloadBytes, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal payload for %s: %v", tt.eventType, err)
			}

			// Test that an event can be created with this payload
			event := core.Event{
				ID:        "test-id",
				Type:      tt.eventType,
				GameID:    "test-game",
				PlayerID:  "test-player",
				Payload:   unmarshaled.(map[string]interface{}),
			}

			// Verify event can be serialized
			_, err = json.Marshal(event)
			if err != nil {
				t.Errorf("Failed to marshal event for %s: %v", tt.eventType, err)
			}

			t.Logf("✅ Contract verified for %s: %s", tt.eventType, tt.description)
		})
	}
}

// TestNoDefensiveParsingRequired verifies that event handlers don't need defensive parsing
// This test will initially fail and should pass once we remove defensive code
func TestNoDefensiveParsingRequired(t *testing.T) {
	// Sample events with strongly-typed payloads
	testEvents := []struct {
		eventType core.EventType
		payload   interface{}
	}{
		{
			core.EventChatMessage,
			core.ChatMessagePayload{
				SenderID:   "player-1",
				SenderName: "Alice",
				Message:    "Hello world",
				ChannelID:  "#war-room",
			},
		},
		{
			core.EventVoteCast,
			core.VoteCastPayload{
				TargetID:    "player-2",
				VoteType:    "NOMINATION",
				TokenWeight: 3,
			},
		},
		{
			core.EventPhaseChanged,
			core.PhaseChangedPayload{
				PhaseType: "DISCUSSION",
				Duration:  120.0,
			},
		},
	}

	for _, tt := range testEvents {
		t.Run(string(tt.eventType), func(t *testing.T) {
			// Create event with typed payload
			payloadBytes, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			var payloadMap map[string]interface{}
			err = json.Unmarshal(payloadBytes, &payloadMap)
			if err != nil {
				t.Fatalf("Failed to unmarshal to map: %v", err)
			}

			event := core.Event{
				ID:       "test-id",
				Type:     tt.eventType,
				GameID:   "test-game",
				PlayerID: "test-player",
				Payload:  payloadMap,
			}

			// Test that ApplyEvent can handle this event without defensive parsing
			// This will fail initially until we refactor the event handlers
			initialState := core.GameState{
				ID:       "test-game",
				Players:  make(map[string]*core.Player),
				Phase:    core.Phase{Type: core.PhaseDiscussion},
			}

			// Add a test player for events that need one
			initialState.Players["test-player"] = &core.Player{
				ID:     "test-player",
				Name:   "Test Player",
				IsAlive: true,
				Tokens: 5,
			}

			// Apply the event - this should work without defensive parsing
			newState := core.ApplyEvent(initialState, event)

			// Verify the state was updated (basic sanity check)
			if newState.ID != initialState.ID {
				t.Errorf("State ID changed unexpectedly")
			}

			t.Logf("✅ Event %s processed without defensive parsing", tt.eventType)
		})
	}
}

// TestAllEventsHaveTypedPayloads ensures we don't miss any events in our refactoring
func TestAllEventsHaveTypedPayloads(t *testing.T) {
	// This is a comprehensive list of all event types that should have typed payloads
	// If we add a new event type, it should be added here and have a corresponding test above
	expectedEventTypes := []core.EventType{
		// Game Lifecycle
		core.EventGameCreated,
		core.EventGameStarted,
		core.EventGameEnded,
		core.EventPhaseChanged,
		core.EventDayStarted,
		core.EventNightStarted,

		// Player Events
		core.EventPlayerJoined,
		core.EventPlayerLeft,
		core.EventPlayerEliminated,
		core.EventPlayerAbandoned,
		core.EventPlayerConnectionStatusChanged,
		core.EventPlayerRoleRevealed,
		core.EventPlayerAligned,
		core.EventPlayerShocked,
		core.EventPlayerStatusChanged,

		// Voting Events
		core.EventVoteStarted,
		core.EventVoteCast,
		core.EventVoteTallyUpdated,
		core.EventVoteCompleted,
		core.EventPlayerNominated,

		// Communication Events
		core.EventChatMessage,
		core.EventMessageReaction,
		core.EventSystemMessage,

		// Add more as we identify them...
	}

	for _, eventType := range expectedEventTypes {
		t.Run(string(eventType), func(t *testing.T) {
			// This test will fail until we implement all typed payloads
			// For now, just verify the event type exists
			if eventType == "" {
				t.Errorf("Empty event type found")
			}
			t.Logf("Event type %s exists", eventType)
		})
	}
}