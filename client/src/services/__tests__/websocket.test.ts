import { describe, it, expect, vi, beforeEach } from 'vitest';
import { WebSocketClient } from '../websocket';
import { ServerEvent, ServerEventType } from '../../types';

// Mock the game engine
vi.mock('../gameEngine', () => ({
  gameEngine: {
    isReady: vi.fn().mockReturnValue(true),
    applyEvent: vi.fn().mockResolvedValue(undefined),
    resetAndLoadState: vi.fn().mockResolvedValue(undefined),
  },
}));

describe('WebSocket Event Handling', () => {
  let client: WebSocketClient;
  let mockHandleServerEvent: any;
  let mockGameEngine: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    // Get the mocked gameEngine
    const { gameEngine } = await import('../gameEngine');
    mockGameEngine = gameEngine;
    
    client = new WebSocketClient();
    
    // Spy on console methods
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});

    // Access the private method for testing
    mockHandleServerEvent = (client as any).handleServerEvent.bind(client);
  });

  it('should ignore deprecated/unknown events and not apply them to game engine', () => {
    // Arrange - Create deprecated/unknown events including some common ones that might exist
    const deprecatedEvents: ServerEvent[] = [
      {
        type: 'DEPRECATED_EVENT' as ServerEventType,
        payload: { some: 'data' },
      },
      {
        type: 'UNKNOWN_EVENT' as ServerEventType,
        payload: { unknown: 'data' },
      },
      {
        type: 'OLD_VOTE_UPDATE' as ServerEventType,
        payload: { votes: {} },
      },
      {
        type: 'LEGACY_PLAYER_ACTION' as ServerEventType,
        payload: { playerId: 'test', action: 'something' },
      },
    ];

    // Act - Process the deprecated events
    deprecatedEvents.forEach(event => {
      mockHandleServerEvent(event);
    });

    // Assert - Game engine should not have been called for deprecated events
    expect(mockGameEngine.applyEvent).not.toHaveBeenCalled();
    expect(mockGameEngine.resetAndLoadState).not.toHaveBeenCalled();

    // Should log unknown event types for each deprecated event
    expect(console.log).toHaveBeenCalledWith(
      expect.stringContaining('Unknown event type: DEPRECATED_EVENT')
    );
    expect(console.log).toHaveBeenCalledWith(
      expect.stringContaining('Unknown event type: UNKNOWN_EVENT')
    );
    expect(console.log).toHaveBeenCalledWith(
      expect.stringContaining('Unknown event type: OLD_VOTE_UPDATE')
    );
    expect(console.log).toHaveBeenCalledWith(
      expect.stringContaining('Unknown event type: LEGACY_PLAYER_ACTION')
    );
  });

  it('should correctly handle NIGHT_ACTIONS_RESOLVED event and apply to game engine', () => {
    // Arrange
    const nightActionsEvent: ServerEvent = {
      type: ServerEventType.NightActionsResolved,
      id: 'event-123',
      gameId: 'test-game',
      timestamp: '2024-01-01T00:00:00Z',
      payload: {
        nightActionResults: [
          {
            id: 'action-1',
            type: 'INVESTIGATION',
            playerName: 'Alice',
            targetName: 'Bob',
            result: 'success',
            description: 'Alice investigated Bob',
            isPublic: false,
          },
        ],
      },
    };

    // Act
    mockHandleServerEvent(nightActionsEvent);

    // Assert - Game engine should have been called to apply the event
    expect(mockGameEngine.applyEvent).toHaveBeenCalledTimes(1);
    expect(mockGameEngine.applyEvent).toHaveBeenCalledWith({
      id: 'event-123',
      type: ServerEventType.NightActionsResolved,
      gameId: 'test-game',
      playerId: '',
      timestamp: '2024-01-01T00:00:00Z',
      payload: {
        nightActionResults: [
          {
            id: 'action-1',
            type: 'INVESTIGATION',
            playerName: 'Alice',
            targetName: 'Bob',
            result: 'success',
            description: 'Alice investigated Bob',
            isPublic: false,
          },
        ],
      },
    });

    // Should log that it's applying the event
    expect(console.log).toHaveBeenCalledWith(
      'Applying granular event NIGHT_ACTIONS_RESOLVED to game engine'
    );
  });

  it('should handle NIGHT_ACTIONS_RESOLVED with multiple action results correctly', () => {
    // Arrange
    const complexNightActionsEvent: ServerEvent = {
      type: ServerEventType.NightActionsResolved,
      id: 'event-456',
      gameId: 'test-game-2',
      timestamp: '2024-01-02T02:00:00Z',
      payload: {
        nightActionResults: [
          {
            id: 'action-1',
            type: 'INVESTIGATION',
            playerName: 'Detective',
            targetName: 'Suspect',
            result: 'success',
            description: 'Detective investigated Suspect and found them suspicious',
            isPublic: false,
          },
          {
            id: 'action-2',
            type: 'CONVERSION',
            playerName: 'AI_Player',
            targetName: 'Human_Target',
            result: 'failed',
            description: 'AI attempted conversion but triggered System Shock',
            isPublic: true,
          },
          {
            id: 'action-3',
            type: 'PROTECTION',
            playerName: 'Protector',
            targetName: 'VIP',
            result: 'success',
            description: 'VIP was successfully protected',
            isPublic: false,
          },
        ],
      },
    };

    // Act
    mockHandleServerEvent(complexNightActionsEvent);

    // Assert
    expect(mockGameEngine.applyEvent).toHaveBeenCalledTimes(1);
    expect(mockGameEngine.applyEvent).toHaveBeenCalledWith({
      id: 'event-456',
      type: ServerEventType.NightActionsResolved,
      gameId: 'test-game-2',
      playerId: '',
      timestamp: '2024-01-02T02:00:00Z',
      payload: {
        nightActionResults: expect.arrayContaining([
          expect.objectContaining({
            type: 'INVESTIGATION',
            result: 'success',
          }),
          expect.objectContaining({
            type: 'CONVERSION',
            result: 'failed',
            description: expect.stringContaining('System Shock'),
          }),
          expect.objectContaining({
            type: 'PROTECTION',
            result: 'success',
          }),
        ]),
      },
    });
  });

  it('should handle NIGHT_ACTIONS_RESOLVED with empty results array', () => {
    // Arrange
    const emptyNightActionsEvent: ServerEvent = {
      type: ServerEventType.NightActionsResolved,
      id: 'event-789',
      gameId: 'test-game-3',
      timestamp: '2024-01-03T03:00:00Z',
      payload: {
        nightActionResults: [],
      },
    };

    // Act
    mockHandleServerEvent(emptyNightActionsEvent);

    // Assert - Should still call game engine even with empty results
    expect(mockGameEngine.applyEvent).toHaveBeenCalledTimes(1);
    expect(mockGameEngine.applyEvent).toHaveBeenCalledWith({
      id: 'event-789',
      type: ServerEventType.NightActionsResolved,
      gameId: 'test-game-3',
      playerId: '',
      timestamp: '2024-01-03T03:00:00Z',
      payload: {
        nightActionResults: [],
      },
    });
  });

  it('should handle events when game engine is not ready', () => {
    // Arrange - Mock game engine as not ready
    mockGameEngine.isReady.mockReturnValue(false);

    const nightActionsEvent: ServerEvent = {
      type: ServerEventType.NightActionsResolved,
      id: 'event-123',
      payload: { nightActionResults: [] },
    };

    // Act
    mockHandleServerEvent(nightActionsEvent);

    // Assert - Should warn but not apply event
    expect(mockGameEngine.applyEvent).not.toHaveBeenCalled();
    expect(console.warn).toHaveBeenCalledWith(
      'Game engine not ready for event NIGHT_ACTIONS_RESOLVED, will buffer for later'
    );
  });

  it('should handle UI-only events without applying to game engine', () => {
    // Arrange
    const uiOnlyEvents: ServerEvent[] = [
      {
        type: ServerEventType.SystemMessage,
        payload: { message: 'System notification' },
      },
      {
        type: ServerEventType.LobbyStateUpdate,
        payload: { players: [] },
      },
      {
        type: ServerEventType.ClientIdentified,
        payload: { playerId: 'test-player' },
      },
    ];

    // Act
    uiOnlyEvents.forEach(event => {
      mockHandleServerEvent(event);
    });

    // Assert - Should not apply to game engine
    expect(mockGameEngine.applyEvent).not.toHaveBeenCalled();
    // Note: These events are handled directly by UI subscribers, not logged as unknown
  });

});