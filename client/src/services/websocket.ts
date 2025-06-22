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

  constructor(url: string = 'ws://localhost:8080/ws') {
    this.url = url;
  }

  connect(gameId?: string, playerId?: string, sessionToken?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {

        let wsUrl = this.url;
        if (gameId && playerId && sessionToken) {
          const params = new URLSearchParams({
            gameId: gameId,
            playerId: playerId,
            sessionToken: sessionToken
          });

          wsUrl = `${this.url}?${params.toString()}`;
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
    this.updateConnectionState({ isConnected: false, isReconnecting: false });
  }

  sendAction(action: ClientAction): void {
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

  private handleServerEvent(event: ServerEvent): void {
    console.log('Received server event:', event.type, event.payload);

    // The only event that should modify the Wasm state is GAME_STATE_UPDATE
    if (event.type === 'GAME_STATE_UPDATE') {
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
    }

    // All other events are emitted to UI listeners.
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
    this.updateConnectionState({ isReconnecting: true });
    this.reconnectInterval = window.setTimeout(() => {
      this.reconnectInterval = null;
      console.log('Attempting to reconnect...');
      this.connect()
        .catch(error => {
          console.error('Reconnection failed:', error);
          setTimeout(() => this.scheduleReconnect(), 5000);
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