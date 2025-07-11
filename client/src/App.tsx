import { useEffect } from "react";
import { AnimatePresence } from "framer-motion";
import { GuardedAppRouter } from "./components/GuardedAppRouter";
import { AchievementNotificationManager } from "./components/AchievementNotification";
import { CommandPalette } from "./components/game/CommandPalette";
import { SettingsModal } from "./components/game/SettingsModal";
import { NotificationManager } from "./components/NotificationManager";
import { NotificationBridge } from "./components/NotificationBridge";
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
  const { state, dispatch, gameEngineLoading, gameEngineError, isConnected, sessionActions } = useSessionManager();

  const { commandPaletteOpen, closeCommandPalette } = useKeyboardShortcuts();
  const { canPlayerAffordAbility, isValidNightActionTarget } = useGameEngineContext();
  const { sendAction } = useWebSocketContext();

  useEffect(() => {
    const handleOpenSettings = () => dispatch({ type: "SET_SETTINGS_MODAL_OPEN", payload: { isOpen: true } });
    window.addEventListener("open-settings-modal", handleOpenSettings);
    return () =>
      window.removeEventListener("open-settings-modal", handleOpenSettings);
  }, [dispatch]);

  const localPlayer =
    (state.gameState?.players && state.appState?.playerId) 
      ? state.gameState.players.find((p) => p.id === state.appState.playerId) || null
      : null;

  const {
    addMessageToBuffer,
    pendingMessages,
    getPendingMessagesForChannel,
    rateLimitError,
    getBufferStatus,
  } = useChatBuffer(localPlayer, state.gameState?.id, sendAction);

  const viewedPlayer =
    (state.gameState?.players && state.gameUIState?.viewedPlayerId) 
      ? state.gameState.players.find((p) => p.id === state.gameUIState.viewedPlayerId) || localPlayer
      : localPlayer;

  // Update viewed player when local player changes
  useEffect(() => {
    if (state.appState?.playerId && state.gameUIState.viewedPlayerId !== state.appState.playerId) {
      dispatch({ type: "SET_VIEWED_PLAYER", payload: { playerId: state.appState.playerId } });
    }
  }, [state.appState?.playerId, state.gameUIState.viewedPlayerId, dispatch]);

  const gameActions = useGameActions({
    gameId: state.gameState?.id || null,
    localPlayer,
    chatInput: state.gameUIState.chatInput,
    setChatInput: (input: string) => dispatch({ type: "SET_CHAT_INPUT", payload: { input } }),
    activeChannel: state.gameUIState.activeChannel,
    miningTarget: state.gameUIState.miningTarget,
    setMiningTarget: (target: string) => dispatch({ type: "SET_MINING_TARGET", payload: { target } }),
    conversionTarget: state.gameUIState.conversionTarget,
    setConversionTarget: (target: string) => dispatch({ type: "SET_CONVERSION_TARGET", payload: { target } }),
    selectedNominee: state.gameUIState.selectedNominee,
    setSelectedNominee: (nominee: string) => dispatch({ type: "SET_SELECTED_NOMINEE", payload: { nominee } }),
    selectedVote: state.gameUIState.selectedVote,
    setSelectedVote: (vote: "GUILTY" | "INNOCENT" | "") => dispatch({ type: "SET_SELECTED_VOTE", payload: { vote } }),
    replyingTo: state.gameUIState.replyingTo,
    setReplyingTo: (replyingTo: { messageId: string; playerName: string; message: string } | null) => dispatch({ type: "SET_REPLYING_TO", payload: { replyingTo } }),
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
    viewedPlayerId: state.gameUIState.viewedPlayerId || "",
    localPlayer,
    viewedPlayer,
    isConnected,
    activeChannel: state.gameUIState.activeChannel,
    skipVoteState: state.gameUIState.skipVoteState,
    sendAction,
    setViewedPlayer: (playerId: string) => dispatch({ type: "SET_VIEWED_PLAYER", payload: { playerId } }),
    setActiveChannel: (channelId: string) => dispatch({ type: "SET_ACTIVE_CHANNEL", payload: { channelId } }),
    chatInput: state.gameUIState.chatInput,
    setChatInput: (input: string) => dispatch({ type: "SET_CHAT_INPUT", payload: { input } }),
    selectedNominee: state.gameUIState.selectedNominee,
    setSelectedNominee: (nominee: string) => dispatch({ type: "SET_SELECTED_NOMINEE", payload: { nominee } }),
    selectedVote: state.gameUIState.selectedVote,
    setSelectedVote: (vote: "GUILTY" | "INNOCENT" | "") => dispatch({ type: "SET_SELECTED_VOTE", payload: { vote } }),
    conversionTarget: state.gameUIState.conversionTarget,
    setConversionTarget: (target: string) => dispatch({ type: "SET_CONVERSION_TARGET", payload: { target } }),
    miningTarget: state.gameUIState.miningTarget,
    setMiningTarget: (target: string) => dispatch({ type: "SET_MINING_TARGET", payload: { target } }),
    replyingTo: state.gameUIState.replyingTo,
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
        <NotificationBridge />
        <AnimatePresence mode="wait">
          <GuardedAppRouter />
        </AnimatePresence>
        <NotificationManager />
        <AchievementNotificationManager />
        <CommandPalette
          isOpen={commandPaletteOpen}
          onClose={closeCommandPalette}
        />
        <SettingsModal
          isOpen={state.gameUIState.settingsModalOpen}
          onClose={() => dispatch({ type: "SET_SETTINGS_MODAL_OPEN", payload: { isOpen: false } })}
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
