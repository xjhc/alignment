import { describe, it, expect } from 'vitest';
import { appReducer, initialAppState, ConsolidatedAppState } from '../appReducer';
import { VoteState, VoteType } from '../../types';

describe('appReducer', () => {
  describe('VOTE_TALLY_UPDATED action', () => {
    it('should update voteState when VOTE_TALLY_UPDATED action is dispatched', () => {
      // Arrange
      const mockVoteState: VoteState = {
        type: VoteType.Nomination,
        votes: {
          'player-1': 'player-2',
          'player-3': 'player-2',
          'player-4': 'player-5'
        },
        tokenWeights: {
          'player-1': 2,
          'player-3': 1,
          'player-4': 3
        },
        results: {
          'player-2': 3, // 2 + 1 token weights
          'player-5': 3  // 3 token weights
        },
        isComplete: false
      };

      const action = {
        type: 'VOTE_TALLY_UPDATED' as const,
        payload: { voteState: mockVoteState }
      };

      // Act
      const result = appReducer(initialAppState, action);

      // Assert
      expect(result.gameState.voteState).toEqual(mockVoteState);
      expect(result.gameState.voteState?.votes).toEqual({
        'player-1': 'player-2',
        'player-3': 'player-2',
        'player-4': 'player-5'
      });
      expect(result.gameState.voteState?.tokenWeights).toEqual({
        'player-1': 2,
        'player-3': 1,
        'player-4': 3
      });
      expect(result.gameState.voteState?.results).toEqual({
        'player-2': 3,
        'player-5': 3
      });
      expect(result.gameState.voteState?.isComplete).toBe(false);
    });

    it('should preserve other gameState properties when updating voteState', () => {
      // Arrange
      const stateWithExistingData: ConsolidatedAppState = {
        ...initialAppState,
        gameState: {
          ...initialAppState.gameState,
          id: 'test-game-123',
          dayNumber: 3,
          chatMessages: [
            {
              id: 'msg-1',
              message: 'Test message',
              playerID: 'player-1',
              playerName: 'TestPlayer',
              timestamp: '2024-01-01T12:00:00Z',
              isSystem: false
            }
          ]
        }
      };

      const mockVoteState: VoteState = {
        type: VoteType.Verdict,
        votes: { 'player-1': 'GUILTY' },
        tokenWeights: { 'player-1': 1 },
        results: { 'GUILTY': 1 },
        isComplete: true
      };

      const action = {
        type: 'VOTE_TALLY_UPDATED' as const,
        payload: { voteState: mockVoteState }
      };

      // Act
      const result = appReducer(stateWithExistingData, action);

      // Assert
      expect(result.gameState.voteState).toEqual(mockVoteState);
      expect(result.gameState.id).toBe('test-game-123');
      expect(result.gameState.dayNumber).toBe(3);
      expect(result.gameState.chatMessages).toHaveLength(1);
      expect(result.gameState.chatMessages[0].message).toBe('Test message');
    });

    it('should handle updating from undefined voteState to a defined voteState', () => {
      // Arrange
      const stateWithoutVoteState: ConsolidatedAppState = {
        ...initialAppState,
        gameState: {
          ...initialAppState.gameState,
          voteState: undefined
        }
      };

      const mockVoteState: VoteState = {
        type: VoteType.Nomination,
        votes: {},
        tokenWeights: {},
        results: {},
        isComplete: false
      };

      const action = {
        type: 'VOTE_TALLY_UPDATED' as const,
        payload: { voteState: mockVoteState }
      };

      // Act
      const result = appReducer(stateWithoutVoteState, action);

      // Assert
      expect(result.gameState.voteState).toEqual(mockVoteState);
      expect(result.gameState.voteState?.isComplete).toBe(false);
    });

    it('should handle replacing existing voteState with new voteState', () => {
      // Arrange
      const existingVoteState: VoteState = {
        type: VoteType.Nomination,
        votes: { 'player-1': 'player-2' },
        tokenWeights: { 'player-1': 1 },
        results: { 'player-2': 1 },
        isComplete: false
      };

      const stateWithExistingVoteState: ConsolidatedAppState = {
        ...initialAppState,
        gameState: {
          ...initialAppState.gameState,
          voteState: existingVoteState
        }
      };

      const newVoteState: VoteState = {
        type: VoteType.Verdict,
        votes: { 
          'player-1': 'GUILTY',
          'player-2': 'INNOCENT'
        },
        tokenWeights: { 
          'player-1': 2,
          'player-2': 1
        },
        results: { 
          'GUILTY': 2,
          'INNOCENT': 1
        },
        isComplete: true
      };

      const action = {
        type: 'VOTE_TALLY_UPDATED' as const,
        payload: { voteState: newVoteState }
      };

      // Act
      const result = appReducer(stateWithExistingVoteState, action);

      // Assert
      expect(result.gameState.voteState).toEqual(newVoteState);
      expect(result.gameState.voteState?.type).toBe(VoteType.Verdict);
      expect(result.gameState.voteState?.isComplete).toBe(true);
      expect(result.gameState.voteState?.results).toEqual({
        'GUILTY': 2,
        'INNOCENT': 1
      });
    });
  });
});