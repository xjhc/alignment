import { createContext, useContext, ReactNode } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { ConnectionState, ClientAction, ServerEvent } from '../types';

interface WebSocketContextType {
  connectionState: ConnectionState;
  isConnected: boolean;
  isReconnecting: boolean;
  lastError?: string;
  connect: (gameId?: string, playerId?: string, sessionToken?: string) => Promise<void>;
  disconnect: () => void;
  sendAction: (action: ClientAction) => void;
  subscribe: (eventType: string, handler: (event: ServerEvent) => void) => () => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

interface WebSocketProviderProps {
  children: ReactNode;
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const webSocketHook = useWebSocket();

  return (
    <WebSocketContext.Provider value={webSocketHook}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocketContext() {
  const context = useContext(WebSocketContext);
  if (context === undefined) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider');
  }
  return context;
}