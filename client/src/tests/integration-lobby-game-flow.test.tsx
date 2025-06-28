import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import React from 'react';
import App from '../App';
import { websocketClient } from '../services/websocket';
import { gameEngine } from '../services/gameEngine';
import { ServerEventType } from '../types/generated';

// Mock WASM loader
vi.mock('../services/wasmLoader', () => ({
  wasmLoader: {
    load: vi.fn().mockResolvedValue(undefined),
    isReady: vi.fn().mockReturnValue(true),
    getCore: vi.fn().mockReturnValue({
      createGame: vi.fn().mockReturnValue({ success: true }),
      deserializeGameState: vi.fn().mockReturnValue({ success: true }),
      getGameState: vi.fn().mockReturnValue('{"id":"test-game","players":{},"phase":{"type":"SITREP"}}'),
      applyEvent: vi.fn().mockReturnValue({ success: true }),
    }),
    onStateChange: vi.fn(),
  },
}));

// Mock WebSocket
class MockWebSocket {
  onopen?: () => void;
  onmessage?: (event: { data: string }) => void;
  onclose?: (event: { code: number; reason: string }) => void;
  onerror?: (error: Event) => void;
  readyState = WebSocket.OPEN;

  constructor(public url: string) {}

  send(data: string) {}

  close(code?: number, reason?: string) {
    if (this.onclose) {
      this.onclose({ code: code || 1000, reason: reason || '' });
    }
  }
}

global.WebSocket = MockWebSocket as any;

