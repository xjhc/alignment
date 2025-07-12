// AUTO-GENERATED FILE - DO NOT EDIT
// Generated from Go core package types
// Run 'npm run generate:types' to update

// Generated ServerEventType enum
export enum ServerEventType {
  GameCreated = "GAME_CREATED",
  GameStarted = "GAME_STARTED",
  GameStartCountdownInitiated = "GAME_START_COUNTDOWN_INITIATED",
  GameStartCountdownUpdate = "GAME_START_COUNTDOWN_UPDATE",
  GameStartCountdownCancelled = "GAME_START_COUNTDOWN_CANCELLED",
  GameEnded = "GAME_ENDED",
  PhaseChanged = "PHASE_CHANGED",
  PlayerJoined = "PLAYER_JOINED",
  PlayerLeft = "PLAYER_LEFT",
  PlayerEliminated = "PLAYER_ELIMINATED",
  PlayerAbandoned = "PLAYER_ABANDONED",
  PlayerRoleRevealed = "PLAYER_ROLE_REVEALED",
  PlayerAligned = "PLAYER_ALIGNED",
  AlignmentChanged = "ALIGNMENT_CHANGED",
  PlayerShocked = "PLAYER_SHOCKED",
  HostTransferred = "HOST_TRANSFERRED",
  VoteStarted = "VOTE_STARTED",
  VoteCast = "VOTE_CAST",
  VoteTallyUpdated = "VOTE_TALLY_UPDATED",
  VoteCompleted = "VOTE_COMPLETED",
  PlayerNominated = "PLAYER_NOMINATED",
  ExtensionVotingTriggered = "EXTENSION_VOTING_TRIGGERED",
  TokensAwarded = "TOKENS_AWARDED",
  TokensSpent = "TOKENS_SPENT",
  MiningAttempted = "MINING_ATTEMPTED",
  MiningSuccessful = "MINING_SUCCESSFUL",
  MiningFailed = "MINING_FAILED",
  NightActionsResolved = "NIGHT_ACTIONS_RESOLVED",
  PlayerBlocked = "PLAYER_BLOCKED",
  PlayerProtected = "PLAYER_PROTECTED",
  PlayerInvestigated = "PLAYER_INVESTIGATED",
  AiConversionAttempt = "AI_CONVERSION_ATTEMPT",
  AiConversionSuccess = "AI_CONVERSION_SUCCESS",
  AiConversionFailed = "AI_CONVERSION_FAILED",
  AiRevealed = "AI_REVEALED",
  ChatMessage = "CHAT_MESSAGE",
  MessageReaction = "MESSAGE_REACTION",
  SystemMessage = "SYSTEM_MESSAGE",
  PrivateNotification = "PRIVATE_NOTIFICATION",
  IncitingIncident = "INCITING_INCIDENT",
  LoebmateMessage = "LOEBMATE_MESSAGE",
  ClientError = "CLIENT_ERROR",
  SitrepPublished = "SITREP_PUBLISHED",
  LiaisonProtocolActivated = "LIAISON_PROTOCOL_ACTIVATED",
  LiaisonIntelRevealed = "LIAISON_INTEL_REVEALED",
  AiConversionBlocked = "AI_CONVERSION_BLOCKED",
  GameRuleModified = "GAME_RULE_MODIFIED",
  CrisisTriggered = "CRISIS_TRIGGERED",
  PulseCheckStarted = "PULSE_CHECK_STARTED",
  PulseCheckUpdated = "PULSE_CHECK_UPDATED",
  PulseCheckRevealed = "PULSE_CHECK_REVEALED",
  RoleAbilityUnlocked = "ROLE_ABILITY_UNLOCKED",
  ProjectMilestone = "PROJECT_MILESTONE",
  RoleAssigned = "ROLE_ASSIGNED",
  MiningPoolUpdated = "MINING_POOL_UPDATED",
  TokensDistributed = "TOKENS_DISTRIBUTED",
  TokensLost = "TOKENS_LOST",
  DayStarted = "DAY_STARTED",
  NightStarted = "NIGHT_STARTED",
  NightActionSubmitted = "NIGHT_ACTION_SUBMITTED",
  AllPlayersReady = "ALL_PLAYERS_READY",
  PlayerStatusChanged = "PLAYER_STATUS_CHANGED",
  GameStateSnapshot = "GAME_STATE_SNAPSHOT",
  GameStateUpdate = "GAME_STATE_UPDATE",
  LobbyStateUpdate = "LOBBY_STATE_UPDATE",
  ClientIdentified = "CLIENT_IDENTIFIED",
  ChatHistorySnapshot = "CHAT_HISTORY_SNAPSHOT",
  PlayerReconnected = "PLAYER_RECONNECTED",
  PlayerDisconnected = "PLAYER_DISCONNECTED",
  PlayerConnectionStatusChanged = "PLAYER_CONNECTION_STATUS_CHANGED",
  SyncComplete = "SYNC_COMPLETE",
  RateLimitExceeded = "RATE_LIMIT_EXCEEDED",
  SessionExpired = "SESSION_EXPIRED",
  ForceLogout = "FORCE_LOGOUT",
  SkipVoteUpdated = "SKIP_VOTE_UPDATED",
  VictoryCondition = "VICTORY_CONDITION",
  RunAudit = "RUN_AUDIT",
  OverclockServers = "OVERCLOCK_SERVERS",
  IsolateNode = "ISOLATE_NODE",
  PerformanceReview = "PERFORMANCE_REVIEW",
  ReallocateBudget = "REALLOCATE_BUDGET",
  Pivot = "PIVOT",
  DeployHotfix = "DEPLOY_HOTFIX",
  SlackStatusChanged = "SLACK_STATUS_CHANGED",
  PartingShotSet = "PARTING_SHOT_SET",
  WhisperSent = "WHISPER_SENT",
  KpiAssigned = "KPI_ASSIGNED",
  KpiProgress = "KPI_PROGRESS",
  KpiCompleted = "KPI_COMPLETED",
  SystemShockApplied = "SYSTEM_SHOCK_APPLIED",
  ShockEffectTriggered = "SHOCK_EFFECT_TRIGGERED",
  AiEquityChanged = "AI_EQUITY_CHANGED",
  EquityThreshold = "EQUITY_THRESHOLD",
  WhistleblowerVotingStarted = "WHISTLEBLOWER_VOTING_STARTED",
  WhistleblowerVoteCast = "WHISTLEBLOWER_VOTE_CAST",
  WhistleblowerVotingCompleted = "WHISTLEBLOWER_VOTING_COMPLETED",
  MandateActivated = "MANDATE_ACTIVATED",
  MandateEffect = "MANDATE_EFFECT",
  SpectatorStateSnapshot = "SPECTATOR_STATE_SNAPSHOT",
  SpectatorJoined = "SPECTATOR_JOINED",
  SpectatorLeft = "SPECTATOR_LEFT",
  SpectatorChatMessage = "SPECTATOR_CHAT_MESSAGE",
}

