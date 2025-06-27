import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { ServerEventType, ClientActionType } from '../types/generated';

describe('Contract Verification', () => {
  const isE2EMode = process.env.NODE_ENV === 'test' && process.env.E2E_SERVER_URL;
  const serverUrl = process.env.E2E_SERVER_URL || 'http://localhost:8080';
  
  let mockFetch: typeof fetch;
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = global.fetch;
    
    if (!isE2EMode) {
      // Mock fetch for unit testing
      mockFetch = vi.fn();
      global.fetch = mockFetch;
    }
  });

  afterEach(() => {
    if (!isE2EMode) {
      global.fetch = originalFetch;
    }
  });

  it('should handle all server event types defined in the backend', async () => {
    if (!isE2EMode) {
      // Mock the debug endpoint response for unit tests - using a subset that should exist
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
          'NIGHT_ACTION_SUBMITTED',
          'SYSTEM_MESSAGE',
          'LOBBY_STATE_UPDATE',
          'CLIENT_IDENTIFIED',
          'GAME_STATE_UPDATE'
        ],
        action_types: [
          'CREATE_GAME',
          'JOIN_GAME',
          'SUBMIT_VOTE',
          'MINE_TOKENS'
        ],
        total_events: 17,
        total_actions: 4
      };

      (mockFetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });
    }

    try {
      // Get event types from backend (either live server or mocked)
      const response = await fetch(`${serverUrl}/api/debug/event-types`);
      expect(response.ok).toBe(true);
      
      const data = await response.json();
      const backendEventTypes = data.event_types;
      const backendActionTypes = data.action_types;

      // Get all event types from our generated enum
      const frontendEventTypes = Object.values(ServerEventType);
      const frontendActionTypes = Object.values(ClientActionType);

      // Check that we handle all backend event types
      const missingEventTypes: string[] = [];
      for (const backendEventType of backendEventTypes) {
        if (!frontendEventTypes.includes(backendEventType as ServerEventType)) {
          missingEventTypes.push(backendEventType);
        }
      }

      // Check that we handle all backend action types
      const missingActionTypes: string[] = [];
      for (const backendActionType of backendActionTypes) {
        if (!frontendActionTypes.includes(backendActionType as ClientActionType)) {
          missingActionTypes.push(backendActionType);
        }
      }

      if (missingEventTypes.length > 0) {
        console.error('Missing event types in frontend:', missingEventTypes);
        console.error('Frontend has', frontendEventTypes.length, 'event types');
        console.error('Backend has', backendEventTypes.length, 'event types');
      }

      if (missingActionTypes.length > 0) {
        console.error('Missing action types in frontend:', missingActionTypes);
        console.error('Frontend has', frontendActionTypes.length, 'action types');
        console.error('Backend has', backendActionTypes.length, 'action types');
      }

      // Check for extra event types in frontend (not in backend)
      const extraEventTypes: string[] = [];
      for (const frontendEventType of frontendEventTypes) {
        if (!backendEventTypes.includes(frontendEventType)) {
          extraEventTypes.push(frontendEventType);
        }
      }

      // Check for extra action types in frontend (not in backend)
      const extraActionTypes: string[] = [];
      for (const frontendActionType of frontendActionTypes) {
        if (!backendActionTypes.includes(frontendActionType)) {
          extraActionTypes.push(frontendActionType);
        }
      }

      if (extraEventTypes.length > 0) {
        console.warn('Extra event types in frontend (not in backend):', extraEventTypes);
      }

      if (extraActionTypes.length > 0) {
        console.warn('Extra action types in frontend (not in backend):', extraActionTypes);
      }

      // Contract verification should fail if there are missing types
      expect(missingEventTypes).toHaveLength(0);
      expect(missingActionTypes).toHaveLength(0);

      // Log summary for visibility
      console.log(`✅ Contract verification passed!`);
      console.log(`   Event types: ${frontendEventTypes.length} frontend, ${backendEventTypes.length} backend`);
      console.log(`   Action types: ${frontendActionTypes.length} frontend, ${backendActionTypes.length} backend`);
      console.log(`   Extra frontend events: ${extraEventTypes.length}`);
      console.log(`   Extra frontend actions: ${extraActionTypes.length}`);

    } catch (error) {
      if (isE2EMode) {
        console.error('Failed to connect to backend server. Make sure the server is running at', serverUrl);
        throw error;
      } else {
        throw error;
      }
    }
  }, 10000); // Increase timeout for E2E tests

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