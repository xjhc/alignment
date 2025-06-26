import { useReducer, useEffect, useCallback } from 'react';
import { BrowserRouter, useLocation } from 'react-router-dom';
import { useWebSocketContext } from './contexts/WebSocketContext';
import { useGameEngineContext } from './contexts/GameEngineContext';
import { useAppNavigation } from './hooks/useAppNavigation';
import { GuardedAppRouter } from './components/GuardedAppRouter';
import { SessionProvider } from './contexts/SessionContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import { GameProvider } from './contexts/GameContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { GameEngineProvider } from './contexts/GameEngineContext';
import { convertToClientTypes } from './utils/coreTypes';
import { appReducer, initialAppState, type RoleAssignment, type PlayerLobbyInfo } from './state/appReducer';


function AppContent() {
  const location = useLocation(); // Get location object for route-aware effects
  const [state, dispatch] = useReducer(appReducer, initialAppState);

  const {
    navigateToLogin,
    navigateToLobbyList,
    navigateToWaiting,
    navigateToRoleReveal,
    navigateToGame,
    navigateToGameOver,
    navigateToAnalysis
  } = useAppNavigation();

  const { connect, disconnect, subscribe, sendAction, isConnected } = useWebSocketContext();
  const {
    isLoading: gameEngineLoading,
    error: gameEngineError,
    gameState: coreGameState
  } = useGameEngineContext();

  // The SINGLE source of truth for UI updates.
  useEffect(() => {
    if (!coreGameState || !state.appState.playerId) {
      return;
    }

    const clientState = convertToClientTypes(coreGameState);

    // Merge avatar information from lobbyState into players
    const playersWithAvatars = clientState.players.map((player: any) => {
      const lobbyInfo = state.lobbyState.playerInfos.find(info => info.id === player.id);
      return {
        ...player,
        avatar: lobbyInfo?.avatar
      };
    });

    const gameStateWithAvatars = {
      ...clientState,
      players: playersWithAvatars
    };

    const localPlayer = clientState.players.find((p: any) => p.id === state.appState.playerId);
    let roleAssignment: RoleAssignment | undefined;
    
    if (localPlayer && localPlayer.role && localPlayer.alignment) {
      // PersonalKPI might be null/undefined, handle gracefully
      roleAssignment = {
        role: localPlayer.role,
        alignment: localPlayer.alignment,
        personalKPI: localPlayer.personalKPI || null
      };
    }

    dispatch({ 
      type: 'UPDATE_GAME_STATE', 
      payload: { 
        gameState: gameStateWithAvatars,
        roleAssignment
      } 
    });

    // Check for game over condition
    if (clientState.winCondition) {
      console.log(`[App] Game over condition met. Winner: ${clientState.winCondition.winner}. Transitioning.`);
      dispatch({ type: 'GAME_OVER', payload: { sessionState: 'POST_GAME' } });
      navigateToGameOver();
    }
  }, [coreGameState, state.appState.playerId, state.lobbyState.playerInfos, navigateToGameOver]);

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

    dispatch({ type: 'UPDATE_LOBBY_STATE', payload });
  }, []);


  const handleSystemMessage = useCallback((event: any) => {
    const payload = event.payload as { message: string, error?: boolean };
    if (payload.error) {
      dispatch({ type: 'SET_CONNECTION_ERROR', payload: { message: payload.message } });
    }
  }, []);

  const handleClientIdentified = useCallback((event: any) => {
    const payload = event.payload as { your_player_id: string };
    const playerId = payload.your_player_id;

    dispatch({ type: 'CLIENT_IDENTIFIED', payload: { playerId } });
  }, []);

  const handleCountdownStart = useCallback((event: any) => {
    const payload = event.payload as { duration: number };
    dispatch({ type: 'COUNTDOWN_START', payload: { duration: payload.duration } });
  }, []);

  const handleCountdownUpdate = useCallback((event: any) => {
    const payload = event.payload as { remaining: number };
    dispatch({ type: 'COUNTDOWN_UPDATE', payload: { remaining: payload.remaining } });
  }, []);

  const handleCountdownCancel = useCallback(() => {
    dispatch({ type: 'COUNTDOWN_CANCEL' });
  }, []);

  const handleChatHistorySnapshot = useCallback((event: any) => {
    const payload = event.payload as { chat_messages: any[] };
    console.log('Received chat history snapshot with', payload.chat_messages?.length || 0, 'messages');
    dispatch({ type: 'LOAD_CHAT_HISTORY', payload: { chatMessages: payload.chat_messages || [] } });
  }, []);

  // Listen for our new private event to get the player ID
  useEffect(() => {
    if (state.isInGameSession) {
      const unsubscribe = subscribe('CLIENT_IDENTIFIED', handleClientIdentified);
      return unsubscribe;
    }
    return () => { };
  }, [state.isInGameSession, subscribe, handleClientIdentified]);

  // Listen for chat history snapshots during reconnection
  useEffect(() => {
    if (state.isInGameSession) {
      const unsubscribe = subscribe('CHAT_HISTORY_SNAPSHOT', handleChatHistorySnapshot);
      return unsubscribe;
    }
    return () => { };
  }, [state.isInGameSession, subscribe, handleChatHistorySnapshot]);

  // Centralized lobby event management
  useEffect(() => {
    // Use location.pathname from the hook, not window.location
    if (location.pathname === '/waiting') {
      const unsubscribers = [
        subscribe('LOBBY_STATE_UPDATE', handleLobbyStateUpdate),
        subscribe('SYSTEM_MESSAGE', handleSystemMessage),
        subscribe('GAME_START_COUNTDOWN_INITIATED', handleCountdownStart),
        subscribe('GAME_START_COUNTDOWN_UPDATE', handleCountdownUpdate),
        subscribe('GAME_START_COUNTDOWN_CANCELLED', handleCountdownCancel),
      ];

      // Reset lobby state when entering waiting screen (preserve playerId)
      dispatch({ type: 'RESET_LOBBY_STATE' });

      return () => {
        unsubscribers.forEach(unsub => unsub());
      };
    }
    return () => { };
    // FIX: Add location.pathname to the dependency array
  }, [location.pathname, subscribe, handleLobbyStateUpdate, handleSystemMessage, handleCountdownStart, handleCountdownUpdate, handleCountdownCancel]);


  // Centralized WebSocket connection logic based on session state
  useEffect(() => {
    if (state.isInGameSession && state.appState.gameId && state.appState.playerId && state.appState.sessionToken) {
      connect(state.appState.gameId, state.appState.playerId, state.appState.sessionToken)
        .catch(error => {
          console.error('Failed to connect to WebSocket:', error);
        });

      // The cleanup function will now only be called when isInGameSession becomes false
      return () => {
        disconnect();
      };
    }
  }, [
    state.isInGameSession, // The primary trigger
    state.appState.gameId,
    state.appState.playerId,
    state.appState.sessionToken,
    connect,
    disconnect
  ]);

  const handleLogin = (playerName: string, avatar: string) => {
    dispatch({ 
      type: 'LOGIN', 
      payload: { 
        playerName, 
        playerAvatar: avatar 
      } 
    });
    navigateToLobbyList();
  };

  const handleJoinLobby = (gameId: string, playerId: string, sessionToken: string) => {
    dispatch({ 
      type: 'JOIN_LOBBY', 
      payload: { gameId, playerId, sessionToken } 
    });
    navigateToWaiting();
  };

  const handleCreateGame = (gameId: string, playerId: string, sessionToken: string) => {
    dispatch({ 
      type: 'CREATE_GAME', 
      payload: { gameId, playerId, sessionToken } 
    });
    navigateToWaiting();
  };


  const handleEnterGame = () => {
    dispatch({ type: 'ENTER_GAME' });
    navigateToGame();
  };

  // Lobby action handlers
  const handleStartGameAction = useCallback(() => {
    if (state.lobbyState.isHost && state.lobbyState.canStart && isConnected && state.appState.gameId) {
      try {
        // Simply send the action. The server will handle the connection handoff.
        sendAction({
          type: 'START_GAME' as any,
          payload: {
            game_id: state.appState.gameId
          }
        });
      } catch (error) {
        console.error('Failed to start game:', error);
        dispatch({ type: 'SET_CONNECTION_ERROR', payload: { message: 'Failed to start game' } });
      }
    }
  }, [state.lobbyState.isHost, state.lobbyState.canStart, isConnected, state.appState.gameId, sendAction]);

  const handleLeaveLobby = useCallback(() => {
    // Explicitly disconnect first
    disconnect();

    // Then update state which will prevent a reconnect attempt
    dispatch({ type: 'LEAVE_LOBBY' });
    navigateToLobbyList();
  }, [disconnect]);

  const handleBackToLogin = () => {
    dispatch({ type: 'BACK_TO_LOGIN' });
    navigateToLogin();
  };

  const handlePlayAgain = () => {
    // End the current session
    disconnect();

    // Reset state and go back to the lobby list
    dispatch({ type: 'PLAY_AGAIN' });
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
      <div className="launch-screen screen-transition animation-fade-in">
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
    appState: state.appState,
    sessionState: state.sessionState,
    lobbyState: state.lobbyState,
    gameState: state.gameState,
    roleAssignment: state.roleAssignment,
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
      <GameProvider gameState={state.gameState} localPlayerId={state.appState.playerId || ''}>
        <GuardedAppRouter />
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