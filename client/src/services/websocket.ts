import { ClientAction, ServerEvent, ConnectionState, CoreEvent } from '../types';
import { gameEngine } from './gameEngine';

export class WebSocketClient {
  private socket: WebSocket | null = null;
  private url: string;
  private currentGameId: string | null = null;
  private eventHandlers: Map<string, ((event: ServerEvent) => void)[]> = new Map();
  private connectionStateHandlers: ((state: ConnectionState) => void)[] = [];
  private connectionState: ConnectionState = { isConnected: false, isReconnecting: false };
  private reconnectInterval: number | null = null;
  private heartbeatInterval: number | null = null;

  constructor(url: string = 'ws://localhost:8080/ws') {
    this.url = url;
  }

  connect(gameId?: string, playerId?: string, sessionToken?: string, lastEventId?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // Store the game ID for later use
        this.currentGameId = gameId || null;
        
        // Build WebSocket URL with query parameters for session-based authentication
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

          // Send RECONNECT if we have a lastEventId (explicit reconnect) 
          // or if we have stored events for this game (auto-reconnect)
          if (gameId && playerId && sessionToken) {
            const storedEventId = this.getStoredEventId(gameId);
            if (lastEventId || storedEventId) {
              const eventId = lastEventId || storedEventId || '';
              this.sendAction({
                type: 'RECONNECT',
                payload: {
                  game_id: gameId,
                  player_id: playerId,
                  session_token: sessionToken,
                  last_event_id: eventId
                }
              });
            }
            // If no stored events, this is likely a fresh lobby connection
            // Let the server handle it through normal lobby flow
          }

          resolve();
        };

