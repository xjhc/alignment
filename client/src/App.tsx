import {
  useReducer,
  useEffect,
  useCallback,
  useState,
  KeyboardEvent,
} from "react";
import { BrowserRouter, useLocation } from "react-router-dom";
import { AnimatePresence } from "framer-motion";
import { ErrorBoundary } from "react-error-boundary";
import { useWebSocketContext } from "./contexts/WebSocketContext";
import { useGameEngineContext } from "./contexts/GameEngineContext";
import { useAppNavigation } from "./hooks/useAppNavigation";
import { GuardedAppRouter } from "./components/GuardedAppRouter";
import { AchievementNotificationManager } from "./components/AchievementNotification";
import { CommandPalette } from "./components/game/CommandPalette";
import { SettingsModal } from "./components/game/SettingsModal";
import {
  SessionProvider,
  GameProvider,
  ThemeProvider,
  GameEngineProvider,
  WebSocketProvider,
  GameContextType,
} from "./contexts";
import { soundManager } from "./services/soundManager";
import { useSound } from "./hooks/useSound";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import { useChatBuffer } from "./hooks/useChatBuffer";
import {
  appReducer,
  initialAppState,
  RoleAssignment,
  PlayerLobbyInfo,
} from "./state/appReducer";
import { ClientActionType, Player } from "./types";

function ErrorFallback({ error, resetErrorBoundary }: any) {
  return (
    <div role="alert" className="launch-screen">
      <div className="launch-form">
        <h2>Something went wrong:</h2>
        <pre style={{ color: "red" }}>{error.message}</pre>
        <button className="btn-primary" onClick={resetErrorBoundary}>
          Try again
        </button>
      </div>
    </div>
  );
}

