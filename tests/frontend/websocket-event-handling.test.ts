import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { websocketClient } from '../../client/src/services/websocket';
import { gameEngine } from '../../client/src/services/gameEngine';
import { ServerEventType } from '../../client/src/types/generated';

// Mock the game engine
vi.mock('../services/gameEngine', () => ({
  gameEngine: {
    isReady: vi.fn(),
    resetAndLoadState: vi.fn(),
    applyEvent: vi.fn(),
    getCurrentState: vi.fn(),
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

describe('WebSocket Event Handling Robustness', () => {
  let mockWebSocket: MockWebSocket;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(gameEngine.isReady).mockReturnValue(true);
    vi.mocked(gameEngine.resetAndLoadState).mockResolvedValue(undefined);
    vi.mocked(gameEngine.applyEvent).mockResolvedValue(undefined);
    vi.mocked(gameEngine.getCurrentState).mockReturnValue(null);
  });

  afterEach(() => {
    if (mockWebSocket) {
      mockWebSocket.close();
    }
    vi.restoreAllMocks();
  });

  const simulateConnection = async () => {
    mockWebSocket = new MockWebSocket('ws://test');
    (websocketClient as any).socket = mockWebSocket;
    await new Promise(resolve => {
      setTimeout(() => {
        mockWebSocket.onopen?.();
        resolve(undefined);
      }, 0);
    });
  };

  const sendEvent = (eventType: string, payload: any = {}) => {
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

  const createMalformedGameState = (type: 'missingPlayers' | 'playersAsNull' | 'playersAsString' | 'invalidStructure') => {
    const base = {
      id: 'test-game',
      phase: { type: 'SITREP' },
      day_number: 1,
      chat_messages: [],
    };

    switch (type) {
      case 'missingPlayers':
        return base; // No players property
      
      case 'playersAsNull':
        return { ...base, players: null };
      
      case 'playersAsString':
        return { ...base, players: 'invalid' };
      
      case 'invalidStructure':
        return { id: 'test-game' }; // Minimal structure
      
      default:
        return base;
    }
  };

  describe('Game State Loading Edge Cases', () => {
    it('should handle GAME_STATE_UPDATE with missing players property', async () => {
      await simulateConnection();
      
      const malformedState = createMalformedGameState('missingPlayers');
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { game_state: malformedState });
      }).not.toThrow();

      expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(malformedState);
    });

    it('should handle GAME_STATE_UPDATE with null players', async () => {
      await simulateConnection();
      
      const malformedState = createMalformedGameState('playersAsNull');
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { game_state: malformedState });
      }).not.toThrow();

      expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(malformedState);
    });

    it('should handle GAME_STATE_UPDATE with players as string', async () => {
      await simulateConnection();
      
      const malformedState = createMalformedGameState('playersAsString');
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { game_state: malformedState });
      }).not.toThrow();

      expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(malformedState);
    });

    it('should handle completely invalid game state structure', async () => {
      await simulateConnection();
      
      const malformedState = createMalformedGameState('invalidStructure');
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { game_state: malformedState });
      }).not.toThrow();

      expect(gameEngine.resetAndLoadState).toHaveBeenCalledWith(malformedState);
    });

    it('should handle GAME_STATE_UPDATE with no game_state payload', async () => {
      await simulateConnection();
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, {}); // Empty payload
      }).not.toThrow();

      expect(gameEngine.resetAndLoadState).not.toHaveBeenCalled();
    });

    it('should handle GAME_STATE_UPDATE when game engine is not ready', async () => {
      vi.mocked(gameEngine.isReady).mockReturnValue(false);
      await simulateConnection();
      
      const gameState = { id: 'test', players: {} };
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
      }).not.toThrow();

      expect(gameEngine.resetAndLoadState).not.toHaveBeenCalled();
    });
  });

  describe('Event Processing Robustness', () => {
    it('should handle GAME_STARTED events without applying to game engine', async () => {
      await simulateConnection();
      
      expect(() => {
        sendEvent(ServerEventType.GameStarted, { game_id: 'test-game' });
      }).not.toThrow();

      // GAME_STARTED should not be applied to game engine
      expect(gameEngine.applyEvent).not.toHaveBeenCalled();
    });

    it('should handle granular events when game engine is not ready', async () => {
      vi.mocked(gameEngine.isReady).mockReturnValue(false);
      await simulateConnection();
      
      expect(() => {
        sendEvent(ServerEventType.PhaseChanged, { phase_type: 'DISCUSSION' });
        sendEvent(ServerEventType.ChatMessage, { sender_id: 'player1', message: 'test' });
        sendEvent(ServerEventType.VoteCast, { voter_id: 'player1', target_id: 'player2' });
      }).not.toThrow();

      expect(gameEngine.applyEvent).not.toHaveBeenCalled();
    });

    it('should handle unknown event types gracefully', async () => {
      await simulateConnection();
      
      expect(() => {
        sendEvent('UNKNOWN_EVENT_TYPE', { some: 'data' });
        sendEvent('ANOTHER_FAKE_EVENT', {});
      }).not.toThrow();
    });

    it('should handle malformed JSON events', async () => {
      await simulateConnection();
      
      expect(() => {
        if (mockWebSocket?.onmessage) {
          mockWebSocket.onmessage({ data: 'invalid json{' });
          mockWebSocket.onmessage({ data: '' });
          mockWebSocket.onmessage({ data: '{}' }); // Valid JSON but missing required fields
        }
      }).not.toThrow();
    });

    it('should handle events with missing required fields', async () => {
      await simulateConnection();
      
      expect(() => {
        // Events missing type
        if (mockWebSocket?.onmessage) {
          mockWebSocket.onmessage({ data: JSON.stringify({ payload: {} }) });
          mockWebSocket.onmessage({ data: JSON.stringify({ type: null, payload: {} }) });
        }
      }).not.toThrow();
    });
  });

  describe('Timing and Sequencing Issues', () => {
    it('should handle rapid event sequences', async () => {
      await simulateConnection();
      
      expect(() => {
        // Send many events rapidly
        for (let i = 0; i < 100; i++) {
          sendEvent(ServerEventType.ChatMessage, { 
            sender_id: `player${i}`, 
            message: `Message ${i}` 
          });
        }
      }).not.toThrow();
    });

    it('should handle events arriving out of sequence', async () => {
      await simulateConnection();
      
      const gameState = {
        id: 'test-game',
        players: { player1: { id: 'player1', name: 'Test' } },
        phase: { type: 'SITREP' },
      };

      expect(() => {
        // Send events in wrong order
        sendEvent(ServerEventType.PhaseChanged, { phase_type: 'DISCUSSION' });
        sendEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
        sendEvent(ServerEventType.GameStarted, { game_id: 'test-game' });
      }).not.toThrow();
    });

    it('should handle duplicate events', async () => {
      await simulateConnection();
      
      const eventPayload = { game_id: 'test-game' };
      
      expect(() => {
        // Send the same event multiple times
        sendEvent(ServerEventType.GameStarted, eventPayload);
        sendEvent(ServerEventType.GameStarted, eventPayload);
        sendEvent(ServerEventType.GameStarted, eventPayload);
      }).not.toThrow();
    });
  });

  describe('Connection State Edge Cases', () => {
    it('should handle events when WebSocket is closed', async () => {
      await simulateConnection();
      mockWebSocket.close();
      
      expect(() => {
        sendEvent(ServerEventType.GameStarted, { game_id: 'test-game' });
      }).not.toThrow();
    });

    it('should handle events during reconnection', async () => {
      await simulateConnection();
      
      // Simulate connection loss
      mockWebSocket.close(1006, 'Connection lost');
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { 
          game_state: { id: 'test', players: {} } 
        });
      }).not.toThrow();
    });

    it('should handle multiple connection attempts', async () => {
      expect(async () => {
        await simulateConnection();
        await simulateConnection(); // Second connection
        await simulateConnection(); // Third connection
      }).not.toThrow();
    });
  });

  describe('Game Engine Interaction Edge Cases', () => {
    it('should handle game engine resetAndLoadState failures', async () => {
      vi.mocked(gameEngine.resetAndLoadState).mockRejectedValue(new Error('WASM error'));
      await simulateConnection();
      
      const gameState = { id: 'test', players: {} };
      
      expect(() => {
        sendEvent(ServerEventType.GameStateUpdate, { game_state: gameState });
      }).not.toThrow();
    });

    it('should handle game engine applyEvent failures', async () => {
      vi.mocked(gameEngine.applyEvent).mockRejectedValue(new Error('Apply event failed'));
      await simulateConnection();
      
      expect(() => {
        sendEvent(ServerEventType.PhaseChanged, { phase_type: 'DISCUSSION' });
      }).not.toThrow();
    });

    it('should handle game engine becoming not ready mid-stream', async () => {
      await simulateConnection();
      
      // Start with ready game engine
      vi.mocked(gameEngine.isReady).mockReturnValue(true);
      sendEvent(ServerEventType.GameStateUpdate, { 
        game_state: { id: 'test', players: {} } 
      });
      
      // Game engine becomes not ready
      vi.mocked(gameEngine.isReady).mockReturnValue(false);
      
      expect(() => {
        sendEvent(ServerEventType.PhaseChanged, { phase_type: 'DISCUSSION' });
      }).not.toThrow();
    });
  });
});