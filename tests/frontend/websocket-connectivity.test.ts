import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocketClient } from '../../client/src/services/websocket';

// Mock WebSocket implementation for testing
class MockWebSocket {
  public readyState: number = WebSocket.CONNECTING;
  public onopen: ((event: Event) => void) | null = null;
  public onclose: ((event: CloseEvent) => void) | null = null;
  public onerror: ((event: Event) => void) | null = null;
  public onmessage: ((event: MessageEvent) => void) | null = null;
  
  constructor(public url: string) {}
  
  send(data: string) {
    // Mock implementation
  }
  
  close(code?: number, reason?: string) {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) {
      this.onclose(new CloseEvent('close', { code: code || 1000, reason: reason || '' }));
    }
  }
  
  // Test helpers
  mockOpen() {
    this.readyState = WebSocket.OPEN;
    if (this.onopen) {
      this.onopen(new Event('open'));
    }
  }
  
  mockError() {
    if (this.onerror) {
      this.onerror(new Event('error'));
    }
  }
  
  mockMessage(data: string) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data }));
    }
  }
}

// Mock the game engine
vi.mock('../gameEngine', () => ({
  gameEngine: {
    isReady: vi.fn().mockReturnValue(true),
    applyEvent: vi.fn().mockResolvedValue(undefined),
    resetAndLoadState: vi.fn().mockResolvedValue(undefined),
  },
}));

