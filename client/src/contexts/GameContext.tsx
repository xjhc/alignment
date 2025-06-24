import { createContext, useContext, ReactNode, useState, useEffect } from 'react';
import { GameState, Player, ClientAction } from '../types';
import { useWebSocketContext } from './WebSocketContext';

interface GameContextType {
  gameState: GameState;
  localPlayerId: string;
  viewedPlayerId: string;
  localPlayer: Player | null;
  viewedPlayer: Player | null;
  isConnected: boolean;
  sendAction: (action: ClientAction) => void;
  setViewedPlayer: (playerId: string) => void;
}

const GameContext = createContext<GameContextType | undefined>(undefined);

interface GameProviderProps {
  children: ReactNode;
  gameState: GameState;
  localPlayerId: string;
}

export function GameProvider({ children, gameState, localPlayerId }: GameProviderProps) {
  const [viewedPlayerId, setViewedPlayerId] = useState(localPlayerId);
  const { isConnected, sendAction } = useWebSocketContext();
  const localPlayer = gameState.players.find(p => p.id === localPlayerId) || null;
  const viewedPlayer = gameState.players.find(p => p.id === viewedPlayerId) || localPlayer;

  useEffect(() => {
    if (localPlayerId) setViewedPlayerId(localPlayerId);
  }, [localPlayerId]);

  const value: GameContextType = {
    gameState,
    localPlayerId,
    viewedPlayerId,
    localPlayer,
    viewedPlayer,
    isConnected,
    sendAction,
    setViewedPlayer: setViewedPlayerId,
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