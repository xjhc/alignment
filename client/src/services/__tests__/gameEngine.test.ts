import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { GameEngine } from '../gameEngine';
import { GeneratedEvent } from '../../types/generated';
import { GameState } from '../../types';
import { ServerEventType, PhaseType, VoteType } from '../../types/generated';

// Mock the WASM loader
let mockStateChangeCallback: ((stateJson: string) => void) | null = null;

vi.mock('../wasmLoader', () => ({
  wasmLoader: {
    load: vi.fn().mockResolvedValue(undefined),
    getCore: vi.fn().mockReturnValue({
      applyEvent: vi.fn().mockImplementation(() => {
        // Simulate state change after applying event
        if (mockStateChangeCallback) {
          const updatedState = '{"id":"test-game","players":[{"id":"player1","name":"Alice","isAlive":true,"tokens":5},{"id":"player2","name":"Bob","isAlive":true,"tokens":3}],"phase":{"type":"VERDICT","startTime":"2024-01-01T00:00:00Z","duration":300},"dayNumber":2,"chatMessages":[],"voteState":{"type":"VERDICT","votes":{"player1":"player2","player2":"player1"},"tokenWeights":{"player1":5,"player2":3},"results":{"player1":3,"player2":5},"isComplete":false}}';
          mockStateChangeCallback(updatedState);
        }
        return { success: true };
      }),
      createGame: vi.fn().mockReturnValue({ success: true }),
      getGameState: vi.fn().mockReturnValue('{"id":"test-game","players":[],"phase":{"type":"DAY","startTime":"2024-01-01T00:00:00Z","duration":300},"dayNumber":1,"chatMessages":[],"voteState":{"type":"NOMINATION","votes":{},"tokenWeights":{},"results":{},"isComplete":false}}'),
      deserializeGameState: vi.fn().mockReturnValue({ success: true }),
    }),
    onStateChange: vi.fn().mockImplementation((callback) => {
      mockStateChangeCallback = callback;
    }),
    isReady: vi.fn().mockReturnValue(true),
  },
  AlignmentCore: class {},
}));