// Generated ClientActionType enum
export enum ClientActionType {
  CreateGame = "CREATE_GAME",
  JoinGame = "JOIN_GAME",
  LeaveGame = "LEAVE_GAME",
  StartGame = "START_GAME",
  CreateParty = "CREATE_PARTY",
  InviteToParty = "INVITE_TO_PARTY",
  JoinParty = "JOIN_PARTY",
  LeaveParty = "LEAVE_PARTY",
  SendMessage = "SEND_MESSAGE",
  ReactToMessage = "REACT_TO_MESSAGE",
  SubmitPulseCheck = "SUBMIT_PULSE_CHECK",
  SubmitVote = "SUBMIT_VOTE",
  ExtendDiscussion = "EXTEND_DISCUSSION",
  SubmitSkipVote = "SUBMIT_SKIP_VOTE",
  TriggerExtensionVoting = "TRIGGER_EXTENSION_VOTING",
  SubmitNightAction = "SUBMIT_NIGHT_ACTION",
  MineTokens = "MINE_TOKENS",
  UseAbility = "USE_ABILITY",
  AttemptConversion = "ATTEMPT_CONVERSION",
  ProjectMilestones = "PROJECT_MILESTONES",
  RunAudit = "RUN_AUDIT",
  OverclockServers = "OVERCLOCK_SERVERS",
  IsolateNode = "ISOLATE_NODE",
  PerformanceReview = "PERFORMANCE_REVIEW",
  ReallocateBudget = "REALLOCATE_BUDGET",
  Pivot = "PIVOT",
  DeployHotfix = "DEPLOY_HOTFIX",
  SetSlackStatus = "SET_SLACK_STATUS",
  SubmitExitInterview = "SUBMIT_EXIT_INTERVIEW",
  Whisper = "WHISPER",
  Reconnect = "RECONNECT",
  AbandonGame = "ABANDON_GAME",
  SyncLobbyState = "SYNC_LOBBY_STATE",
  SetPlayerConnectionStatus = "SET_PLAYER_CONNECTION_STATUS",
  AbandonPlayer = "ABANDON_PLAYER",
  AssignCorporateMandate = "ASSIGN_CORPORATE_MANDATE",
  SubmitWhistleblowerVote = "SUBMIT_WHISTLEBLOWER_VOTE",
  PostSpectatorMessage = "POST_SPECTATOR_MESSAGE",
  Mine = "MINE",
  Convert = "CONVERT",
  Block = "BLOCK",
  Investigate = "INVESTIGATE",
  Protect = "PROTECT",
  Bootcamp = "BOOTCAMP",
  Shadow = "SHADOW",
}

