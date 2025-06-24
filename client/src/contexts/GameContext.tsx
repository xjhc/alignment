import { createContext, useContext, ReactNode } from 'react';
import { GameState, Player } from '../types';

interface GameContextType {
  gameState: GameState;
  localPlayerId: string;
  localPlayer: Player | null;
  isConnected: boolean;
}

const GameContext = createContext<GameContextType | undefined>(undefined);

interface GameProviderProps {
  children: ReactNode;
  gameState: GameState;
  localPlayerId: string;
  isConnected: boolean;
}

export function GameProvider({ children, gameState, localPlayerId, isConnected }: GameProviderProps) {
  const localPlayer = gameState.players.find(p => p.id === localPlayerId) || null;

  const value: GameContextType = {
    gameState,
    localPlayerId,
    localPlayer,
    isConnected,
  };

  return (
    <GameContext.Provider value={value}>
      {children}
    </GameContext.Provider>
  );
}

export function useGameContext() {
  const context = useContext(GameContext);
  if (context === undefined) {
    throw new Error('useGameContext must be used within a GameProvider');
  }
  return context;
}