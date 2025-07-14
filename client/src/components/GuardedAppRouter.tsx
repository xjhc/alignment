import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { useSessionContext } from '../contexts/SessionContext';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import { Button } from './ui/Button';

// Import all screen components
import { LoginScreen } from './LoginScreen';
import { LobbyListScreen } from './LobbyListScreen';
import { WaitingScreen } from './WaitingScreen';
import { RoleRevealScreen } from './RoleRevealScreen';
import { GameScreen } from './GameScreen';
import { GameOverScreen } from './GameOverScreen';
import { PostGameAnalysis } from './PostGameAnalysis';
import { WasmTestScreen } from './WasmTestScreen';
import { JoinLobbyScreen } from './JoinLobbyScreen';
import { PartyJoinScreen } from './PartyJoinScreen';
import { ReconnectionOverlay } from './ReconnectionOverlay';


export function GuardedAppRouter() {
  const {
    sessionState,
    appState,
    gameState,
    gameAnalysis,
    onLogin,
    onBackToLogin,
    onJoinLobby,
    onCreateGame,
    onSpectateGame,
    onLeaveLobby,
    onEnterGame
  } = useSessionContext();
  const { isReconnecting, lastError, isConnected } = useWebSocketContext();
  const location = useLocation();

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  // Show a loading screen until the initial session check from localStorage is complete.
  if (!appState.sessionChecked) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <div className="animate-pulse text-lg font-mono tracking-widest">INITIALIZING...</div>
      </div>
    );
  }

  // If the user has an invalid session (e.g., from an old game), show an error with a recovery option.
  if (lastError?.includes('Session expired') || lastError?.includes('invalid session')) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <div className="bg-background-secondary border border-border p-8 rounded-lg text-center">
          <h2 className="text-xl font-bold text-danger mb-4">Session Expired</h2>
          <p className="text-text-secondary mb-6">Your session has expired or is invalid. Please return to the lobby list.</p>
          <Button variant="primary" onClick={onLeaveLobby}>
            Return to Lobbies
          </Button>
        </div>
      </div>
    );
  }

  // Force unauthenticated users back to the login screen.
  if (!appState.playerName && location.pathname !== '/login') {
    return <Navigate to="/login" replace />;
  }

  // For session restores, show a "Syncing..." overlay while waiting for the first state update.
  const showSyncingScreen = 
    !appState.isNewJoin && // This is a restore, not a fresh join.
    !appState.hasSyncedInitialState && // We haven't received state yet.
    (sessionState === 'IN_LOBBY' || sessionState === 'IN_GAME'); // And we expect to be in a session.

  if (showSyncingScreen) {
    return (
      <div className="screen-transition animation-fade-in">
        <ReconnectionOverlay show={true} />
      </div>
    );
  }

  // The "State Guardian": Once state is synced, this logic ensures the user is on the correct screen
  // for their current session state, preventing manual navigation to invalid pages.
  switch (sessionState) {
    case 'IN_LOBBY':
      if (location.pathname !== '/waiting') {
        return <Navigate to="/waiting" replace />;
      }
      break;
    case 'IN_GAME':
      // Allow access to game-related routes, but redirect from lobby/login.
      if (location.pathname !== '/game' && location.pathname !== '/role-reveal') {
        // If we have a role, we've started. Go to role reveal.
        // The RoleRevealScreen will handle navigating to /game.
        return <Navigate to="/role-reveal" replace />;
      }
      break;
    case 'POST_GAME':
      if (location.pathname !== '/game-over' && location.pathname !== '/analysis') {
        return <Navigate to="/game-over" replace />;
      }
      break;
    case 'IDLE':
      if (appState.playerName && location.pathname !== '/lobby-list' && !location.pathname.startsWith('/join')) {
        return <Navigate to="/lobby-list" replace />;
      }
      break;
  }

  // If all checks pass, render the router and handle reconnection overlays.
  return (
    <div className="screen-transition animation-fade-in">
      <Routes>
        <Route path="/login" element={<LoginScreen onLogin={onLogin} />} />
        <Route
          path="/lobby-list"
          element={
            <LobbyListScreen
              playerName={appState.playerName}
              playerAvatar={appState.playerAvatar}
              onJoinLobby={onJoinLobby}
              onCreateGame={onCreateGame}
              onSpectateGame={onSpectateGame}
              onBack={onBackToLogin}
            />
          }
        />
        <Route path="/waiting" element={<WaitingScreen />} />
        <Route path="/role-reveal" element={<RoleRevealScreen onEnterGame={onEnterGame} />} />
        <Route path="/game" element={<GameScreen />} />
        <Route path="/game-over" element={<GameOverScreen />} />
        <Route path="/analysis" element={<PostGameAnalysis analysisData={gameAnalysis} />} />
        <Route path="/join/:lobbyId" element={<JoinLobbyScreen />} />
        <Route path="/party/join/:inviteCode" element={<PartyJoinScreen />} />

        {/* Default route */}
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
      
      {/* Reconnection overlay - shows on top of any page during reconnection */}
      {/* This overlay is for mid-session network drops, only show if session was already synced */}
      <ReconnectionOverlay show={isReconnecting && !!appState.hasSyncedInitialState} />
    </div>
  );
}