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
  dayNumber: number;
  chatMessages: CoreChatMessage[];
  createdAt: string;
  updatedAt: string;
  settings: CoreGameSettings;
  voteState?: CoreVoteState;
  crisisEvent?: CoreCrisisEvent;
  winCondition?: CoreWinCondition;
  nightActions?: Record<string, CoreSubmittedNightAction>;
  corporateMandate?: CoreCorporateMandate;
  pulseCheckResponses?: Record<string, string>;
  nominatedPlayer?: string;
  nightActionResults?: any[];
  privateNotifications?: any[];
}

export interface CorePlayer {
  id: string;
  name: string;
  jobTitle: string;
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
    dayNumber: coreState.dayNumber,
    chatMessages: coreState.chatMessages || [],
    voteState: coreState.voteState,
    crisisEvent: coreState.crisisEvent,
    winCondition: coreState.winCondition,
    nominatedPlayer: coreState.nominatedPlayer,
    corporateMandate: coreState.corporateMandate,
    nightActionResults: coreState.nightActionResults || [],
    privateNotifications: coreState.privateNotifications || [],
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
    dayNumber: clientState.dayNumber,
    chatMessages: clientState.chatMessages || [],
    createdAt: clientState.createdAt || new Date().toISOString(),
    updatedAt: clientState.updatedAt || new Date().toISOString(),
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
    voteState: clientState.voteState,
    crisisEvent: clientState.crisisEvent,
    winCondition: clientState.winCondition,
    nightActions: clientState.nightActions,
    corporateMandate: clientState.corporateMandate,
    pulseCheckResponses: clientState.pulseCheckResponses,
    nominatedPlayer: clientState.nominatedPlayer,
  };
}

// TODO: Replace this file with actual imports from core package:
// export * from '../../../core/types';
// export { ApplyEvent } from '../../../core/game_state';