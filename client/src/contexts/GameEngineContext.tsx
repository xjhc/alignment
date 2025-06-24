import { createContext, useContext, ReactNode } from 'react';
import { useGameEngine } from '../hooks/useGameEngine';

interface GameEngineContextType {
  isLoaded: boolean;
  isLoading: boolean;
  error: string | null;
  gameState: any | null;
  isReady: boolean;
  canPlayerAffordAbility: (playerId: string) => boolean;
  isValidNightActionTarget: (actorId: string, targetId: string, actionType: string) => boolean;
  canPlayerVote: (playerId: string, phaseType: string) => boolean;
  checkWinCondition: () => any;
  calculateMiningSuccess: (playerId: string, difficulty?: number) => boolean;
  getVoteWinner: (threshold?: number) => { winner: string; hasWinner: boolean };
  isGamePhaseOver: () => boolean;
  clearError: () => void;
}

const GameEngineContext = createContext<GameEngineContextType | undefined>(undefined);

interface GameEngineProviderProps {
  children: ReactNode;
}

export function GameEngineProvider({ children }: GameEngineProviderProps) {
  const gameEngineHook = useGameEngine();

  const value: GameEngineContextType = {
    isLoaded: gameEngineHook.isLoaded,
    isLoading: gameEngineHook.isLoading,
    error: gameEngineHook.error,
    gameState: gameEngineHook.gameState,
    isReady: gameEngineHook.isReady,
    canPlayerAffordAbility: gameEngineHook.canPlayerAffordAbility,
    isValidNightActionTarget: gameEngineHook.isValidNightActionTarget,
    canPlayerVote: gameEngineHook.canPlayerVote,
    checkWinCondition: gameEngineHook.checkWinCondition,
    calculateMiningSuccess: gameEngineHook.calculateMiningSuccess,
    getVoteWinner: gameEngineHook.getVoteWinner,
    isGamePhaseOver: gameEngineHook.isGamePhaseOver,
    clearError: gameEngineHook.clearError,
  };

  return (
    <GameEngineContext.Provider value={value}>
      {children}
    </GameEngineContext.Provider>
  );
}

export function useGameEngineContext() {
  const context = useContext(GameEngineContext);
  if (context === undefined) {
    throw new Error('useGameEngineContext must be used within a GameEngineProvider');
  }
  return context;
}