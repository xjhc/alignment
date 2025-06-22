import { wasmLoader, AlignmentCore } from './wasmLoader';
import { CoreGameState, CoreEvent, CoreAction } from '../utils/coreTypes';

export class GameEngine {
  private core: AlignmentCore | null = null;
  private stateChangeListeners: ((state: CoreGameState) => void)[] = [];

  async initialize(): Promise<void> {
    try {
      await wasmLoader.load();
      this.core = wasmLoader.getCore();

      wasmLoader.onStateChange((stateJson: string) => {
        try {
          const state: CoreGameState = JSON.parse(stateJson);
          this.notifyStateChange(state);
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

  resetAndLoadState(state: CoreGameState): Promise<void> {
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
          // Manually trigger a state change notification since the Go side won't
          this.notifyStateChange(state);
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

  applyEvent(event: CoreEvent): Promise<void> {
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

  getCurrentState(): CoreGameState | null {
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

  loadState(state: CoreGameState): Promise<void> {
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

  submitPlayerAction(action: CoreAction): Promise<CoreEvent[]> {
    return new Promise((resolve, reject) => {
      const events = this.actionToEvents(action);

      Promise.all(events.map(event => this.applyEvent(event)))
        .then(() => resolve(events))
        .catch(reject);
    });
  }

  private actionToEvents(action: CoreAction): CoreEvent[] {
    const baseEvent: Partial<CoreEvent> = {
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
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
        } as CoreEvent];

      case 'SEND_MESSAGE':
        return [{
          ...baseEvent,
          type: 'CHAT_MESSAGE',
        } as CoreEvent];

      case 'MINE_TOKENS':
        return [{
          ...baseEvent,
          type: 'MINING_ATTEMPTED',
        } as CoreEvent];

      case 'USE_ABILITY':
        return [{
          ...baseEvent,
          type: 'NIGHT_ACTION_SUBMITTED',
        } as CoreEvent];

      default:
        console.warn('Unknown action type:', action.type);
        return [];
    }
  }

  onStateChange(callback: (state: CoreGameState) => void): () => void {
    this.stateChangeListeners.push(callback);

    return () => {
      const index = this.stateChangeListeners.indexOf(callback);
      if (index > -1) {
        this.stateChangeListeners.splice(index, 1);
      }
    };
  }

  private notifyStateChange(state: CoreGameState): void {
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