describe('WebSocket Connectivity', () => {
  let mockWebSocket: MockWebSocket;
  let client: WebSocketClient;
  
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock WebSocket globally
    global.WebSocket = vi.fn().mockImplementation((url: string) => {
      mockWebSocket = new MockWebSocket(url);
      return mockWebSocket;
    });
    
    // Spy on console methods
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
    vi.spyOn(console, 'warn').mockImplementation(() => {});
  });
  
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Dynamic URL Construction', () => {
    it('should construct WebSocket URL from window location', () => {
      // Mock window.location
      Object.defineProperty(window, 'location', {
        value: {
          protocol: 'http:',
          hostname: '172.17.0.3',
        },
        writable: true,
      });
      
      client = new WebSocketClient();
      
      const gameId = 'test-game';
      const playerId = 'test-player';
      const sessionToken = 'test-token';
      
      client.connect(gameId, playerId, sessionToken);
      
      expect(global.WebSocket).toHaveBeenCalledWith(
        'ws://172.17.0.3:8080/ws?gameId=test-game&playerId=test-player&sessionToken=test-token'
      );
    });
    
    it('should use wss for HTTPS origins', () => {
      Object.defineProperty(window, 'location', {
        value: {
          protocol: 'https:',
          hostname: 'example.com',
        },
        writable: true,
      });
      
      client = new WebSocketClient();
      
      client.connect('game', 'player', 'token');
      
      expect(global.WebSocket).toHaveBeenCalledWith(
        'wss://example.com:8080/ws?gameId=game&playerId=player&sessionToken=token'
      );
    });
    
    it('should allow custom URL override', () => {
      client = new WebSocketClient('ws://custom-server:9090/ws');
      
      client.connect('game', 'player', 'token');
      
      expect(global.WebSocket).toHaveBeenCalledWith(
        'ws://custom-server:9090/ws?gameId=game&playerId=player&sessionToken=token'
      );
    });
  });

  describe('Connection Management', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'location', {
        value: { protocol: 'http:', hostname: 'localhost' },
        writable: true,
      });
      client = new WebSocketClient();
    });
    
    it('should successfully connect and update connection state', async () => {
      const connectPromise = client.connect('game', 'player', 'token');
      
      expect(client.getConnectionState().isConnected).toBe(false);
      
      // Simulate successful connection
      mockWebSocket.mockOpen();
      
      await connectPromise;
      
      expect(client.getConnectionState().isConnected).toBe(true);
      expect(client.getConnectionState().isReconnecting).toBe(false);
      expect(console.log).toHaveBeenCalledWith('WebSocket connected');
    });
    
    it('should handle connection timeout', async () => {
      vi.useFakeTimers();
      
      const connectPromise = client.connect('game', 'player', 'token');
      
      // Fast-forward past the 10 second timeout
      vi.advanceTimersByTime(10000);
      
      await expect(connectPromise).rejects.toThrow('Connection timeout - unable to connect to server');
      
      vi.useRealTimers();
    });
    
    it('should handle connection errors', async () => {
      const connectPromise = client.connect('game', 'player', 'token');
      
      // Simulate connection error
      mockWebSocket.mockError();
      
      await expect(connectPromise).rejects.toThrow();
      
      expect(client.getConnectionState().isConnected).toBe(false);
      expect(client.getConnectionState().lastError).toBe('Connection error');
    });
    
    it('should validate connection parameters', async () => {
      await expect(client.connect('', 'player', 'token')).rejects.toThrow(
        'Invalid connection parameters'
      );
      
      await expect(client.connect('game', '', 'token')).rejects.toThrow(
        'Invalid connection parameters'
      );
      
      await expect(client.connect('game', 'player', '')).rejects.toThrow(
        'Invalid connection parameters'
      );
    });
  });

  describe('Message Handling', () => {
    beforeEach(async () => {
      Object.defineProperty(window, 'location', {
        value: { protocol: 'http:', hostname: 'localhost' },
        writable: true,
      });
      client = new WebSocketClient();
      
      // Establish connection
      const connectPromise = client.connect('game', 'player', 'token');
      mockWebSocket.mockOpen();
      await connectPromise;
    });
    
    it('should handle LOBBY_STATE_UPDATE events', () => {
      const eventHandler = vi.fn();
      client.on('LOBBY_STATE_UPDATE', eventHandler);
      
      const lobbyUpdateEvent = {
        type: 'LOBBY_STATE_UPDATE',
        payload: {
          players: [{ id: 'player1', name: 'Test Player' }],
          host_id: 'player1',
          can_start: false,
        },
      };
      
      mockWebSocket.mockMessage(JSON.stringify(lobbyUpdateEvent));
      
      expect(eventHandler).toHaveBeenCalledWith(lobbyUpdateEvent);
    });
    
    it('should handle CLIENT_IDENTIFIED events', () => {
      const eventHandler = vi.fn();
      client.on('CLIENT_IDENTIFIED', eventHandler);
      
      const clientIdentifiedEvent = {
        type: 'CLIENT_IDENTIFIED',
        payload: { your_player_id: 'player-123' },
      };
      
      mockWebSocket.mockMessage(JSON.stringify(clientIdentifiedEvent));
      
      expect(eventHandler).toHaveBeenCalledWith(clientIdentifiedEvent);
    });
    
    it('should handle multiple events in a single message', () => {
      const eventHandler = vi.fn();
      client.on('LOBBY_STATE_UPDATE', eventHandler);
      
      const multipleEvents = [
        { type: 'LOBBY_STATE_UPDATE', payload: { players: [] } },
        { type: 'LOBBY_STATE_UPDATE', payload: { players: [{ id: 'p1' }] } },
      ];
      
      const messageData = multipleEvents.map(e => JSON.stringify(e)).join('\n');
      mockWebSocket.mockMessage(messageData);
      
      expect(eventHandler).toHaveBeenCalledTimes(2);
    });
    
    it('should handle malformed JSON gracefully', () => {
      const eventHandler = vi.fn();
      client.on('LOBBY_STATE_UPDATE', eventHandler);
      
      mockWebSocket.mockMessage('invalid json {');
      
      // Should not crash and should not call the handler
      expect(eventHandler).not.toHaveBeenCalled();
      expect(console.error).toHaveBeenCalledWith(
        'Failed to parse WebSocket message:',
        expect.any(Error)
      );
    });
  });

  describe('Reconnection Logic', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'location', {
        value: { protocol: 'http:', hostname: 'localhost' },
        writable: true,
      });
      client = new WebSocketClient();
      vi.useFakeTimers();
    });
    
    afterEach(() => {
      vi.useRealTimers();
    });
    
    it('should schedule reconnection on unexpected close', async () => {
      // Initial connection
      const connectPromise = client.connect('game', 'player', 'token');
      mockWebSocket.mockOpen();
      await connectPromise;
      
      expect(client.getConnectionState().isConnected).toBe(true);
      
      // Simulate unexpected close (not code 1000)
      if (mockWebSocket.onclose) {
        mockWebSocket.onclose(new CloseEvent('close', { code: 1006, reason: 'Connection lost' }));
      }
      
      expect(client.getConnectionState().isConnected).toBe(false);
      expect(client.getConnectionState().isReconnecting).toBe(true);
      
      // Fast-forward to trigger reconnection
      vi.advanceTimersByTime(2000);
      
      // Should attempt to reconnect with same credentials
      expect(global.WebSocket).toHaveBeenCalledTimes(2);
    });
    
    it('should not reconnect on normal close (code 1000)', async () => {
      const connectPromise = client.connect('game', 'player', 'token');
      mockWebSocket.mockOpen();
      await connectPromise;
      
      // Simulate normal close
      if (mockWebSocket.onclose) {
        mockWebSocket.onclose(new CloseEvent('close', { code: 1000, reason: 'Normal close' }));
      }
      
      expect(client.getConnectionState().isReconnecting).toBe(false);
      
      vi.advanceTimersByTime(5000);
      
      // Should not attempt reconnection
      expect(global.WebSocket).toHaveBeenCalledTimes(1);
    });
    
    it('should stop reconnecting if credentials are cleared', async () => {
      const connectPromise = client.connect('game', 'player', 'token');
      mockWebSocket.mockOpen();
      await connectPromise;
      
      // Disconnect explicitly (clears credentials)
      client.disconnect();
      
      expect(client.getConnectionState().isReconnecting).toBe(false);
      
      vi.advanceTimersByTime(5000);
      
      // Should not attempt reconnection
      expect(global.WebSocket).toHaveBeenCalledTimes(1);
    });
  });

  describe('Action Sending', () => {
    beforeEach(async () => {
      Object.defineProperty(window, 'location', {
        value: { protocol: 'http:', hostname: 'localhost' },
        writable: true,
      });
      client = new WebSocketClient();
      
      // Establish connection
      const connectPromise = client.connect('game', 'player', 'token');
      mockWebSocket.mockOpen();
      await connectPromise;
    });
    
    it('should send actions when connected', () => {
      const sendSpy = vi.spyOn(mockWebSocket, 'send');
      
      const action = {
        type: 'SUBMIT_VOTE' as const,
        gameId: 'game',
        playerId: 'player',
        payload: { targetId: 'target' },
      };
      
      client.sendAction(action);
      
      expect(sendSpy).toHaveBeenCalledWith(JSON.stringify(action));
    });
    
    it('should not send actions when not connected', () => {
      // Disconnect first
      client.disconnect();
      
      const sendSpy = vi.spyOn(mockWebSocket, 'send');
      
      const action = {
        type: 'SUBMIT_VOTE' as const,
        gameId: 'game',
        playerId: 'player',
        payload: { targetId: 'target' },
      };
      
      client.sendAction(action);
      
      expect(sendSpy).not.toHaveBeenCalled();
      expect(console.warn).toHaveBeenCalledWith(
        'Cannot send action: Invalid connection state'
      );
    });
  });
});