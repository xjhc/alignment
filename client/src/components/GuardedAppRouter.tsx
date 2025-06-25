import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { useSessionContext } from '../contexts/SessionContext';

// Import all screen components
import { LoginScreen } from './LoginScreen';
import { LobbyListScreen } from './LobbyListScreen';
import { WaitingScreen } from './WaitingScreen';
import { RoleRevealScreen } from './RoleRevealScreen';
import { GameScreen } from './GameScreen';
import { GameOverScreen } from './GameOverScreen';
import { PostGameAnalysis } from './PostGameAnalysis';
import { WasmTestScreen } from './WasmTestScreen';


export function GuardedAppRouter() {
  const { 
    sessionState, 
    appState, 
    onLogin, 
    onBackToLogin, 
    onJoinLobby, 
    onCreateGame,
    onEnterGame
  } = useSessionContext();
  const location = useLocation();

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  // --- The "State Guardian" Logic ---
  // If the user is in an active session (lobby or game), they should not be able to
  // manually navigate back to the /login or /lobby-list pages.
  if (sessionState === 'IN_LOBBY' || sessionState === 'IN_GAME') {
    if (location.pathname.startsWith('/login') || location.pathname.startsWith('/lobby-list')) {
      // The internal state says we're in a game, but the URL is for login/lobbies.
      // The state wins. Force redirect back to the active game.
      return <Navigate to="/waiting" replace />;
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
        <Route path="/analysis" element={<PostGameAnalysis />} />
        
        {/* Default route */}
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </div>
  );
}