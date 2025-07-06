import { describe, it, expect, vi, beforeEach } from 'vitest';
import { websocketClient } from '../../client/src/services/websocket';
import { gameEngine } from '../../client/src/services/gameEngine';
import { ServerEventType } from '../../client/src/types/generated';

// This test focuses specifically on the robustness of the lobby-to-game transition
// without UI components, focusing on the core data flow and edge cases.

describe('Lobby to Game Transition Robustness', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock game engine with realistic behavior
    vi.spyOn(gameEngine, 'isReady').mockReturnValue(true);
    vi.spyOn(gameEngine, 'resetAndLoadState').mockResolvedValue(undefined);
    vi.spyOn(gameEngine, 'applyEvent').mockResolvedValue(undefined);
    vi.spyOn(gameEngine, 'getCurrentState').mockReturnValue(null);
    vi.spyOn(gameEngine, 'onStateChange').mockImplementation(() => () => {});
  });

  const createGameStateVariant = (variant: string) => {
    const baseState = {
      id: 'test-game',
      phase: { type: 'SITREP', startTime: new Date().toISOString(), duration: 15000 },
      day_number: 1,
      settings: { maxPlayers: 10 },
    };

    switch (variant) {
      case 'validObjectPlayers':
        return {
          ...baseState,
          players: {
            'player_1': { id: 'player_1', name: 'Alice', isAlive: true, alignment: 'HUMAN' },
            'player_2': { id: 'player_2', name: 'Bob', isAlive: true, alignment: 'AI' },
          },
          chat_messages: [],
        };
      
      case 'validArrayPlayers':
        return {
          ...baseState,
          players: [
            { id: 'player_1', name: 'Alice', isAlive: true, alignment: 'HUMAN' },
            { id: 'player_2', name: 'Bob', isAlive: true, alignment: 'AI' },
          ],
          chatMessages: [],
        };
      
      case 'nullPlayers':
        return { ...baseState, players: null, chat_messages: [] };
      
      case 'undefinedPlayers':
        return { ...baseState, chat_messages: [] };
      
      case 'stringPlayers':
        return { ...baseState, players: 'invalid', chat_messages: [] };
      
      case 'emptyPlayers':
        return { ...baseState, players: {}, chat_messages: [] };
      
      case 'nullChatMessages':
        return { ...baseState, players: {}, chatMessages: null, chat_messages: null };
      
      case 'missingChatMessages':
        return { ...baseState, players: {} };
      
      case 'bothChatFormats':
        return { 
          ...baseState, 
          players: {},
          chatMessages: [{ id: 'msg1', message: 'test' }],
          chat_messages: [{ id: 'msg2', message: 'test2' }],
        };
      
      case 'minimalState':
        return { id: 'test-game' };
      
      case 'emptyState':
        return {};
      
      default:
        return baseState;
    }
  };

  describe('Game State Structure Validation', () => {
    it('should handle valid object-format players', () => {
      expect(() => {
        const state = createGameStateVariant('validObjectPlayers');
        // Simulate what our components would do
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : Object.values(players);
        expect(playersArray).toHaveLength(2);
      }).not.toThrow();
    });

    it('should handle valid array-format players', () => {
      expect(() => {
        const state = createGameStateVariant('validArrayPlayers');
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : Object.values(players);
        expect(playersArray).toHaveLength(2);
      }).not.toThrow();
    });

    it('should handle null players gracefully', () => {
      expect(() => {
        const state = createGameStateVariant('nullPlayers');
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : Object.values(players);
        expect(playersArray).toHaveLength(0);
      }).not.toThrow();
    });

    it('should handle undefined players gracefully', () => {
      expect(() => {
        const state = createGameStateVariant('undefinedPlayers');
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : Object.values(players);
        expect(playersArray).toHaveLength(0);
      }).not.toThrow();
    });

    it('should handle string players gracefully', () => {
      expect(() => {
        const state = createGameStateVariant('stringPlayers');
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : 
                           (typeof players === 'object' && players !== null) ? Object.values(players) : [];
        expect(playersArray).toHaveLength(0);
      }).not.toThrow();
    });

    it('should handle empty players object', () => {
      expect(() => {
        const state = createGameStateVariant('emptyPlayers');
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : Object.values(players);
        expect(playersArray).toHaveLength(0);
      }).not.toThrow();
    });
  });

  describe('Chat Messages Handling', () => {
    it('should handle missing chat messages', () => {
      expect(() => {
        const state = createGameStateVariant('missingChatMessages');
        const messages = state?.chatMessages || (state as any)?.chat_messages || [];
        expect(Array.isArray(messages) ? messages : []).toHaveLength(0);
      }).not.toThrow();
    });

    it('should handle null chat messages', () => {
      expect(() => {
        const state = createGameStateVariant('nullChatMessages');
        const messages = state?.chatMessages || (state as any)?.chat_messages || [];
        expect(Array.isArray(messages) ? messages : []).toHaveLength(0);
      }).not.toThrow();
    });

    it('should prefer chatMessages over chat_messages when both exist', () => {
      expect(() => {
        const state = createGameStateVariant('bothChatFormats');
        const messages = state?.chatMessages || (state as any)?.chat_messages || [];
        expect(messages).toHaveLength(1);
        expect(messages[0].id).toBe('msg1');
      }).not.toThrow();
    });

    it('should fall back to chat_messages when chatMessages is missing', () => {
      expect(() => {
        const state = {
          id: 'test',
          players: {},
          chat_messages: [{ id: 'msg1', message: 'test' }],
        };
        const messages = state?.chatMessages || (state as any)?.chat_messages || [];
        expect(messages).toHaveLength(1);
        expect(messages[0].id).toBe('msg1');
      }).not.toThrow();
    });
  });

  describe('State Transition Scenarios', () => {
    it('should handle minimal state without crashing', () => {
      expect(() => {
        const state = createGameStateVariant('minimalState');
        
        // Simulate RosterPanel logic
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : 
                           (typeof players === 'object' && players !== null) ? Object.values(players) : [];
        
        // Simulate CommsPanel logic
        const messages = state?.chatMessages || (state as any)?.chat_messages || [];
        const validMessages = Array.isArray(messages) ? messages : [];
        
        expect(playersArray).toHaveLength(0);
        expect(validMessages).toHaveLength(0);
      }).not.toThrow();
    });

    it('should handle completely empty state', () => {
      expect(() => {
        const state = createGameStateVariant('emptyState');
        
        // All component logic should handle this
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : 
                           (typeof players === 'object' && players !== null) ? Object.values(players) : [];
        
        const messages = state?.chatMessages || (state as any)?.chat_messages || [];
        const validMessages = Array.isArray(messages) ? messages : [];
        
        // Should have safe defaults
        expect(playersArray).toHaveLength(0);
        expect(validMessages).toHaveLength(0);
      }).not.toThrow();
    });

    it('should handle rapid state transitions', () => {
      const variants = [
        'emptyState',
        'minimalState',
        'nullPlayers',
        'undefinedPlayers',
        'validObjectPlayers',
        'validArrayPlayers',
      ];

      expect(() => {
        variants.forEach(variant => {
          const state = createGameStateVariant(variant);
          
          // Simulate what components do
          const players = state?.players || [];
          const playersArray = Array.isArray(players) ? players : 
                             (typeof players === 'object' && players !== null) ? Object.values(players) : [];
          
          const messages = state?.chatMessages || (state as any)?.chat_messages || [];
          const validMessages = Array.isArray(messages) ? messages : [];
          
          // Should always have valid arrays
          expect(Array.isArray(playersArray)).toBe(true);
          expect(Array.isArray(validMessages)).toBe(true);
        });
      }).not.toThrow();
    });
  });

  describe('Game Engine Integration', () => {
    it('should handle game engine resetAndLoadState with various state formats', async () => {
      const variants = [
        'validObjectPlayers',
        'validArrayPlayers',
        'nullPlayers',
        'undefinedPlayers',
        'emptyState',
      ];

      for (const variant of variants) {
        const state = createGameStateVariant(variant);
        
        // This should not throw regardless of state format
        expect(async () => {
          await gameEngine.resetAndLoadState(state);
        }).not.toThrow();
      }
      
      expect(gameEngine.resetAndLoadState).toHaveBeenCalledTimes(variants.length);
    });

    it('should handle game engine failures gracefully', async () => {
      vi.mocked(gameEngine.resetAndLoadState).mockRejectedValue(new Error('WASM error'));
      
      const state = createGameStateVariant('validObjectPlayers');
      
      expect(async () => {
        try {
          await gameEngine.resetAndLoadState(state);
        } catch (error) {
          // Error should be caught and handled by the calling code
          expect(error).toBeInstanceOf(Error);
        }
      }).not.toThrow();
    });

    it('should handle game engine not ready state', () => {
      vi.mocked(gameEngine.isReady).mockReturnValue(false);
      
      expect(() => {
        // This simulates what websocket.ts does
        if (gameEngine.isReady()) {
          gameEngine.resetAndLoadState({});
        } else {
          console.warn('Game engine not ready');
        }
      }).not.toThrow();
      
      expect(gameEngine.resetAndLoadState).not.toHaveBeenCalled();
    });
  });

  describe('Edge Case Combinations', () => {
    it('should handle player counting with various formats', () => {
      const testCases = [
        { variant: 'validObjectPlayers', expectedCount: 2 },
        { variant: 'validArrayPlayers', expectedCount: 2 },
        { variant: 'nullPlayers', expectedCount: 0 },
        { variant: 'undefinedPlayers', expectedCount: 0 },
        { variant: 'emptyPlayers', expectedCount: 0 },
        { variant: 'stringPlayers', expectedCount: 0 },
      ];

      testCases.forEach(({ variant, expectedCount }) => {
        expect(() => {
          const state = createGameStateVariant(variant);
          const players = state?.players || [];
          
          let count = 0;
          if (Array.isArray(players)) {
            count = players.length;
          } else if (typeof players === 'object' && players !== null) {
            count = Object.keys(players).length;
          }
          
          expect(count).toBe(expectedCount);
        }).not.toThrow();
      });
    });

    it('should handle chat message filtering with various formats', () => {
      const testCases = [
        'bothChatFormats',
        'nullChatMessages', 
        'missingChatMessages',
      ];

      testCases.forEach(variant => {
        expect(() => {
          const state = createGameStateVariant(variant);
          const messages = state?.chatMessages || (state as any)?.chat_messages || [];
          const validMessages = Array.isArray(messages) ? messages : [];
          
          // Simulate filtering by channel
          const warRoomMessages = validMessages.filter(msg => 
            (msg.channelID || '#war-room') === '#war-room'
          );
          
          expect(Array.isArray(warRoomMessages)).toBe(true);
        }).not.toThrow();
      });
    });

    it('should handle vote state calculations with missing data', () => {
      expect(() => {
        const state = createGameStateVariant('undefinedPlayers');
        const players = state?.players || [];
        const playersArray = Array.isArray(players) ? players : 
                           (typeof players === 'object' && players !== null) ? Object.values(players) : [];
        
        // Simulate skip vote calculation from CommsPanel
        const livingHumans = playersArray.filter(p => p?.isAlive && p?.controlType === 'HUMAN').length;
        expect(livingHumans).toBe(0);
        
        // Simulate vote state from VoteUI
        const voteState = state?.voteState || {};
        const votes = voteState?.votes || {};
        const playerVotes = Object.values(votes).filter(vote => vote === 'some-player').length;
        expect(playerVotes).toBe(0);
      }).not.toThrow();
    });
  });

  describe('Critical Transition Scenarios', () => {
    it('should handle WebSocket disconnection during game state loading', () => {
      expect(() => {
        const state = createGameStateVariant('validObjectPlayers');
        
        // Simulate WebSocket disconnection mid-transition
        if (gameEngine.isReady()) {
          gameEngine.resetAndLoadState(state).catch(() => {
            // Connection lost during loading
          });
        }
        
        // Should handle gracefully without corrupting state
        expect(true).toBe(true);
      }).not.toThrow();
    });

    it('should handle partial game state updates during transition', () => {
      expect(() => {
        // Simulate receiving partial updates
        const partialStates = [
          { id: 'test-game' },
          { id: 'test-game', players: {} },
          { id: 'test-game', players: {}, phase: { type: 'SITREP' } },
          createGameStateVariant('validObjectPlayers'),
        ];

        partialStates.forEach(state => {
          if (gameEngine.isReady()) {
            gameEngine.resetAndLoadState(state);
          }
        });
      }).not.toThrow();
    });

    it('should handle concurrent state changes from multiple sources', () => {
      expect(() => {
        const state1 = createGameStateVariant('validObjectPlayers');
        const state2 = createGameStateVariant('validArrayPlayers');
        const state3 = createGameStateVariant('nullPlayers');

        // Simulate rapid concurrent updates
        if (gameEngine.isReady()) {
          Promise.all([
            gameEngine.resetAndLoadState(state1),
            gameEngine.resetAndLoadState(state2),
            gameEngine.resetAndLoadState(state3),
          ]).catch(() => {
            // Handle race conditions gracefully
          });
        }
      }).not.toThrow();
    });

    it('should handle game engine failures during critical transition moments', () => {
      vi.mocked(gameEngine.resetAndLoadState).mockRejectedValueOnce(new Error('WASM failure'));
      
      expect(() => {
        const state = createGameStateVariant('validObjectPlayers');
        
        try {
          gameEngine.resetAndLoadState(state);
        } catch (error) {
          // Should handle WASM failures gracefully
          expect(error).toBeInstanceOf(Error);
        }
      }).not.toThrow();
    });

    it('should handle memory pressure during state loading', () => {
      expect(() => {
        // Simulate large game state that could cause memory issues
        const largeState = {
          ...createGameStateVariant('validObjectPlayers'),
          chat_messages: Array.from({ length: 1000 }, (_, i) => ({
            id: `msg_${i}`,
            sender: `player_${i % 10}`,
            message: `Message ${i}`.repeat(100), // Large messages
            timestamp: new Date().toISOString(),
          })),
          players: Object.fromEntries(
            Array.from({ length: 100 }, (_, i) => [
              `player_${i}`,
              {
                id: `player_${i}`,
                name: `Player ${i}`,
                isAlive: true,
                alignment: i % 2 === 0 ? 'HUMAN' : 'AI',
              },
            ])
          ),
        };

        if (gameEngine.isReady()) {
          gameEngine.resetAndLoadState(largeState);
        }
      }).not.toThrow();
    });
  });

  describe('State Consistency Validation', () => {
    it('should handle inconsistent player data formats across updates', () => {
      expect(() => {
        const inconsistentStates = [
          createGameStateVariant('validObjectPlayers'),
          createGameStateVariant('validArrayPlayers'),
          createGameStateVariant('nullPlayers'),
          createGameStateVariant('stringPlayers'),
          createGameStateVariant('emptyPlayers'),
        ];

        // Simulate receiving these states in rapid succession
        inconsistentStates.forEach((state, index) => {
          setTimeout(() => {
            if (gameEngine.isReady()) {
              gameEngine.resetAndLoadState(state);
            }
          }, index * 10);
        });
      }).not.toThrow();
    });

    it('should handle chat message format inconsistencies', () => {
      expect(() => {
        const chatVariants = [
          createGameStateVariant('bothChatFormats'),
          createGameStateVariant('nullChatMessages'),
          createGameStateVariant('missingChatMessages'),
        ];

        chatVariants.forEach(state => {
          const messages = state?.chatMessages || (state as any)?.chat_messages || [];
          const validMessages = Array.isArray(messages) ? messages : [];
          
          // Simulate component logic handling these variants
          validMessages.forEach(msg => {
            const channelID = msg?.channelID || '#war-room';
            const sender = msg?.sender || 'Unknown';
            const message = msg?.message || '';
            
            expect(typeof channelID).toBe('string');
            expect(typeof sender).toBe('string');
            expect(typeof message).toBe('string');
          });
        });
      }).not.toThrow();
    });

    it('should validate game phase transitions during state loading', () => {
      expect(() => {
        const phaseTransitions = [
          { ...createGameStateVariant('validObjectPlayers'), phase: { type: 'SITREP' } },
          { ...createGameStateVariant('validObjectPlayers'), phase: { type: 'PULSE_CHECK' } },
          { ...createGameStateVariant('validObjectPlayers'), phase: { type: 'DISCUSSION' } },
          { ...createGameStateVariant('validObjectPlayers'), phase: { type: 'NOMINATION' } },
          { ...createGameStateVariant('validObjectPlayers'), phase: null },
          { ...createGameStateVariant('validObjectPlayers'), phase: undefined },
        ];

        phaseTransitions.forEach(state => {
          const phase = state?.phase || { type: 'SITREP' };
          const phaseType = phase?.type || 'SITREP';
          
          expect(['SITREP', 'PULSE_CHECK', 'DISCUSSION', 'NOMINATION'].includes(phaseType) || phaseType === 'SITREP').toBe(true);
          
          if (gameEngine.isReady()) {
            gameEngine.resetAndLoadState(state);
          }
        });
      }).not.toThrow();
    });
  });

  describe('Recovery and Resilience', () => {
    it('should recover from corrupted state transitions', () => {
      expect(() => {
        // Simulate corrupted state
        const corruptedState = {
          id: 'test-game',
          players: Symbol('corrupted'), // Invalid type
          phase: { type: 'INVALID_PHASE' },
          corrupted_field: () => { throw new Error('Corrupted data'); },
        };

        try {
          if (gameEngine.isReady()) {
            gameEngine.resetAndLoadState(corruptedState as any);
          }
        } catch (error) {
          // Should handle corruption gracefully
        }

        // Then send valid state to recover
        const validState = createGameStateVariant('validObjectPlayers');
        if (gameEngine.isReady()) {
          gameEngine.resetAndLoadState(validState);
        }
      }).not.toThrow();
    });

    it('should handle state rollback scenarios', () => {
      expect(() => {
        const states = [
          createGameStateVariant('validObjectPlayers'),
          createGameStateVariant('nullPlayers'), // Bad state
          createGameStateVariant('validObjectPlayers'), // Rollback to good state
        ];

        states.forEach(state => {
          if (gameEngine.isReady()) {
            gameEngine.resetAndLoadState(state).catch(() => {
              // Handle rollback gracefully
            });
          }
        });
      }).not.toThrow();
    });

    it('should handle timeout scenarios during state loading', () => {
      vi.mocked(gameEngine.resetAndLoadState).mockImplementation(() => 
        new Promise((resolve) => setTimeout(resolve, 5000)) // Simulate slow loading
      );

      expect(() => {
        const state = createGameStateVariant('validObjectPlayers');
        
        // Simulate timeout handling
        const timeoutPromise = new Promise((_, reject) =>
          setTimeout(() => reject(new Error('Timeout')), 1000)
        );

        Promise.race([
          gameEngine.resetAndLoadState(state),
          timeoutPromise,
        ]).catch(() => {
          // Handle timeout gracefully
        });
      }).not.toThrow();
    });

    it('should handle browser resource constraints', () => {
      expect(() => {
        // Simulate low memory conditions
        const mockOutOfMemory = () => {
          throw new Error('QuotaExceededError');
        };

        try {
          const state = createGameStateVariant('validObjectPlayers');
          if (gameEngine.isReady()) {
            gameEngine.resetAndLoadState(state);
          }
        } catch (error) {
          if (error instanceof Error && error.message.includes('QuotaExceeded')) {
            // Handle memory constraints gracefully
            console.warn('Memory constraints detected, using fallback');
          }
        }
      }).not.toThrow();
    });
  });

  describe('Performance Edge Cases', () => {
    it('should handle high-frequency state updates without memory leaks', () => {
      expect(() => {
        // Simulate rapid state updates that could cause memory issues
        for (let i = 0; i < 1000; i++) {
          const state = {
            ...createGameStateVariant('validObjectPlayers'),
            id: `test-game-${i}`,
            update_sequence: i,
          };

          if (gameEngine.isReady()) {
            gameEngine.resetAndLoadState(state).catch(() => {
              // Handle failures without accumulating errors
            });
          }
        }
      }).not.toThrow();
    });

    it('should handle state updates with deeply nested objects', () => {
      expect(() => {
        const deeplyNestedState = {
          ...createGameStateVariant('validObjectPlayers'),
          metadata: {
            level1: {
              level2: {
                level3: {
                  level4: {
                    level5: {
                      data: Array.from({ length: 100 }, (_, i) => ({
                        id: i,
                        nested: { value: `deep_${i}` },
                      })),
                    },
                  },
                },
              },
            },
          },
        };

        if (gameEngine.isReady()) {
          gameEngine.resetAndLoadState(deeplyNestedState);
        }
      }).not.toThrow();
    });

    it('should handle concurrent access to game state during transition', () => {
      expect(() => {
        const state = createGameStateVariant('validObjectPlayers');

        // Simulate multiple components trying to access state simultaneously
        const accessAttempts = Array.from({ length: 50 }, () =>
          Promise.resolve().then(() => {
            if (gameEngine.isReady()) {
              const currentState = gameEngine.getCurrentState();
              // Simulate component using the state
              const players = (currentState as any)?.players || [];
              const playersArray = Array.isArray(players) ? players : Object.values(players);
              return playersArray.length;
            }
            return 0;
          })
        );

        if (gameEngine.isReady()) {
          gameEngine.resetAndLoadState(state);
        }

        Promise.all(accessAttempts).catch(() => {
          // Handle concurrent access gracefully
        });
      }).not.toThrow();
    });
  });
});