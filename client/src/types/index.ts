// Re-export types from the core package via WASM bridge
import * as CoreTypes from '../utils/coreTypes';
export * from '../utils/coreTypes';

// WebSocket message types
export interface WebSocketMessage {
  type: string;
  payload?: any;
}

// Client-to-server actions
export interface ClientAction extends WebSocketMessage {
  type: 'RECONNECT' | 'CREATE_GAME' | 'JOIN_GAME' | 'START_GAME' | 'LEAVE_GAME' |
        'POST_CHAT_MESSAGE' | 'UPDATE_STATUS' | 'SUBMIT_NIGHT_ACTION' | 
        'SUBMIT_VOTE' | 'SUBMIT_PULSE_CHECK' | 'SUBMIT_EXIT_INTERVIEW' |
        'REQUEST_LOBBY_LIST';
}

// Server-to-client events
export interface ServerEvent extends WebSocketMessage {
  type: 'PLAYER_JOINED' | 'PLAYER_LEFT' | 'PLAYER_DEACTIVATED' | 
        'ROLES_ASSIGNED' | 'ALIGNMENT_CHANGED' | 'PHASE_CHANGED' | 
        'CHAT_MESSAGE_POSTED' | 'PULSE_CHECK_SUBMITTED' | 
        'NIGHT_ACTIONS_RESOLVED' | 'GAME_ENDED' | 'SYNC_COMPLETE' | 
        'PRIVATE_NOTIFICATION' | 'LOBBY_LIST_UPDATE' | 'LOBBY_STATE_UPDATE' |
        'GAME_STARTED';
  id?: string;        // Event ID for tracking
  game_id?: string;   // Game ID for storage
}

// Type aliases for compatibility with existing code
export type Player = CoreTypes.CorePlayer;
export type Role = CoreTypes.CoreRole;
export type Ability = CoreTypes.CoreAbility;
export type PersonalKPI = CoreTypes.CorePersonalKPI;
export type SystemShock = CoreTypes.CoreSystemShock;
export type NightAction = CoreTypes.CoreNightAction;
export type ChatMessage = CoreTypes.CoreChatMessage;
export type Phase = CoreTypes.CorePhase;
export type VoteState = CoreTypes.CoreVoteState;
export type CrisisEvent = CoreTypes.CoreCrisisEvent;
export type WinCondition = CoreTypes.CoreWinCondition;

// GameState with client-friendly structure (array instead of map for players)
export interface GameState {
  id: string;
  players: Player[];
  phase: Phase;
  dayNumber: number;
  chatMessages: ChatMessage[];
  voteState?: VoteState;
  crisisEvent?: CrisisEvent;
  winCondition?: WinCondition;
  nominatedPlayer?: string;
  corporateMandate?: CorporateMandate;
  nightActionResults?: NightActionResult[];
  privateNotifications?: PrivateNotification[];
}

// Corporate Mandate information
export interface CorporateMandate {
  type: string;
  name: string;
  description: string;
  effects: Record<string, any>;
  isActive: boolean;
}

// Night action results for SITREP display
export interface NightActionResult {
  id: string;
  type: string;
  playerName: string;
  targetName?: string;
  result: 'success' | 'failed' | 'blocked';
  description: string;
  isPublic: boolean;
}

// Private notifications for individual players
export interface PrivateNotification {
  id: string;
  type: 'system_shock' | 'kpi_progress' | 'role_ability' | 'conversion' | 'investigation';
  title: string;
  message: string;
  timestamp: string;
  isRead: boolean;
  priority: 'low' | 'medium' | 'high';
}

// Connection state
export interface ConnectionState {
  isConnected: boolean;
  isReconnecting: boolean;
  lastError?: string;
}

// Lobby information for optimistic UI
export interface LobbyInfo {
  id: string;
  name: string;
  player_count: number;
  max_players: number;
  min_players: number;
  status: string;
  can_join: boolean;
  created_at: string;
}

// Application state
export interface AppState {
  currentScreen: 'login' | 'lobby-list' | 'waiting' | 'game' | 'role-reveal';
  playerName: string;
  playerAvatar?: string;
  gameId?: string;
  playerId?: string;
  joinToken?: string;
  sessionToken?: string;
  lastEventId?: string;
  lobbyInfo?: LobbyInfo;
}