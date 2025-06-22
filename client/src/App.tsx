import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { useGameEngine } from './hooks/useGameEngine';
import { LoginScreen } from './components/LoginScreen';
import { LobbyListScreen } from './components/LobbyListScreen';
import { WaitingScreen } from './components/WaitingScreen';
import { RoleRevealScreen } from './components/RoleRevealScreen';
import { GameScreen } from './components/GameScreen';
import { WasmTestScreen } from './components/WasmTestScreen';
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

function App() {
  const [appState, setAppState] = useState<AppState>({
    currentScreen: 'login',
    playerName: '',
  });

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

  const { connect, disconnect, subscribe, sendAction, isConnected } = useWebSocket();
  const {
    isLoading: gameEngineLoading,
    error: gameEngineError,
    gameState: coreGameState
  } = useGameEngine();

  // The SINGLE source of truth for UI updates.
  useEffect(() => {
    if (!coreGameState || !appState.playerId) {
      // Don't do anything until both the Wasm state and the player's identity are known.
      return;
    }

    // 1. Sync the core state to our React state
    const clientState = convertToClientTypes(coreGameState);
    setGameState(clientState);

    // 2. Try to extract the role assignment for the local player
    const localPlayer = clientState.players.find((p: any) => p.id === appState.playerId);
    if (localPlayer && localPlayer.role && localPlayer.alignment && localPlayer.personalKPI) {
      const newAssignment = { role: localPlayer.role, alignment: localPlayer.alignment, personalKPI: localPlayer.personalKPI };
      setRoleAssignment(newAssignment);
    }

    // 3. Self-healing state transition logic
    if (clientState.phase?.type !== 'LOBBY' && appState.currentScreen === 'waiting') {
      console.log(`[App] Game state advanced to ${clientState.phase?.type} while in lobby. Transitioning to role-reveal.`);
      setAppState(prev => ({ ...prev, currentScreen: 'role-reveal' }));
    }
  }, [coreGameState, appState.playerId, appState.currentScreen]);


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
    if (isInGameSession) {
      const unsubscribe = subscribe('CLIENT_IDENTIFIED', handleClientIdentified);
      return unsubscribe;
    }
    return () => { };
  }, [isInGameSession, subscribe, handleClientIdentified]);

  // Centralized lobby event management
  useEffect(() => {
    if (appState.currentScreen === 'waiting') {
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
  }, [appState.currentScreen, subscribe, handleLobbyStateUpdate, handleSystemMessage]);


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
      playerAvatar: avatar,
      currentScreen: 'lobby-list'
    }));
    setIsInGameSession(false); // Ensure session is not active
  };

  const handleJoinLobby = (gameId: string, playerId: string, sessionToken: string) => {
    setAppState(prev => ({
      ...prev,
      gameId,
      playerId,
      sessionToken,
      currentScreen: 'waiting'
    }));
    setLobbyState(prev => ({ ...prev, playerId }));
    setIsInGameSession(true); // START the session
  };

  const handleCreateGame = (gameId: string, playerId: string, sessionToken: string) => {
    setAppState(prev => ({
      ...prev,
      gameId,
      playerId,
      sessionToken,
      currentScreen: 'waiting'
    }));
    setLobbyState(prev => ({ ...prev, playerId }));
    setIsInGameSession(true); // START the session
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
  }, [disconnect]);

  const handleBackToLogin = () => {
    setIsInGameSession(false); // END the session
    setAppState({
      currentScreen: 'login',
      playerName: '',
    });
  };

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
      return (
        <RoleRevealScreen
          assignment={roleAssignment}
          onEnterGame={handleEnterGame}
        />
      );

    case 'game':
      return (
        <GameScreen
          gameState={gameState}
          playerId={appState.playerId || 'unknown'}
          isChatHistoryLoading={false} // Chat history is not yet implemented
        />
      );

    default:
      return <LoginScreen onLogin={handleLogin} />;
  }
}

export default App;