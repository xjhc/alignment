import { useReducer, useEffect, useCallback, useRef } from "react";
import { useLocation } from "react-router-dom";
import { useWebSocketContext } from "../contexts/WebSocketContext";
import { useGameEngineContext } from "../contexts/GameEngineContext";
import { useAppNavigation } from "./useAppNavigation";
import { soundManager } from "../services/soundManager";
import { gameEngine } from "../services/gameEngine";
import {
  appReducer,
  initialAppState,
  RoleAssignment,
} from "../state/appReducer";
import { ClientActionType, ServerEventType, ServerEvent } from "../types";

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
    resetAndLoadState,
  } = useGameEngineContext();

  // Authentication and session restoration (run-once)
  useEffect(() => {
    if (didRestoreSession.current) {
      return; // Exit early if we've already run this
    }
    didRestoreSession.current = true;

    const checkAuthAndSession = async () => {
      // PRIORITY 1: Check for an existing game session first.
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
              "[App] Found existing session, restoring:",
              sessionData
            );
            dispatch({ 
              type: "RESTORE_SESSION", 
              payload: {
                ...sessionData,
                hasSeenRoleReveal: sessionData.hasSeenRoleReveal || false,
              }
            });
            
            // Immediately attempt to reconnect to the WebSocket
            console.log("[App] Attempting to reconnect to WebSocket with restored session");
            try {
              await connect(sessionData.gameId, sessionData.playerId, sessionData.sessionToken);
              console.log("[App] WebSocket reconnection successful");
            } catch (error) {
              console.error("[App] WebSocket reconnection failed:", error);
              // Don't clear the session here - let the WebSocket handler deal with it
            }
            
            // Session restored, we can exit. The router will handle navigation.
            return;
          }
        } catch (error) {
          console.error("[App] Failed to parse saved session data:", error);
          localStorage.removeItem("alignmentGameSession");
        }
      }

      // PRIORITY 2: If no session, check for an authenticated user (e.g., from Discord login).
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
                playerAvatar: userData.avatar || "👤",
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

      // PRIORITY 3: If no session and no auth, we're a new guest. Mark check as complete.
      // Finally, dispatch that the check is complete
      dispatch({ type: "SESSION_CHECK_COMPLETE" });
    };
    checkAuthAndSession();
  }, [navigateToLobbyList, connect]);

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
        hasSeenRoleReveal: state.appState.hasSeenRoleReveal || false,
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
    state.appState.hasSeenRoleReveal,
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
        console.log(
          "[useSessionManager] coreGameState.chatMessages:",
          coreGameState.chatMessages.length,
          "messages"
        );
        coreGameState.chatMessages.forEach((msg: any, index: number) => {
          console.log(
            `[useSessionManager] Message ${index}: id="${msg.id}", playerName="${msg.playerName}", reactions:`,
            msg.reactions?.length || 0,
            msg.reactions
          );
          if (msg.reactions && msg.reactions.length > 0) {
            console.log(
              `[useSessionManager] ✅ Message ${msg.id} has ${msg.reactions.length} reactions:`,
              msg.reactions
            );
          } else {
            console.log(
              `[useSessionManager] ❌ Message ${msg.id} has NO reactions`
            );
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

  const handleGameStarted = useCallback((event: any) => {
    const gameId = event.payload?.game_id;
    if (gameId) {
      console.log("[SessionManager] Game started, transitioning state to IN_GAME");
      dispatch({ type: "GAME_STARTED", payload: { gameId } });
      navigateToRoleReveal();
    } else {
      console.error("[SessionManager] GAME_STARTED event missing game_id");
    }
  }, [navigateToRoleReveal]);
  const handleRoleAssigned = useCallback((event: any) => {
    console.log(
      "[SessionManager] ROLE_ASSIGNED received:",
      event.payload
    );
    dispatch({
      type: "ROLE_ASSIGNED",
      payload: { roleAssignment: event.payload },
    });
  }, []);
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
          // Use resetAndLoadState to update the core with the reconnection snapshot
          await resetAndLoadState(event.payload.game_state);
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
    [resetAndLoadState]
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
        subscribe("GAME_STARTED", handleGameStarted),
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
    handleGameStarted,
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

  // NEW: Unified handler for all granular game events
  const handleGameEvent = useCallback(
    async (event: ServerEvent) => {
      if (!gameEngine.isReady()) {
        console.warn(`Game engine not ready, skipping event ${event.type}`);
        return;
      }

      try {
        console.log(
          `[SessionManager] Applying granular event ${event.type} to game engine`
        );

        // Convert ServerEvent to CoreEvent format
        const coreEvent = {
          id: event.id || `event_${Date.now()}`,
          type: event.type,
          gameId: event.gameId || event.game_id || "",
          playerId: event.playerId || "",
          timestamp: event.timestamp || new Date().toISOString(),
          payload: event.payload || {},
        };

        // Special handling for chat messages to ensure proper format
        if (event.type === ServerEventType.ChatMessage) {
          // Backend sends chat messages with this payload structure:
          // payload: { sender_id, sender_name, message, phase, day_number, channel_id }
          // We need to make sure the playerId is set from sender_id
          if (event.payload?.sender_id) {
            coreEvent.playerId = event.payload.sender_id;
          }
        }

        // The applyEvent now returns the new state directly
        const newGameState = await gameEngine.applyEvent(coreEvent as any);

        // Find the local player's role for the roleAssignment piece of state
        const playersArray: any[] = Array.isArray(newGameState.players)
          ? newGameState.players
          : Object.values(newGameState.players || {});
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

        // Add avatars to the new game state
        const playersWithAvatars = playersArray.map((player: any) => ({
          ...player,
          avatar: state.lobbyState.playerInfos.find(
            (info) => info.id === player.id
          )?.avatar,
        }));
        const gameStateWithAvatars = {
          ...newGameState,
          players: playersWithAvatars,
        };

        // Dispatch the updated state to the reducer
        dispatch({
          type: "UPDATE_GAME_STATE",
          payload: { gameState: gameStateWithAvatars, roleAssignment },
        });

        // If this event was a phase change, also reset the skip vote UI state.
        // This consolidates all logic for a PHASE_CHANGED event into one handler.
        if (event.type === ServerEventType.PhaseChanged) {
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
          console.log(
            "[SessionManager] Skip vote state reset due to phase change."
          );
        }

        // Check for game over condition
        if (gameStateWithAvatars.winCondition) {
          dispatch({
            type: "GAME_OVER",
            payload: { sessionState: "POST_GAME" },
          });
          navigateToGameOver();
        }
      } catch (error) {
        console.error(`Failed to apply event ${event.type}:`, error);
      }
    },
    [
      gameEngine.isReady,
      state.appState.playerId,
      state.lobbyState.playerInfos,
      navigateToGameOver,
    ]
  );

  // Game event subscriptions
  useEffect(() => {
    if (!isConnected || location.pathname === "/waiting") return;

    // List of all events that should trigger a state update
    const stateChangingEvents: string[] = [
      ServerEventType.PhaseChanged,
      ServerEventType.ChatMessage,
      ServerEventType.MessageReaction,
      ServerEventType.VoteCast,
      ServerEventType.ExtensionVotingTriggered,
      ServerEventType.NightActionSubmitted,
      ServerEventType.NightActionsResolved,
      ServerEventType.PlayerLeft,
      ServerEventType.PlayerEliminated,
      ServerEventType.MandateActivated,
      ServerEventType.RoleAssigned,
      ServerEventType.IncitingIncident,
      ServerEventType.LoebmateMessage,
      // Add any other event that modifies core.GameState
    ];

    const unsubscribers = stateChangingEvents.map((eventType) =>
      subscribe(eventType, handleGameEvent)
    );

    // Continue to handle full state snapshots and non-state events separately
    unsubscribers.push(
      subscribe(ServerEventType.GameStateUpdate, handleGameStateUpdate)
    );
    unsubscribers.push(
      subscribe(ServerEventType.PulseCheckUpdated, handlePulseCheckUpdated)
    );
    unsubscribers.push(
      subscribe(ServerEventType.SkipVoteUpdated, handleSkipVoteUpdated)
    );
    unsubscribers.push(
      subscribe(ServerEventType.RoleAssigned, handleRoleAssigned)
    );

    return () => unsubscribers.forEach((unsub) => unsub());
  }, [
    isConnected,
    location.pathname,
    subscribe,
    handleGameEvent,
    handleGameStateUpdate,
    handlePulseCheckUpdated,
    handleSkipVoteUpdated,
    handleRoleAssigned,
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
    // This is called from RoleRevealScreen to navigate to the main game view.
    // Dispatch ACKNOWLEDGE_ROLE to mark that the player has seen the role reveal.
    dispatch({ type: "ACKNOWLEDGE_ROLE" });
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

  // Centralized SESSION_EXPIRED handling with comprehensive reason-based routing
  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("SESSION_EXPIRED", (event: any) => {
      const reason = event.payload?.reason;
      const message = event.payload?.message || "Session expired";
      
      console.log(`[SessionManager] SESSION_EXPIRED received: reason="${reason}", message="${message}"`);
      
      // Clear localStorage immediately for all session expired events
      localStorage.removeItem("alignmentGameSession");
      
      switch (reason) {
        case 'session_invalid':
          // Unrecoverable error - the session token is bad or expired
          console.log("[SessionManager] Session invalid - returning to login");
          dispatch({ type: "BACK_TO_LOGIN" });
          disconnect();
          navigateToLogin();
          break;
          
        case 'lobby_not_found':
        case 'game_not_found':
          // The specific lobby/game is gone, but the user's identity is fine
          console.log(`[SessionManager] ${reason} - returning to lobby list`);
          dispatch({ type: "LEAVE_LOBBY" });
          disconnect();
          navigateToLobbyList();
          break;
          
        case 'transition_failed':
        case 'reconnection_failed':
          // Failed to transition or reconnect to game state
          console.log(`[SessionManager] ${reason} - attempting to return to lobby list`);
          dispatch({ type: "LEAVE_LOBBY" });
          disconnect();
          navigateToLobbyList();
          break;
          
        case 'server_error':
        case 'join_failed':
          // Server-side errors
          console.log(`[SessionManager] ${reason} - returning to lobby list`);
          dispatch({ type: "LEAVE_LOBBY" });
          disconnect();
          navigateToLobbyList();
          break;
          
        default:
          // Unknown reason - assume recoverable and go to lobby list
          console.log(`[SessionManager] Unknown session expiry reason "${reason}" - returning to lobby list`);
          dispatch({ type: "LEAVE_LOBBY" });
          disconnect();
          navigateToLobbyList();
          break;
      }
    });
    return unsubscribe;
  }, [isConnected, subscribe, disconnect, navigateToLogin, navigateToLobbyList]);

  // Player abandoned event handling (when local player abandons)
  useEffect(() => {
    if (!isConnected) return;
    const unsubscribe = subscribe("PLAYER_ABANDONED", (event: any) => {
      // Check if the abandoned player is the local player
      // CRITICAL FIX: Check the correct event property for player ID
      const abandonedPlayerId =
        event.playerId || event.payload?.player_id || event.payload?.playerId;
      if (abandonedPlayerId === state.appState.playerId) {
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