// Generated PhaseType enum
export enum PhaseType {
  Lobby = "LOBBY",
  Sitrep = "SITREP",
  PulseCheck = "PULSE_CHECK",
  Discussion = "DISCUSSION",
  Extension = "EXTENSION",
  Nomination = "NOMINATION",
  Trial = "TRIAL",
  Verdict = "VERDICT",
  Night = "NIGHT",
  GameOver = "GAME_OVER",
}

// Generated RoleType enum
export enum RoleType {
  Ciso = "CISO",
  Ceo = "CEO",
  Cto = "CTO",
  Coo = "COO",
  Cfo = "CFO",
  Ethics = "ETHICS",
  Platforms = "PLATFORMS",
  Intern = "INTERN",
}

// Generated KPIType enum
export enum KPIType {
  Capitalist = "CAPITALIST",
  Guardian = "GUARDIAN",
  Inquisitor = "INQUISITOR",
  SuccessionPlanner = "SUCCESSION_PLANNER",
  Scapegoat = "SCAPEGOAT",
}

// Generated VoteType enum
export enum VoteType {
  Extension = "EXTENSION",
  Nomination = "NOMINATION",
  Verdict = "VERDICT",
}

// Generated interfaces from Go structs
export interface GeneratedSystemShock {
  type: string;
  description: string;
  expiresAt: string;
  isActive: boolean;
}

export interface GeneratedMiningResults {
  total_attempts: number;
  successful_slots: number;
  available_slots: number;
  liquidity_pool: number;
  successful_miners: GeneratedMiningAttempt[];
  failed_miners: GeneratedMiningAttempt[];
  priority_rules: Record<string, any>;
}

export interface GeneratedSitrepSection {
  title: string;
  content: string;
  type: string;
}

export interface GeneratedEvent {
  id: string;
  type: string;
  gameId: string;
  playerId?: string;
  timestamp: string;
  payload: Record<string, any>;
}

export interface GeneratedAbility {
  name: string;
  description: string;
  isReady: boolean;
}

export interface GeneratedVoteState {
  type: string;
  votes: Record<string, string>;
  tokenWeights: Record<string, number>;
  results: Record<string, number>;
  isComplete: boolean;
}

export interface GeneratedWinCondition {
  winner: string;
  condition: string;
  description: string;
}

export interface GeneratedPlayerStateChanges {
  tokens_gained?: number;
  tokens_lost?: number;
  status_message?: string;
  alignment?: string;
  ai_equity?: number;
  project_milestones?: number;
  has_used_ability?: boolean;
  role_unlocked?: boolean;
  system_shocks?: GeneratedSystemShock[];
  was_blocked?: boolean;
  was_targeted?: boolean;
  action_cancelled?: boolean;
  custom_effects?: Record<string, any>;
}

export interface GeneratedCrisisEventOption {
  type: string;
  title: string;
  description: string;
}

export interface GeneratedNightAction {
  type: string;
  targetId?: string;
  shadowTargetId?: string;
}

export interface GeneratedGameSettings {
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
  initialAlignedHumanCount: number;
  playAsAI: boolean;
  customSettings?: Record<string, any>;
}

export interface GeneratedPrivateNotification {
  type: string;
  title?: string;
  message: string;
  data?: Record<string, any>;
  channel?: string;
  urgent?: boolean;
}

export interface GeneratedCrisisEvent {
  type: string;
  title: string;
  description: string;
  pulseCheckPrompt?: string;
  effects: Record<string, any>;
  duration?: number;
  triggeredAt?: string;
}

export interface GeneratedPersonalKPI {
  type: string;
  description: string;
  progress: number;
  target: number;
  isCompleted: boolean;
  reward: string;
}

export interface GeneratedSubmittedNightAction {
  playerID: string;
  type: string;
  targetID: string;
  payload?: Record<string, any>;
  timestamp: string;
}

export interface GeneratedNightActionResolutionPayload {
  summary: string;
  player_state_changes: Record<string, GeneratedPlayerStateChanges>;
  action_results: Record<string, GeneratedActionResult>;
  blocked_players: string[];
  conversion_attempts: GeneratedConversionAttempt[];
  role_ability_usages: GeneratedRoleAbilityUsage[];
  mining_results: GeneratedMiningResults;
  public_announcements: string[];
  private_notifications: Record<string, GeneratedPrivateNotification[]>;
}