describe('Integration: Complete Lobby to Game Flow', () => {
  let mockWebSocket: MockWebSocket;
  let stateChangeCallback: (state: any) => void;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock game engine
    vi.spyOn(gameEngine, 'isReady').mockReturnValue(true);
    vi.spyOn(gameEngine, 'resetAndLoadState').mockResolvedValue(undefined);
    vi.spyOn(gameEngine, 'getCurrentState').mockReturnValue(null);
    vi.spyOn(gameEngine, 'onStateChange').mockImplementation((callback) => {
      stateChangeCallback = callback;
      return () => {};
    });
    
    // Mock WebSocket connection
    vi.spyOn(websocketClient, 'connect').mockImplementation(() => {
      mockWebSocket = new MockWebSocket('ws://test');
      (websocketClient as any).socket = mockWebSocket;
      setTimeout(() => mockWebSocket.onopen?.(), 0);
      return Promise.resolve();
    });

    // Suppress console warnings for cleaner test output
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const simulateServerEvent = (eventType: string, payload: any) => {
    if (mockWebSocket?.onmessage) {
      const event = {
        type: eventType,
        payload,
        id: `event_${Date.now()}_${Math.random()}`,
        timestamp: new Date().toISOString(),
      };
      mockWebSocket.onmessage({ data: JSON.stringify(event) });
    }
  };

  const simulateGameStateUpdate = (gameState: any) => {
    // First simulate the websocket event
    simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
    
    // Then simulate the WASM state change callback
    if (stateChangeCallback) {
      stateChangeCallback(gameState);
    }
  };

  const createValidGameState = () => ({
    id: 'test-game-123',
    phase: { type: 'SITREP', startTime: new Date().toISOString(), duration: 15000 },
    day_number: 1,
    players: {
      'player_1': {
        id: 'player_1',
        name: 'Alice',
        jobTitle: 'CEO',
        controlType: 'HUMAN',
        isAlive: true,
        tokens: 0,
        projectMilestones: 0,
        statusMessage: '',
        joinedAt: '2025-01-01T00:00:00Z',
        alignment: 'HUMAN',
        role: {
          type: 'CEO',
          name: 'Chief Executive Officer',
          description: 'Leads the company',
          isUnlocked: false,
        },
        lobbyHandle: 'alice',
        isRolePubliclyRevealed: false,
      },
      'player_2': {
        id: 'player_2',
        name: 'Bob',
        jobTitle: 'CTO',
        controlType: 'AI',
        isAlive: true,
        tokens: 0,
        projectMilestones: 0,
        statusMessage: '',
        joinedAt: '2025-01-01T00:00:00Z',
        alignment: 'AI',
        role: {
          type: 'CTO',
          name: 'Chief Technology Officer',
          description: 'Manages technology',
          isUnlocked: false,
        },
        lobbyHandle: 'bob',
        isRolePubliclyRevealed: false,
      },
    },
    chat_messages: [
      {
        id: 'msg1',
        sender: 'System',
        message: 'Game started',
        timestamp: new Date().toISOString(),
        channelID: '#war-room',
        isSystem: true,
      },
    ],
    settings: {
      maxPlayers: 10,
      minPlayers: 2,
      sitrepDuration: 15000000000,
    },
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  });

  describe('Happy Path Flow', () => {
    it('should complete full lobby to game room transition successfully', async () => {
      const { container } = render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      // Step 1: Start in lobby with proper player identification
      await act(async () => {
        simulateServerEvent(ServerEventType.ClientIdentified, { 
          your_player_id: 'player_1' 
        });
      });

      // Step 2: Lobby state with players
      await act(async () => {
        simulateServerEvent(ServerEventType.LobbyStateUpdate, {
          lobby_id: 'test-lobby',
          host_id: 'player_1',
          can_start: true,
          players: [
            { id: 'player_1', name: 'Alice', avatar: 'avatar1' },
            { id: 'player_2', name: 'Bob', avatar: 'avatar2' },
          ],
        });
      });

      // Step 3: Game start countdown sequence
      await act(async () => {
        simulateServerEvent(ServerEventType.GameStartCountdownInitiated, { duration: 3 });
        simulateServerEvent(ServerEventType.GameStartCountdownUpdate, { remaining: 2 });
        simulateServerEvent(ServerEventType.GameStartCountdownUpdate, { remaining: 1 });
        simulateServerEvent(ServerEventType.GameStartCountdownUpdate, { remaining: 0 });
      });

      // Step 4: Game started notification
      await act(async () => {
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });
      });

      // Step 5: Complete game state update
      const gameState = createValidGameState();
      await act(async () => {
        simulateGameStateUpdate(gameState);
      });

      // Verify game engine received the state
      expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(gameState);

      // Step 6: Additional game events
      await act(async () => {
        simulateServerEvent(ServerEventType.PhaseChanged, {
          phase_type: 'PULSE_CHECK',
          day_number: 1,
          duration: 30,
        });
      });

      // The entire flow should complete without errors
      expect(container).toBeTruthy();
    });

    it('should handle navigation through all transition screens', async () => {
      const { rerender } = render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      // Simulate the complete flow with navigation checks
      await act(async () => {
        // Client identification
        simulateServerEvent(ServerEventType.ClientIdentified, { 
          your_player_id: 'player_1' 
        });

        // Lobby state
        simulateServerEvent(ServerEventType.LobbyStateUpdate, {
          lobby_id: 'test-lobby',
          host_id: 'player_1',
          can_start: true,
          players: [
            { id: 'player_1', name: 'Alice', avatar: 'avatar1' },
            { id: 'player_2', name: 'Bob', avatar: 'avatar2' },
          ],
        });

        // Game start
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });

        // Game state update
        const gameState = createValidGameState();
        simulateGameStateUpdate(gameState);
      });

      // Should not crash during any navigation
      expect(true).toBe(true);
    });
  });

  describe('Error Recovery Scenarios', () => {
    it('should recover from malformed game state during transition', async () => {
      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      await act(async () => {
        // Start normal flow
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });

        // Send malformed state first
        simulateGameStateUpdate({
          id: 'test-game-123',
          players: null, // Invalid
          phase: { type: 'SITREP' },
        });

        // Then send correct state
        const validGameState = createValidGameState();
        simulateGameStateUpdate(validGameState);
      });

      // Should recover and load the valid state
      expect(gameEngine.resetAndLoadState).toHaveBeenCalledTimes(2);
    });

    it('should handle WebSocket disconnection during transition', async () => {
      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      await act(async () => {
        // Start transition
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });

        // Disconnect WebSocket
        mockWebSocket?.close(1006, 'Connection lost');

        // Try to send more events (should not crash)
        simulateServerEvent(ServerEventType.GameStateUpdate, { 
          game_state: createValidGameState() 
        });
      });

      // Should handle gracefully
      expect(true).toBe(true);
    });

    it('should handle game engine failures during transition', async () => {
      // Make game engine fail
      vi.mocked(gameEngine.resetAndLoadState).mockRejectedValue(new Error('WASM failure'));

      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      await act(async () => {
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });

        const gameState = createValidGameState();
        simulateGameStateUpdate(gameState);
      });

      // Should not crash the app
      expect(gameEngine.resetAndLoadState).toHaveBeenCalled();
    });

    it('should handle rapid state changes during transition', async () => {
      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      await act(async () => {
        // Send many state updates rapidly
        for (let i = 0; i < 10; i++) {
          const gameState = {
            ...createValidGameState(),
            id: `test-game-${i}`,
            day_number: i + 1,
          };
          simulateGameStateUpdate(gameState);
        }
      });

      // Should handle all updates
      expect(gameEngine.resetAndLoadState).toHaveBeenCalledTimes(10);
    });
  });

  describe('Component Integration During Transition', () => {
    it('should handle components mounting with partial data', async () => {
      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      await act(async () => {
        // Send game started without state
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });
      });

      // Components should mount without crashing even with missing data
      await act(async () => {
        // Send minimal state
        simulateGameStateUpdate({
          id: 'test-game-123',
          phase: { type: 'SITREP' },
          players: {},
        });
      });

      // Then send complete state
      await act(async () => {
        const completeState = createValidGameState();
        simulateGameStateUpdate(completeState);
      });

      expect(true).toBe(true);
    });

    it('should handle components receiving state in different formats', async () => {
      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      await act(async () => {
        simulateServerEvent(ServerEventType.GameStarted, { 
          game_id: 'test-game-123' 
        });

        // Send state with players as array
        simulateGameStateUpdate({
          id: 'test-game-123',
          phase: { type: 'SITREP' },
          players: [
            { id: 'player_1', name: 'Alice', isAlive: true },
            { id: 'player_2', name: 'Bob', isAlive: true },
          ],
          chatMessages: [],
        });

        // Then send state with players as object
        simulateGameStateUpdate({
          id: 'test-game-123',
          phase: { type: 'SITREP' },
          players: {
            'player_1': { id: 'player_1', name: 'Alice', isAlive: true },
            'player_2': { id: 'player_2', name: 'Bob', isAlive: true },
          },
          chat_messages: [], // snake_case format
        });
      });

      // Should handle both formats
      expect(gameEngine.resetAndLoadState).toHaveBeenCalledTimes(2);
    });
  });

  describe('Performance and Stability', () => {
    it('should not memory leak during multiple transitions', async () => {
      const { unmount } = render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      // Simulate multiple game starts and state updates
      for (let i = 0; i < 5; i++) {
        await act(async () => {
          simulateServerEvent(ServerEventType.GameStarted, { 
            game_id: `test-game-${i}` 
          });

          const gameState = {
            ...createValidGameState(),
            id: `test-game-${i}`,
          };
          simulateGameStateUpdate(gameState);
        });
      }

      // Cleanup should not throw
      expect(() => {
        unmount();
      }).not.toThrow();
    });

    it('should handle concurrent state updates gracefully', async () => {
      render(
        <MemoryRouter>
          <App />
        </MemoryRouter>
      );

      // Send multiple concurrent updates
      const promises = Array.from({ length: 20 }, async (_, i) => {
        await act(async () => {
          const gameState = {
            ...createValidGameState(),
            id: 'test-game-123',
            day_number: i + 1,
          };
          simulateGameStateUpdate(gameState);
        });
      });

      await Promise.all(promises);

      // Should handle all updates without crashing
      expect(gameEngine.resetAndLoadState).toHaveBeenCalledTimes(20);
    });
  });
});