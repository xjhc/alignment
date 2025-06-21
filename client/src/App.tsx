import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { useGameEngine } from './hooks/useGameEngine';
import { LoginScreen } from './components/LoginScreen';
import { LobbyListScreen } from './components/LobbyListScreen';
import { WaitingScreen } from './components/WaitingScreen';
import { RoleRevealScreen } from './components/RoleRevealScreen';
import { GameScreen } from './components/GameScreen';
import { WasmTestScreen } from './components/WasmTestScreen';
import { AppState, GameState } from './types';
import { convertToClientTypes } from './utils/coreTypes';

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

function App() {
  const [appState, setAppState] = useState<AppState>({
    currentScreen: 'login',
    playerName: '',
  });

  const [gameState, setGameState] = useState<GameState>({
    id: '',
    players: [],
    phase: { type: 'LOBBY', startTime: new Date().toISOString(), duration: 0 },
    dayNumber: 1,
    chatMessages: [],
  });

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

  const { connect, disconnect, subscribe, sendAction, isConnected } = useWebSocket();
  const {
    isLoading: gameEngineLoading,
    error: gameEngineError,
    gameState: coreGameState
  } = useGameEngine();

  // Set dark theme by default
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', 'dark');
  }, []);

  // Sync core game state with React state
  useEffect(() => {
    if (coreGameState) {
      const clientState = convertToClientTypes(coreGameState);
      setGameState(clientState);
    }
  }, [coreGameState]);

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
      // FIXED: Use current playerId from lobby state instead of stale appState
      isHost: prev.playerId === payload.host_id,
      connectionError: null,
    }));
  }, []); // No dependencies = no stale closures

  const handleGameStart = useCallback(() => {
    setAppState(prev => ({ ...prev, currentScreen: 'role-reveal' }));
  }, []);

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
    // Only subscribe if we are in a state that requires a connection
    if (appState.currentScreen === 'waiting' || appState.currentScreen === 'game' || appState.currentScreen === 'role-reveal') {
      const unsubscribe = subscribe('CLIENT_IDENTIFIED', handleClientIdentified);
      return unsubscribe;
    }
    // Always return a function to maintain hooks consistency
    return () => { };
  }, [appState.currentScreen, subscribe, handleClientIdentified]);

  // Centralized lobby event management
  useEffect(() => {
    if (appState.currentScreen === 'waiting') {
      const unsubscribers = [
        subscribe('LOBBY_STATE_UPDATE', handleLobbyStateUpdate),
        subscribe('GAME_STARTED', handleGameStart),
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
    // Always return a function to maintain hooks consistency
    return () => { };
  }, [appState.currentScreen, subscribe, handleLobbyStateUpdate, handleGameStart, handleSystemMessage]);

  // REVISED: Centralized WebSocket connection logic
  useEffect(() => {
    let shouldConnect = false;

    if ((appState.currentScreen === 'waiting' || appState.currentScreen === 'game' || appState.currentScreen === 'role-reveal') &&
      appState.gameId && appState.playerId && appState.sessionToken
    ) {
      // Session-based connection: works for both lobby and game states
      connect(appState.gameId, appState.playerId, appState.sessionToken, appState.lastEventId)
        .catch(error => {
          console.error('Failed to connect to WebSocket:', error);
          // TODO: Handle connection error appropriately
        });
      shouldConnect = true;
    }

    // Cleanup function: disconnect when dependencies change or component unmounts
    return () => {
      if (shouldConnect) {
        disconnect();
      }
    };
  }, [
    appState.currentScreen,
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
      playerAvatar: avatar,
      currentScreen: 'lobby-list'
    }));
  };

  const handleJoinLobby = (gameId: string, playerId: string, sessionToken: string) => {
    setAppState(prev => ({
      ...prev,
      gameId,
      playerId,
      sessionToken,
      currentScreen: 'waiting'
    }));
    // Immediately set playerId in lobbyState too
    setLobbyState(prev => ({ ...prev, playerId }));
  };

  const handleCreateGame = (gameId: string, playerId: string, sessionToken: string) => {
    setAppState(prev => ({
      ...prev,
      gameId,
      playerId,
      sessionToken,
      currentScreen: 'waiting'
    }));
    // Immediately set playerId in lobbyState too
    setLobbyState(prev => ({ ...prev, playerId }));
  };


  const handleEnterGame = () => {
    setAppState(prev => ({
      ...prev,
      currentScreen: 'game'
    }));
  };

  // Lobby action handlers
  const handleStartGameAction = useCallback(() => {
    if (lobbyState.isHost && lobbyState.canStart && isConnected && appState.gameId) {
      try {
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
    try {
      if (isConnected) {
        sendAction({
          type: 'LEAVE_GAME',
          payload: {}
        });
      }
      // disconnect is now handled by App.tsx on screen change
      setAppState(prev => ({
        ...prev,
        gameId: undefined,
        playerId: undefined,
        joinToken: undefined,
        sessionToken: undefined,
        currentScreen: 'lobby-list'
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
    } catch (error) {
      console.error('Failed to leave lobby:', error);
      // Still navigate away even if action fails
      setAppState(prev => ({
        ...prev,
        gameId: undefined,
        playerId: undefined,
        joinToken: undefined,
        sessionToken: undefined,
        currentScreen: 'lobby-list'
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
    }
  }, [isConnected, sendAction]);

  const handleBackToLogin = () => {
    setAppState({
      currentScreen: 'login',
      playerName: '',
    });
  };

  // handleUpdateGameState removed - state is now managed by gameEngine

  // Show loading screen while game engine is loading
  if (gameEngineLoading) {
    return (
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Loading game engine...</h2>
          <div className="loading-spinner">⏳</div>
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

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  // Render the appropriate screen based on current state
  switch (appState.currentScreen) {
    case 'login':
      return <LoginScreen onLogin={handleLogin} />;

    case 'lobby-list':
      return (
        <LobbyListScreen
          playerName={appState.playerName}
          playerAvatar={appState.playerAvatar}
          onJoinLobby={handleJoinLobby}
          onCreateGame={handleCreateGame}
          onBack={handleBackToLogin}
        />
      );

    case 'waiting':
      return (
        <WaitingScreen
          gameId={appState.gameId || 'unknown'}
          playerId={appState.playerId}
          lobbyState={lobbyState}
          isConnected={isConnected}
          onStartGame={handleStartGameAction}
          onLeaveLobby={handleLeaveLobby}
        />
      );

    case 'role-reveal':
      return <RoleRevealScreen onEnterGame={handleEnterGame} />;

    case 'game':
      return (
        <GameScreen
          gameState={gameState}
          playerId={appState.playerId || 'unknown'}
        />
      );

    default:
      return <LoginScreen onLogin={handleLogin} />;
  }
}

export default App;