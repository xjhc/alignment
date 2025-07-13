import { useReducer, useEffect, useCallback, useRef } from "react";
import { useLocation } from "react-router-dom";
import { useWebSocketContext } from "../contexts/WebSocketContext";
import { useGameEngineContext } from "../contexts/GameEngineContext";
import { useAppNavigation } from "./useAppNavigation";
import { soundManager } from "../services/soundManager";
import {
  appReducer,
  initialAppState,
  RoleAssignment,
} from "../state/appReducer";
import { ClientActionType, ServerEventType } from "../types";

export function useSessionManager() {
  const location = useLocation();
  const [state, dispatch] = useReducer(appReducer, initialAppState);
  const didRestoreSession = useRef(false);

  const {
    navigateToLogin,
    navigateToLobbyList,
    navigateToWaiting,
    navigateToRoleReveal,
    navigateToGame,
    navigateToGameOver,
    navigateToAnalysis,
  } = useAppNavigation();

  const { connect, disconnect, subscribe, sendAction, isConnected } =
    useWebSocketContext();
  const {
    gameState: coreGameState,
    isLoading: gameEngineLoading,
    error: gameEngineError,
    loadGameState,
  } = useGameEngineContext();

  // Authentication and session restoration (run-once)
  useEffect(() => {
    if (didRestoreSession.current) {
      return; // Exit early if we've already run this
    }
    didRestoreSession.current = true;

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
                playerAvatar: userData.avatar || "üë§",
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
            // Check for existing session before navigating to lobby list
            const savedSession = localStorage.getItem("alignmentGameSession");
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
                    "[App] Authenticated user has existing session, restoring:",
                    sessionData
                  );
                  dispatch({
                    type: "RESTORE_SESSION",
                    payload: {
                      gameId: sessionData.gameId,
                      playerId: sessionData.playerId,
                      sessionToken: sessionData.sessionToken,
                      sessionState: sessionData.sessionState,
                      lobbyName: sessionData.lobbyName,
                    },
                  });
                  return;
                }
              } catch (error) {
                console.error(
                  "[App] Failed to parse authenticated user's session data:",
                  error
                );
                localStorage.removeItem("alignmentGameSession");
              }
            }
            navigateToLobbyList();
            return;
          }
        }
      } catch (error) {
        console.log("[App] No authenticated user:", error);
      }

      const savedSession = localStorage.getItem("alignmentGameSession");
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
              "[App] Restoring session from localStorage:",
              sessionData
            );
            dispatch({
              type: "RESTORE_SESSION",
              payload: {
                gameId: sessionData.gameId,
                playerId: sessionData.playerId,
                sessionToken: sessionData.sessionToken,
                sessionState: sessionData.sessionState,
                lobbyName: sessionData.lobbyName,
              },
            });
          }
        } catch (error) {
          console.error("[App] Failed to parse saved session data:", error);
          localStorage.removeItem("alignmentGameSession");
        }
      }
      
      // Finally, dispatch that the check is complete
      dispatch({ type: "SESSION_CHECK_COMPLETE" });
    };
    checkAuthAndSession();
  }, [navigateToLobbyList]);

  // Session persistence (runs whenever session state changes)
  useEffect(() => {
    if (
      state.isInGameSession &&
      state.appState.gameId &&
      state.appState.playerId &&
      state.appState.sessionToken
    ) {
      const sessionData = {
        gameId: state.appState.gameId,
        playerId: state.appState.playerId,
        sessionToken: state.appState.sessionToken,
        sessionState: state.sessionState,
        lobbyName: state.lobbyState.lobbyName || "",
        ...(state.appState.isSpectating && { isSpectating: true }),
      };
      localStorage.setItem("alignmentGameSession", JSON.stringify(sessionData));
    } else if (!state.isInGameSession) {
      // If we are not in a session, ensure localStorage is clean
      localStorage.removeItem("alignmentGameSession");
    }
  }, [
    state.isInGameSession,
    state.sessionState,
    state.appState.gameId,
    state.appState.playerId,
    state.appState.sessionToken,
    state.lobbyState.lobbyName,
    state.appState.isSpectating,
  ]);

  // Game state synchronization
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

  // Victory condition subscription
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

  // Game phase navigation
  useEffect(() => {
    if (!coreGameState) return;
    if (coreGameState.phase && location.pathname === "/waiting") {
      const phaseType = coreGameState.phase.type;

      console.log(
        `[SessionManager] Game phase navigation: phase=${phaseType}, location=${location.pathname}`
      );

      // Navigate to appropriate screen based on game phase
      if (phaseType === "LOBBY") {
        // Stay on waiting screen - this is correct
        return;
      } else if (phaseType === "SITREP" || phaseType === "PULSE_CHECK") {
        // These are the initial game phases - go to role reveal first
        console.log(
          "[SessionManager] Navigating to role reveal for phase:",
          phaseType
        );
        navigateToRoleReveal();
      } else if (
        phaseType === "DISCUSSION" ||
        phaseType === "EXTENSION" ||
        phaseType === "NOMINATION" ||
        phaseType === "TRIAL" ||
        phaseType === "VERDICT" ||
        phaseType === "NIGHT"
      ) {
        // These are active game phases - go directly to game screen
        console.log(
          "[SessionManager] Navigating to game screen for phase:",
          phaseType
        );
        navigateToGame();
      } else if (phaseType === "GAME_OVER") {
        // Game is over
        console.log(
          "[SessionManager] Navigating to game over for phase:",
          phaseType
        );
        navigateToGameOver();
      } else {
        // Unknown phase - default to role reveal
        console.log(
          "[SessionManager] Unknown phase, defaulting to role reveal:",
          phaseType
        );
        navigateToRoleReveal();
      }
    }
  }, [
    coreGameState,
    location.pathname,
    navigateToRoleReveal,
    navigateToGame,
    navigateToGameOver,
  ]);

  // Theme setup
  useEffect(() => {
    document.documentElement.setAttribute("data-theme", "dark");
  }, []);

  // Conversion visual effect
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

  // Debug: Log when coreGameState changes
  useEffect(() => {
    if (coreGameState) {
      console.log("[useSessionManager] coreGameState updated:", coreGameState);
      if (coreGameState.chatMessages) {
        console.log("[useSessionManager] coreGameState.chatMessages:", coreGameState.chatMessages.length, "messages");
        coreGameState.chatMessages.forEach((msg: any, index: number) => {
          console.log(`[useSessionManager] Message ${index}: id="${msg.id}", playerName="${msg.playerName}", reactions:`, msg.reactions?.length || 0, msg.reactions);
          if (msg.reactions && msg.reactions.length > 0) {
            console.log(`[useSessionManager] ✅ Message ${msg.id} has ${msg.reactions.length} reactions:`, msg.reactions);
          } else {
            console.log(`[useSessionManager] ❌ Message ${msg.id} has NO reactions`);
          }
        });
      }
    }
  }, [coreGameState]);

  // Music management
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

  // Event handler creators
  const handleLobbyStateUpdate = useCallback((event: any) => {
    console.log("[SessionManager] LOBBY_STATE_UPDATE received:", event.payload);
    console.log(
      "[SessionManager] Player data in payload:",
      event.payload.players ||
        event.payload.connected_players ||
        event.payload.player_infos ||
        "NOT FOUND"
    );
    dispatch({ type: "UPDATE_LOBBY_STATE", payload: event.payload });
  }, []);

  const handlePlayerConnectionStatusChanged = useCallback((event: any) => {
    console.log(
      "[SessionManager] PLAYER_CONNECTION_STATUS_CHANGED received:",
      event.payload
    );
    dispatch({
      type: "UPDATE_PLAYER_CONNECTION_STATUS",
      payload: event.payload,
    });
  }, []);
  const handleClientError = useCallback((event: any) => {
    console.warn("[SessionManager] Client error received:", event.payload);
    // Don't show generic client errors as connection errors unless they're severe
    if (event.payload?.error_code === "CONNECTION_ERROR") {
      dispatch({
        type: "SET_CONNECTION_ERROR",
        payload: {
          message: event.payload.message || "A connection error occurred.",
        },
      });
    }
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
  const handleGameStateUpdate = useCallback(
    async (event: any) => {
      console.log(
        "[SessionManager] GAME_STATE_UPDATE received:",
        event.payload
      );
      if (event.payload?.game_state) {
        try {
          // Use loadGameState (which calls resetAndLoadState) to update the core with the reconnection snapshot
          await loadGameState(event.payload.game_state);
          console.log(
            "[SessionManager] Successfully loaded game state from reconnection snapshot"
          );
        } catch (error) {
          console.error(
            "[SessionManager] Failed to load game state from reconnection snapshot:",
            error
          );
        }
      }
    },
    [loadGameState]
  );
  const handlePulseCheckUpdated = useCallback(
    (event: any) =>
      dispatch({
        type: "PULSE_CHECK_UPDATED",
        payload: { player_id: event.payload.player_id },
      }),
    []
  );

  // Lobby event subscriptions
  useEffect(() => {
    if (location.pathname === "/waiting") {
      const unsubscribers = [
        subscribe("CLIENT_IDENTIFIED", handleClientIdentified),
        subscribe("CHAT_HISTORY_SNAPSHOT", handleChatHistorySnapshot),
        subscribe("GAME_STATE_UPDATE", handleGameStateUpdate),
        subscribe("LOBBY_STATE_UPDATE", handleLobbyStateUpdate),
        subscribe(
          "PLAYER_CONNECTION_STATUS_CHANGED",
          handlePlayerConnectionStatusChanged
        ),
        subscribe("CLIENT_ERROR", handleClientError),
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
    handleGameStateUpdate,
    handleLobbyStateUpdate,
    handlePlayerConnectionStatusChanged,
    handleClientError,
    handleCountdownStart,
    handleCountdownUpdate,
    handleCountdownCancel,
    handleHostTransferred,
  ]);

  // Skip vote event handler
  const handleSkipVoteUpdated = useCallback((event: any) => {
    console.log("[SessionManager] Skip vote updated:", event.payload);
    dispatch({
      type: "UPDATE_SKIP_VOTES",
      payload: {
        skipVoteState: {
          currentVotes: event.payload.current_votes || 0,
          requiredVotes: event.payload.required_votes || 0,
          voters: event.payload.voters || [],
        },
      },
    });
  }, []);

  const handlePhaseChanged = useCallback(() => {
    // Reset skip vote state for new phase with empty state
    dispatch({
      type: "UPDATE_SKIP_VOTES",
      payload: {
        skipVoteState: {
          currentVotes: 0,
          requiredVotes: 0,
          voters: [],
        },
      },
    });
  }, [dispatch]);


  // Game event subscriptions
  useEffect(() => {
    if (!isConnected || location.pathname === "/waiting") return;
    const unsubscribers = [
      subscribe(ServerEventType.PulseCheckUpdated, handlePulseCheckUpdated),
      subscribe(ServerEventType.GameStateUpdate, handleGameStateUpdate),
      subscribe(ServerEventType.SkipVoteUpdated, handleSkipVoteUpdated),
      subscribe(ServerEventType.PhaseChanged, handlePhaseChanged),
    ];
    return () => unsubscribers.forEach((unsub) => unsub());
  }, [
    isConnected,
    location.pathname,
    subscribe,
    handlePulseCheckUpdated,
    handleGameStateUpdate,
    handleSkipVoteUpdated,
    handlePhaseChanged,
  ]);

  // WebSocket connection
  useEffect(() => {
    if (
      state.isInGameSession &&
      state.appState.gameId &&
      state.appState.playerId &&
      state.appState.sessionToken
    ) {
      console.log(
        "[SessionManager] Initiating WebSocket connection with credentials:",
        {
          gameId: state.appState.gameId,
          playerId: state.appState.playerId,
          sessionToken: state.appState.sessionToken ? "***" : "missing",
        }
      );

      connect(
        state.appState.gameId,
        state.appState.playerId,
        state.appState.sessionToken
      ).catch((error) => {
        console.error("[SessionManager] WebSocket connection failed:", error);
        dispatch({
          type: "SET_CONNECTION_ERROR",
          payload: { message: error.message || "Failed to connect to lobby." },
        });
      });
      return () => disconnect();
    } else if (state.isInGameSession) {
      console.warn(
        "[SessionManager] In game session but missing required credentials:",
        {
          gameId: state.appState.gameId || "missing",
          playerId: state.appState.playerId || "missing",
          sessionToken: state.appState.sessionToken ? "present" : "missing",
        }
      );
    }
  }, [
    state.isInGameSession,
    state.appState.gameId,
    state.appState.playerId,
    state.appState.sessionToken,
    connect,
    disconnect,
  ]);

  // Session action handlers
  const handleLogin = (playerName: string, avatar: string) => {
    dispatch({ type: "LOGIN", payload: { playerName, playerAvatar: avatar } });
    navigateToLobbyList();
  };

  const handleJoinLobby = (
    gameId: string,
    playerId: string,
    sessionToken: string,
    lobbyName?: string
  ) => {
    dispatch({
      type: "JOIN_LOBBY",
      payload: { gameId, playerId, sessionToken, lobbyName, isNewJoin: true },
    });
    navigateToWaiting();
  };

  const handleCreateGame = (
    gameId: string,
    playerId: string,
    sessionToken: string,
    lobbyName?: string
  ) => {
    dispatch({
      type: "CREATE_GAME",
      payload: { gameId, playerId, sessionToken, lobbyName, isNewJoin: true },
    });
    navigateToWaiting();
  };

  const handleSpectateGame = (
    gameId: string,
    playerId: string,
    sessionToken: string,
    lobbyName?: string
  ) => {
    dispatch({
      type: "SPECTATE_GAME",
      payload: { gameId, playerId, sessionToken, lobbyName, isNewJoin: true },
    });
    navigateToGame();
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

  // Session expiry handling
  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("SESSION_EXPIRED", (event: any) => {
      console.log(
        "Session expired, clearing session and returning to lobby list"
      );
      // Clear local storage
      localStorage.removeItem("alignmentGameSession");
      // Clear the current session state
      dispatch({ type: "LEAVE_LOBBY" });
      // Disconnect WebSocket
      disconnect();
      // Navigate to lobby list
      navigateToLobbyList();
    });
    return unsubscribe;
  }, [isConnected, subscribe, disconnect, navigateToLobbyList]);

  // Server-forced logout handling
  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("FORCE_LOGOUT", (event: any) => {
      console.log(
        "Server forced logout, clearing all session data and returning to login"
      );
      // Clear all session data
      localStorage.removeItem("alignmentGameSession");
      localStorage.removeItem("wsConnectionCredentials");
      // Clear the current session state completely
      dispatch({ type: "BACK_TO_LOGIN" });
      // Disconnect WebSocket
      disconnect();
      // Navigate to login screen
      navigateToLogin();
    });
    return unsubscribe;
  }, [isConnected, subscribe, disconnect, navigateToLogin]);

  // Player abandoned event handling (when local player abandons)
  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("PLAYER_ABANDONED", (event: any) => {
      // Check if the abandoned player is the local player
      if (event.playerId === state.appState.playerId) {
        console.log(
          "Player abandoned game, clearing session and returning to lobby list"
        );
        // Clear session data
        localStorage.removeItem("alignmentGameSession");
        // Clear the current session state
        dispatch({ type: "LEAVE_LOBBY" });
        // Disconnect WebSocket
        disconnect();
        // Navigate to lobby list
        navigateToLobbyList();
      }
    });
    return unsubscribe;
  }, [
    isConnected,
    subscribe,
    disconnect,
    navigateToLobbyList,
    state.appState.playerId,
  ]);

  // Alignment changed event handling (for AI conversions)
  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("ALIGNMENT_CHANGED", (event: any) => {
      console.log("Alignment changed event received:", event.payload);
      // Update the local player's alignment in the central state
      dispatch({
        type: "UPDATE_LOCAL_PLAYER_ALIGNMENT",
        payload: {
          newAlignment: event.payload.new_alignment,
          message: event.payload.message,
        },
      });
    });
    return unsubscribe;
  }, [isConnected, subscribe]);

  return {
    state,
    dispatch,
    gameEngineLoading,
    gameEngineError,
    isConnected,
    sessionActions: {
      onLogin: handleLogin,
      onJoinLobby: handleJoinLobby,
      onCreateGame: handleCreateGame,
      onSpectateGame: handleSpectateGame,
      onBackToLogin: handleBackToLogin,
      onStartGame: handleStartGameAction,
      onLeaveLobby: handleLeaveLobby,
      onEnterGame: handleEnterGame,
      onViewAnalysis: handleViewAnalysis,
      onPlayAgain: handlePlayAgain,
      onBackToResults: handleBackToResults,
    },
  };
}
