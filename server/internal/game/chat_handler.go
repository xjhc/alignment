package game

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// ChatHandler handles chat-related actions
type ChatHandler struct {
	spectatorBroadcaster SpectatorBroadcaster
}

// SpectatorBroadcaster interface for broadcasting to spectators
type SpectatorBroadcaster interface {
	BroadcastToSpectators(events []core.Event)
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(spectatorBroadcaster SpectatorBroadcaster) *ChatHandler {
	return &ChatHandler{
		spectatorBroadcaster: spectatorBroadcaster,
	}
}

// Handle processes chat actions
func (h *ChatHandler) Handle(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	switch action.Type {
	case core.ActionSendMessage:
		return h.handleChatMessage(state, action, acker)
	case core.ActionPostSpectatorMessage:
		return h.handleSpectatorMessage(state, action, acker)
	case core.ActionReactToMessage:
		return h.handleReaction(state, action, acker)
	default:
		return nil, fmt.Errorf("unsupported chat action type: %s", action.Type)
	}
}

// handleChatMessage processes regular chat messages
func (h *ChatHandler) handleChatMessage(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate that the player exists and is alive
	player, exists := state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot send messages")
	}

	// Check corporate mandate communication restrictions
	if state.CorporateMandate != nil && state.CorporateMandate.IsActive {
		if noDirectMessages, exists := state.CorporateMandate.Effects["no_direct_messages"]; exists {
			if restricted, ok := noDirectMessages.(bool); ok && restricted {
				// Check if this is a private message (not in public channels)
				if channelID, exists := action.Payload["channel_id"].(string); exists {
					if channelID != "#war-room" && channelID != "#aligned" {
						return nil, fmt.Errorf("private communications suspended due to Total Transparency mandate")
					}
				}
			}
		}
	}

	// Extract messages from payload (PlayerActor already processed and validated this)
	messagesInterface, ok := action.Payload["messages"].([]map[string]interface{})
	if !ok {
		// Try legacy format as fallback
		if legacyMessages, legacyOk := action.Payload["messages"].([]string); legacyOk {
			return h.handleLegacyMessageFormat(state, action, legacyMessages, acker)
		}
		return nil, fmt.Errorf("invalid or missing messages array in payload")
	}

	if len(messagesInterface) == 0 {
		return nil, fmt.Errorf("messages array cannot be empty")
	}

	// Process each structured message and collect events
	var events []core.Event
	for i, msgData := range messagesInterface {
		message, ok := msgData["message"].(string)
		if !ok || message == "" {
			return nil, fmt.Errorf("message at index %d is missing or empty", i)
		}

		// Check for /help command
		if message == "/help" {
			helpEvents, err := h.handleHelpCommand(state, action.PlayerID, acker)
			if err != nil {
				return nil, fmt.Errorf("failed to handle help command: %v", err)
			}
			events = append(events, helpEvents...)
			continue
		}

		// Check for /status command
		if len(message) > 8 && message[:8] == "/status " {
			statusMessage := message[8:] // Remove "/status " prefix
			statusEvents, err := h.handleStatusCommand(state, action.PlayerID, statusMessage, acker)
			if err != nil {
				return nil, fmt.Errorf("failed to handle status command: %v", err)
			}
			events = append(events, statusEvents...)
			continue
		}

		// Create a copy of the payload for this specific message
		messagePayload := make(map[string]interface{})
		for k, v := range action.Payload {
			messagePayload[k] = v
		}

		// Add the specific client_message_id for this message if available
		if clientMessageId, exists := msgData["client_message_id"].(string); exists && clientMessageId != "" {
			messagePayload["client_message_id"] = clientMessageId
		}

		// Create individual event for this message
		event, err := h.createChatMessageEvent(state, action.PlayerID, message, messagePayload, acker)
		if err != nil {
			return nil, fmt.Errorf("failed to create event for message %d: %v", i, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// handleSpectatorMessage handles spectator chat messages
func (h *ChatHandler) handleSpectatorMessage(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate that the sender is a spectator
	spectator, exists := state.Spectators[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("spectator %s not found", action.PlayerID)
	}

	// Extract message content
	content, ok := action.Payload["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content field is required and must be a string")
	}

	// Validate message length
	if len(content) == 0 {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	if len(content) > 500 {
		return nil, fmt.Errorf("message too long: %d characters (max 500)", len(content))
	}

	// Create chat message
	chatMessage := core.ChatMessage{
		ID:         fmt.Sprintf("spectator-msg-%d", time.Now().UnixNano()),
		PlayerID:   spectator.ID,
		PlayerName: spectator.Name,
		Message:    content,
		Timestamp:  time.Now(),
		IsSystem:   false,
		ChannelID:  "#spectators",
	}

	// Create event
	event := core.Event{
		ID:        fmt.Sprintf("spectator-chat-%d", time.Now().UnixNano()),
		Type:      core.EventSpectatorChatMessage,
		GameID:    acker.GetGameID(),
		PlayerID:  "", // Broadcast to spectators only
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": chatMessage,
		},
	}

	// Broadcast only to spectators
	h.spectatorBroadcaster.BroadcastToSpectators([]core.Event{event})

	return []core.Event{event}, nil
}

// handleReaction processes message reactions
func (h *ChatHandler) handleReaction(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate that the player exists and is alive
	player, exists := state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot react to messages")
	}

	// Extract reaction data from payload
	messageID, ok := action.Payload["message_id"].(string)
	if !ok || messageID == "" {
		return nil, fmt.Errorf("invalid or missing message_id in payload")
	}

	emoji, ok := action.Payload["emoji"].(string)
	if !ok || emoji == "" {
		return nil, fmt.Errorf("invalid or missing emoji in payload")
	}

	// Validate emoji is from allowed set (as per design doc)
	allowedEmojis := map[string]bool{
		"👍": true, "👎": true, "🤔": true, "👀": true, "😂": true, "🔥": true,
		"thinking_face": true, "thumbs_up": true, "thumbs_down": true, "eyes": true, "joy": true, "fire": true,
	}
	if !allowedEmojis[emoji] {
		return nil, fmt.Errorf("emoji %s not allowed", emoji)
	}

	// Check channel permissions (reactions follow same rules as messages)
	channelID, ok := action.Payload["channel_id"].(string)
	if !ok || channelID == "" {
		channelID = "#war-room"
	}

	if !core.CanPlayerSendMessageInChannel(*player, channelID, state.Phase.Type, time.Now()) {
		return nil, fmt.Errorf("cannot react in channel %s during %s phase", channelID, state.Phase.Type)
	}

	// Create reaction event
	payload := map[string]interface{}{
		"player_id":   action.PlayerID,
		"player_name": player.Name,
		"message_id":  messageID,
		"emoji":       emoji,
		"channel_id":  channelID,
		"phase":       string(state.Phase.Type),
		"day_number":  state.DayNumber,
	}

	// For #aligned channel, restrict visibility to AI faction members
	if channelID == "#aligned" {
		payload["restricted_to_alignment"] = "ALIGNED"
	}

	event := core.Event{
		ID:        fmt.Sprintf("reaction_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventMessageReaction,
		GameID:    acker.GetGameID(),
		PlayerID:  "", // Public event by default
		Timestamp: time.Now(),
		Payload:   payload,
	}

	return []core.Event{event}, nil
}

// handleLegacyMessageFormat handles the old string array format for backward compatibility
func (h *ChatHandler) handleLegacyMessageFormat(state *core.GameState, action core.Action, messages []string, acker ActionAcker) ([]core.Event, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages array cannot be empty")
	}

	// Extract client message IDs (optional)
	clientMessageIds, _ := action.Payload["client_message_ids"].([]string)

	// Process each message and collect events
	var events []core.Event
	for i, message := range messages {
		if message == "" {
			return nil, fmt.Errorf("message at index %d is empty", i)
		}

		// Check for /help command
		if message == "/help" {
			helpEvents, err := h.handleHelpCommand(state, action.PlayerID, acker)
			if err != nil {
				return nil, fmt.Errorf("failed to handle help command: %v", err)
			}
			events = append(events, helpEvents...)
			continue
		}

		// Create a copy of the payload for this specific message
		messagePayload := make(map[string]interface{})
		for k, v := range action.Payload {
			messagePayload[k] = v
		}

		// Add the specific client_message_id for this message if available
		if i < len(clientMessageIds) && clientMessageIds[i] != "" {
			messagePayload["client_message_id"] = clientMessageIds[i]
		}

		// Create individual event for this message
		event, err := h.createChatMessageEvent(state, action.PlayerID, message, messagePayload, acker)
		if err != nil {
			return nil, fmt.Errorf("failed to create event for message %d: %v", i, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// handleHelpCommand handles the /help command
func (h *ChatHandler) handleHelpCommand(state *core.GameState, playerID string, acker ActionAcker) ([]core.Event, error) {
	player := state.Players[playerID]
	if player == nil {
		return nil, fmt.Errorf("player %s not found", playerID)
	}

	// Generate help message based on player's role and current phase
	helpMessage := h.generateHelpMessage(state, player)

	// Create private help response event
	event := core.Event{
		ID:        fmt.Sprintf("help_response_%s_%d", playerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    acker.GetGameID(),
		PlayerID:  playerID, // Private message to the requesting player
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message":      helpMessage,
			"sender_name":  "System",
			"sender_id":    "system",
			"channel":      "#help",
			"message_type": "help_response",
		},
	}

	return []core.Event{event}, nil
}

// generateHelpMessage generates a help message for a player
func (h *ChatHandler) generateHelpMessage(state *core.GameState, player *core.Player) string {
	helpText := fmt.Sprintf("**Help for %s**\n\n", player.Name)
	helpText += "**Available Commands:**\n"
	helpText += "• `/help` - Show this help message\n"
	helpText += "• `/status <message>` - Set your status message\n"
	helpText += "• Type messages to communicate with other players\n"
	helpText += "• Use reactions to respond to messages\n\n"
	
	helpText += "**Current Phase:** " + string(state.Phase.Type) + "\n"
	helpText += "**Day Number:** " + fmt.Sprintf("%d", state.DayNumber) + "\n\n"
	
	if player.Role != nil {
		helpText += "**Your Role:** " + h.getRoleName(player.Role.Type) + "\n"
		helpText += "**Role Description:** " + h.getRoleDescription(player.Role.Type) + "\n"
	}
	
	return helpText
}

// handleStatusCommand handles the /status command
func (h *ChatHandler) handleStatusCommand(state *core.GameState, playerID string, statusMessage string, acker ActionAcker) ([]core.Event, error) {
	player := state.Players[playerID]
	if player == nil {
		return nil, fmt.Errorf("player %s not found", playerID)
	}

	// Validate status message length
	if len(statusMessage) > 100 {
		return nil, fmt.Errorf("status message too long (max 100 characters)")
	}

	// Generate status changed event
	event := core.Event{
		ID:        fmt.Sprintf("status_changed_%s_%d", playerID, time.Now().UnixNano()),
		Type:      core.EventSlackStatusChanged,
		GameID:    acker.GetGameID(),
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"status": statusMessage,
		},
	}

	return []core.Event{event}, nil
}

// createChatMessageEvent creates a chat message event
func (h *ChatHandler) createChatMessageEvent(state *core.GameState, playerID, message string, actionPayload map[string]interface{}, acker ActionAcker) (core.Event, error) {
	player := state.Players[playerID]
	if player == nil {
		return core.Event{}, fmt.Errorf("player %s not found", playerID)
	}

	// Extract client message ID if available
	clientMessageID, _ := actionPayload["client_message_id"].(string)

	// Determine channel (default to #war-room)
	channel := "#war-room"
	if channelID, ok := actionPayload["channel_id"].(string); ok && channelID != "" {
		channel = channelID
	}

	// Validate channel permissions
	if !core.CanPlayerSendMessageInChannel(*player, channel, state.Phase.Type, time.Now()) {
		return core.Event{}, fmt.Errorf("cannot send message in channel %s during %s phase", channel, state.Phase.Type)
	}

	// Create event payload as a flat map, not a nested struct
	eventPayload := map[string]interface{}{
		"id":          fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		"sender_id":   playerID,
		"sender_name": player.Name,
		"message":     message,
		"timestamp":   time.Now().Format(time.RFC3339Nano),
		"isSystem":    false,
		"channel_id":  channel,
		"phase":       string(state.Phase.Type),
		"day_number":  state.DayNumber,
	}

	// Include client message ID if provided for confirmation
	if clientMessageID != "" {
		eventPayload["client_message_id"] = clientMessageID
	}

	// For #aligned channel, restrict visibility to AI faction members
	if channel == "#aligned" {
		eventPayload["restricted_to_alignment"] = "ALIGNED"
	}

	event := core.Event{
		ID:        fmt.Sprintf("chat_%s_%d", playerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    acker.GetGameID(),
		PlayerID:  "", // Public event by default
		Timestamp: time.Now(),
		Payload:   eventPayload,
	}

	return event, nil
}

// getRoleName returns the display name for a role type
func (h *ChatHandler) getRoleName(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "Chief Information Security Officer"
	case core.RoleCTO:
		return "Chief Technology Officer"
	case core.RoleCOO:
		return "Chief Operating Officer"
	case core.RoleCFO:
		return "Chief Financial Officer"
	case core.RoleCEO:
		return "Chief Executive Officer"
	case core.RoleEthics:
		return "VP, Ethics & Alignment"
	case core.RolePlatforms:
		return "VP, Platforms"
	case core.RoleIntern:
		return "Intern"
	default:
		return "Unknown Role"
	}
}

// getRoleDescription returns the description for a role type
func (h *ChatHandler) getRoleDescription(roleType core.RoleType) string {
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
	case core.RoleIntern:
		return "Shadows experienced employees to learn their abilities. Use BOOTCAMP to gain Bootcamp Points, then SHADOW other players to copy their role abilities."
	default:
		return "Unknown role"
	}
}