describe('GameEngine', () => {
  let gameEngine: GameEngine;
  let mockCore: any;

  beforeEach(async () => {
    // Clear all mocks
    vi.clearAllMocks();
    
    // Create a fresh instance for each test
    gameEngine = new GameEngine();
    await gameEngine.initialize();
    
    // Get the mock core for assertions
    const { wasmLoader } = await import('../wasmLoader');
    mockCore = wasmLoader.getCore();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should handle VOTE_TALLY_UPDATED event and update voteState', async () => {
    // Arrange
    const initialState: GameState = {
      id: 'test-game',
      players: [
        { id: 'player1', name: 'Alice', jobTitle: 'Employee', controlType: 'HUMAN', isAlive: true, tokens: 5, projectMilestones: 0, statusMessage: '', joinedAt: '2024-01-01T00:00:00Z' },
        { id: 'player2', name: 'Bob', jobTitle: 'Employee', controlType: 'HUMAN', isAlive: true, tokens: 3, projectMilestones: 0, statusMessage: '', joinedAt: '2024-01-01T00:00:00Z' },
      ],
      phase: { type: PhaseType.Verdict, startTime: '2024-01-01T00:00:00Z', duration: 300 },
      dayNumber: 2,
      chatMessages: [],
      voteState: {
        type: VoteType.Verdict,
        votes: { 'player1': 'player2' },
        tokenWeights: { 'player1': 5 },
        results: { 'player2': 5 },
        isComplete: false,
      },
    };

    // Mock the core to return updated state after applying event
    const updatedStateJson = JSON.stringify({
      ...initialState,
      voteState: {
        type: VoteType.Verdict,
        votes: { 'player1': 'player2', 'player2': 'player1' },
        tokenWeights: { 'player1': 5, 'player2': 3 },
        results: { 'player1': 3, 'player2': 5 },
        isComplete: false,
      },
    });
    
    mockCore.getGameState.mockReturnValue(updatedStateJson);

    // Set up state change listener to capture the update
    let capturedState: GameState | null = null;
    const unsubscribe = gameEngine.onStateChange((state) => {
      capturedState = state;
    });

    // Load initial state
    await gameEngine.loadState(initialState);

    // Act - Apply VOTE_TALLY_UPDATED event
    const voteTallyEvent: GeneratedEvent = {
      id: 'event-123',
      type: ServerEventType.VoteTallyUpdated,
      gameId: 'test-game',
      playerId: 'player2',
      timestamp: '2024-01-01T00:05:00Z',
      payload: {
        voteState: {
          type: VoteType.Verdict,
          votes: { 'player1': 'player2', 'player2': 'player1' },
          tokenWeights: { 'player1': 5, 'player2': 3 },
          results: { 'player1': 3, 'player2': 5 },
          isComplete: false,
        },
      },
    };

    await gameEngine.applyEvent(voteTallyEvent);

    // Assert
    expect(mockCore.applyEvent).toHaveBeenCalledWith(JSON.stringify(voteTallyEvent));
    expect(mockCore.applyEvent).toHaveBeenCalledTimes(1);
    expect(capturedState).not.toBeNull();
    expect((capturedState as any)?.voteState).toEqual({
      type: VoteType.Verdict,
      votes: { 'player1': 'player2', 'player2': 'player1' },
      tokenWeights: { 'player1': 5, 'player2': 3 },
      results: { 'player1': 3, 'player2': 5 },
      isComplete: false,
    });

    // Cleanup
    unsubscribe();
  });

  it('should handle VOTE_TALLY_UPDATED event with vote completion', async () => {
    // Mock the applyEvent to return completed vote state
    mockCore.applyEvent.mockImplementation(() => {
      if (mockStateChangeCallback) {
        const completedState = '{"id":"test-game","players":[{"id":"player1","name":"Alice","isAlive":true,"tokens":5},{"id":"player2","name":"Bob","isAlive":true,"tokens":3}],"phase":{"type":"VERDICT","startTime":"2024-01-01T00:00:00Z","duration":300},"dayNumber":2,"chatMessages":[],"voteState":{"type":"VERDICT","votes":{"player1":"player2","player2":"player1"},"tokenWeights":{"player1":5,"player2":3},"results":{"player1":3,"player2":5},"isComplete":true}}';
        mockStateChangeCallback(completedState);
      }
      return { success: true };
    });

    // Set up state change listener
    let capturedState: GameState | null = null;
    const unsubscribe = gameEngine.onStateChange((state) => {
      capturedState = state;
    });

    // Act - Apply VOTE_TALLY_UPDATED event that completes the vote
    const voteTallyEvent: GeneratedEvent = {
      id: 'event-124',
      type: ServerEventType.VoteTallyUpdated,
      gameId: 'test-game',
      playerId: 'system',
      timestamp: '2024-01-01T00:05:00Z',
      payload: {
        voteState: {
          type: VoteType.Verdict,
          votes: { 'player1': 'player2', 'player2': 'player1' },
          tokenWeights: { 'player1': 5, 'player2': 3 },
          results: { 'player1': 3, 'player2': 5 },
          isComplete: true,
        },
      },
    };

    await gameEngine.applyEvent(voteTallyEvent);

    // Assert
    expect(mockCore.applyEvent).toHaveBeenCalledWith(JSON.stringify(voteTallyEvent));
    expect((capturedState as any)?.voteState?.isComplete).toBe(true);
    expect((capturedState as any)?.voteState?.results).toEqual({ 'player1': 3, 'player2': 5 });

    // Cleanup
    unsubscribe();
  });

  it('should handle errors when applying VOTE_TALLY_UPDATED event', async () => {
    // Arrange
    mockCore.applyEvent.mockReturnValue({ success: false, error: 'Invalid vote state' });

    const voteTallyEvent: GeneratedEvent = {
      id: 'event-125',
      type: ServerEventType.VoteTallyUpdated,
      gameId: 'test-game',
      playerId: 'player1',
      timestamp: '2024-01-01T00:05:00Z',
      payload: {
        voteState: {
          type: 'INVALID_TYPE' as any,
          votes: {},
          tokenWeights: {},
          results: {},
          isComplete: false,
        },
      },
    };

    // Act & Assert
    await expect(gameEngine.applyEvent(voteTallyEvent)).rejects.toThrow('Invalid vote state');
    expect(mockCore.applyEvent).toHaveBeenCalledWith(JSON.stringify(voteTallyEvent));
  });
});