import {
  AppState,
  GameState,
  Role,
  PersonalKPI,
  VoteState,
  PhaseType,
  UserIdentity,
  GameSettings,
} from "../types";
import {
  updateGuestProfile,
  getCurrentUserIdentity,
} from "../services/guestIdentity";

// Define the possible states of the user's session
export type SessionState = "IDLE" | "IN_LOBBY" | "IN_GAME" | "POST_GAME";

// Centralized lobby state interface
export interface PlayerLobbyInfo {
  id: string;
  name: string;
  avatar: string;
  joinedAt: string;
  connectionStatus: string; // "CONNECTED" | "DISCONNECTED"
}

export interface LobbyState {
  playerId?: string;
  playerInfos: PlayerLobbyInfo[];
  isHost: boolean;
  canStart: boolean;
  hostId: string;
  lobbyName: string;
  maxPlayers: number;
  gameSettings?: Partial<GameSettings>;
  connectionError: string | null;
  countdown: {
    isActive: boolean;
    remaining: number;
    duration: number;
  } | null;
}

export interface RoleAssignment {
  role: Role;
  alignment: string;
  personalKPI: PersonalKPI | null;
}

// Skip vote state for real-time updates
export interface SkipVoteState {
  currentVotes: number;
  requiredVotes: number;
  voters: string[];
}

// UI State for game interactions
export interface GameUIState {
  viewedPlayerId: string | null;
  activeChannel: string;
  chatInput: string;
  selectedNominee: string;
  selectedVote: "GUILTY" | "INNOCENT" | "";
  conversionTarget: string;
  miningTarget: string;
  replyingTo: {
    messageId: string;
    playerName: string;
    message: string;
  } | null;
  settingsModalOpen: boolean;
  skipVoteState: SkipVoteState | null;
}

// Consolidated app state
export interface ConsolidatedAppState {
  appState: AppState;
  sessionState: SessionState;
  lobbyState: LobbyState;
  gameState: GameState;
  roleAssignment: RoleAssignment | null;
  gameAnalysis: any | null; // Analysis data from VICTORY_CONDITION event
  isInGameSession: boolean;
  gameUIState: GameUIState;
}

// Action types
export type AppAction =
  | { type: "LOGIN"; payload: { playerName: string; playerAvatar: string } }
  | { type: "SET_VIEWED_PLAYER"; payload: { playerId: string | null } }
  | { type: "SET_ACTIVE_CHANNEL"; payload: { channelId: string } }
  | { type: "SET_CHAT_INPUT"; payload: { input: string } }
  | { type: "SET_SELECTED_NOMINEE"; payload: { nominee: string } }
  | { type: "SET_SELECTED_VOTE"; payload: { vote: "GUILTY" | "INNOCENT" | "" } }
  | { type: "SET_CONVERSION_TARGET"; payload: { target: string } }
  | { type: "SET_MINING_TARGET"; payload: { target: string } }
  | { type: "SET_REPLYING_TO"; payload: { replyingTo: { messageId: string; playerName: string; message: string } | null } }
  | { type: "SET_SETTINGS_MODAL_OPEN"; payload: { isOpen: boolean } }
  | { type: "SESSION_CHECK_COMPLETE" }
  | {
      type: "JOIN_LOBBY";
      payload: { gameId: string; playerId: string; sessionToken: string; lobbyName?: string; isNewJoin?: boolean };
    }
  | {
      type: "CREATE_GAME";
      payload: { gameId: string; playerId: string; sessionToken: string; lobbyName?: string; isNewJoin?: boolean };
    }
  | {
      type: "SPECTATE_GAME";
      payload: { gameId: string; playerId: string; sessionToken: string; lobbyName?: string; isNewJoin?: boolean };
    }
  | { type: "LEAVE_LOBBY" }
  | { type: "BACK_TO_LOGIN" }
  | { type: "ENTER_GAME" }
  | { type: "PLAY_AGAIN" }
  | {
      type: "RESTORE_SESSION";
      payload: {
        gameId: string;
        playerId: string;
        sessionToken: string;
        sessionState: SessionState;
        lobbyName?: string;
      };
    }
  | {
      type: "UPDATE_LOBBY_STATE";
      payload: {
        players: PlayerLobbyInfo[];
        host_id: string;
        can_start: boolean;
        lobby_id: string;
        name: string;
        max_players: number;
        game_settings?: Partial<GameSettings>;
      };
    }
  | {
      type: "UPDATE_PLAYER_CONNECTION_STATUS";
      payload: {
        player_id: string;
        connection_status: string;
      };
    }
  | { type: "SET_CONNECTION_ERROR"; payload: { message: string } }
  | { type: "CLEAR_CONNECTION_ERROR" }
  | { type: "CLIENT_IDENTIFIED"; payload: { playerId: string } }
  | {
      type: "UPDATE_GAME_STATE";
      payload: { gameState: GameState; roleAssignment?: RoleAssignment };
    }
  | {
      type: "UPDATE_SKIP_VOTES";
      payload: { skipVoteState: SkipVoteState };
    }
  | {
      type: "GAME_OVER";
      payload: { sessionState: SessionState; gameAnalysis?: any };
    }
  | { type: "RESET_LOBBY_STATE" }
  | { type: "COUNTDOWN_START"; payload: { duration: number } }
  | { type: "COUNTDOWN_UPDATE"; payload: { remaining: number } }
  | { type: "COUNTDOWN_CANCEL" }
  | {
      type: "HOST_TRANSFERRED";
      payload: { newHostId: string; previousHostId: string };
    }
  | { type: "LOAD_CHAT_HISTORY"; payload: { chatMessages: any[] } }
  | { type: "MESSAGE_REACTION"; payload: { message_id: string; emoji: string; player_id: string; player_name: string } }
  | { type: "VOTE_TALLY_UPDATED"; payload: { voteState: VoteState } }
  | { type: "PULSE_CHECK_UPDATED"; payload: { player_id: string } }
  | { type: "LOBBY_LOADING_TIMEOUT" }
  | { type: "UPDATE_LOCAL_PLAYER_ALIGNMENT"; payload: { newAlignment: string; message: string } };

