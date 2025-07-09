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
    onEnterGame
  } = useSessionContext();
  const { isReconnecting, lastError } = useWebSocketContext();
  const location = useLocation();

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  // --- Authentication Guardian ---
  // If the user has no name (is not logged in) and is not on the login page,
  // force them back to the login page. This is the highest priority rule.
  if (!appState.playerName && location.pathname !== '/login') {
    return <Navigate to="/login" replace />;
  }

  // --- The "State Guardian" Logic ---
  // If the user is in an active session (lobby or game), they should not be able to
  // manually navigate back to the /login or /lobby-list pages.
  if (sessionState === 'IN_LOBBY' || sessionState === 'IN_GAME') {
    if (location.pathname.startsWith('/login') || location.pathname.startsWith('/lobby-list')) {
      // Only redirect if we have valid session data (gameId and playerId)
      // Otherwise, the session state might be stale and we should go to login
      if (appState.gameId && appState.playerId) {
        // If we're currently reconnecting, prevent any navigation changes
        // and let the reconnection overlay handle the user experience
        if (isReconnecting) {
          console.log('[GuardedAppRouter] Reconnecting - preventing navigation, showing overlay');
          // Block the redirect during reconnection - the overlay will handle this
          return (
            <div className="screen-transition animation-fade-in">
              <div className="flex items-center justify-center min-h-screen bg-background-primary">
                <div className="text-center">
                  <h2 className="text-lg font-semibold text-text-primary mb-2">
                    Reconnecting to your game...
                  </h2>
                  <p className="text-text-secondary">Please wait while we restore your session.</p>
                </div>
              </div>
              <ReconnectionOverlay 
                show={true}
              />
            </div>
          );
        } else {
          // The internal state says we're in a game, but the URL is for login/lobbies.
          // The state wins. Navigate to the appropriate screen based on game phase.
          
          // If we have game state and it's not in LOBBY phase, go to the appropriate screen
          if (gameState && gameState.phase) {
            console.log(`[GuardedAppRouter] Navigating based on game phase: ${gameState.phase.type}`);
            switch (gameState.phase.type) {
              case 'LOBBY':
                return <Navigate to="/waiting" replace />;
              case 'SITREP':
              case 'PULSE_CHECK':
              case 'DISCUSSION':
              case 'EXTENSION':
              case 'NOMINATION':
              case 'TRIAL':
              case 'VERDICT':
              case 'NIGHT':
                return <Navigate to="/game" replace />;
              case 'GAME_OVER':
                return <Navigate to="/game-over" replace />;
              default:
                return <Navigate to="/waiting" replace />;
            }
          } else {
            // No game state yet, default to waiting
            console.log(`[GuardedAppRouter] No game state available, defaulting to waiting. sessionState: ${sessionState}`);
            return <Navigate to="/waiting" replace />;
          }
        }
      } else {
        // Session state indicates active session but we don't have valid session data
        // This suggests stale/invalid state, so redirect to login
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
      {/* Only show overlay if we're not in the special reconnection screen above */}
      <ReconnectionOverlay 
        show={isReconnecting && (sessionState === 'IN_LOBBY' || sessionState === 'IN_GAME') && 
              !(location.pathname.startsWith('/login') || location.pathname.startsWith('/lobby-list'))} 
      />
    </div>
  );
}