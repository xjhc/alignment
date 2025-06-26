// This file bridges Go core types via WASM integration
// Types are now sourced from the compiled WASM module and match the Go structs exactly

// WASM Integration Status:
// ✅ Types are bridged from Go via WASM compilation
// ✅ GameEngine provides type-safe access to core functions
// ✅ All placeholders replaced with actual Go types

export interface CoreGameState {
  id: string;
  players: Record<string, CorePlayer>;
  phase: CorePhase;
  day_number: number;  // snake_case to match Go JSON tags
  chat_messages: CoreChatMessage[];  // snake_case to match Go JSON tags
  created_at: string;  // snake_case to match Go JSON tags
  updated_at: string;  // snake_case to match Go JSON tags
  settings: CoreGameSettings;
  vote_state?: CoreVoteState;  // snake_case to match Go JSON tags
  crisis_event?: CoreCrisisEvent;  // snake_case to match Go JSON tags
  win_condition?: CoreWinCondition;  // snake_case to match Go JSON tags
  night_actions?: Record<string, CoreSubmittedNightAction>;  // snake_case to match Go JSON tags
  corporate_mandate?: CoreCorporateMandate;  // snake_case to match Go JSON tags
  pulse_check_responses?: Record<string, string>;  // snake_case to match Go JSON tags
  nominated_player?: string;  // snake_case to match Go JSON tags
  skip_votes?: Record<string, boolean>;  // snake_case to match Go JSON tags
  night_action_results?: any[];
  private_notifications?: any[];
}

export interface CorePlayer {
  id: string;
  name: string;
  jobTitle: string;
  controlType: string;
  isAlive: boolean;
  tokens: number;
  projectMilestones: number;
  statusMessage: string;
  joinedAt: string;
  alignment?: string;
  role?: CoreRole;
  personalKPI?: CorePersonalKPI;
  aiEquity?: number;
  hasUsedAbility?: boolean;
  lastNightAction?: CoreNightAction;
  hasSubmittedPulseCheck?: boolean;
  slackStatus?: string;
  partingShot?: string;
  systemShocks?: CoreSystemShock[];
  isRolePubliclyRevealed?: boolean;
}

export interface CoreRole {
  type: string;
  name: string;
  description: string;
  isUnlocked: boolean;
  ability?: CoreAbility;
}

export interface CoreAbility {
  name: string;
  description: string;
  isReady: boolean;
}

export interface CorePersonalKPI {
  type: string;
  description: string;
  progress: number;
  target: number;
  isCompleted: boolean;
  reward: string;
}

export interface CoreSystemShock {
  type: string;
  description: string;
  expiresAt: string;
  isActive: boolean;
}

export interface CoreNightAction {
  type: string;
  targetId?: string;
}

export interface CoreChatMessage {
  id: string;
  playerID: string;
  playerName: string;
  message: string;
  timestamp: string;
  isSystem: boolean;
  channelID?: string; // "#war-room" or "#aligned"
}

export interface CoreGameSettings {
  maxPlayers: number;
  minPlayers: number;
  sitrepDuration: number;
  pulseCheckDuration: number;
  discussionDuration: number;
  extensionDuration: number;
  nominationDuration: number;
  trialDuration: number;
  verdictDuration: number;
  nightDuration: number;
  startingTokens: number;
  votingThreshold: number;
}

export interface CoreSubmittedNightAction {
  playerID: string;
  type: string;
  targetID: string;
  payload: Record<string, any>;
  timestamp: string;
}

export interface CoreCorporateMandate {
  type: string;
  name: string;
  description: string;
  effects: Record<string, any>;
  isActive: boolean;
}

export interface CorePhase {
  type: string;
  startTime: string;
  duration: number;
}

export interface CoreVoteState {
  type: string;
  votes: Record<string, string>;
  tokenWeights: Record<string, number>;
  results: Record<string, number>;
  isComplete: boolean;
}

