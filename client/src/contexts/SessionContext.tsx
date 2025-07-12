import { createContext, useContext, ReactNode } from 'react';
import { AppState, GameState } from '../types';
import { SessionState, LobbyState, RoleAssignment, GameUIState } from '../state/appReducer';

// Define the shape of the context data
export interface SessionContextType {
    appState: AppState;
    sessionState: SessionState;
    lobbyState: LobbyState;
    gameState: GameState;
    roleAssignment: RoleAssignment | null;
    gameAnalysis: any | null;
    gameUIState: GameUIState;
    isConnected: boolean;
    dispatch: any;
    gameEngineLoading: boolean;
    gameEngineError: string | null;
    onLogin: (playerName: string, avatar: string) => void;
    onJoinLobby: (gameId: string, playerId: string, sessionToken: string, lobbyName?: string) => void;
    onCreateGame: (gameId: string, playerId: string, sessionToken: string, lobbyName?: string) => void;
    onSpectateGame: (gameId: string, playerId: string, sessionToken: string, lobbyName?: string) => void;
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