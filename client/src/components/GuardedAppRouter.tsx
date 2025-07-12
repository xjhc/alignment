import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { useSessionContext } from '../contexts/SessionContext';
import { useWebSocketContext } from '../contexts/WebSocketContext';

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
    onEnterGame
  } = useSessionContext();
  const { isReconnecting, lastError, isConnected } = useWebSocketContext();
  const location = useLocation();

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  // FIX: Add a loading state guard until the initial session check is complete.
  if (!appState.sessionChecked) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <div className="animate-pulse text-lg font-mono tracking-widest">LOADING SESSION...</div>
      </div>
    );
  }

  // --- Authentication Guardian ---
  // If the user has no name (is not logged in) and is not on the login page,
  // force them back to the login page. This is the highest priority rule.
  if (!appState.playerName && location.pathname !== '/login') {
    return <Navigate to="/login" replace />;
  }

  // --- DEFINITIVE: Reconnection Guardian ---
  // Show syncing screen ONLY for session restoration, not for fresh joins
  const showSyncingScreen = 
    !appState.isNewJoin && // Not a fresh join (i.e., this is a restore)
    !appState.hasSyncedInitialState && // We haven't received the first update yet
    (sessionState === 'IN_LOBBY' || sessionState === 'IN_GAME'); // And we're in a session

  if (showSyncingScreen) {
    // This now correctly shows ONLY on session restoration, not on new lobby creation
    return (
      <div className="screen-transition animation-fade-in">
        <ReconnectionOverlay show={true} />
      </div>
    );
  }

  // --- The "State Guardian" Logic ---
  // Once state is loaded, prevent navigating to incorrect pages.
  if (sessionState !== 'IDLE') {
    if (location.pathname.startsWith('/login') || location.pathname.startsWith('/lobby-list')) {
      if (sessionState === 'IN_GAME') {
        return <Navigate to="/game" replace />;
      } else if (sessionState === 'IN_LOBBY') {
        return <Navigate to="/waiting" replace />;
      } else if (sessionState === 'POST_GAME') {
        return <Navigate to="/game-over" replace />;
      } else {
        // Fallback for any other state
        return <Navigate to="/login" replace />;
      }
    }
  }

  // If the user is in post-game state, they should not be able to navigate to active game URLs
  if (sessionState === 'POST_GAME') {
    if (location.pathname.startsWith('/login') ||
      location.pathname.startsWith('/lobby-list') ||
      location.pathname.startsWith('/waiting') ||
      location.pathname.startsWith('/role-reveal') ||
      location.pathname === '/game') {
      // The state says we're in post-game, redirect to game over
      return <Navigate to="/game-over" replace />;
    }
  }

  // If the user is NOT in a session, they should not be able to access game URLs.
  if (sessionState === 'IDLE') {
    if (location.pathname.startsWith('/waiting') ||
      location.pathname.startsWith('/role-reveal') ||
      location.pathname.startsWith('/game') ||
      location.pathname.startsWith('/analysis')) {
      // The URL is for a game, but our state says we're not in one.
      // The state wins. Force redirect back to the lobby list.
      return <Navigate to="/lobby-list" replace />;
    }
  }

  // If state and URL are consistent, render the routes normally.
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