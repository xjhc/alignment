import { ClientAction, ServerEvent, ConnectionState } from '../types';
import { gameEngine } from './gameEngine';

export class WebSocketClient {
  private socket: WebSocket | null = null;
  private url: string;
  private eventHandlers: Map<string, ((event: ServerEvent) => void)[]> = new Map();
  private connectionStateHandlers: ((state: ConnectionState) => void)[] = [];
  private connectionState: ConnectionState = { isConnected: false, isReconnecting: false };
  private reconnectInterval: number | null = null;
  private heartbeatInterval: number | null = null;
  private connectionCredentials: { gameId: string; playerId: string; sessionToken: string; connectedAt: Date } | null = null;

  constructor(url: string = 'ws://localhost:8080/ws') {
    this.url = url;
  }

  connect(gameId?: string, playerId?: string, sessionToken?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        let wsUrl = this.url;
        if (gameId && playerId && sessionToken) {
          // Ensure proper URL encoding and validation
          const params = new URLSearchParams();
          params.set('gameId', gameId.trim());
          params.set('playerId', playerId.trim());
          params.set('sessionToken', sessionToken.trim());

          // Validate required parameters
          if (!params.get('gameId') || !params.get('playerId') || !params.get('sessionToken')) {
            throw new Error('Invalid connection parameters: gameId, playerId, and sessionToken must be non-empty');
          }

          wsUrl = `${this.url}?${params.toString()}`;

          // Store credentials for reconnection
          this.connectionCredentials = {
            gameId: params.get('gameId')!,
            playerId: params.get('playerId')!,
            sessionToken: params.get('sessionToken')!,
            connectedAt: new Date()
          };
        }

        this.socket = new WebSocket(wsUrl);

        this.socket.onopen = () => {
          console.log('WebSocket connected');
          this.updateConnectionState({ isConnected: true, isReconnecting: false });
          this.startHeartbeat();

          resolve();
        };

        this.socket.onmessage = (event) => {
          try {
            const messageData = event.data as string;
            const messages = messageData.split('\n');

            for (const messageStr of messages) {
              if (messageStr.trim() === '') continue;

              const message: ServerEvent = JSON.parse(messageStr);
              this.handleServerEvent(message);
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        this.socket.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason);
          this.updateConnectionState({
            isConnected: false,
            isReconnecting: false,
            lastError: event.reason || 'Connection closed'
          });
          this.stopHeartbeat();

          if (event.code !== 1000) {
            this.scheduleReconnect();
          }
        };

        this.socket.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.updateConnectionState({
            isConnected: false,
            isReconnecting: false,
            lastError: 'Connection error'
          });
          reject(error);
        };

      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.reconnectInterval) {
      clearTimeout(this.reconnectInterval);
      this.reconnectInterval = null;
    }
    this.stopHeartbeat();
    if (this.socket) {
      this.socket.close(1000, 'Client disconnect');
      this.socket = null;
    }
    this.connectionCredentials = null; // Clear stored credentials
    this.updateConnectionState({ isConnected: false, isReconnecting: false });
  }

  sendAction(action: ClientAction): void {
    if (!this.isValidConnection()) {
      console.warn('Cannot send action: Invalid connection state');
      return;
    }

    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      try {
        this.socket.send(JSON.stringify(action));
      } catch (error) {
        console.error('Failed to send action:', error);
      }
    } else {
      console.warn('Cannot send action: WebSocket not connected');
    }
  }

  on(eventType: string, handler: (event: ServerEvent) => void): void {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    this.eventHandlers.get(eventType)!.push(handler);
  }

  off(eventType: string, handler: (event: ServerEvent) => void): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  onConnectionStateChange(handler: (state: ConnectionState) => void): void {
    this.connectionStateHandlers.push(handler);
  }

  offConnectionStateChange(handler: (state: ConnectionState) => void): void {
    const index = this.connectionStateHandlers.indexOf(handler);
    if (index > -1) {
      this.connectionStateHandlers.splice(index, 1);
    }
  }

  getConnectionState(): ConnectionState {
    return { ...this.connectionState };
  }

  isValidConnection(): boolean {
    return this.connectionState.isConnected &&
      !this.connectionState.isReconnecting &&
      this.connectionCredentials !== null;
  }

  getConnectionAge(): number | null {
    if (!this.connectionCredentials) {
      return null;
    }
    return Date.now() - this.connectionCredentials.connectedAt.getTime();
  }

  isConnectionNearExpiry(): boolean {
    const age = this.getConnectionAge();
    // Warn if connection is older than 23 hours (tokens expire at 24h)
    return age !== null && age > 23 * 60 * 60 * 1000;
  }

  private handleServerEvent(event: ServerEvent): void {
    console.log('Received server event:', event.type, event.payload);

    // Handle events using a switch statement
    switch (event.type) {
      case 'GAME_STATE_UPDATE':
        // Full state sync - only used for initial game transition
        if (gameEngine.isReady()) {
          const gameState = event.payload?.game_state;
          if (gameState) {
            console.log('Loading core state from GAME_STATE_UPDATE...');
            gameEngine.resetAndLoadState(gameState)
              .catch(err => console.error('Failed to load game state:', err));
          }
        } else {
          console.warn('Game engine not ready for GAME_STATE_UPDATE, will retry when ready');
        }
        break;

      case 'ROLE_ASSIGNED':
      case 'GAME_STARTED':
      case 'PHASE_CHANGED':
      case 'CHAT_MESSAGE':
      case 'VOTE_CAST':
      case 'NIGHT_ACTION_SUBMITTED':
      case 'PLAYER_LEFT':
      case 'PLAYER_ELIMINATED':
        // Granular events - apply to game engine if available
        if (gameEngine.isReady()) {
          console.log(`Applying granular event ${event.type} to game engine`);
          // Convert ServerEvent to CoreEvent format
          const coreEvent = {
            id: event.id || `event_${Date.now()}`,
            type: event.type,
            gameId: event.gameId || event.game_id || '',
            playerId: event.playerId || '',
            timestamp: event.timestamp || new Date().toISOString(),
            payload: event.payload || {}
          };
          gameEngine.applyEvent(coreEvent)
            .catch(err => console.error(`Failed to apply event ${event.type}:`, err));
        } else {
          console.warn(`Game engine not ready for event ${event.type}, will buffer for later`);
          // Could implement event buffering here if needed
        }
        break;
      case 'LOBBY_STATE_UPDATE':
      case 'CLIENT_IDENTIFIED':
      case 'SYSTEM_MESSAGE':
        // These events are handled directly by UI subscribers in App.tsx.
        // The game engine doesn't need to process them.
        // We add them here to prevent the "Unknown event" log.
        break;

      default:
        // Unknown event types - just log and pass to subscribers
        console.log(`Unknown event type: ${event.type}, passing to subscribers only`);
        break;
    }

    // Always emit to UI subscribers for additional handling
    this.emitToSubscribers(event);
  }

  private emitToSubscribers(event: ServerEvent) {
    const handlers = this.eventHandlers.get(event.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(event);
        } catch (error) {
          console.error(`Error in event handler for ${event.type}:`, error);
        }
      });
    }
  }

  private updateConnectionState(newState: Partial<ConnectionState>): void {
    this.connectionState = { ...this.connectionState, ...newState };
    this.connectionStateHandlers.forEach(handler => {
      try {
        handler(this.connectionState);
      } catch (error) {
        console.error('Error in connection state handler:', error);
      }
    });
  }

  private scheduleReconnect(): void {
    if (this.reconnectInterval) {
      return;
    }

    // Don't reconnect if we don't have credentials
    if (!this.connectionCredentials) {
      console.log('No credentials available for reconnection');
      return;
    }

    this.updateConnectionState({ isReconnecting: true });
    this.reconnectInterval = window.setTimeout(() => {
      this.reconnectInterval = null;
      console.log('Attempting to reconnect...');
      const creds = this.connectionCredentials!;
      this.connect(creds.gameId, creds.playerId, creds.sessionToken)
        .catch(error => {
          console.error('Reconnection failed:', error);
          // If token is invalid, clear credentials and stop reconnecting
          if (error.message?.includes('Invalid session') || error.message?.includes('Unauthorized')) {
            console.log('Session expired, clearing credentials');
            this.connectionCredentials = null;
            this.updateConnectionState({
              isConnected: false,
              isReconnecting: false,
              lastError: 'Session expired'
            });
          } else {
            // Retry for other errors
            setTimeout(() => this.scheduleReconnect(), 5000);
          }
        });
    }, 2000);
  }

  private startHeartbeat(): void {
    this.heartbeatInterval = window.setInterval(() => {
      if (this.socket && this.socket.readyState === WebSocket.OPEN) {
        this.socket.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

}

// Singleton instance
export const websocketClient = new WebSocketClient();