        this.socket.onmessage = (event) => {
          try {
            const messageData = event.data as string;
            // Split by newline to handle batched messages
            const messages = messageData.split('\n');

            for (const messageStr of messages) {
              if (messageStr.trim() === '') continue; // Skip empty strings

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

          // Attempt to reconnect if not a clean close
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

  // Event subscription methods
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

    // Convert WebSocket event to Core event and apply to game engine
    this.applyEventToGameEngine(event);

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

    // Store event ID after successful processing
    this.updateLastEventId(event);
  }

  private applyEventToGameEngine(serverEvent: ServerEvent): void {
    // Only apply game-changing events to the engine
    if (!this.shouldApplyToGameEngine(serverEvent.type)) {
      console.log('Skipping engine application for event type:', serverEvent.type);
      return;
    }

    try {
      // Convert server event to core event format
      const coreEvent: CoreEvent = {
        id: `ws-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: this.mapEventType(serverEvent.type),
        gameId: this.currentGameId || serverEvent.payload?.game_id || 'unknown',
        playerId: serverEvent.payload?.player_id || '',
        timestamp: new Date().toISOString(),
        payload: serverEvent.payload || {},
      };

      console.log('Applying event to game engine:', coreEvent.type, 'for game:', coreEvent.gameId);

      // Apply to game engine, creating game if needed
      if (gameEngine.isReady()) {
        // Try to apply the event
        gameEngine.applyEvent(coreEvent).catch(async (error) => {
          console.log('Game engine error:', error.message);
          // If no game state initialized, try to create one
          if (error.message?.includes('No game state initialized')) {
            try {
              console.log('Auto-creating game state for:', coreEvent.gameId);
              await gameEngine.createGame(coreEvent.gameId);
              console.log('Game created successfully, retrying event application');
              // Retry applying the event
              await gameEngine.applyEvent(coreEvent);
              console.log('Event applied successfully after game creation');
            } catch (createError) {
              console.error('Failed to create game and apply event:', createError);
            }
          } else {
            console.error('Failed to apply event to game engine:', error);
          }
        });
      } else {
        console.log('Game engine not ready, skipping event application');
      }
    } catch (error) {
      console.error('Error converting server event to core event:', error);
    }
  }

  private shouldApplyToGameEngine(eventType: string): boolean {
    // Only apply events that affect game state
    const gameStateEvents = [
      'PLAYER_JOINED',
      'PLAYER_LEFT', 
      'PLAYER_DEACTIVATED',
      'ROLE_ASSIGNED',
      'ROLES_ASSIGNED',
      'ALIGNMENT_CHANGED',
      'PHASE_CHANGED',
      'CHAT_MESSAGE_POSTED',
      'PULSE_CHECK_SUBMITTED',
      'NIGHT_ACTIONS_RESOLVED',
      'GAME_ENDED',
      'GAME_STARTED',
      'VOTE_CAST',
      'VOTE_STARTED',
      'VOTE_COMPLETED',
      'TOKENS_AWARDED',
      'MINING_SUCCESSFUL',
      'MINING_FAILED',
      'CRISIS_TRIGGERED',
    ];

    return gameStateEvents.includes(eventType);
  }

  private mapEventType(serverEventType: string): string {
    // Map WebSocket event types to Core event types
    const eventTypeMap: Record<string, string> = {
      'PLAYER_JOINED': 'PLAYER_JOINED',
      'PLAYER_LEFT': 'PLAYER_LEFT',
      'PLAYER_DEACTIVATED': 'PLAYER_ELIMINATED',
      'ROLE_ASSIGNED': 'ROLE_ASSIGNED',
      'ROLES_ASSIGNED': 'ROLE_ASSIGNED',
      'ALIGNMENT_CHANGED': 'PLAYER_ALIGNED',
      'PHASE_CHANGED': 'PHASE_CHANGED',
      'CHAT_MESSAGE_POSTED': 'CHAT_MESSAGE',
      'PULSE_CHECK_SUBMITTED': 'PULSE_CHECK_SUBMITTED',
      'NIGHT_ACTIONS_RESOLVED': 'NIGHT_ACTIONS_RESOLVED',
      'GAME_ENDED': 'GAME_ENDED',
      'GAME_STARTED': 'GAME_STARTED',
      'VOTE_CAST': 'VOTE_CAST',
      'VOTE_STARTED': 'VOTE_STARTED',
      'VOTE_COMPLETED': 'VOTE_COMPLETED',
      'TOKENS_AWARDED': 'TOKENS_AWARDED',
      'MINING_SUCCESSFUL': 'MINING_SUCCESSFUL',
      'MINING_FAILED': 'MINING_FAILED',
      'CRISIS_TRIGGERED': 'CRISIS_TRIGGERED',
    };

    return eventTypeMap[serverEventType] || serverEventType;
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
      return; // Already scheduled
    }

    this.updateConnectionState({ isReconnecting: true });

    this.reconnectInterval = window.setTimeout(() => {
      this.reconnectInterval = null;
      console.log('Attempting to reconnect...');

      this.connect()
        .catch(error => {
          console.error('Reconnection failed:', error);
          // Try again in a longer interval
          setTimeout(() => this.scheduleReconnect(), 5000);
        });
    }, 2000);
  }

  private startHeartbeat(): void {
    this.heartbeatInterval = window.setInterval(() => {
      if (this.socket && this.socket.readyState === WebSocket.OPEN) {
        this.socket.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000); // Send ping every 30 seconds
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private getStoredEventId(gameId: string): string | null {
    try {
      return localStorage.getItem(`alignment_last_event_${gameId}`);
    } catch (error) {
      console.error('Failed to get stored event ID:', error);
      return null;
    }
  }

  private storeEventId(gameId: string, eventId: string): void {
    try {
      localStorage.setItem(`alignment_last_event_${gameId}`, eventId);
    } catch (error) {
      console.error('Failed to store event ID:', error);
    }
  }

  private updateLastEventId(event: ServerEvent): void {
    // Update stored event ID after successfully processing each event
    if (event.id && event.game_id) {
      this.storeEventId(event.game_id, event.id);
    }
  }
}

// Singleton instance
export const websocketClient = new WebSocketClient();