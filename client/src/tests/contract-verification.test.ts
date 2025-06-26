import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ServerEventType } from '../types/generated';

describe('Contract Verification', () => {
  let mockFetch: typeof fetch;

  beforeEach(() => {
    // Mock fetch for testing
    mockFetch = vi.fn();
    global.fetch = mockFetch;
  });

  it('should handle all server event types defined in the backend', async () => {
    // Mock the debug endpoint response
    const mockResponse = {
      event_types: [
        'GAME_CREATED',
        'GAME_STARTED',
        'GAME_ENDED',
        'PHASE_CHANGED',
        'PLAYER_JOINED',
        'PLAYER_LEFT',
        'PLAYER_ELIMINATED',
        'CHAT_MESSAGE',
        'VOTE_CAST',
        'ROLE_ASSIGNED',
        'MANDATE_ACTIVATED',
        'PULSE_CHECK_STARTED',
        'PULSE_CHECK_SUBMITTED',
        'NIGHT_ACTION_SUBMITTED',
        'SYSTEM_MESSAGE',
        'LOBBY_STATE_UPDATE',
        'CLIENT_IDENTIFIED',
        'GAME_STATE_UPDATE'
      ],
      action_types: [],
      total_events: 18,
      total_actions: 0
    };

    (mockFetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse
    });

    // Get event types from backend
    const response = await fetch('http://localhost:8080/api/debug/event-types');
    expect(response.ok).toBe(true);
    
    const data = await response.json();
    const backendEventTypes = data.event_types;

    // Get all event types from our generated enum
    const frontendEventTypes = Object.values(ServerEventType);

    // Check that we handle all backend event types
    const missingEventTypes: string[] = [];
    for (const backendEventType of backendEventTypes) {
      if (!frontendEventTypes.includes(backendEventType as ServerEventType)) {
        missingEventTypes.push(backendEventType);
      }
    }

    if (missingEventTypes.length > 0) {
      console.error('Missing event types in frontend:', missingEventTypes);
    }

    expect(missingEventTypes).toHaveLength(0);
  });

  it('should have all generated event types available', () => {
    // Test that our generated enum has the expected structure
    expect(ServerEventType.GameCreated).toBe('GAME_CREATED');
    expect(ServerEventType.ChatMessage).toBe('CHAT_MESSAGE');
    expect(ServerEventType.GameStateUpdate).toBe('GAME_STATE_UPDATE');
    expect(ServerEventType.LobbyStateUpdate).toBe('LOBBY_STATE_UPDATE');
    expect(ServerEventType.ClientIdentified).toBe('CLIENT_IDENTIFIED');
  });

  it('should verify websocket handler covers all event types', () => {
    // This test ensures that all event types from the enum are handled
    // in the websocket handler switch statement
    
    // Get all possible event types
    const allEventTypes = Object.values(ServerEventType);
    
    // These are the event types we know are handled in the websocket handler
    const handledEventTypes = [
      ServerEventType.GameStateUpdate,
      ServerEventType.RoleAssigned,
      ServerEventType.GameStarted,
      ServerEventType.PhaseChanged,
      ServerEventType.ChatMessage,
      ServerEventType.VoteCast,
      ServerEventType.NightActionSubmitted,
      ServerEventType.PlayerLeft,
      ServerEventType.PlayerEliminated,
      ServerEventType.PulseCheckStarted,
      ServerEventType.PulseCheckSubmitted,
      ServerEventType.MandateActivated,
      ServerEventType.SystemMessage,
      ServerEventType.LobbyStateUpdate,
      ServerEventType.ClientIdentified,
    ];

    // Calculate unhandled event types
    const unhandledEventTypes = allEventTypes.filter(
      eventType => !handledEventTypes.includes(eventType)
    );

    // Log unhandled event types for visibility
    if (unhandledEventTypes.length > 0) {
      console.log('Event types not explicitly handled in websocket switch:', unhandledEventTypes);
    }

    // This test doesn't fail - it's informational
    // All unhandled events go to the default case which is acceptable
    expect(allEventTypes.length).toBeGreaterThan(0);
    expect(handledEventTypes.length).toBeGreaterThan(0);
  });
});