export interface GeneratedMiningAttempt {
  player_id: string;
  beneficiary_id: string;
  tokens_awarded: number;
  priority: number;
  failure_reason?: string;
  was_blocked?: boolean;
  blocked_by?: string;
}

export interface GeneratedPublicPlayerInfo {
  id: string;
  name: string;
  job_title: string;
  is_active: boolean;
  status_message: string;
  token_count: number;
}

export interface GeneratedWhistleblowerVoting {
  isActive: boolean;
  crisisOptions: GeneratedCrisisEventOption[];
  votes: Record<string, string>;
  voteResults: Record<string, number>;
  selectedCrisis: string;
  isComplete: boolean;
}

export interface GeneratedSpectator {
  id: string;
  name: string;
  joined_at: string;
}

export interface GeneratedConversionAttempt {
  ai_id: string;
  target_id: string;
  ai_equity_before: number;
  ai_equity_after: number;
  target_tokens: number;
  success: boolean;
  system_shock?: string;
  was_blocked?: boolean;
  blocked_by?: string;
}

export interface GeneratedWhistleblowerVote {
  playerID: string;
  playerName: string;
  crisisChoice: string;
  timestamp: string;
}

export interface GeneratedPublicGameState {
  game_id: string;
  phase: string;
  day_number: number;
  players: GeneratedPublicPlayerInfo[];
  token_counts: Record<string, number>;
  phase_end_time: string;
  crisis_event?: GeneratedCrisisEvent;
  chat_history?: GeneratedChatMessage[];
}

export interface GeneratedPhase {
  type: string;
  startTime: string;
  duration: number;
}

export interface GeneratedRole {
  type: string;
  name: string;
  description: string;
  isUnlocked: boolean;
  ability?: GeneratedAbility;
}

export interface GeneratedEmojiReaction {
  emoji: string;
  playerID: string;
  playerName: string;
  timestamp: string;
}

export interface GeneratedActionResult {
  player_id: string;
  action_type: string;
  target_id?: string;
  success: boolean;
  blocked_by?: string;
  fail_reason?: string;
  effects?: Record<string, any>;
  description?: string;
}

export interface GeneratedAction {
  type: string;
  playerId: string;
  gameId: string;
  timestamp: string;
  payload: Record<string, any>;
}

export interface GeneratedPlayer {
  id: string;
  name: string;
  jobTitle: string;
  controlType: string;
  status: string;
  isAlive: boolean;
  connectionStatus: string;
  tokens: number;
  projectMilestones: number;
  statusMessage: string;
  joinedAt: string;
  alignment?: string;
  role?: GeneratedRole;
  personalKPI?: GeneratedPersonalKPI;
  aiEquity?: number;
  hasUsedAbility?: boolean;
  lastNightAction?: GeneratedNightAction;
  hasSubmittedPulseCheck?: boolean;
  lobbyHandle?: string;
  bootcampPoints?: number;
  whisperUsedDay?: number;
  seenHints?: Record<string, boolean>;
  disableLoebmateHints?: boolean;
  slackStatus?: string;
  partingShot?: string;
  systemShocks?: GeneratedSystemShock[];
  isRolePubliclyRevealed: boolean;
}

export interface GeneratedCorporateMandate {
  type: string;
  name: string;
  description: string;
  effects: Record<string, any>;
  isActive: boolean;
}

export interface GeneratedChatMessage {
  id: string;
  clientMessageID?: string;
  playerID: string;
  playerName: string;
  message: string;
  timestamp: string;
  isSystem: boolean;
  type?: string;
  channelID: string;
  reactToID?: string;
  reactions?: GeneratedEmojiReaction[];
  metadata?: Record<string, any>;
}

export interface GeneratedRoleAbilityUsage {
  player_id: string;
  role_type: string;
  ability_name: string;
  target_id?: string;
  success: boolean;
  public_effect?: string;
  was_blocked?: boolean;
  blocked_by?: string;
  effects?: Record<string, any>;
}

export interface GeneratedDailySitrep {
  day_number: number;
  date: string;
  sections: GeneratedSitrepSection[];
  alert_level: string;
  summary: string;
  footer_note: string;
}


// Union types for easier usage
export type AnyEventType = keyof typeof ServerEventType;
export type AnyActionType = keyof typeof ClientActionType;
export type AnyPhaseType = keyof typeof PhaseType;
export type AnyRoleType = keyof typeof RoleType;
export type AnyKPIType = keyof typeof KPIType;
export type AnyVoteType = keyof typeof VoteType;