function AppContent() {
  const location = useLocation();
  const [state, dispatch] = useReducer(appReducer, initialAppState);

  const { commandPaletteOpen, closeCommandPalette } = useKeyboardShortcuts();
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);

  const {
    navigateToLogin,
    navigateToLobbyList,
    navigateToWaiting,
    navigateToRoleReveal,
    navigateToGame,
    navigateToGameOver,
    navigateToAnalysis,
  } = useAppNavigation();

  useEffect(() => {
    const handleOpenSettings = () => setSettingsModalOpen(true);
    window.addEventListener("open-settings-modal", handleOpenSettings);
    return () =>
      window.removeEventListener("open-settings-modal", handleOpenSettings);
  }, []);

  useEffect(() => {
    const checkAuthAndSession = async () => {
      try {
        const response = await fetch("/api/me");
        if (response.ok) {
          const userData = await response.json();
          if (userData.is_authenticated) {
            console.log("[App] User is authenticated:", userData);
            dispatch({
              type: "LOGIN",
              payload: {
                playerName: userData.name,
                playerAvatar: userData.avatar || "👤",
              },
            });
            const returnUrl = localStorage.getItem("discord_login_return_url");
            if (returnUrl) {
              localStorage.removeItem("discord_login_return_url");
              setTimeout(() => {
                navigateToLobbyList();
                setTimeout(() => {
                  window.location.pathname = returnUrl;
                }, 100);
              }, 100);
              return;
            }
            navigateToLobbyList();
            return;
          }
        }
      } catch (error) {
        console.log("[App] No authenticated user:", error);
      }

      const savedSession = sessionStorage.getItem("alignmentGameSession");
      if (savedSession) {
        try {
          const sessionData = JSON.parse(savedSession);
          if (
            sessionData.gameId &&
            sessionData.playerId &&
            sessionData.sessionToken &&
            sessionData.sessionState
          ) {
            console.log(
              "[App] Restoring session from sessionStorage:",
              sessionData
            );
            dispatch({
              type: "RESTORE_SESSION",
              payload: {
                gameId: sessionData.gameId,
                playerId: sessionData.playerId,
                sessionToken: sessionData.sessionToken,
                sessionState: sessionData.sessionState,
              },
            });
          }
        } catch (error) {
          console.error("[App] Failed to parse saved session data:", error);
          sessionStorage.removeItem("alignmentGameSession");
        }
      }
    };
    checkAuthAndSession();
  }, [navigateToLobbyList]);

  const { connect, disconnect, subscribe, sendAction, isConnected } =
    useWebSocketContext();
  const {
    gameState: coreGameState,
    isLoading: gameEngineLoading,
    error: gameEngineError,
    canPlayerAffordAbility,
    isValidNightActionTarget,
  } = useGameEngineContext();
  const { playSound } = useSound();
  const localPlayer =
    state.gameState?.players.find((p) => p.id === state.appState.playerId) ||
    null;
  const {
    addMessageToBuffer,
    pendingMessages,
    rateLimitError,
    getBufferStatus,
  } = useChatBuffer(localPlayer);

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

  const getPhaseDisplayName = useCallback((phaseType: string) => {
    switch (phaseType) {
      case "SITREP":
        return "SITREP";
      case "PULSE_CHECK":
        return "PULSE CHECK";
      case "DISCUSSION":
        return "DISCUSSION";
      case "NOMINATION":
        return "NOMINATION";
      case "TRIAL":
        return "TRIAL";
      case "VERDICT":
        return "VERDICT";
      case "NIGHT":
        return "NIGHT PHASE";
      case "GAME_OVER":
        return "GAME OVER";
      default:
        return phaseType;
    }
  }, []);

  const handleSendMessage = useCallback(async () => {
    if (
      !chatInput.trim() ||
      !localPlayer ||
      !isConnected ||
      !state.gameState.id
    )
      return;
    try {
      let message = chatInput.trim();
      const statusMatch = message.match(/^\/status\s+(.+)$/);
      if (statusMatch) {
        const statusMessage = statusMatch[1].trim();
        sendAction({
          type: ClientActionType.SetSlackStatus,
          payload: {
            game_id: state.gameState.id,
            player_id: localPlayer.id,
            status_message: statusMessage,
          },
        });
        setChatInput("");
        setReplyingTo(null);
        return;
      }
      if (replyingTo) {
        message = `[quote=${replyingTo.playerName}]${replyingTo.message}[/quote]\n${message}`;
      }
      addMessageToBuffer(message);
      setChatInput("");
      setReplyingTo(null);
      playSound("message");
    } catch (error) {
      console.error("Failed to buffer message:", error);
      playSound("error");
    }
  }, [
    chatInput,
    localPlayer,
    isConnected,
    addMessageToBuffer,
    replyingTo,
    playSound,
    sendAction,
    state.gameState.id,
  ]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        handleSendMessage();
      } else if (e.key === "Escape" && replyingTo) {
        setReplyingTo(null);
      }
    },
    [handleSendMessage, replyingTo]
  );

  const startReply = useCallback(
    (messageId: string, playerName: string, message: string) => {
      setReplyingTo({ messageId, playerName, message });
    },
    []
  );

  const cancelReply = useCallback(() => {
    setReplyingTo(null);
  }, []);

  const handleMineTokens = useCallback(async () => {
    if (!localPlayer || !miningTarget || !isConnected || !state.gameState.id)
      return;
    sendAction({
      type: ClientActionType.SubmitNightAction,
      payload: {
        game_id: state.gameState.id,
        player_id: localPlayer.id,
        action_type: "MINE_TOKENS",
        target_player_id: miningTarget,
      },
    });
    setMiningTarget("");
  }, [localPlayer, miningTarget, isConnected, state.gameState.id, sendAction]);

  const handleUseAbility = useCallback(
    async (targetId?: string) => {
      if (
        !localPlayer ||
        !canPlayerAffordAbility(localPlayer.id) ||
        !isConnected ||
        !state.gameState.id
      )
        return;
      sendAction({
        type: ClientActionType.SubmitNightAction,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer.id,
          action_type: "USE_ABILITY",
          ability_type: localPlayer.role?.type || "UNKNOWN",
          target_id: targetId,
        },
      });
    },
    [
      localPlayer,
      canPlayerAffordAbility,
      isConnected,
      state.gameState.id,
      sendAction,
    ]
  );

  const handleProjectMilestones = useCallback(async () => {
    if (!localPlayer || !isConnected || !state.gameState.id) return;
    sendAction({
      type: ClientActionType.SubmitNightAction,
      payload: {
        game_id: state.gameState.id,
        player_id: localPlayer.id,
        action_type: "PROJECT_MILESTONES",
      },
    });
  }, [localPlayer, isConnected, state.gameState.id, sendAction]);

  const handleConversionAttempt = useCallback(async () => {
    if (
      !localPlayer ||
      !conversionTarget ||
      !isConnected ||
      !state.gameState.id
    )
      return;
    if (
      !isValidNightActionTarget(
        localPlayer.id,
        conversionTarget,
        "ATTEMPT_CONVERSION"
      )
    )
      return;
    sendAction({
      type: ClientActionType.SubmitNightAction,
      payload: {
        game_id: state.gameState.id,
        player_id: localPlayer.id,
        action_type: "ATTEMPT_CONVERSION",
        target_player_id: conversionTarget,
      },
    });
    setConversionTarget("");
  }, [
    localPlayer,
    conversionTarget,
    isConnected,
    isValidNightActionTarget,
    state.gameState.id,
    sendAction,
  ]);

  const handleNominate = useCallback(async () => {
    if (!selectedNominee || !isConnected || !state.gameState.id) return;
    sendAction({
      type: ClientActionType.SubmitVote,
      payload: {
        game_id: state.gameState.id,
        player_id: localPlayer?.id,
        target_id: selectedNominee,
        vote_type: "NOMINATION",
      },
    });
    setSelectedNominee("");
  }, [
    selectedNominee,
    isConnected,
    state.gameState.id,
    localPlayer?.id,
    sendAction,
  ]);

  const handleVote = useCallback(async () => {
    if (!selectedVote || !isConnected || !state.gameState.id) return;
    sendAction({
      type: ClientActionType.SubmitVote,
      payload: {
        game_id: state.gameState.id,
        player_id: localPlayer?.id,
        target_id: selectedVote,
        vote_type: "VERDICT",
      },
    });
    setSelectedVote("");
    playSound("vote");
  }, [
    selectedVote,
    isConnected,
    state.gameState.id,
    localPlayer?.id,
    sendAction,
    playSound,
  ]);

  const handleExtensionVote = useCallback(
    async (choice: "EXTEND" | "NOMINATE") => {
      if (!isConnected || !state.gameState.id) return;
      sendAction({
        type: ClientActionType.SubmitVote,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer?.id,
          target_id: choice,
          vote_type: "EXTENSION",
        },
      });
    },
    [isConnected, state.gameState.id, localPlayer?.id, sendAction]
  );

  const handlePulseCheck = useCallback(
    async (response: string) => {
      if (!isConnected || !state.gameState.id) return;
      sendAction({
        type: ClientActionType.SubmitPulseCheck,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer?.id,
          response,
        },
      });
    },
    [isConnected, state.gameState.id, localPlayer?.id, sendAction]
  );

  const handleSkipPhase = useCallback(async () => {
    if (!localPlayer || !isConnected || !state.gameState.id) return;
    sendAction({
      type: ClientActionType.SubmitSkipVote,
      payload: { game_id: state.gameState.id, player_id: localPlayer.id },
    });
  }, [localPlayer, isConnected, state.gameState.id, sendAction]);

  const handleEmojiReaction = useCallback(
    async (
      messageId: string,
      emoji: string,
      channelId: string = "#war-room"
    ) => {
      if (!localPlayer || !isConnected || !state.gameState.id) return;
      sendAction({
        type: ClientActionType.ReactToMessage,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer.id,
          message_id: messageId,
          emoji,
          channel_id: channelId,
        },
      });
    },
    [localPlayer, isConnected, state.gameState.id, sendAction]
  );

  const handleSubmitPartingShot = useCallback(
    async (partingShot: string) => {
      if (!isConnected || !partingShot.trim() || !state.gameState.id) return;
      sendAction({
        type: ClientActionType.SubmitExitInterview,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer?.id,
          parting_shot: partingShot.trim(),
        },
      });
    },
    [isConnected, state.gameState.id, localPlayer?.id, sendAction]
  );

  const submitWhistleblowerVote = useCallback(
    async (crisisChoice: string) => {
      if (!isConnected || !crisisChoice.trim() || !state.gameState.id) return;
      sendAction({
        type: ClientActionType.SubmitWhistleblowerVote,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer?.id,
          crisis_choice: crisisChoice,
        },
      });
    },
    [isConnected, state.gameState.id, localPlayer?.id, sendAction]
  );

  useEffect(() => {
    const handleBufferFlush = (event: CustomEvent<{ messages: string[] }>) => {
      if (!localPlayer || !isConnected || !state.gameState.id) return;
      const { messages } = event.detail;
      sendAction({
        type: ClientActionType.SendMessage,
        payload: {
          game_id: state.gameState.id,
          player_id: localPlayer.id,
          messages,
          player_name: localPlayer.name,
        },
      });
    };
    window.addEventListener(
      "flushChatBuffer",
      handleBufferFlush as EventListener
    );
    return () =>
      window.removeEventListener(
        "flushChatBuffer",
        handleBufferFlush as EventListener
      );
  }, [localPlayer, isConnected, state.gameState.id, sendAction]);

  useEffect(() => {
    if (!coreGameState || !state.appState.playerId) return;
    const clientState = coreGameState;
    let playersArray: any[] = Array.isArray(clientState.players)
      ? clientState.players
      : Object.values(clientState.players || {});
    const playersWithAvatars = playersArray.map((player: any) => ({
      ...player,
      avatar: state.lobbyState.playerInfos.find((info) => info.id === player.id)
        ?.avatar,
    }));
    const gameStateWithAvatars = {
      ...clientState,
      players: playersWithAvatars,
    };
    const localPlayerInState = playersArray.find(
      (p: any) => p.id === state.appState.playerId
    );
    let roleAssignment: RoleAssignment | undefined;
    if (
      localPlayerInState &&
      localPlayerInState.role &&
      localPlayerInState.alignment
    ) {
      roleAssignment = {
        role: localPlayerInState.role,
        alignment: localPlayerInState.alignment,
        personalKPI: localPlayerInState.personalKPI || null,
      };
    }
    dispatch({
      type: "UPDATE_GAME_STATE",
      payload: { gameState: gameStateWithAvatars, roleAssignment },
    });
    if (gameStateWithAvatars.winCondition) {
      dispatch({ type: "GAME_OVER", payload: { sessionState: "POST_GAME" } });
      navigateToGameOver();
    }
  }, [
    coreGameState,
    state.appState.playerId,
    state.lobbyState.playerInfos,
    navigateToGameOver,
  ]);

  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("VICTORY_CONDITION", (event: any) => {
      const analysis = event.payload?.analysis;
      dispatch({
        type: "GAME_OVER",
        payload: { sessionState: "POST_GAME", gameAnalysis: analysis },
      });
      navigateToGameOver();
    });
    return unsubscribe;
  }, [isConnected, subscribe, navigateToGameOver]);

  useEffect(() => {
    if (!coreGameState) return;
    if (
      coreGameState.phase &&
      coreGameState.phase.type !== "LOBBY" &&
      location.pathname === "/waiting"
    ) {
      navigateToRoleReveal();
    }
  }, [coreGameState, location.pathname, navigateToRoleReveal]);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", "dark");
  }, []);

  useEffect(() => {
    if (!coreGameState || !state.appState.playerId || !state.roleAssignment)
      return;
    const playersArray = Array.isArray(coreGameState.players)
      ? coreGameState.players
      : Object.values(coreGameState.players || {});
    const newLocalPlayer = playersArray.find(
      (p: any) => p.id === state.appState.playerId
    );
    const wasHuman = state.roleAssignment.alignment === "HUMAN";
    const isNowAligned =
      newLocalPlayer?.alignment === "ALIGNED" ||
      newLocalPlayer?.alignment === "AI";
    if (wasHuman && isNowAligned) {
      const rootElement = document.getElementById("root");
      if (rootElement) {
        rootElement.classList.add("conversion-glitch-overlay");
        setTimeout(
          () => rootElement.classList.remove("conversion-glitch-overlay"),
          500
        );
      }
    }
  }, [coreGameState, state.appState.playerId, state.roleAssignment]);

  useEffect(() => {
    if (
      location.pathname === "/waiting" ||
      location.pathname === "/lobby-list"
    ) {
      soundManager.playMusic("lobby");
      return;
    }
    if (coreGameState?.phase?.type) {
      const phaseType = coreGameState.phase.type;
      switch (phaseType) {
        case "DAY":
        case "DISCUSSION":
        case "VOTING":
          soundManager.playMusic("day");
          break;
        case "NIGHT":
        case "NIGHT_ACTIONS":
          soundManager.playMusic("night");
          break;
        case "GAME_OVER":
          if (coreGameState.winCondition) {
            soundManager.playSound("victory");
          }
          soundManager.stopMusic();
          break;
        default:
          if (location.pathname === "/game") soundManager.playMusic("day");
      }
    }
  }, [
    coreGameState?.phase?.type,
    coreGameState?.winCondition,
    location.pathname,
  ]);

  const handleLobbyStateUpdate = useCallback(
    (event: any) =>
      dispatch({ type: "UPDATE_LOBBY_STATE", payload: event.payload }),
    []
  );
  const handleSystemMessage = useCallback((event: any) => {
    if (event.payload.error)
      dispatch({
        type: "SET_CONNECTION_ERROR",
        payload: { message: event.payload.message },
      });
  }, []);
  const handleClientIdentified = useCallback(
    (event: any) =>
      dispatch({
        type: "CLIENT_IDENTIFIED",
        payload: { playerId: event.payload.your_player_id },
      }),
    []
  );
  const handleCountdownStart = useCallback(
    (event: any) =>
      dispatch({
        type: "COUNTDOWN_START",
        payload: { duration: event.payload.duration },
      }),
    []
  );
  const handleCountdownUpdate = useCallback(
    (event: any) =>
      dispatch({
        type: "COUNTDOWN_UPDATE",
        payload: { remaining: event.payload.remaining },
      }),
    []
  );
  const handleCountdownCancel = useCallback(
    () => dispatch({ type: "COUNTDOWN_CANCEL" }),
    []
  );
  const handleHostTransferred = useCallback(
    (event: any) =>
      dispatch({
        type: "HOST_TRANSFERRED",
        payload: {
          newHostId: event.payload.new_host_id,
          previousHostId: event.payload.previous_host_id,
        },
      }),
    []
  );
  const handleChatHistorySnapshot = useCallback(
    (event: any) =>
      dispatch({
        type: "LOAD_CHAT_HISTORY",
        payload: { chatMessages: event.payload.chat_messages || [] },
      }),
    []
  );
  const handlePulseCheckUpdated = useCallback(
    (event: any) =>
      dispatch({
        type: "PULSE_CHECK_UPDATED",
        payload: { player_id: event.payload.player_id },
      }),
    []
  );

  useEffect(() => {
    if (location.pathname === "/waiting") {
      const unsubscribers = [
        subscribe("CLIENT_IDENTIFIED", handleClientIdentified),
        subscribe("CHAT_HISTORY_SNAPSHOT", handleChatHistorySnapshot),
        subscribe("LOBBY_STATE_UPDATE", handleLobbyStateUpdate),
        subscribe("SYSTEM_MESSAGE", handleSystemMessage),
        subscribe("GAME_START_COUNTDOWN_INITIATED", handleCountdownStart),
        subscribe("GAME_START_COUNTDOWN_UPDATE", handleCountdownUpdate),
        subscribe("GAME_START_COUNTDOWN_CANCELLED", handleCountdownCancel),
        subscribe("HOST_TRANSFERRED", handleHostTransferred),
      ];
      dispatch({ type: "RESET_LOBBY_STATE" });
      return () => unsubscribers.forEach((unsub) => unsub());
    }
    return () => {};
  }, [
    location.pathname,
    subscribe,
    handleClientIdentified,
    handleChatHistorySnapshot,
    handleLobbyStateUpdate,
    handleSystemMessage,
    handleCountdownStart,
    handleCountdownUpdate,
    handleCountdownCancel,
    handleHostTransferred,
  ]);

  useEffect(() => {
    if (!isConnected || location.pathname === "/waiting") return;
    const unsubscribers = [
      subscribe("PULSE_CHECK_UPDATED", handlePulseCheckUpdated),
    ];
    return () => unsubscribers.forEach((unsub) => unsub());
  }, [isConnected, location.pathname, subscribe, handlePulseCheckUpdated]);

  useEffect(() => {
    if (
      state.isInGameSession &&
      state.appState.gameId &&
      state.appState.playerId &&
      state.appState.sessionToken
    ) {
      connect(
        state.appState.gameId,
        state.appState.playerId,
        state.appState.sessionToken
      ).catch((error) =>
        dispatch({
          type: "SET_CONNECTION_ERROR",
          payload: { message: error.message || "Failed to connect to lobby." },
        })
      );
      return () => disconnect();
    }
  }, [
    state.isInGameSession,
    state.appState.gameId,
    state.appState.playerId,
    state.appState.sessionToken,
    connect,
    disconnect,
  ]);

  const handleLogin = (playerName: string, avatar: string) => {
    dispatch({ type: "LOGIN", payload: { playerName, playerAvatar: avatar } });
    navigateToLobbyList();
  };
  const handleJoinLobby = (
    gameId: string,
    playerId: string,
    sessionToken: string
  ) => {
    dispatch({
      type: "JOIN_LOBBY",
      payload: { gameId, playerId, sessionToken },
    });
    navigateToWaiting();
  };
  const handleCreateGame = (
    gameId: string,
    playerId: string,
    sessionToken: string
  ) => {
    dispatch({
      type: "CREATE_GAME",
      payload: { gameId, playerId, sessionToken },
    });
    navigateToWaiting();
  };
  const handleEnterGame = () => {
    dispatch({ type: "ENTER_GAME" });
    navigateToGame();
  };
  const handleStartGameAction = useCallback(() => {
    if (
      state.lobbyState.isHost &&
      state.lobbyState.canStart &&
      isConnected &&
      state.appState.gameId
    ) {
      sendAction({
        type: ClientActionType.StartGame,
        payload: { game_id: state.appState.gameId },
      });
    }
  }, [
    state.lobbyState.isHost,
    state.lobbyState.canStart,
    isConnected,
    state.appState.gameId,
    sendAction,
  ]);
  const handleLeaveLobby = useCallback(() => {
    disconnect();
    dispatch({ type: "LEAVE_LOBBY" });
    navigateToLobbyList();
  }, [disconnect, navigateToLobbyList]);
  const handleBackToLogin = () => {
    dispatch({ type: "BACK_TO_LOGIN" });
    navigateToLogin();
  };
  const handlePlayAgain = () => {
    disconnect();
    dispatch({ type: "PLAY_AGAIN" });
    navigateToLobbyList();
  };
  const handleViewAnalysis = () => navigateToAnalysis();
  const handleBackToResults = () => navigateToGameOver();

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
    onLogin: handleLogin,
    onJoinLobby: handleJoinLobby,
    onCreateGame: handleCreateGame,
    onBackToLogin: handleBackToLogin,
    onStartGame: handleStartGameAction,
    onLeaveLobby: handleLeaveLobby,
    onEnterGame: handleEnterGame,
    onViewAnalysis: handleViewAnalysis,
    onPlayAgain: handlePlayAgain,
    onBackToResults: handleBackToResults,
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
    handleSendMessage,
    handleMineTokens,
    handleUseAbility,
    handleProjectMilestones,
    handleConversionAttempt,
    handleNominate,
    handleVote,
    handleExtensionVote,
    handlePulseCheck,
    handleKeyDown: handleKeyDown as (e: KeyboardEvent<any>) => void,
    startReply,
    cancelReply,
    handleSkipPhase,
    handleEmojiReaction,
    handleSubmitPartingShot,
    submitWhistleblowerVote,
    getPhaseDisplayName,
    canPlayerAffordAbility: (id: string) => canPlayerAffordAbility(id),
    isValidNightActionTarget: (actorId, targetId, actionType) =>
      isValidNightActionTarget(actorId, targetId, actionType),
    pendingMessages,
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
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onReset={() => window.location.reload()}
    >
      <BrowserRouter>
        <ThemeProvider>
          <GameEngineProvider>
            <WebSocketProvider>
              <AppContent />
            </WebSocketProvider>
          </GameEngineProvider>
        </ThemeProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
