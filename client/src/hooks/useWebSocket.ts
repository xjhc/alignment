// client/src/hooks/useWebSocket.ts
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
    (gameId?: string, playerId?: string, sessionToken?: string) => {
      return websocketClient.connect(gameId, playerId, sessionToken);
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
