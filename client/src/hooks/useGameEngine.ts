import { useState, useEffect, useCallback } from 'react';
import { gameEngine } from '../services/gameEngine';
import { CoreGameState, CoreEvent, CoreAction } from '../utils/coreTypes';

export interface GameEngineState {
  isLoaded: boolean;
  isLoading: boolean;
  error: string | null;
  gameState: CoreGameState | null;
}

export function useGameEngine() {
  const [state, setState] = useState<GameEngineState>({
    isLoaded: false,
    isLoading: false,
    error: null,
    gameState: null,
  });

  // Initialize the game engine
  useEffect(() => {
    let mounted = true;

    const initializeEngine = async () => {
      if (state.isLoading || state.isLoaded) {
        return;
      }

      setState(prev => ({ ...prev, isLoading: true, error: null }));

      try {
        await gameEngine.initialize();
        
        if (mounted) {
          setState(prev => ({ 
            ...prev, 
            isLoaded: true, 
            isLoading: false,
            gameState: gameEngine.getCurrentState()
          }));
        }
      } catch (error) {
        if (mounted) {
          setState(prev => ({ 
            ...prev, 
            isLoading: false, 
            error: error instanceof Error ? error.message : 'Failed to initialize game engine'
          }));
        }
      }
    };

    initializeEngine();

    return () => {
      mounted = false;
    };
  }, []);

  // Listen for state changes from the game engine
  useEffect(() => {
    if (!state.isLoaded) {
      return;
    }

    const unsubscribe = gameEngine.onStateChange((newState) => {
      setState(prev => ({ ...prev, gameState: newState }));
    });

    return unsubscribe;
  }, [state.isLoaded]);

  // Game engine methods
  const createGame = useCallback(async (gameId: string) => {
    if (!state.isLoaded) {
      throw new Error('Game engine not loaded');
    }

    try {
      await gameEngine.createGame(gameId);
      setState(prev => ({ ...prev, gameState: gameEngine.getCurrentState() }));
    } catch (error) {
      setState(prev => ({ 
        ...prev, 
        error: error instanceof Error ? error.message : 'Failed to create game'
      }));
      throw error;
    }
  }, [state.isLoaded]);

  const applyEvent = useCallback(async (event: CoreEvent) => {
    if (!state.isLoaded) {
      throw new Error('Game engine not loaded');
    }

    try {
      await gameEngine.applyEvent(event);
      // State will be updated automatically via the state change listener
    } catch (error) {
      setState(prev => ({ 
        ...prev, 
        error: error instanceof Error ? error.message : 'Failed to apply event'
      }));
      throw error;
    }
  }, [state.isLoaded]);

  const submitAction = useCallback(async (action: CoreAction) => {
    if (!state.isLoaded) {
      throw new Error('Game engine not loaded');
    }

    try {
      const events = await gameEngine.submitPlayerAction(action);
      return events;
    } catch (error) {
      setState(prev => ({ 
        ...prev, 
        error: error instanceof Error ? error.message : 'Failed to submit action'
      }));
      throw error;
    }
  }, [state.isLoaded]);

  const loadGameState = useCallback(async (gameState: CoreGameState) => {
    if (!state.isLoaded) {
      throw new Error('Game engine not loaded');
    }

    try {
      await gameEngine.loadState(gameState);
      setState(prev => ({ ...prev, gameState: gameEngine.getCurrentState() }));
    } catch (error) {
      setState(prev => ({ 
        ...prev, 
        error: error instanceof Error ? error.message : 'Failed to load game state'
      }));
      throw error;
    }
  }, [state.isLoaded]);

  // Game rule methods
  const canPlayerVote = useCallback((playerId: string, phaseType: string): boolean => {
    if (!state.isLoaded) {
      return false;
    }
    return gameEngine.canPlayerVote(playerId, phaseType);
  }, [state.isLoaded]);

  const canPlayerAffordAbility = useCallback((playerId: string): boolean => {
    if (!state.isLoaded) {
      return false;
    }
    return gameEngine.canPlayerAffordAbility(playerId);
  }, [state.isLoaded]);

  const checkWinCondition = useCallback(() => {
    if (!state.isLoaded) {
      return null;
    }
    return gameEngine.checkWinCondition();
  }, [state.isLoaded]);

  const calculateMiningSuccess = useCallback((playerId: string, difficulty?: number): boolean => {
    if (!state.isLoaded) {
      return false;
    }
    return gameEngine.calculateMiningSuccess(playerId, difficulty);
  }, [state.isLoaded]);

  const isValidNightActionTarget = useCallback((actorId: string, targetId: string, actionType: string): boolean => {
    if (!state.isLoaded) {
      return false;
    }
    return gameEngine.isValidNightActionTarget(actorId, targetId, actionType);
  }, [state.isLoaded]);

  const getVoteWinner = useCallback((threshold?: number) => {
    if (!state.isLoaded) {
      return { winner: '', hasWinner: false };
    }
    return gameEngine.getVoteWinner(threshold);
  }, [state.isLoaded]);

  const isGamePhaseOver = useCallback((): boolean => {
    if (!state.isLoaded) {
      return false;
    }
    return gameEngine.isGamePhaseOver();
  }, [state.isLoaded]);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  return {
    // State
    isLoaded: state.isLoaded,
    isLoading: state.isLoading,
    error: state.error,
    gameState: state.gameState,

    // Actions
    createGame,
    applyEvent,
    submitAction,
    loadGameState,
    clearError,

    // Game rules
    canPlayerVote,
    canPlayerAffordAbility,
    checkWinCondition,
    calculateMiningSuccess,
    isValidNightActionTarget,
    getVoteWinner,
    isGamePhaseOver,

    // Helper
    isReady: state.isLoaded && gameEngine.isReady(),
  };
}