import { useReducer, useEffect, useCallback } from "react";
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
import { ClientActionType } from "../types";

export function useSessionManager() {
  const location = useLocation();
  const [state, dispatch] = useReducer(appReducer, initialAppState);

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
  } = useGameEngineContext();

  // Authentication and session restoration
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
            // Check for existing session before navigating to lobby list
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
                    },
                  });
                  return;
                }
              } catch (error) {
                console.error("[App] Failed to parse authenticated user's session data:", error);
                sessionStorage.removeItem("alignmentGameSession");
              }
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
    if (
      coreGameState.phase &&
      coreGameState.phase.type !== "LOBBY" &&
      location.pathname === "/waiting"
    ) {
      navigateToRoleReveal();
    }
  }, [coreGameState, location.pathname, navigateToRoleReveal]);

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
  const handleLobbyStateUpdate = useCallback(
    (event: any) =>
      dispatch({ type: "UPDATE_LOBBY_STATE", payload: event.payload }),
    []
  );
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

  // Lobby event subscriptions
  useEffect(() => {
    if (location.pathname === "/waiting") {
      const unsubscribers = [
        subscribe("CLIENT_IDENTIFIED", handleClientIdentified),
        subscribe("CHAT_HISTORY_SNAPSHOT", handleChatHistorySnapshot),
        subscribe("LOBBY_STATE_UPDATE", handleLobbyStateUpdate),
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
    handleCountdownStart,
    handleCountdownUpdate,
    handleCountdownCancel,
    handleHostTransferred,
  ]);

  // Game event subscriptions
  useEffect(() => {
    if (!isConnected || location.pathname === "/waiting") return;
    const unsubscribers = [
      subscribe("PULSE_CHECK_UPDATED", handlePulseCheckUpdated),
    ];
    return () => unsubscribers.forEach((unsub) => unsub());
  }, [isConnected, location.pathname, subscribe, handlePulseCheckUpdated]);

  // WebSocket connection
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

  // Session action handlers
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