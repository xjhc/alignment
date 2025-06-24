import { useState, useEffect, useCallback } from 'react';
import { BrowserRouter, useLocation } from 'react-router-dom';
import { useWebSocketContext } from './contexts/WebSocketContext';
import { useGameEngineContext } from './contexts/GameEngineContext';
import { useAppNavigation } from './hooks/useAppNavigation';
import { AppRouter } from './components/AppRouter';
import { SessionProvider } from './contexts/SessionContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import { GameProvider } from './contexts/GameContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { GameEngineProvider } from './contexts/GameEngineContext';
import { AppState, GameState, Role, PersonalKPI } from './types';
import { convertToClientTypes } from './utils/coreTypes';

interface RoleAssignment {
  role: Role;
  alignment: string;
  personalKPI: PersonalKPI;
}

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

function AppContent() {
  const location = useLocation(); // Get location object for route-aware effects

  const [appState, setAppState] = useState<AppState>({
    playerName: '',
  });

  const {
    navigateToLogin,
    navigateToLobbyList,
    navigateToWaiting,
    navigateToRoleReveal,
    navigateToGame,
    navigateToGameOver,
    navigateToAnalysis
  } = useAppNavigation();

  const [isInGameSession, setIsInGameSession] = useState(false);

  const [gameState, setGameState] = useState<GameState>({
    id: '',
    players: [],
    phase: { type: 'LOBBY', startTime: new Date().toISOString(), duration: 0 },
    dayNumber: 1,
    chatMessages: [],
  });

  const [roleAssignment, setRoleAssignment] = useState<RoleAssignment | null>(null);

  // Centralized lobby state
  const [lobbyState, setLobbyState] = useState<LobbyState>({
    playerId: undefined,
    playerInfos: [],
    isHost: false,
    canStart: false,
    hostId: '',
    lobbyName: '',
    maxPlayers: 8,
    connectionError: null,
  });

  const { connect, disconnect, subscribe, sendAction, isConnected } = useWebSocketContext();
  const {
    isLoading: gameEngineLoading,
    error: gameEngineError,
    gameState: coreGameState
  } = useGameEngineContext();

  // The SINGLE source of truth for UI updates.
  useEffect(() => {
    if (!coreGameState || !appState.playerId) {
      return;
    }

    const clientState = convertToClientTypes(coreGameState);

    // Merge avatar information from lobbyState into players
    const playersWithAvatars = clientState.players.map((player: any) => {
      const lobbyInfo = lobbyState.playerInfos.find(info => info.id === player.id);
      return {
        ...player,
        avatar: lobbyInfo?.avatar
      };
    });

    setGameState({
      ...clientState,
      players: playersWithAvatars
    });

    const localPlayer = clientState.players.find((p: any) => p.id === appState.playerId);
    if (localPlayer && localPlayer.role && localPlayer.alignment) {
      // PersonalKPI might be null/undefined, handle gracefully
      const newAssignment = {
        role: localPlayer.role,
        alignment: localPlayer.alignment,
        personalKPI: localPlayer.personalKPI || null
      };
      setRoleAssignment(newAssignment);
    }

    // Check for game over condition
    if (clientState.winCondition) {
      console.log(`[App] Game over condition met. Winner: ${clientState.winCondition.winner}. Transitioning.`);
      navigateToGameOver();
    }
  }, [coreGameState, appState.playerId, lobbyState.playerInfos, navigateToGameOver]);

  // Wait for BOTH phase change AND role assignment before transitioning
  useEffect(() => {
    // Centralized game start event management
    // This event is the single trigger to move from Lobby to Role Reveal.
    const handleGameStarted = () => {
      console.log('[App] Game has started. Navigating to role reveal.');
      navigateToRoleReveal();
    };

    const unsubscribe = subscribe('GAME_STARTED', handleGameStarted);
    return () => unsubscribe();
  }, [subscribe, navigateToRoleReveal]);

  // Set dark theme by default
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', 'dark');
  }, []);

  // Stabilized event handlers using useCallback
  const handleLobbyStateUpdate = useCallback((event: any) => {
    const payload = event.payload as {
      players: PlayerLobbyInfo[],
      host_id: string,
      can_start: boolean,
      lobby_id: string,
      name: string,
      max_players: number
    };

    setLobbyState(prev => ({
      ...prev,
      playerInfos: payload.players,
      hostId: payload.host_id,
      canStart: payload.can_start,
      lobbyName: payload.name,
      maxPlayers: payload.max_players,
      isHost: prev.playerId === payload.host_id,
      connectionError: null,
    }));
  }, [appState.playerId]); // FIX: Add dependency to prevent stale closure for `isHost` logic


  const handleSystemMessage = useCallback((event: any) => {
    const payload = event.payload as { message: string, error?: boolean };
    if (payload.error) {
      setLobbyState(prev => ({ ...prev, connectionError: payload.message }));
    }
  }, []);

  const handleClientIdentified = useCallback((event: any) => {
    const payload = event.payload as { your_player_id: string };
    const playerId = payload.your_player_id;

    // Update both appState and lobbyState synchronously
    setAppState(prev => ({ ...prev, playerId }));
    setLobbyState(prev => ({ ...prev, playerId }));
  }, []);

  // Listen for our new private event to get the player ID
  useEffect(() => {
    if (isInGameSession) {
      const unsubscribe = subscribe('CLIENT_IDENTIFIED', handleClientIdentified);
      return unsubscribe;
    }
    return () => { };
  }, [isInGameSession, subscribe, handleClientIdentified]);

  // Centralized lobby event management
  useEffect(() => {
    // Use location.pathname from the hook, not window.location
    if (location.pathname === '/waiting') {
      const unsubscribers = [
        subscribe('LOBBY_STATE_UPDATE', handleLobbyStateUpdate),
        subscribe('SYSTEM_MESSAGE', handleSystemMessage),
      ];

      // Reset lobby state when entering waiting screen (preserve playerId)
      setLobbyState(prev => ({
        ...prev,
        playerInfos: [],
        isHost: false,
        canStart: false,
        hostId: '',
        lobbyName: '',
        connectionError: null,
      }));

      return () => {
        unsubscribers.forEach(unsub => unsub());
      };
    }
    return () => { };
    // FIX: Add location.pathname to the dependency array
  }, [location.pathname, subscribe, handleLobbyStateUpdate, handleSystemMessage]);


  // Centralized WebSocket connection logic based on session state
  useEffect(() => {
    if (isInGameSession && appState.gameId && appState.playerId && appState.sessionToken) {
      connect(appState.gameId, appState.playerId, appState.sessionToken)
        .catch(error => {
          console.error('Failed to connect to WebSocket:', error);
        });

      // The cleanup function will now only be called when isInGameSession becomes false
      return () => {
        disconnect();
      };
    }
  }, [
    isInGameSession, // The primary trigger
    appState.gameId,
    appState.playerId,
    appState.sessionToken,
    connect,
    disconnect
  ]);

  const handleLogin = (playerName: string, avatar: string) => {
    setAppState(prev => ({
      ...prev,
      playerName,
      playerAvatar: avatar
    }));
    setIsInGameSession(false); // Ensure session is not active
    navigateToLobbyList();
  };

  const handleJoinLobby = (gameId: string, playerId: string, sessionToken: string) => {
    setAppState(prev => ({
      ...prev,
      gameId,
      playerId,
      sessionToken
    }));
    setLobbyState(prev => ({ ...prev, playerId }));
    setIsInGameSession(true); // START the session
    navigateToWaiting();
  };

  const handleCreateGame = (gameId: string, playerId: string, sessionToken: string) => {
    setAppState(prev => ({
      ...prev,
      gameId,
      playerId,
      sessionToken
    }));
    setLobbyState(prev => ({ ...prev, playerId }));
    setIsInGameSession(true); // START the session
    navigateToWaiting();
  };


  const handleEnterGame = () => {
    navigateToGame();
  };

  // Lobby action handlers
  const handleStartGameAction = useCallback(() => {
    if (lobbyState.isHost && lobbyState.canStart && isConnected && appState.gameId) {
      try {
        // Simply send the action. The server will handle the connection handoff.
        sendAction({
          type: 'START_GAME',
          payload: {
            game_id: appState.gameId
          }
        });
      } catch (error) {
        console.error('Failed to start game:', error);
        setLobbyState(prev => ({ ...prev, connectionError: 'Failed to start game' }));
      }
    }
  }, [lobbyState.isHost, lobbyState.canStart, isConnected, appState.gameId, sendAction]);

  const handleLeaveLobby = useCallback(() => {
    // Explicitly disconnect first
    disconnect();

    // Then update state which will prevent a reconnect attempt
    setIsInGameSession(false);

    setAppState(prev => ({
      ...prev,
      gameId: undefined,
      playerId: undefined,
      joinToken: undefined,
      sessionToken: undefined
    }));
    // Reset lobby state completely when leaving
    setLobbyState({
      playerId: undefined,
      playerInfos: [],
      isHost: false,
      canStart: false,
      hostId: '',
      lobbyName: '',
      maxPlayers: 8,
      connectionError: null,
    });
    navigateToLobbyList();
  }, [disconnect]);

  const handleBackToLogin = () => {
    setIsInGameSession(false); // END the session
    setAppState({
      playerName: '',
    });
    navigateToLogin();
  };

  const handlePlayAgain = () => {
    // End the current session
    setIsInGameSession(false);
    disconnect();

    // Reset state and go back to the lobby list
    setAppState(prev => ({
      ...prev,
      gameId: undefined,
      sessionToken: undefined,
    }));
    setRoleAssignment(null);
    navigateToLobbyList();
  };

  const handleViewAnalysis = () => {
    navigateToAnalysis();
  };

  const handleBackToResults = () => {
    navigateToGameOver();
  };

  // Show loading screen while game engine is loading
  if (gameEngineLoading) {
    return (
      <div className="launch-screen screen-transition animate-fade-in">
        <div className="launch-form">
          <h2>Loading game engine...</h2>
          <div className="loading-spinner large"></div>
        </div>
      </div>
    );
  }

  // Show error screen if game engine failed to load
  if (gameEngineError) {
    return (
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Game Engine Error</h2>
          <p style={{ color: 'var(--color-danger)' }}>{gameEngineError}</p>
          <button
            className="btn-primary"
            onClick={() => window.location.reload()}
          >
            Reload Page
          </button>
        </div>
      </div>
    );
  }

  const sessionContextValue = {
    appState,
    lobbyState,
    gameState,
    roleAssignment,
    isConnected,
    onLogin: handleLogin,
    onJoinLobby: handleJoinLobby,
    onCreateGame: handleCreateGame,
    onBackToLogin: handleBackToLogin,
    onStartGame: handleStartGameAction,
    onLeaveLobby: handleLeaveLobby,
    onEnterGame: handleEnterGame,
    onViewAnalysis: handleViewAnalysis,
    onPlayAgain: handlePlayAgain,
    onBackToResults: handleBackToResults,
  };
  return (
    <SessionProvider value={sessionContextValue}>
      <GameProvider gameState={gameState} localPlayerId={appState.playerId || ''}>
        <AppRouter />
      </GameProvider>
    </SessionProvider>
  );
}

function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <GameEngineProvider>
          <WebSocketProvider>
            <AppContent />
          </WebSocketProvider>
        </GameEngineProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}

export default App;