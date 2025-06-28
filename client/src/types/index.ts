// Re-export generated types from the core package
export * from './generated';
import { 
  ServerEventType, 
  ClientActionType, 
  GeneratedPlayer,
  GeneratedRole,
  GeneratedAbility,
  GeneratedPersonalKPI,
  GeneratedSystemShock,
  GeneratedNightAction,
  GeneratedChatMessage,
  GeneratedPhase,
  GeneratedVoteState,
  GeneratedCrisisEvent,
  GeneratedWinCondition
} from './generated';

// WebSocket message types
export interface WebSocketMessage {
  type: string;
  payload?: any;
}

// Client-to-server actions using generated enum
export interface ClientAction extends WebSocketMessage {
  type: ClientActionType;
}

// Server-to-client events using generated enum
export interface ServerEvent extends WebSocketMessage {
  type: ServerEventType;
  id?: string;        // Event ID for tracking
  game_id?: string;   // Game ID for storage
  gameId?: string;    // Alternative naming for compatibility
  timestamp?: string; // Timestamp for events
  playerId?: string;  // Player ID for events
}

// Type aliases for compatibility with existing code using generated types
export interface Player extends GeneratedPlayer {
  avatar?: string;
  isRolePubliclyRevealed?: boolean;
}
export type Role = GeneratedRole;
export type Ability = GeneratedAbility;
export type PersonalKPI = GeneratedPersonalKPI;
export type SystemShock = GeneratedSystemShock;
export type NightAction = GeneratedNightAction;

// EmojiReaction type - define it here since not yet in core types
export interface EmojiReaction {
  emoji: string;
  playerID: string;
  playerName: string;
  timestamp: string;
}

// Enhanced ChatMessage with specialized message types
export interface ChatMessage extends GeneratedChatMessage {
  type?: 'SITREP' | 'VOTE_RESULT' | 'PULSE_CHECK' | 'PULSE_CHECK_SUBMISSION' | 'INCITING_INCIDENT' | 'LOEBMATE_MESSAGE' | 'REGULAR';
  reactions?: EmojiReaction[];
  metadata?: {
    nightActions?: any[];
    playerHeadcount?: {
      humans: number;
      aligned: number;
      dead: number;
    };
    crisisEvent?: CrisisEvent;
    voteResult?: {
      question: string;
      outcome: string;
      votes: Record<string, string>;
      tokenWeights: Record<string, number>;
      results: Record<string, number>;
      eliminatedPlayer?: {
        id: string;
        name: string;
        role: string;
        alignment: string;
      };
    };
    pulseCheckResponses?: Record<string, string>;
    player_responses?: Record<string, string>;
    question?: string;
    total_responses?: number;
    player_name?: string;
    response?: string;
    // New properties for inciting incident messages
    from?: string;
    to?: string;
    subject?: string;
    body?: string;
  };
}
export type Phase = GeneratedPhase;
export type VoteState = GeneratedVoteState;
export type CrisisEvent = GeneratedCrisisEvent;
export type WinCondition = GeneratedWinCondition;

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
  skipVotes?: Record<string, boolean>;
  pulseCheckResponses?: Record<string, string>;
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
  playerName: string;
  playerAvatar?: string;
  gameId?: string;
  playerId?: string;
  joinToken?: string;
  sessionToken?: string;
  lobbyInfo?: LobbyInfo;
}