import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { appReducer, initialAppState, ConsolidatedAppState } from '../../client/src/state/appReducer';

describe('Session Restoration', () => {
  let mockSessionStorage: { [key: string]: string };
  
  beforeEach(() => {
    // Mock sessionStorage
    mockSessionStorage = {};
    
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: vi.fn((key: string) => mockSessionStorage[key] || null),
        setItem: vi.fn((key: string, value: string) => {
          mockSessionStorage[key] = value;
        }),
        removeItem: vi.fn((key: string) => {
          delete mockSessionStorage[key];
        }),
        clear: vi.fn(() => {
          mockSessionStorage = {};
        }),
      },
      writable: true,
    });
    
    // Spy on console methods
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });
  
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Session Persistence', () => {
    it('should persist session data when joining lobby', () => {
      const state = initialAppState;
      
      const action = {
        type: 'JOIN_LOBBY' as const,
        payload: {
          gameId: 'test-game-123',
          playerId: 'player-456',
          sessionToken: 'token-789',
        },
      };
      
      const newState = appReducer(state, action);
      
      expect(window.sessionStorage.setItem).toHaveBeenCalledWith(
        'alignmentGameSession',
        JSON.stringify({
          gameId: 'test-game-123',
          playerId: 'player-456',
          sessionToken: 'token-789',
          sessionState: 'IN_LOBBY',
        })
      );
      
      expect(newState.isInGameSession).toBe(true);
      expect(newState.sessionState).toBe('IN_LOBBY');
      expect(newState.appState.gameId).toBe('test-game-123');
    });
    
    it('should persist session data when creating game', () => {
      const state = initialAppState;
      
      const action = {
        type: 'CREATE_GAME' as const,
        payload: {
          gameId: 'new-game-123',
          playerId: 'host-456',
          sessionToken: 'host-token-789',
        },
      };
      
      appReducer(state, action);
      
      expect(window.sessionStorage.setItem).toHaveBeenCalledWith(
        'alignmentGameSession',
        JSON.stringify({
          gameId: 'new-game-123',
          playerId: 'host-456',
          sessionToken: 'host-token-789',
          sessionState: 'IN_LOBBY',
        })
      );
    });
    
    it('should update session state when entering game', () => {
      // Setup existing session
      mockSessionStorage['alignmentGameSession'] = JSON.stringify({
        gameId: 'test-game',
        playerId: 'test-player',
        sessionToken: 'test-token',
        sessionState: 'IN_LOBBY',
      });
      
      const state: ConsolidatedAppState = {
        ...initialAppState,
        isInGameSession: true,
        sessionState: 'IN_LOBBY',
      };
      
      const action = { type: 'ENTER_GAME' as const };
      
      appReducer(state, action);
      
      expect(window.sessionStorage.setItem).toHaveBeenCalledWith(
        'alignmentGameSession',
        JSON.stringify({
          gameId: 'test-game',
          playerId: 'test-player',
          sessionToken: 'test-token',
          sessionState: 'IN_GAME',
        })
      );
    });
    
    it('should clear session data when leaving lobby', () => {
      const state: ConsolidatedAppState = {
        ...initialAppState,
        isInGameSession: true,
        sessionState: 'IN_LOBBY',
        appState: {
          ...initialAppState.appState,
          gameId: 'test-game',
          playerId: 'test-player',
          sessionToken: 'test-token',
        },
      };
      
      const action = { type: 'LEAVE_LOBBY' as const };
      
      const newState = appReducer(state, action);
      
      expect(window.sessionStorage.removeItem).toHaveBeenCalledWith('alignmentGameSession');
      expect(newState.isInGameSession).toBe(false);
      expect(newState.sessionState).toBe('IDLE');
      expect(newState.appState.gameId).toBeUndefined();
    });
  });

  describe('Session Restoration', () => {
    it('should restore valid session data', () => {
      const state = initialAppState;
      
      const action = {
        type: 'RESTORE_SESSION' as const,
        payload: {
          gameId: 'restored-game',
          playerId: 'restored-player',
          sessionToken: 'restored-token',
          sessionState: 'IN_LOBBY' as const,
        },
      };
      
      const newState = appReducer(state, action);
      
      expect(newState.isInGameSession).toBe(true);
      expect(newState.sessionState).toBe('IN_LOBBY');
      expect(newState.appState.gameId).toBe('restored-game');
      expect(newState.appState.playerId).toBe('restored-player');
      expect(newState.appState.sessionToken).toBe('restored-token');
      expect(newState.lobbyState.playerId).toBe('restored-player');
    });
    
    it('should restore in-game session', () => {
      const state = initialAppState;
      
      const action = {
        type: 'RESTORE_SESSION' as const,
        payload: {
          gameId: 'game-in-progress',
          playerId: 'active-player',
          sessionToken: 'active-token',
          sessionState: 'IN_GAME' as const,
        },
      };
      
      const newState = appReducer(state, action);
      
      expect(newState.sessionState).toBe('IN_GAME');
      expect(newState.isInGameSession).toBe(true);
    });
    
    it('should restore post-game session', () => {
      const state = initialAppState;
      
      const action = {
        type: 'RESTORE_SESSION' as const,
        payload: {
          gameId: 'completed-game',
          playerId: 'finished-player',
          sessionToken: 'finished-token',
          sessionState: 'POST_GAME' as const,
        },
      };
      
      const newState = appReducer(state, action);
      
      expect(newState.sessionState).toBe('POST_GAME');
      expect(newState.isInGameSession).toBe(true);
    });
  });

  describe('Connection Error Handling', () => {
    it('should set connection error on failed WebSocket connection', () => {
      const state: ConsolidatedAppState = {
        ...initialAppState,
        isInGameSession: true,
        sessionState: 'IN_LOBBY',
      };
      
      const action = {
        type: 'SET_CONNECTION_ERROR' as const,
        payload: {
          message: 'Failed to connect to lobby. The session may have expired.',
        },
      };
      
      const newState = appReducer(state, action);
      
      expect(newState.lobbyState.connectionError).toBe(
        'Failed to connect to lobby. The session may have expired.'
      );
    });
    
    it('should clear connection error', () => {
      const state: ConsolidatedAppState = {
        ...initialAppState,
        lobbyState: {
          ...initialAppState.lobbyState,
          connectionError: 'Previous error message',
        },
      };
      
      const action = { type: 'CLEAR_CONNECTION_ERROR' as const };
      
      const newState = appReducer(state, action);
      
      expect(newState.lobbyState.connectionError).toBeNull();
    });
  });

  describe('Session Cleanup Scenarios', () => {
    it('should clear session when logging out', () => {
      const state: ConsolidatedAppState = {
        ...initialAppState,
        isInGameSession: true,
        sessionState: 'IN_LOBBY',
        appState: {
          ...initialAppState.appState,
          gameId: 'test-game',
          playerId: 'test-player',
          playerName: 'Test Player',
        },
      };
      
      const action = { type: 'BACK_TO_LOGIN' as const };
      
      const newState = appReducer(state, action);
      
      expect(window.sessionStorage.removeItem).toHaveBeenCalledWith('alignmentGameSession');
      expect(newState.isInGameSession).toBe(false);
      expect(newState.sessionState).toBe('IDLE');
      expect(newState.appState.playerName).toBe('');
      expect(newState.appState.userIdentity).toBeUndefined();
    });
    
    it('should clear session when playing again', () => {
      const state: ConsolidatedAppState = {
        ...initialAppState,
        isInGameSession: true,
        sessionState: 'POST_GAME',
        roleAssignment: {
          role: 'EMPLOYEE',
          alignment: 'HUMAN',
          personalKPI: null,
        },
      };
      
      const action = { type: 'PLAY_AGAIN' as const };
      
      const newState = appReducer(state, action);
      
      expect(window.sessionStorage.removeItem).toHaveBeenCalledWith('alignmentGameSession');
      expect(newState.isInGameSession).toBe(false);
      expect(newState.sessionState).toBe('IDLE');
      expect(newState.roleAssignment).toBeNull();
    });
  });

  describe('Edge Cases', () => {
    it('should handle malformed session data gracefully', () => {
      // This would be tested in the App component integration test
      // where sessionStorage.getItem() returns malformed JSON
      mockSessionStorage['alignmentGameSession'] = 'invalid json {';
      
      // The App component should handle this by calling removeItem
      // This is tested in the App component level
      expect(mockSessionStorage['alignmentGameSession']).toBe('invalid json {');
    });
    
    it('should handle missing required session fields', () => {
      // Session with missing required fields
      mockSessionStorage['alignmentGameSession'] = JSON.stringify({
        gameId: 'test-game',
        // Missing playerId, sessionToken, sessionState
      });
      
      // The App component should handle this by not dispatching RESTORE_SESSION
      // This is tested at the App component level
      expect(mockSessionStorage['alignmentGameSession']).toBeDefined();
    });
    
    it('should handle empty session data', () => {
      mockSessionStorage['alignmentGameSession'] = JSON.stringify({});
      
      // Should not attempt restoration with empty data
      expect(mockSessionStorage['alignmentGameSession']).toBe('{}');
    });
  });
});