import { ClientAction, ServerEvent, ConnectionState, ServerEventType } from '../types';
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

  constructor(url?: string) {
    // Auto-detect the WebSocket URL based on current location
    if (!url) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.hostname;
      const port = '8080'; // Backend port
      this.url = `${protocol}//${host}:${port}/ws`;
    } else {
      this.url = url;
    }
  }

  connect(gameId?: string, playerId?: string, sessionToken?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // Validate required parameters upfront
        if (!gameId || !playerId || !sessionToken || 
            gameId.trim() === '' || playerId.trim() === '' || sessionToken.trim() === '') {
          console.error('WebSocket connect called with invalid parameters:', {
            gameId: gameId || 'undefined',
            playerId: playerId || 'undefined', 
            sessionToken: sessionToken ? 'present' : 'undefined'
          });
          reject(new Error('No credentials available for WebSocket connection'));
          return;
        }

        let wsUrl = this.url;
        if (gameId && playerId && sessionToken) {
          // Ensure proper URL encoding and validation
          const params = new URLSearchParams();
          params.set('gameId', gameId.trim());
          params.set('playerId', playerId.trim());
          params.set('sessionToken', sessionToken.trim());

          // Double check after trimming
          if (!params.get('gameId') || !params.get('playerId') || !params.get('sessionToken')) {
            reject(new Error('Invalid connection parameters'));
            return;
          }

          wsUrl = `${this.url}?${params.toString()}`;

          // Store credentials for reconnection (persist to sessionStorage as backup)
          const credentials = {
            gameId: params.get('gameId')!,
            playerId: params.get('playerId')!,
            sessionToken: params.get('sessionToken')!,
            connectedAt: new Date()
          };
          
          this.connectionCredentials = credentials;
          
          // Also store in sessionStorage as a backup in case of memory loss
          try {
            sessionStorage.setItem('wsConnectionCredentials', JSON.stringify({
              gameId: credentials.gameId,
              playerId: credentials.playerId,
              sessionToken: credentials.sessionToken,
              connectedAt: credentials.connectedAt.toISOString()
            }));
          } catch (e) {
            console.warn('Failed to persist WebSocket credentials to sessionStorage:', e);
          }
        }

        // Add connection timeout
        const connectionTimeout = setTimeout(() => {
          reject(new Error('Connection timeout - unable to connect to server'));
        }, 10000); // 10 second timeout

        this.socket = new WebSocket(wsUrl);

        this.socket.onopen = () => {
          console.log('WebSocket connected');
          clearTimeout(connectionTimeout);
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

          // Only reconnect if it wasn't a normal closure (1000) and we have credentials
          if (event.code !== 1000 && this.connectionCredentials) {
            // Reset reconnect attempts on unexpected disconnection
            this.reconnectAttempts = 0;
            this.scheduleReconnect();
          }
        };

        this.socket.onerror = (error) => {
          console.error('WebSocket error:', error);
          clearTimeout(connectionTimeout);
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
    this.reconnectAttempts = 0; // Reset reconnect attempts
    
    // Clear backup credentials from sessionStorage
    try {
      sessionStorage.removeItem('wsConnectionCredentials');
    } catch (e) {
      console.warn('Failed to clear WebSocket credentials from sessionStorage:', e);
    }
    
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

  // Add manual reconnection method
  forceReconnect(): void {
    if (this.connectionCredentials) {
      console.log('Forcing manual reconnection...');
      this.reconnectAttempts = 0; // Reset attempts for manual reconnection
      this.scheduleReconnect();
    } else {
      console.warn('Cannot force reconnect: No credentials available');
    }
  }
  
  // Add reconnection listener management
  onReconnect(listener: (isReconnecting: boolean, attempts: number) => void): void {
    this.reconnectListeners.push(listener);
  }
  
  offReconnect(listener: (isReconnecting: boolean, attempts: number) => void): void {
    const index = this.reconnectListeners.indexOf(listener);
    if (index > -1) {
      this.reconnectListeners.splice(index, 1);
    }
  }
  
  private notifyReconnectListeners(isReconnecting: boolean, attempts: number): void {
    this.reconnectListeners.forEach(listener => {
      try {
        listener(isReconnecting, attempts);
      } catch (error) {
        console.error('Error in reconnection listener:', error);
      }
    });
  }
  
  // Get reconnection status
  getReconnectionStatus(): { isReconnecting: boolean; attempts: number; maxAttempts: number } {
    return {
      isReconnecting: this.connectionState.isReconnecting,
      attempts: this.reconnectAttempts,
      maxAttempts: this.maxReconnectAttempts
    };
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

    // Handle events using a switch statement with generated enum
    switch (event.type) {
      case ServerEventType.GameStateUpdate:
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

      case ServerEventType.RoleAssigned:
      case ServerEventType.PhaseChanged:
      case ServerEventType.ChatMessage:
      case ServerEventType.MessageReaction:
      case ServerEventType.IncitingIncident:
      case ServerEventType.LoebmateMessage:
      case ServerEventType.VoteCast:
      case ServerEventType.NightActionSubmitted:
      case ServerEventType.NightActionsResolved:
      case ServerEventType.PlayerLeft:
      case ServerEventType.PlayerEliminated:
      case ServerEventType.PulseCheckStarted:
      case ServerEventType.PulseCheckUpdated:
      case ServerEventType.MandateActivated:
        // Granular events - apply to game engine if available
        if (gameEngine.isReady()) {
          console.log(`Applying granular event ${event.type} to game engine`);

          // Convert ServerEvent to CoreEvent format
          let coreEvent = {
            id: event.id || `event_${Date.now()}`,
            type: event.type,
            gameId: event.gameId || event.game_id || '',
            playerId: event.playerId || '',
            timestamp: event.timestamp || new Date().toISOString(),
            payload: event.payload || {}
          };

          // Special handling for chat messages to ensure proper format
          if (event.type === ServerEventType.ChatMessage) {
            // Backend sends chat messages with this payload structure:
            // payload: { sender_id, sender_name, message, phase, day_number, channel_id }
            // We need to make sure the playerId is set from sender_id
            if (event.payload?.sender_id) {
              coreEvent.playerId = event.payload.sender_id;
            }
          }

          gameEngine.applyEvent(coreEvent)
            .catch(err => console.error(`Failed to apply event ${event.type}:`, err));
        } else {
          console.warn(`Game engine not ready for event ${event.type}, will buffer for later`);
          // Could implement event buffering here if needed
        }
        break;

      case ServerEventType.GameStarted:
        // GAME_STARTED events don't need to be applied to the game engine
        // since they're just notifications. The actual game state will come
        // via GAME_STATE_UPDATE event which loads the full initial state.
        console.log('Game started event received - awaiting state update');
        break;
        
      // Events handled directly by UI subscribers
      case ServerEventType.SystemMessage:
      case ServerEventType.LobbyStateUpdate:
      case ServerEventType.ClientIdentified:
      case ServerEventType.ChatHistorySnapshot:
        // These events are handled directly by UI subscribers in App.tsx.
        // The game engine doesn't need to process them.
        break;

      case ServerEventType.SessionExpired:
        // Handle session expiry by clearing credentials and stopping reconnection
        console.log('Session expired:', event.payload?.message || 'Session has expired');
        this.connectionCredentials = null;
        
        // Clear backup credentials from sessionStorage
        try {
          sessionStorage.removeItem('wsConnectionCredentials');
        } catch (e) {
          console.warn('Failed to clear WebSocket credentials from sessionStorage:', e);
        }
        
        this.updateConnectionState({
          isConnected: false,
          isReconnecting: false,
          lastError: event.payload?.message || 'Session expired'
        });
        break;

      case ServerEventType.ForceLogout:
        // Handle server-forced logout due to unrecoverable state
        console.log('Server forced logout:', event.payload?.message || 'Forced logout by server');
        this.connectionCredentials = null;
        
        // Clear backup credentials from sessionStorage
        try {
          sessionStorage.removeItem('wsConnectionCredentials');
        } catch (e) {
          console.warn('Failed to clear WebSocket credentials from sessionStorage:', e);
        }
        
        this.updateConnectionState({
          isConnected: false,
          isReconnecting: false,
          lastError: event.payload?.message || 'Forced logout by server'
        });
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

  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectListeners: ((isReconnecting: boolean, attempts: number) => void)[] = [];

  private scheduleReconnect(): void {
    if (this.reconnectInterval) {
      return;
    }

    // Try to restore credentials from sessionStorage if not in memory
    if (!this.connectionCredentials) {
      try {
        const storedCredentials = sessionStorage.getItem('wsConnectionCredentials');
        if (storedCredentials) {
          const parsed = JSON.parse(storedCredentials);
          this.connectionCredentials = {
            gameId: parsed.gameId,
            playerId: parsed.playerId,
            sessionToken: parsed.sessionToken,
            connectedAt: new Date(parsed.connectedAt)
          };
          console.log('Restored WebSocket credentials from sessionStorage');
        }
      } catch (e) {
        console.warn('Failed to restore WebSocket credentials from sessionStorage:', e);
      }
    }

    // Don't reconnect if we still don't have credentials
    if (!this.connectionCredentials) {
      console.log('No credentials available for reconnection');
      return;
    }

    // Check if we've exceeded max reconnect attempts
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Maximum reconnection attempts reached, stopping reconnection');
      this.updateConnectionState({
        isConnected: false,
        isReconnecting: false,
        lastError: 'Maximum reconnection attempts reached'
      });
      return;
    }

    this.updateConnectionState({ isReconnecting: true });
    
    // Notify reconnection listeners
    this.notifyReconnectListeners(true, this.reconnectAttempts);
    
    // Calculate exponential backoff with jitter
    const baseDelay = 1000;
    const maxDelay = 30000;
    const jitter = Math.random() * 1000;
    const delay = Math.min(baseDelay * Math.pow(2, this.reconnectAttempts), maxDelay) + jitter;
    
    console.log(`Attempting to reconnect in ${Math.round(delay / 1000)} seconds... (attempt ${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);
    
    this.reconnectInterval = window.setTimeout(() => {
      this.reconnectInterval = null;
      this.reconnectAttempts++;
      console.log('Attempting to reconnect...');
      const creds = this.connectionCredentials!;
      this.connect(creds.gameId, creds.playerId, creds.sessionToken)
        .then(() => {
          // Reset reconnect attempts on successful connection
          this.reconnectAttempts = 0;
          console.log('Successfully reconnected!');
          
          // Notify reconnection listeners
          this.notifyReconnectListeners(false, 0);
        })
        .catch(error => {
          console.error('Reconnection failed:', error);
          // If token is invalid, clear credentials and stop reconnecting
          if (error.message?.includes('Invalid session') || 
              error.message?.includes('Unauthorized') ||
              error.message?.includes('Session expired')) {
            console.log('Session expired or invalid, clearing credentials');
            this.connectionCredentials = null;
            this.reconnectAttempts = 0;
            this.updateConnectionState({
              isConnected: false,
              isReconnecting: false,
              lastError: 'Session expired'
            });
            
            // Notify reconnection listeners
            this.notifyReconnectListeners(false, this.reconnectAttempts);
          } else {
            // Schedule next reconnection attempt
            this.scheduleReconnect();
          }
        });
        
        // Notify reconnection listeners about failure
        this.notifyReconnectListeners(false, this.reconnectAttempts);
    }, delay);
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