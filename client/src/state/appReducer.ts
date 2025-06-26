import { AppState, GameState, Role, PersonalKPI } from '../types';

// Define the possible states of the user's session
export type SessionState = 'IDLE' | 'IN_LOBBY' | 'IN_GAME' | 'POST_GAME';

// Centralized lobby state interface
export interface PlayerLobbyInfo {
  id: string;
  name: string;
  avatar: string;
}

export interface LobbyState {
  playerId?: string;
  playerInfos: PlayerLobbyInfo[];
  isHost: boolean;
  canStart: boolean;
  hostId: string;
  lobbyName: string;
  maxPlayers: number;
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

// Consolidated app state
export interface ConsolidatedAppState {
  appState: AppState;
  sessionState: SessionState;
  lobbyState: LobbyState;
  gameState: GameState;
  roleAssignment: RoleAssignment | null;
  isInGameSession: boolean;
}

// Action types
export type AppAction = 
  | { type: 'LOGIN'; payload: { playerName: string; playerAvatar: string } }
  | { type: 'JOIN_LOBBY'; payload: { gameId: string; playerId: string; sessionToken: string } }
  | { type: 'CREATE_GAME'; payload: { gameId: string; playerId: string; sessionToken: string } }
  | { type: 'LEAVE_LOBBY' }
  | { type: 'BACK_TO_LOGIN' }
  | { type: 'ENTER_GAME' }
  | { type: 'PLAY_AGAIN' }
  | { type: 'UPDATE_LOBBY_STATE'; payload: { 
      players: PlayerLobbyInfo[];
      host_id: string;
      can_start: boolean;
      lobby_id: string;
      name: string;
      max_players: number;
    }}
  | { type: 'SET_CONNECTION_ERROR'; payload: { message: string } }
  | { type: 'CLEAR_CONNECTION_ERROR' }
  | { type: 'CLIENT_IDENTIFIED'; payload: { playerId: string } }
  | { type: 'UPDATE_GAME_STATE'; payload: { gameState: GameState; roleAssignment?: RoleAssignment } }
  | { type: 'GAME_OVER'; payload: { sessionState: SessionState } }
  | { type: 'RESET_LOBBY_STATE' }
  | { type: 'COUNTDOWN_START'; payload: { duration: number } }
  | { type: 'COUNTDOWN_UPDATE'; payload: { remaining: number } }
  | { type: 'COUNTDOWN_CANCEL' }
  | { type: 'LOAD_CHAT_HISTORY'; payload: { chatMessages: any[] } };

// Initial state
export const initialAppState: ConsolidatedAppState = {
  appState: {
    playerName: '',
  },
  sessionState: 'IDLE',
  lobbyState: {
    playerId: undefined,
    playerInfos: [],
    isHost: false,
    canStart: false,
    hostId: '',
    lobbyName: '',
    maxPlayers: 8,
    connectionError: null,
    countdown: null,
  },
  gameState: {
    id: '',
    players: [],
    phase: { type: 'LOBBY', startTime: new Date().toISOString(), duration: 0 },
    dayNumber: 1,
    chatMessages: [],
  },
  roleAssignment: null,
  isInGameSession: false,
};

// Reducer function
export function appReducer(state: ConsolidatedAppState, action: AppAction): ConsolidatedAppState {
  switch (action.type) {
    case 'LOGIN':
      return {
        ...state,
        appState: {
          ...state.appState,
          playerName: action.payload.playerName,
          playerAvatar: action.payload.playerAvatar,
        },
        sessionState: 'IDLE',
        isInGameSession: false,
      };

    case 'JOIN_LOBBY':
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: action.payload.gameId,
          playerId: action.payload.playerId,
          sessionToken: action.payload.sessionToken,
        },
        lobbyState: {
          ...state.lobbyState,
          playerId: action.payload.playerId,
          countdown: null,
        },
        sessionState: 'IN_LOBBY',
        isInGameSession: true,
      };

    case 'CREATE_GAME':
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: action.payload.gameId,
          playerId: action.payload.playerId,
          sessionToken: action.payload.sessionToken,
        },
        lobbyState: {
          ...state.lobbyState,
          playerId: action.payload.playerId,
          countdown: null,
        },
        sessionState: 'IN_LOBBY',
        isInGameSession: true,
      };

    case 'LEAVE_LOBBY':
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: undefined,
          playerId: undefined,
          joinToken: undefined,
          sessionToken: undefined,
        },
        lobbyState: {
          playerId: undefined,
          playerInfos: [],
          isHost: false,
          canStart: false,
          hostId: '',
          lobbyName: '',
          maxPlayers: 8,
          connectionError: null,
          countdown: null,
        },
        sessionState: 'IDLE',
        isInGameSession: false,
      };

    case 'BACK_TO_LOGIN':
      return {
        ...state,
        appState: {
          playerName: '',
        },
        sessionState: 'IDLE',
        isInGameSession: false,
      };

    case 'ENTER_GAME':
      return {
        ...state,
        sessionState: 'IN_GAME',
      };

    case 'PLAY_AGAIN':
      return {
        ...state,
        appState: {
          ...state.appState,
          gameId: undefined,
          sessionToken: undefined,
        },
        sessionState: 'IDLE',
        roleAssignment: null,
        isInGameSession: false,
      };

    case 'UPDATE_LOBBY_STATE':
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          playerInfos: action.payload.players,
          hostId: action.payload.host_id,
          canStart: action.payload.can_start,
          lobbyName: action.payload.name,
          maxPlayers: action.payload.max_players,
          isHost: state.lobbyState.playerId === action.payload.host_id,
          connectionError: null,
        },
      };

    case 'SET_CONNECTION_ERROR':
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          connectionError: action.payload.message,
        },
      };

    case 'CLEAR_CONNECTION_ERROR':
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          connectionError: null,
        },
      };

    case 'CLIENT_IDENTIFIED':
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
      };

    case 'UPDATE_GAME_STATE':
      return {
        ...state,
        gameState: action.payload.gameState,
        ...(action.payload.roleAssignment && {
          roleAssignment: action.payload.roleAssignment,
        }),
      };

    case 'GAME_OVER':
      return {
        ...state,
        sessionState: action.payload.sessionState,
      };

    case 'RESET_LOBBY_STATE':
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          playerInfos: [],
          isHost: false,
          canStart: false,
          hostId: '',
          lobbyName: '',
          connectionError: null,
          countdown: null,
        },
      };

    case 'COUNTDOWN_START':
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

    case 'COUNTDOWN_UPDATE':
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          countdown: state.lobbyState.countdown ? {
            ...state.lobbyState.countdown,
            remaining: action.payload.remaining,
          } : null,
        },
      };

    case 'COUNTDOWN_CANCEL':
      return {
        ...state,
        lobbyState: {
          ...state.lobbyState,
          countdown: null,
        },
      };

    case 'LOAD_CHAT_HISTORY':
      return {
        ...state,
        gameState: {
          ...state.gameState,
          chatMessages: action.payload.chatMessages,
        },
      };

    default:
      return state;
  }
}