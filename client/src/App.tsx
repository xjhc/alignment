import { useState, useEffect } from "react";
import { AnimatePresence } from "framer-motion";
import { GuardedAppRouter } from "./components/GuardedAppRouter";
import { AchievementNotificationManager } from "./components/AchievementNotification";
import { CommandPalette } from "./components/game/CommandPalette";
import { SettingsModal } from "./components/game/SettingsModal";
import { AppProviders } from "./contexts/AppProviders";
import {
  SessionProvider,
  GameProvider,
  GameContextType,
} from "./contexts";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import { useChatBuffer } from "./hooks/useChatBuffer";
import { useSessionManager } from "./hooks/useSessionManager";
import { useGameActions } from "./hooks/useGameActions";
import { useGameEngineContext } from "./contexts/GameEngineContext";
import { useWebSocketContext } from "./contexts/WebSocketContext";

function AppContent() {
  const { state, gameEngineLoading, gameEngineError, isConnected, sessionActions } = useSessionManager();

  const { commandPaletteOpen, closeCommandPalette } = useKeyboardShortcuts();
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);
  const { canPlayerAffordAbility, isValidNightActionTarget } = useGameEngineContext();
  const { sendAction } = useWebSocketContext();

  useEffect(() => {
    const handleOpenSettings = () => setSettingsModalOpen(true);
    window.addEventListener("open-settings-modal", handleOpenSettings);
    return () =>
      window.removeEventListener("open-settings-modal", handleOpenSettings);
  }, []);

  const localPlayer =
    state.gameState?.players.find((p) => p.id === state.appState.playerId) ||
    null;

  const {
    addMessageToBuffer,
    pendingMessages,
    getPendingMessagesForChannel,
    rateLimitError,
    getBufferStatus,
  } = useChatBuffer(localPlayer, state.gameState?.id, sendAction);

  const [viewedPlayerId, setViewedPlayerId] = useState(
    state.appState.playerId || ""
  );
  const [activeChannel, setActiveChannel] = useState("#war-room");
  const [chatInput, setChatInput] = useState("");
  const [selectedNominee, setSelectedNominee] = useState<string>("");
  const [selectedVote, setSelectedVote] = useState<"GUILTY" | "INNOCENT" | "">(
    ""
  );
  const [conversionTarget, setConversionTarget] = useState<string>("");
  const [miningTarget, setMiningTarget] = useState<string>("");
  const [replyingTo, setReplyingTo] = useState<{
    messageId: string;
    playerName: string;
    message: string;
  } | null>(null);

  const viewedPlayer =
    state.gameState?.players.find((p) => p.id === viewedPlayerId) ||
    localPlayer;

  useEffect(() => {
    if (state.appState.playerId) setViewedPlayerId(state.appState.playerId);
  }, [state.appState.playerId]);

  const gameActions = useGameActions({
    gameId: state.gameState?.id || null,
    localPlayer,
    chatInput,
    setChatInput,
    activeChannel,
    miningTarget,
    setMiningTarget,
    conversionTarget,
    setConversionTarget,
    selectedNominee,
    setSelectedNominee,
    selectedVote,
    setSelectedVote,
    replyingTo,
    setReplyingTo,
    addMessageToBuffer,
  });


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
  if (gameEngineError) {
    return (
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Game Engine Error</h2>
          <p style={{ color: "var(--color-danger)" }}>{gameEngineError}</p>
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
    gameAnalysis: state.gameAnalysis,
    isConnected,
    ...sessionActions,
  };

  const gameContextValue: GameContextType = {
    gameState: state.gameState,
    localPlayerId: state.appState.playerId || "",
    viewedPlayerId,
    localPlayer,
    viewedPlayer,
    isConnected,
    activeChannel,
    sendAction,
    setViewedPlayer: setViewedPlayerId,
    setActiveChannel,
    chatInput,
    setChatInput,
    selectedNominee,
    setSelectedNominee,
    selectedVote,
    setSelectedVote,
    conversionTarget,
    setConversionTarget,
    miningTarget,
    setMiningTarget,
    replyingTo,
    ...gameActions,
    canPlayerAffordAbility: (id: string) => canPlayerAffordAbility(id),
    isValidNightActionTarget: (actorId, targetId, actionType) =>
      isValidNightActionTarget(actorId, targetId, actionType),
    pendingMessages,
    getPendingMessagesForChannel,
    rateLimitError,
    getBufferStatus,
  };

  return (
    <SessionProvider value={sessionContextValue}>
      <GameProvider value={gameContextValue}>
        <AnimatePresence mode="wait">
          <GuardedAppRouter />
        </AnimatePresence>
        <AchievementNotificationManager />
        <CommandPalette
          isOpen={commandPaletteOpen}
          onClose={closeCommandPalette}
        />
        <SettingsModal
          isOpen={settingsModalOpen}
          onClose={() => setSettingsModalOpen(false)}
        />
      </GameProvider>
    </SessionProvider>
  );
}

function App() {
  return (
    <AppProviders>
      <AppContent />
    </AppProviders>
  );
}

export default App;
