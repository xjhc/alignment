import React from 'react';
import { GameProvider, GameContextType } from '../../contexts/GameContext';
import { SessionProvider, SessionContextType } from '../../contexts/SessionContext';
import { GameState } from '../../types';
import { PhaseType, VoteType } from '../../types/generated';
import { SessionState, LobbyState, GameUIState } from '../../state/appReducer';

// Helper to create a default game state for stories
const createMockGameState = (overrides: Partial<GameState> = {}): GameState => ({
  id: 'story-game-1',
  players: [],
  phase: { 
    type: PhaseType.Discussion, 
    startTime: new Date().toISOString(), 
    duration: 180000000000 
  },
  dayNumber: 1,
  chatMessages: [],
  voteState: { 
    votes: {}, 
    tokenWeights: {}, 
    results: {}, 
    isComplete: false, 
    type: VoteType.Nomination 
  },
  ...overrides,
});

// Create a default mock that satisfies the full context type
const createMockGameContext = (gameState: GameState, localPlayerId: string): GameContextType => {
  const localPlayer = gameState.players.find(p => p.id === localPlayerId) || null;
  return {
    gameState,
    localPlayerId,
    viewedPlayerId: localPlayerId,
    localPlayer,
    viewedPlayer: localPlayer,
    isConnected: true,
    activeChannel: '#war-room',
    skipVoteState: { currentVotes: 1, requiredVotes: 5, voters: [] },
    // Mock all functions to prevent "is not a function" errors
    sendAction: (action) => console.log('[Mock SendAction]', action),
    setViewedPlayer: (id) => console.log('[Mock setViewedPlayer]', id),
    setActiveChannel: (id) => console.log('[Mock setActiveChannel]', id),
    chatInput: '',
    setChatInput: () => {},
    selectedNominee: '',
    setSelectedNominee: () => {},
    selectedVote: '',
    setSelectedVote: () => {},
    conversionTarget: '',
    setConversionTarget: () => {},
    miningTarget: '',
    setMiningTarget: () => {},
    replyingTo: null,
    // All other functions are mocked to do nothing but log to console
    handleSendMessage: async () => console.log('handleSendMessage'),
    handleMineTokens: async () => console.log('handleMineTokens'),
    handleUseAbility: async () => console.log('handleUseAbility'),
    handleProjectMilestones: async () => console.log('handleProjectMilestones'),
    handleConversionAttempt: async () => console.log('handleConversionAttempt'),
    handleNominate: async () => console.log('handleNominate'),
    handleVote: async () => console.log('handleVote'),
    handleExtensionVote: async () => console.log('handleExtensionVote'),
    handlePulseCheck: async () => console.log('handlePulseCheck'),
    handleKeyDown: () => {},
    startReply: () => {},
    cancelReply: () => {},
    handleSkipPhase: async () => console.log('handleSkipPhase'),
    handleEmojiReaction: async () => console.log('handleEmojiReaction'),
    handleSubmitPartingShot: async () => console.log('handleSubmitPartingShot'),
    submitWhistleblowerVote: async () => console.log('submitWhistleblowerVote'),
    getPhaseDisplayName: (phase) => phase.replace('_', ' '),
    canPlayerAffordAbility: () => true,
    isValidNightActionTarget: () => true,
    pendingMessages: {},
    getPendingMessagesForChannel: () => [],
    retryMessage: () => {},
    rateLimitError: null,
    getBufferStatus: () => ({ bufferLength: 0, hasPendingMessages: false, nextFlushETA: 0 }),
  };
};

// Create a default mock for SessionContext
const createMockSessionContext = (localPlayerId: string, overrides: Partial<SessionContextType> = {}): SessionContextType => ({
  appState: { 
    isSpectating: false, 
    playerId: localPlayerId,
    playerName: 'Story Player',
    sessionChecked: true
  },
  sessionState: 'IN_GAME' as SessionState,
  lobbyState: { 
    playerInfos: [], 
    gameSettings: {}, 
    isHost: false, 
    canStart: false,
    hostId: 'host-1',
    lobbyName: 'Story Lobby',
    maxPlayers: 8,
    connectionError: null,
    countdown: null
  } as LobbyState,
  gameState: createMockGameState(),
  roleAssignment: null,
  gameAnalysis: null,
  gameUIState: { 
    viewedPlayerId: null,
    activeChannel: '#war-room',
    chatInput: '',
    selectedNominee: '',
    selectedVote: '',
    conversionTarget: '',
    miningTarget: '',
    replyingTo: null,
    settingsModalOpen: false,
    skipVoteState: null
  } as GameUIState,
  isConnected: true,
  dispatch: () => {},
  gameEngineLoading: false,
  gameEngineError: null,
  onLogin: () => {},
  onJoinLobby: () => {},
  onCreateGame: () => {},
  onSpectateGame: () => {},
  onBackToLogin: () => {},
  onStartGame: () => {},
  onLeaveLobby: () => {},
  onEnterGame: () => {},
  onViewAnalysis: () => {},
  onPlayAgain: () => {},
  onBackToResults: () => {},
  ...overrides,
});

interface StorybookProvidersProps {
  children: React.ReactNode;
  // Allow stories to pass in custom game state
  storyGameState?: Partial<GameState>;
  localPlayerId?: string;
  // Allow stories to pass in custom session context
  storySessionContext?: Partial<SessionContextType>;
}

// The "Super Provider" component
export const StorybookProviders: React.FC<StorybookProvidersProps> = ({
  children,
  storyGameState,
  localPlayerId = 'p-1',
  storySessionContext,
}) => {
  // Create the final, merged game state for the story
  const gameState = createMockGameState(storyGameState);
  
  // Create the mock context values
  const gameContextValue = createMockGameContext(gameState, localPlayerId);
  const sessionContextValue = createMockSessionContext(localPlayerId, storySessionContext);

  return (
    <SessionProvider value={sessionContextValue}>
      <GameProvider value={gameContextValue}>
        {children}
      </GameProvider>
    </SessionProvider>
  );
};