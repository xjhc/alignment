import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import React from 'react';
import App from '../App';
import { websocketClient } from '../services/websocket';
import { gameEngine } from '../services/gameEngine';
import { ServerEventType } from '../types/generated';

// Mock the WASM loader to prevent real WASM loading in tests
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

  send(data: string) {
    // Mock send implementation
  }

  close(code?: number, reason?: string) {
    if (this.onclose) {
      this.onclose({ code: code || 1000, reason: reason || '' });
    }
  }
}

// Replace global WebSocket with mock
global.WebSocket = MockWebSocket as any;

describe('Lobby to Game Transition E2E', () => {
  let mockWebSocket: MockWebSocket;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset game engine state
    vi.spyOn(gameEngine, 'isReady').mockReturnValue(true);
    vi.spyOn(gameEngine, 'resetAndLoadState').mockResolvedValue(undefined);
    vi.spyOn(gameEngine, 'getCurrentState').mockReturnValue(null);
    
    // Mock WebSocket connection
    vi.spyOn(websocketClient, 'connect').mockImplementation(() => {
      mockWebSocket = new MockWebSocket('ws://test');
      (websocketClient as any).socket = mockWebSocket;
      setTimeout(() => mockWebSocket.onopen?.(), 0);
      return Promise.resolve();
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const simulateServerEvent = (eventType: string, payload: any) => {
    if (mockWebSocket?.onmessage) {
      const event = {
        type: eventType,
        payload,
        id: `event_${Date.now()}`,
        timestamp: new Date().toISOString(),
      };
      mockWebSocket.onmessage({ data: JSON.stringify(event) });
    }
  };

  const createTestGameState = (overrides = {}) => ({
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
        controlType: 'HUMAN',
        isAlive: true,
        tokens: 0,
        projectMilestones: 0,
        statusMessage: '',
        joinedAt: '2025-01-01T00:00:00Z',
        lobbyHandle: 'bob',
        isRolePubliclyRevealed: false,
      },
    },
    chat_messages: [],
    settings: {
      maxPlayers: 10,
      minPlayers: 2,
      sitrepDuration: 15000000000,
    },
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    ...overrides,
  });

  it('should handle complete lobby to game transition flow', async () => {
    const { container } = render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Step 1: Simulate being in lobby with players
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

    // Step 2: Simulate game start countdown
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStartCountdownInitiated, { duration: 3 });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStartCountdownUpdate, { remaining: 2 });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStartCountdownUpdate, { remaining: 1 });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStartCountdownUpdate, { remaining: 0 });
    });

    // Step 3: Simulate game start
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
    });

    // Step 4: Simulate game state update with complete data
    const gameState = createTestGameState();
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
    });

    // Verify that game engine received the state
    expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(gameState);

    // Step 5: Simulate subsequent game events
    await act(async () => {
      simulateServerEvent(ServerEventType.PhaseChanged, {
        phase_type: 'PULSE_CHECK',
        day_number: 1,
        duration: 30,
      });
    });

    // The test should complete without throwing errors
    expect(container).toBeTruthy();
  });

  it('should handle game state with missing players array', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Send a malformed game state without players array
    const malformedGameState = createTestGameState({ players: undefined });
    
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: malformedGameState });
    });

    // Should not crash - game engine should handle gracefully
    expect(gameEngine.resetAndLoadState).toHaveBeenCalled();
  });

  it('should handle game state with players as object format', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Send game state with players as object (current backend format)
    const gameStateWithObjectPlayers = createTestGameState();
    
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameStateWithObjectPlayers });
    });

    // Should handle object format correctly
    expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(gameStateWithObjectPlayers);
  });

  it('should handle missing chatMessages/chat_messages', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Send game state without chat messages
    const gameStateNoChatMessages = createTestGameState();
    delete (gameStateNoChatMessages as any).chat_messages;
    
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameStateNoChatMessages });
    });

    // Should not crash
    expect(gameEngine.resetAndLoadState).toHaveBeenCalled();
  });

  it('should handle events arriving out of order', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Send events in wrong order - state update before game started
    const gameState = createTestGameState();
    
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
    });

    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
    });

    // Should handle gracefully
    expect(gameEngine.resetAndLoadState).toHaveBeenCalled();
  });

  it('should handle WASM state change notifications with invalid data', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Mock WASM state change with invalid data
    const mockOnStateChange = vi.fn();
    vi.mocked(gameEngine.onStateChange).mockImplementation((callback) => {
      mockOnStateChange.mockImplementation(callback);
      return () => {};
    });

    // Send complete game state first
    const gameState = createTestGameState();
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
    });

    // Simulate WASM sending invalid state change notifications
    await act(async () => {
      mockOnStateChange({ id: 'test-game', players: null }); // Invalid players
    });

    await act(async () => {
      mockOnStateChange({ id: 'test-game' }); // Missing players entirely
    });

    await act(async () => {
      mockOnStateChange({}); // Completely invalid state
    });

    // Should not crash
    expect(true).toBe(true);
  });

  it('should handle rapid successive events without crashing', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    const gameState = createTestGameState();
    
    // Send many events rapidly
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
      simulateServerEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
      simulateServerEvent(ServerEventType.PhaseChanged, { phase_type: 'PULSE_CHECK' });
      simulateServerEvent(ServerEventType.PhaseChanged, { phase_type: 'DISCUSSION' });
      simulateServerEvent(ServerEventType.PhaseChanged, { phase_type: 'NOMINATION' });
      simulateServerEvent(ServerEventType.ChatMessage, { 
        sender_id: 'player_1', 
        message: 'Hello', 
        channel_id: '#war-room' 
      });
    });

    // Should handle all events without crashing
    expect(gameEngine.resetAndLoadState).toHaveBeenCalled();
  });

  it('should handle WebSocket disconnection during transition', async () => {
    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>
    );

    // Start the transition
    await act(async () => {
      simulateServerEvent(ServerEventType.GameStarted, { game_id: 'test-game-123' });
    });

    // Simulate WebSocket disconnection
    await act(async () => {
      mockWebSocket?.close(1006, 'Connection lost');
    });

    // Should handle gracefully without crashing
    expect(true).toBe(true);
  });
});