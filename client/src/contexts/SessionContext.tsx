import { createContext, useContext, ReactNode } from 'react';
import { AppState, GameState, Role, PersonalKPI } from '../types';

// Centralized lobby state interface
interface PlayerLobbyInfo {
  id: string;
  name: string;
  avatar: string;
}

interface LobbyState {
  playerId?: string;
  playerInfos: PlayerLobbyInfo[];
  isHost: boolean;
  canStart: boolean;
  hostId: string;
  lobbyName: string;
  maxPlayers: number;
  connectionError: string | null;
}

interface RoleAssignment {
  role: Role;
  alignment: string;
  personalKPI: PersonalKPI | null;
}

// Define the shape of the context data
interface SessionContextType {
    appState: AppState;
    lobbyState: LobbyState;
    gameState: GameState;
    roleAssignment: RoleAssignment | null;
    isConnected: boolean;
    onLogin: (playerName: string, avatar: string) => void;
    onJoinLobby: (gameId: string, playerId: string, sessionToken: string) => void;
    onCreateGame: (gameId: string, playerId: string, sessionToken: string) => void;
    onBackToLogin: () => void;
    onStartGame: () => void;
    onLeaveLobby: () => void;
    onEnterGame: () => void;
    onViewAnalysis: () => void;
    onPlayAgain: () => void;
    onBackToResults: () => void;
}

const SessionContext = createContext<SessionContextType | undefined>(undefined);

interface SessionProviderProps {
    children: ReactNode;
    value: SessionContextType;
}

export function SessionProvider({ children, value }: SessionProviderProps) {
    return (
        <SessionContext.Provider value={value}>
            {children}
        </SessionContext.Provider>
    );
}

export function useSessionContext() {
    const context = useContext(SessionContext);
    if (context === undefined) {
        throw new Error('useSessionContext must be used within a SessionProvider');
    }
    return context;
}