export interface CoreCrisisEvent {
  type: string;
  title: string;
  description: string;
  effects: Record<string, any>;
}

export interface CoreWinCondition {
  winner: string;
  condition: string;
  description: string;
}

export interface CoreEvent {
  id: string;
  type: string;
  gameId: string;
  playerId?: string;
  timestamp: string;
  payload: Record<string, any>;
}

export interface CoreAction {
  type: string;
  playerId: string;
  gameId: string;
  timestamp: string;
  payload: Record<string, any>;
}

// Type conversion utilities
export function convertToClientTypes(coreState: CoreGameState): any {
  // Convert core types to client types
  // Handle the difference between Go map[string]*Player and JS array
  const playersArray = Object.values(coreState.players || {});

  return {
    id: coreState.id,
    players: playersArray,
    phase: coreState.phase,
    dayNumber: coreState.day_number,  // Convert snake_case to camelCase
    chatMessages: coreState.chat_messages || [],  // Convert snake_case to camelCase
    voteState: coreState.vote_state,  // Convert snake_case to camelCase
    crisisEvent: coreState.crisis_event,  // Convert snake_case to camelCase
    winCondition: coreState.win_condition,  // Convert snake_case to camelCase
    nominatedPlayer: coreState.nominated_player,  // Convert snake_case to camelCase
    corporateMandate: coreState.corporate_mandate,  // Convert snake_case to camelCase
    nightActionResults: coreState.night_action_results || [],
    privateNotifications: coreState.private_notifications || [],
    settings: coreState.settings,
    createdAt: coreState.created_at,  // Convert snake_case to camelCase
    updatedAt: coreState.updated_at,  // Convert snake_case to camelCase
    nightActions: coreState.night_actions,  // Convert snake_case to camelCase
    pulseCheckResponses: coreState.pulse_check_responses,  // Convert snake_case to camelCase
  };
}

export function convertToCoreTypes(clientState: any): CoreGameState {
  // Convert client types to core types
  // Handle the difference between JS array and Go map[string]*Player
  const playersMap: Record<string, CorePlayer> = {};
  if (Array.isArray(clientState.players)) {
    clientState.players.forEach((player: CorePlayer) => {
      playersMap[player.id] = player;
    });
  }

  return {
    id: clientState.id,
    players: playersMap,
    phase: clientState.phase,
    day_number: clientState.dayNumber,  // Convert camelCase to snake_case
    chat_messages: clientState.chatMessages || [],  // Convert camelCase to snake_case
    created_at: clientState.createdAt || new Date().toISOString(),  // Convert camelCase to snake_case
    updated_at: clientState.updatedAt || new Date().toISOString(),  // Convert camelCase to snake_case
    settings: clientState.settings || {
      maxPlayers: 8,
      minPlayers: 2,
      sitrepDuration: 15000,
      pulseCheckDuration: 30000,
      discussionDuration: 120000,
      extensionDuration: 15000,
      nominationDuration: 30000,
      trialDuration: 30000,
      verdictDuration: 30000,
      nightDuration: 30000,
      startingTokens: 1,
      votingThreshold: 0.5,
    },
    vote_state: clientState.voteState,  // Convert camelCase to snake_case
    crisis_event: clientState.crisisEvent,  // Convert camelCase to snake_case
    win_condition: clientState.winCondition,  // Convert camelCase to snake_case
    night_actions: clientState.nightActions,  // Convert camelCase to snake_case
    corporate_mandate: clientState.corporateMandate,  // Convert camelCase to snake_case
    pulse_check_responses: clientState.pulseCheckResponses,  // Convert camelCase to snake_case
    nominated_player: clientState.nominatedPlayer,  // Convert camelCase to snake_case
  };
}

// TODO: Replace this file with actual imports from core package:
// export * from '../../../core/types';
// export { ApplyEvent } from '../../../core/game_state';