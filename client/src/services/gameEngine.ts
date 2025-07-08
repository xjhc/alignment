import { wasmLoader, AlignmentCore } from './wasmLoader';
import { GeneratedEvent, GeneratedAction } from '../types/generated';
import { GameState } from '../types';

export class GameEngine {
  private core: AlignmentCore | null = null;
  private stateChangeListeners: ((state: GameState) => void)[] = [];

  async initialize(): Promise<void> {
    try {
      await wasmLoader.load();
      this.core = wasmLoader.getCore();

      wasmLoader.onStateChange((stateJson: string) => {
        try {
          const state: GameState = JSON.parse(stateJson);
          
          // Validate the state before notifying - only notify if we have essential data
          if (state && state.id && state.players && 
              (Array.isArray(state.players) || typeof state.players === 'object')) {
            this.notifyStateChange(state);
          } else {
            console.warn('WASM sent invalid or incomplete state, ignoring notification');
          }
        } catch (error) {
          console.error('Failed to parse game state from WASM:', error);
        }
      });

      console.log('Game engine initialized successfully');
    } catch (error) {
      console.error('Failed to initialize game engine:', error);
      throw error;
    }
  }

  resetAndLoadState(state: GameState): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.core) {
        reject(new Error('Game engine not initialized'));
        return;
      }

      try {
        // Step 1: Re-create the game state in WASM to discard any old/stale state.
        const createResult = this.core.createGame(state.id);
        if (!createResult.success) {
          throw new Error(createResult.error || `Failed to re-initialize game state for ${state.id}`);
        }

        // Step 2: Load the snapshot into the now-pristine state object.
        const stateJson = JSON.stringify(state);
        const loadResult = this.core.deserializeGameState(stateJson);
        if (loadResult.success) {
          // Get the actual current state after loading to verify it worked
          const currentState = this.getCurrentState();
          
          // Use the actual current state instead of the input state for notification
          // to ensure we're sending what was actually loaded into WASM
          if (currentState) {
            this.notifyStateChange(currentState);
          } else {
            console.warn('getCurrentState returned null after successful load');
            this.notifyStateChange(state);
          }
          resolve();
        } else {
          reject(new Error(loadResult.error || 'Failed to load state from snapshot'));
        }
      } catch (error) {
        reject(error);
      }
    });
  }

  createGame(gameId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.core) {
        reject(new Error('Game engine not initialized'));
        return;
      }

      const result = this.core.createGame(gameId);
      if (result.success) {
        // After creating, immediately get the new state and notify listeners.
        const newState = this.getCurrentState();
        if (newState) {
          this.notifyStateChange(newState);
        }
        resolve();
      } else {
        reject(new Error(result.error || 'Failed to create game'));
      }
    });
  }

  applyEvent(event: GeneratedEvent): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.core) {
        reject(new Error('Game engine not initialized'));
        return;
      }

      try {
        const eventJson = JSON.stringify(event);
        const result = this.core.applyEvent(eventJson);

        if (result.success) {
          resolve();
        } else {
          reject(new Error(result.error || 'Failed to apply event'));
        }
      } catch (error) {
        reject(error);
      }
    });
  }

  getCurrentState(): GameState | null {
    if (!this.core) {
      return null;
    }

    try {
      const result = this.core.getGameState();

      if (typeof result === 'object' && result !== null && 'error' in result) {
        return null;
      }

      if (typeof result === 'string') {
        return JSON.parse(result);
      }

      console.warn('Unexpected result type from getGameState:', typeof result, result);
      return null;
    } catch (error) {
      console.error('Failed to get current state:', error);
      return null;
    }
  }

  loadState(state: GameState): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.core) {
        reject(new Error('Game engine not initialized'));
        return;
      }

      try {
        const stateJson = JSON.stringify(state);
        const result = this.core.deserializeGameState(stateJson);

        if (result.success) {
          resolve();
        } else {
          reject(new Error(result.error || 'Failed to load state'));
        }
      } catch (error) {
        reject(error);
      }
    });
  }

  canPlayerVote(playerId: string, phaseType: string): boolean {
    if (!this.core) {
      return false;
    }
    return this.core.canPlayerVote(playerId, phaseType);
  }

  canPlayerAffordAbility(playerId: string): boolean {
    if (!this.core) {
      return false;
    }
    return this.core.canPlayerAffordAbility(playerId);
  }

  checkWinCondition(): any | null {
    if (!this.core) {
      return null;
    }

    const winJson = this.core.checkWinCondition();
    if (winJson) {
      try {
        return JSON.parse(winJson);
      } catch (error) {
        console.error('Failed to parse win condition:', error);
        return null;
      }
    }
    return null;
  }

  calculateMiningSuccess(playerId: string, difficulty: number = 0.3): boolean {
    if (!this.core) {
      return false;
    }
    return this.core.calculateMiningSuccess(playerId, difficulty);
  }

  isValidNightActionTarget(actorId: string, targetId: string, actionType: string): boolean {
    if (!this.core) {
      return false;
    }
    return this.core.isValidNightActionTarget(actorId, targetId, actionType);
  }

  getVoteWinner(threshold: number = 0.5): { winner: string; hasWinner: boolean } {
    if (!this.core) {
      return { winner: '', hasWinner: false };
    }
    return this.core.getVoteWinner(threshold);
  }

  isGamePhaseOver(): boolean {
    if (!this.core) {
      return false;
    }
    return this.core.isGamePhaseOver();
  }

  submitPlayerAction(action: GeneratedAction): Promise<GeneratedEvent[]> {
    return new Promise((resolve, reject) => {
      const events = this.actionToEvents(action);

      Promise.all(events.map(event => this.applyEvent(event)))
        .then(() => resolve(events))
        .catch(reject);
    });
  }

  private actionToEvents(action: GeneratedAction): GeneratedEvent[] {
    const baseEvent: Partial<GeneratedEvent> = {
      id: `${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
      gameId: action.gameId,
      playerId: action.playerId,
      timestamp: new Date().toISOString(),
      payload: action.payload,
    };

    switch (action.type) {
      case 'SUBMIT_VOTE':
        return [{
          ...baseEvent,
          type: 'VOTE_CAST',
        } as GeneratedEvent];

      case 'SEND_MESSAGE':
        return [{
          ...baseEvent,
          type: 'CHAT_MESSAGE',
        } as GeneratedEvent];

      case 'MINE_TOKENS':
        return [{
          ...baseEvent,
          type: 'MINING_ATTEMPTED',
        } as GeneratedEvent];

      case 'USE_ABILITY':
        return [{
          ...baseEvent,
          type: 'NIGHT_ACTION_SUBMITTED',
        } as GeneratedEvent];

      default:
        console.warn('Unknown action type:', action.type);
        return [];
    }
  }

  onStateChange(callback: (state: GameState) => void): () => void {
    this.stateChangeListeners.push(callback);

    return () => {
      const index = this.stateChangeListeners.indexOf(callback);
      if (index > -1) {
        this.stateChangeListeners.splice(index, 1);
      }
    };
  }

  private notifyStateChange(state: GameState): void {
    this.stateChangeListeners.forEach(callback => {
      try {
        callback(state);
      } catch (error) {
        console.error('Error in state change listener:', error);
      }
    });
  }

  isReady(): boolean {
    return wasmLoader.isReady() && this.core !== null;
  }
}

// Singleton instance
export const gameEngine = new GameEngine();