import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { useGameEngine } from './hooks/useGameEngine';
import { LoginScreen } from './components/LoginScreen';
import { LobbyListScreen } from './components/LobbyListScreen';
import { WaitingScreen } from './components/WaitingScreen';
import { RoleRevealScreen } from './components/RoleRevealScreen';
import { GameScreen } from './components/GameScreen';
import { GameOverScreen } from './components/GameOverScreen';
import { PostGameAnalysis } from './components/PostGameAnalysis';
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
    if (clientState.winCondition && appState.currentScreen !== 'game-over' && appState.currentScreen !== 'analysis') {
      console.log(`[App] Game over condition met. Winner: ${clientState.winCondition.winner}. Transitioning.`);
      setAppState(prev => ({...prev, currentScreen: 'game-over'}));
    }
  }, [coreGameState, appState.playerId, lobbyState.playerInfos, appState.currentScreen]);

  // Wait for BOTH phase change AND role assignment before transitioning
  useEffect(() => {
    // This effect should ONLY trigger the transition from 'waiting' to 'role-reveal'
    // and it should only be able to do so once.
    const isReadyToReveal =
      appState.currentScreen === 'waiting' &&
      gameState.phase.type !== 'LOBBY' &&
      roleAssignment !== null;

    if (isReadyToReveal) {
      console.log(`[App] Game phase is now '${gameState.phase.type}' and role is assigned. Transitioning to 'role-reveal'.`);
      setAppState(prev => ({ ...prev, currentScreen: 'role-reveal' }));
    }
    // By keeping the dependency array the same, we still react to the right changes,
    // but the more robust condition prevents it from re-triggering incorrectly.
  }, [appState.currentScreen, gameState.phase.type, roleAssignment]);

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

  const handlePlayAgain = () => {
    // End the current session
    setIsInGameSession(false);
    disconnect();
    
    // Reset state and go back to the lobby list
    setAppState(prev => ({
      ...prev,
      currentScreen: 'lobby-list',
      gameId: undefined,
      sessionToken: undefined,
    }));
    setRoleAssignment(null);
  };

  const handleViewAnalysis = () => {
    setAppState(prev => ({ ...prev, currentScreen: 'analysis' }));
  };

  const handleBackToResults = () => {
    setAppState(prev => ({ ...prev, currentScreen: 'game-over' }));
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

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  // Render the appropriate screen based on current state
  const screenClass = "screen-transition animate-fade-in";
  
  switch (appState.currentScreen) {
    case 'login':
      return (
        <div className={screenClass}>
          <LoginScreen onLogin={handleLogin} />
        </div>
      );

    case 'lobby-list':
      return (
        <div className={screenClass}>
          <LobbyListScreen
            playerName={appState.playerName}
            playerAvatar={appState.playerAvatar}
            onJoinLobby={handleJoinLobby}
            onCreateGame={handleCreateGame}
            onBack={handleBackToLogin}
          />
        </div>
      );

    case 'waiting':
      return (
        <div className={screenClass}>
          <WaitingScreen
            gameId={appState.gameId || 'unknown'}
            playerId={appState.playerId}
            lobbyState={lobbyState}
            isConnected={isConnected}
            onStartGame={handleStartGameAction}
            onLeaveLobby={handleLeaveLobby}
          />
        </div>
      );

    case 'role-reveal':
      return (
        <div className={screenClass}>
          <RoleRevealScreen
            assignment={roleAssignment}
            onEnterGame={handleEnterGame}
          />
        </div>
      );

    case 'game':
      return (
        <div className={screenClass}>
          <GameScreen
            gameState={gameState}
            playerId={appState.playerId || 'unknown'}
            isChatHistoryLoading={false} // Chat history is not yet implemented
          />
        </div>
      );

    case 'game-over':
      return (
        <div className={screenClass}>
          <GameOverScreen 
            gameState={gameState} 
            onViewAnalysis={handleViewAnalysis}
            onPlayAgain={handlePlayAgain}
          />
        </div>
      );

    case 'analysis':
      return (
        <div className={screenClass}>
          <PostGameAnalysis 
            gameState={gameState} 
            onBackToResults={handleBackToResults}
            onPlayAgain={handlePlayAgain}
          />
        </div>
      );

    default:
      return (
        <div className={screenClass}>
          <LoginScreen onLogin={handleLogin} />
        </div>
      );
  }
}

export default App;