// Helper function to create initial app state with guest identity
function createInitialAppState(): ConsolidatedAppState {
  const existingIdentity = getCurrentUserIdentity();

  return {
    appState: {
      playerName: existingIdentity?.name || "",
      playerAvatar: existingIdentity?.avatar,
      userIdentity: existingIdentity || undefined,
      hasSyncedInitialState: false,
      isNewJoin: false,
      sessionChecked: false, // Start as false
    },
    sessionState: "IDLE",
    lobbyState: {
      playerId: undefined,
      playerInfos: [],
      isHost: false,
      canStart: false,
      hostId: "",
      lobbyName: "",
      maxPlayers: 8,
      gameSettings: {},
      connectionError: null,
      countdown: null,
    },
    gameState: {
      id: "",
      players: [],
      phase: {
        type: PhaseType.Lobby,
        startTime: new Date().toISOString(),
        duration: 0,
      },
      dayNumber: 1,
      chatMessages: [],
    },
    roleAssignment: null,
    gameAnalysis: null,
    isInGameSession: false,
    gameUIState: {
      viewedPlayerId: null,
      activeChannel: "#war-room",
      chatInput: "",
      selectedNominee: "",
      selectedVote: "",
      conversionTarget: "",
      miningTarget: "",
      replyingTo: null,
      settingsModalOpen: false,
      skipVoteState: null,
    },
  };
}

// Initial state
export const initialAppState: ConsolidatedAppState = createInitialAppState();

