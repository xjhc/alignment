import { useEffect, useState, useCallback, useRef } from 'react';
import { websocketClient } from '../services/websocket';
import { ServerEvent, ConnectionState, ClientAction } from '../types';

export function useWebSocket() {
  const [connectionState, setConnectionState] = useState<ConnectionState>(
    websocketClient.getConnectionState()
  );

  useEffect(() => {
    const handleConnectionStateChange = (state: ConnectionState) => {
      setConnectionState(state);
    };

    websocketClient.onConnectionStateChange(handleConnectionStateChange);

    return () => {
      websocketClient.offConnectionStateChange(handleConnectionStateChange);
    };
  }, []);

  const connect = useCallback(
    (gameId?: string, playerId?: string, sessionToken?: string, lastEventId?: string) => {
      return websocketClient.connect(gameId, playerId, sessionToken, lastEventId);
    },
    []
  );

  const disconnect = useCallback(() => {
    websocketClient.disconnect();
  }, []);

  const sendAction = useCallback((action: ClientAction) => {
    websocketClient.sendAction(action);
  }, []);

  const subscribe = useCallback((eventType: string, handler: (event: ServerEvent) => void) => {
    websocketClient.on(eventType, handler);
    return () => websocketClient.off(eventType, handler);
  }, []);

  return {
    connectionState,
    connect,
    disconnect,
    sendAction,
    subscribe,
    isConnected: connectionState.isConnected,
    isReconnecting: connectionState.isReconnecting,
    lastError: connectionState.lastError
  };
}

// Hook for subscribing to specific event types
export function useWebSocketEvent<T = any>(
  eventType: string,
  handler: (payload: T, event: ServerEvent) => void
) {
  const { subscribe } = useWebSocket();
  const savedHandler = useRef(handler);

  // Keep the handler ref up-to-date
  useEffect(() => {
    savedHandler.current = handler;
  }, [handler]);

  useEffect(() => {
    const eventHandler = (event: ServerEvent) => savedHandler.current(event.payload, event);
    const unsubscribe = subscribe(eventType, eventHandler);
    return unsubscribe;
  }, [eventType, subscribe]); // Now the effect only re-runs if eventType or subscribe change
}

// Hook for handling the full suite of game events with automatic game engine integration
export function useGameEvents() {
  const { subscribe, isConnected } = useWebSocket();

  useEffect(() => {
    if (!isConnected) return;

    const unsubscribeFunctions: (() => void)[] = [];

    // Game lifecycle events
    unsubscribeFunctions.push(
      subscribe('GAME_STARTED', (event) => {
        console.log('Game started:', event.payload);
      }),
      subscribe('GAME_ENDED', (event) => {
        console.log('Game ended:', event.payload);
      }),
      subscribe('PHASE_CHANGED', (event) => {
        console.log('Phase changed:', event.payload);
      })
    );

    // Player events
    unsubscribeFunctions.push(
      subscribe('PLAYER_JOINED', (event) => {
        console.log('Player joined:', event.payload);
      }),
      subscribe('PLAYER_LEFT', (event) => {
        console.log('Player left:', event.payload);
      }),
      subscribe('PLAYER_DEACTIVATED', (event) => {
        console.log('Player eliminated:', event.payload);
      }),
      subscribe('ROLES_ASSIGNED', (event) => {
        console.log('Roles assigned:', event.payload);
      }),
      subscribe('ALIGNMENT_CHANGED', (event) => {
        console.log('Player alignment changed:', event.payload);
      })
    );

    // Chat and communication events
    unsubscribeFunctions.push(
      subscribe('CHAT_MESSAGE_POSTED', (event) => {
        console.log('Chat message posted:', event.payload);
      }),
      subscribe('PRIVATE_NOTIFICATION', (event) => {
        console.log('Private notification:', event.payload);
      })
    );

    // Voting events
    unsubscribeFunctions.push(
      subscribe('VOTE_CAST', (event) => {
        console.log('Vote cast:', event.payload);
      }),
      subscribe('VOTE_STARTED', (event) => {
        console.log('Vote started:', event.payload);
      }),
      subscribe('VOTE_COMPLETED', (event) => {
        console.log('Vote completed:', event.payload);
      })
    );

    // Token and mining events
    unsubscribeFunctions.push(
      subscribe('TOKENS_AWARDED', (event) => {
        console.log('Tokens awarded:', event.payload);
      }),
      subscribe('MINING_SUCCESSFUL', (event) => {
        console.log('Mining successful:', event.payload);
      }),
      subscribe('MINING_FAILED', (event) => {
        console.log('Mining failed:', event.payload);
      })
    );

    // Night action and crisis events
    unsubscribeFunctions.push(
      subscribe('NIGHT_ACTIONS_RESOLVED', (event) => {
        console.log('Night actions resolved:', event.payload);
      }),
      subscribe('PULSE_CHECK_SUBMITTED', (event) => {
        console.log('Pulse check submitted:', event.payload);
      }),
      subscribe('CRISIS_TRIGGERED', (event) => {
        console.log('Crisis triggered:', event.payload);
      })
    );

    // Sync and lobby events
    unsubscribeFunctions.push(
      subscribe('SYNC_COMPLETE', (event) => {
        console.log('Sync complete:', event.payload);
      }),
      subscribe('LOBBY_LIST_UPDATE', (event) => {
        console.log('Lobby list updated:', event.payload);
      }),
      subscribe('LOBBY_STATE_UPDATE', (event) => {
        console.log('Lobby state updated:', event.payload);
      })
    );

    // Return cleanup function
    return () => {
      unsubscribeFunctions.forEach(fn => fn());
    };
  }, [subscribe, isConnected]);

  return { isConnected };
}