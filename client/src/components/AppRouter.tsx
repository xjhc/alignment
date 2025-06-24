import { Routes, Route, Navigate } from 'react-router-dom';
import { LoginScreen } from './LoginScreen';
import { LobbyListScreen } from './LobbyListScreen';
import { WaitingScreen } from './WaitingScreen';
import { RoleRevealScreen } from './RoleRevealScreen';
import { GameScreen } from './GameScreen';
import { GameOverScreen } from './GameOverScreen';
import { PostGameAnalysis } from './PostGameAnalysis';
import { WasmTestScreen } from './WasmTestScreen';
import { useSessionContext } from '../contexts/SessionContext';


export function AppRouter() {
  // Get all state and callbacks from the new context
  const {
    appState,
    onLogin,
    onJoinLobby,
    onCreateGame,
    onBackToLogin,
    onEnterGame,
  } = useSessionContext();
  const screenClass = "screen-transition animate-fade-in";

  // Show WASM test screen if query parameter is present
  if (window.location.search.includes('test=wasm')) {
    return <WasmTestScreen />;
  }

  return (
    <div className={screenClass}>
      <Routes>
        <Route
          path="/login"
          element={<LoginScreen onLogin={onLogin} />}
        />

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

        <Route
          path="/waiting"
          element={<WaitingScreen />}
        />

        <Route
          path="/role-reveal"
          element={<RoleRevealScreen onEnterGame={onEnterGame} />}
        />

        <Route
          path="/game"
          element={<GameScreen />}
        />

        <Route
          path="/game-over"
          element={<GameOverScreen />}
        />

        <Route
          path="/analysis"
          element={<PostGameAnalysis />}
        />

        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </div>
  );
}