// Reducer function
export function appReducer(
  state: ConsolidatedAppState,
  action: AppAction
): ConsolidatedAppState {
  switch (action.type) {
    case "LOGIN":
      // Update guest profile with the new name and avatar
      const guestProfile = updateGuestProfile({
        name: action.payload.playerName,
        avatar: action.payload.playerAvatar,
      });

      const userIdentity: UserIdentity = {
        id: guestProfile.id,
        name: guestProfile.name,
        avatar: guestProfile.avatar,
        isAuthenticated: false,
      };

      return {
        ...state,
        appState: {
          ...state.appState,
          playerName: action.payload.playerName,
          playerAvatar: action.payload.playerAvatar,
          userIdentity,
        },
        sessionState: "IDLE",
        isInGameSession: false,
      };

    case "JOIN_LOBBY":
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: action.payload.gameId,
          playerId: action.payload.playerId,
          sessionToken: action.payload.sessionToken,
          hasSyncedInitialState: false, // Reset sync flag to wait for initial state
          isNewJoin: true, // Mark as fresh join
        },
        lobbyState: {
          ...state.lobbyState,
          playerId: action.payload.playerId,
          lobbyName: action.payload.lobbyName || state.lobbyState.lobbyName,
          countdown: null,
        },
        sessionState: "IN_LOBBY",
        isInGameSession: true,
      };

    case "CREATE_GAME":
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: action.payload.gameId,
          playerId: action.payload.playerId,
          sessionToken: action.payload.sessionToken,
          hasSyncedInitialState: false, // Reset sync flag to wait for initial state
          isNewJoin: true, // Mark as fresh join
        },
        lobbyState: {
          ...state.lobbyState,
          playerId: action.payload.playerId,
          lobbyName: action.payload.lobbyName || state.lobbyState.lobbyName,
          countdown: null,
        },
        sessionState: "IN_LOBBY",
        isInGameSession: true,
      };

    case "SPECTATE_GAME":
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: action.payload.gameId,
          playerId: action.payload.playerId,
          sessionToken: action.payload.sessionToken,
          isSpectating: true,
          hasSyncedInitialState: false, // Reset sync flag to wait for initial state
          isNewJoin: true, // Mark as fresh join
        },
        sessionState: "IN_GAME",
        isInGameSession: true,
        gameUIState: {
          ...state.gameUIState,
          activeChannel: "#spectators",
        },
      };

    case "LEAVE_LOBBY":
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: undefined,
          playerId: undefined,
          joinToken: undefined,
          sessionToken: undefined,
          hasSyncedInitialState: false, // Reset sync flag
          isNewJoin: false, // Reset new join flag
        },
        lobbyState: {
          playerId: undefined,
          playerInfos: [],
          isHost: false,
          canStart: false,
          hostId: "",
          lobbyName: "",
          maxPlayers: 8,
          connectionError: null,
          countdown: null,
        },
        sessionState: "IDLE",
        isInGameSession: false,
      };

    case "BACK_TO_LOGIN":
      return {
        ...state,
        appState: {
          playerName: "",
          userIdentity: undefined,
        },
        sessionState: "IDLE",
        isInGameSession: false,
      };

    case "ENTER_GAME":
      return {
        ...state,
        sessionState: "IN_GAME",
      };

    case "PLAY_AGAIN":
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: undefined,
          sessionToken: undefined,
        },
        sessionState: "IDLE",
        roleAssignment: null,
        isInGameSession: false,
      };

    case "UPDATE_LOBBY_STATE":
      return {
        ...state,
        appState: {
          ...state.appState,
          hasSyncedInitialState: true, // Mark that we've received initial state
          isNewJoin: false, // Reset after first sync
        },
        lobbyState: {
          ...state.lobbyState,
          playerInfos: action.payload.players,
          hostId: action.payload.host_id,
          canStart: action.payload.can_start,
          lobbyName: action.payload.name,
          maxPlayers: action.payload.max_players,
          gameSettings:
            action.payload.game_settings || state.lobbyState.gameSettings,
          isHost: state.lobbyState.playerId === action.payload.host_id,
          connectionError: null,
        },
      };

    case "UPDATE_PLAYER_CONNECTION_STATUS":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          playerInfos: state.lobbyState.playerInfos.map(player =>
            player.id === action.payload.player_id
              ? { ...player, connectionStatus: action.payload.connection_status }
              : player
          ),
        },
      };

    case "SET_CONNECTION_ERROR":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          connectionError: action.payload.message,
        },
      };

    case "CLEAR_CONNECTION_ERROR":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          connectionError: null,
        },
      };

    case "CLIENT_IDENTIFIED":
      return {
        ...state,
        appState: {
          ...state.appState,
          playerId: action.payload.playerId,
        },
        lobbyState: {
          ...state.lobbyState,
          playerId: action.payload.playerId,
          countdown: null,
        },
        gameUIState: {
          ...state.gameUIState,
          viewedPlayerId: action.payload.playerId,
        },
      };

    case "UPDATE_GAME_STATE":
      return {
        ...state,
        appState: {
          ...state.appState,
          hasSyncedInitialState: true, // Mark that we've received initial state
          isNewJoin: false, // Reset after first sync
        },
        gameState: action.payload.gameState,
        ...(action.payload.roleAssignment && {
          roleAssignment: action.payload.roleAssignment,
        }),
      };

    case "UPDATE_SKIP_VOTES":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          skipVoteState: action.payload.skipVoteState,
        },
      };

    case "GAME_OVER":
      return {
        ...state,
        sessionState: action.payload.sessionState,
        gameAnalysis: action.payload.gameAnalysis || null,
      };

    case "RESET_LOBBY_STATE":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          playerInfos: [],
          isHost: false,
          canStart: false,
          hostId: "",
          // Preserve lobby name if we have session data (for reconnects)
          lobbyName: state.appState.gameId ? state.lobbyState.lobbyName : "",
          connectionError: null,
          gameSettings: {},
          countdown: null,
        },
      };

    case "COUNTDOWN_START":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          countdown: {
            isActive: true,
            remaining: action.payload.duration,
            duration: action.payload.duration,
          },
        },
      };

    case "COUNTDOWN_UPDATE":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          countdown: state.lobbyState.countdown
            ? {
                ...state.lobbyState.countdown,
                remaining: action.payload.remaining,
              }
            : null,
        },
      };

    case "COUNTDOWN_CANCEL":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          countdown: null,
        },
      };

    case "HOST_TRANSFERRED":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          hostId: action.payload.newHostId,
          isHost: state.lobbyState.playerId === action.payload.newHostId,
        },
      };

    case "LOAD_CHAT_HISTORY":
      return {
        ...state,
        gameState: {
          ...state.gameState,
          chatMessages: action.payload.chatMessages,
        },
      };

    case "MESSAGE_REACTION": {
      const { message_id, emoji, player_id, player_name } = action.payload;
      const updatedMessages = state.gameState.chatMessages.map(message => {
        if (message.id === message_id) {
          const existingReactions = message.reactions || [];
          
          // Check if this player already reacted with this emoji
          const existingReactionIndex = existingReactions.findIndex(
            reaction => reaction.playerID === player_id && reaction.emoji === emoji
          );
          
          let newReactions;
          if (existingReactionIndex >= 0) {
            // Toggle off - remove the reaction
            newReactions = existingReactions.filter((_, index) => index !== existingReactionIndex);
          } else {
            // Add new reaction
            newReactions = [...existingReactions, {
              emoji,
              playerID: player_id,
              playerName: player_name,
              timestamp: new Date().toISOString()
            }];
          }
          
          return {
            ...message,
            reactions: newReactions
          };
        }
        return message;
      });
      
      return {
        ...state,
        gameState: {
          ...state.gameState,
          chatMessages: updatedMessages,
        },
      };
    }

    case "VOTE_TALLY_UPDATED":
      return {
        ...state,
        gameState: {
          ...state.gameState,
          voteState: action.payload.voteState,
        },
      };

    case "PULSE_CHECK_UPDATED":
      return {
        ...state,
        gameState: {
          ...state.gameState,
          players: state.gameState.players.map((player) =>
            player.id === action.payload.player_id
              ? { ...player, hasSubmittedPulseCheck: true }
              : player
          ),
        },
      };

    case "RESTORE_SESSION":
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: action.payload.gameId,
          playerId: action.payload.playerId,
          sessionToken: action.payload.sessionToken,
          hasSyncedInitialState: false, // Reset sync flag to wait for initial state
          isNewJoin: false, // Mark as session restoration
        },
        lobbyState: {
          ...state.lobbyState,
          playerId: action.payload.playerId,
          lobbyName: action.payload.lobbyName || state.lobbyState.lobbyName,
          countdown: null,
        },
        sessionState: action.payload.sessionState,
        isInGameSession: true,
      };

    case "LOBBY_LOADING_TIMEOUT":
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          connectionError: "Unable to load lobby information. The lobby may no longer exist or there may be a connection issue.",
        },
      };

    case "SET_VIEWED_PLAYER":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          viewedPlayerId: action.payload.playerId,
        },
      };

    case "SET_ACTIVE_CHANNEL":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          activeChannel: action.payload.channelId,
        },
      };

    case "SET_CHAT_INPUT":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          chatInput: action.payload.input,
        },
      };

    case "SET_SELECTED_NOMINEE":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          selectedNominee: action.payload.nominee,
        },
      };

    case "SET_SELECTED_VOTE":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          selectedVote: action.payload.vote,
        },
      };

    case "SET_CONVERSION_TARGET":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          conversionTarget: action.payload.target,
        },
      };

    case "SET_MINING_TARGET":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          miningTarget: action.payload.target,
        },
      };

    case "SET_REPLYING_TO":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          replyingTo: action.payload.replyingTo,
        },
      };

    case "SET_SETTINGS_MODAL_OPEN":
      return {
        ...state,
        gameUIState: {
          ...state.gameUIState,
          settingsModalOpen: action.payload.isOpen,
        },
      };

    case "UPDATE_LOCAL_PLAYER_ALIGNMENT":
      return {
        ...state,
        roleAssignment: state.roleAssignment
          ? {
              ...state.roleAssignment,
              alignment: action.payload.newAlignment,
            }
          : null,
      };

    case "SESSION_CHECK_COMPLETE":
      return {
        ...state,
        appState: {
          ...state.appState,
          sessionChecked: true,
        },
      };

    default:
      return state;
  }
}
