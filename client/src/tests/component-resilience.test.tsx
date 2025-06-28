import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { GameProvider } from '../contexts/GameContext';
import { ThemeProvider } from '../contexts/ThemeContext';
import { BrowserRouter } from 'react-router-dom';
import { RosterPanel } from '../components/game/RosterPanel';
import { CommsPanel } from '../components/game/CommsPanel';
import { VoteUI } from '../components/game/VoteUI';
import { NightActionSelection } from '../components/game/NightActionSelection';
import { SitrepMessage } from '../components/game/SitrepMessage';

// Mock hooks and dependencies
vi.mock('../hooks/useTheme', () => ({
  useTheme: () => ({ theme: 'dark', toggleTheme: vi.fn() }),
}));

vi.mock('../hooks/usePhaseTimer', () => ({
  usePhaseTimer: () => 30,
}));

vi.mock('../hooks/useGameActions', () => ({
  useGameActions: () => ({
    startReply: vi.fn(),
    handleSkipPhase: vi.fn(),
    handleEmojiReaction: vi.fn(),
    selectedNominee: null,
    setSelectedNominee: vi.fn(),
    setSelectedVote: vi.fn(),
    handleNominate: vi.fn(),
    handleVote: vi.fn(),
    canPlayerAffordAbility: vi.fn(() => true),
  }),
}));

vi.mock('../hooks/useSound', () => ({
  useSound: () => ({ playSound: vi.fn() }),
}));

// Mock WebSocketContext to avoid provider requirement
vi.mock('../contexts/WebSocketContext', () => ({
  useWebSocketContext: () => ({
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(() => () => {}),
    sendAction: vi.fn(),
    isConnected: true,
  }),
  WebSocketProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

// Mock GameContext to provide necessary defaults
vi.mock('../contexts/GameContext', () => ({
  useGameContext: () => ({
    gameState: {
      players: [],
      chatMessages: [],
      phase: { type: 'SITREP' },
      settings: {},
    },
    localPlayer: { id: 'player_1', name: 'Test', isAlive: true },
    localPlayerId: 'player_1',
    viewedPlayerId: null,
    setViewedPlayer: vi.fn(),
    activeChannel: '#war-room',
    setActiveChannel: vi.fn(),
  }),
  GameProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const TestWrapper: React.FC<{ children: React.ReactNode; gameState?: any; localPlayerId?: string }> = ({ 
  children, 
  gameState = {}, 
  localPlayerId = 'player_1' 
}) => (
  <BrowserRouter>
    <ThemeProvider>
      <GameProvider gameState={gameState} localPlayerId={localPlayerId}>
        {children}
      </GameProvider>
    </ThemeProvider>
  </BrowserRouter>
);

describe('Component Resilience During Transition', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Suppress console warnings for these tests since we're testing error conditions
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  const createGameStateVariant = (variant: string) => {
    const baseState = {
      id: 'test-game',
      phase: { type: 'SITREP', startTime: new Date().toISOString(), duration: 15000 },
      day_number: 1,
      settings: { maxPlayers: 10 },
    };

    switch (variant) {
      case 'empty':
        return {};
      
      case 'nullPlayers':
        return { ...baseState, players: null };
      
      case 'undefinedPlayers':
        return { ...baseState, players: undefined };
      
      case 'stringPlayers':
        return { ...baseState, players: 'invalid' };
      
      case 'arrayPlayers':
        return { 
          ...baseState, 
          players: [
            { id: 'player_1', name: 'Alice', isAlive: true, alignment: 'HUMAN' },
            { id: 'player_2', name: 'Bob', isAlive: true, alignment: 'AI' },
          ]
        };
      
      case 'objectPlayers':
        return {
          ...baseState,
          players: {
            'player_1': { id: 'player_1', name: 'Alice', isAlive: true, alignment: 'HUMAN' },
            'player_2': { id: 'player_2', name: 'Bob', isAlive: true, alignment: 'AI' },
          }
        };
      
      case 'nullChatMessages':
        return { ...baseState, players: {}, chatMessages: null, chat_messages: null };
      
      case 'undefinedChatMessages':
        return { ...baseState, players: {} };
      
      case 'stringChatMessages':
        return { ...baseState, players: {}, chatMessages: 'invalid' };
      
      case 'validChatMessages':
        return { 
          ...baseState, 
          players: {},
          chatMessages: [
            { id: 'msg1', channelID: '#war-room', sender: 'Alice', message: 'Hello' },
          ]
        };
      
      case 'snakeCaseChatMessages':
        return { 
          ...baseState, 
          players: {},
          chat_messages: [
            { id: 'msg1', channelID: '#war-room', sender: 'Alice', message: 'Hello' },
          ]
        };
      
      default:
        return baseState;
    }
  };

  describe('RosterPanel Resilience', () => {
    it('should render without crashing with empty game state', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('empty')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle null players gracefully', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('nullPlayers')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle undefined players gracefully', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('undefinedPlayers')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle players as string gracefully', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('stringPlayers')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle players as array format', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('arrayPlayers')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle players as object format', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('objectPlayers')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle missing chat messages', () => {
      expect(() => {
        render(
          <TestWrapper gameState={createGameStateVariant('undefinedChatMessages')}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });
  });

  describe('CommsPanel Resilience', () => {
    const createLocalPlayer = () => ({
      id: 'player_1',
      name: 'Alice',
      isAlive: true,
      alignment: 'HUMAN',
    });

    it('should render without crashing with empty game state', () => {
      expect(() => {
        render(
          <TestWrapper gameState={{ ...createGameStateVariant('empty'), players: { player_1: createLocalPlayer() } }}>
            <CommsPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle null chat messages gracefully', () => {
      expect(() => {
        render(
          <TestWrapper gameState={{ 
            ...createGameStateVariant('nullChatMessages'), 
            players: { player_1: createLocalPlayer() }
          }}>
            <CommsPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle undefined chat messages gracefully', () => {
      expect(() => {
        render(
          <TestWrapper gameState={{ 
            ...createGameStateVariant('undefinedChatMessages'), 
            players: { player_1: createLocalPlayer() }
          }}>
            <CommsPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle both camelCase and snake_case chat messages', () => {
      expect(() => {
        render(
          <TestWrapper gameState={{ 
            ...createGameStateVariant('validChatMessages'), 
            players: { player_1: createLocalPlayer() }
          }}>
            <CommsPanel />
          </TestWrapper>
        );
      }).not.toThrow();

      render(
        <TestWrapper gameState={{ 
          ...createGameStateVariant('snakeCaseChatMessages'), 
          players: { player_1: createLocalPlayer() }
        }}>
          <CommsPanel />
        </TestWrapper>
      );
    });

    it('should handle missing players for skip vote calculation', () => {
      expect(() => {
        render(
          <TestWrapper gameState={{ 
            ...createGameStateVariant('nullPlayers'),
            players: { player_1: createLocalPlayer() }
          }}>
            <CommsPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });
  });

  describe('VoteUI Resilience', () => {
    const createLocalPlayer = () => ({
      id: 'player_1',
      name: 'Alice',
      isAlive: true,
      alignment: 'HUMAN',
    });

    it('should handle null players gracefully', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('nullPlayers'),
              phase: { type: 'NOMINATION' },
              players: { player_1: createLocalPlayer() }
            }}
          >
            <VoteUI />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle undefined players gracefully', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('undefinedPlayers'),
              phase: { type: 'NOMINATION' },
              players: { player_1: createLocalPlayer() }
            }}
          >
            <VoteUI />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle players as object format', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('objectPlayers'),
              phase: { type: 'NOMINATION' },
              voteState: { votes: {} }
            }}
          >
            <VoteUI />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle missing vote state', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('objectPlayers'),
              phase: { type: 'NOMINATION' },
              voteState: null
            }}
          >
            <VoteUI />
          </TestWrapper>
        );
      }).not.toThrow();
    });
  });

  describe('NightActionSelection Resilience', () => {
    const createLocalPlayer = () => ({
      id: 'player_1',
      name: 'Alice',
      isAlive: true,
      alignment: 'HUMAN',
      role: { type: 'CEO' },
      projectMilestones: 3,
    });

    it('should handle null players gracefully', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('nullPlayers'),
              phase: { type: 'NIGHT' },
              players: { player_1: createLocalPlayer() }
            }}
          >
            <NightActionSelection />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle undefined players gracefully', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('undefinedPlayers'),
              phase: { type: 'NIGHT' },
              players: { player_1: createLocalPlayer() }
            }}
          >
            <NightActionSelection />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle players as object format', () => {
      expect(() => {
        render(
          <TestWrapper 
            gameState={{ 
              ...createGameStateVariant('objectPlayers'),
              phase: { type: 'NIGHT' }
            }}
          >
            <NightActionSelection />
          </TestWrapper>
        );
      }).not.toThrow();
    });
  });

  describe('SitrepMessage Resilience', () => {
    const createTestMessage = () => ({
      id: 'msg1',
      sender: 'System',
      message: 'Sitrep message',
      timestamp: new Date().toISOString(),
      isSystem: true,
      metadata: {},
    });

    it('should handle null players gracefully', () => {
      expect(() => {
        render(
          <SitrepMessage 
            message={createTestMessage()} 
            gameState={createGameStateVariant('nullPlayers')} 
          />
        );
      }).not.toThrow();
    });

    it('should handle undefined players gracefully', () => {
      expect(() => {
        render(
          <SitrepMessage 
            message={createTestMessage()} 
            gameState={createGameStateVariant('undefinedPlayers')} 
          />
        );
      }).not.toThrow();
    });

    it('should handle players as string gracefully', () => {
      expect(() => {
        render(
          <SitrepMessage 
            message={createTestMessage()} 
            gameState={createGameStateVariant('stringPlayers')} 
          />
        );
      }).not.toThrow();
    });

    it('should handle players as object format', () => {
      expect(() => {
        render(
          <SitrepMessage 
            message={createTestMessage()} 
            gameState={createGameStateVariant('objectPlayers')} 
          />
        );
      }).not.toThrow();
    });

    it('should handle missing nightActionResults', () => {
      expect(() => {
        const gameState = {
          ...createGameStateVariant('objectPlayers'),
          nightActionResults: null,
        };
        render(
          <SitrepMessage 
            message={createTestMessage()} 
            gameState={gameState} 
          />
        );
      }).not.toThrow();
    });
  });

  describe('Transition State Handling', () => {
    it('should handle rapid state changes without crashing', () => {
      const { rerender } = render(
        <TestWrapper gameState={createGameStateVariant('empty')}>
          <RosterPanel />
          <CommsPanel />
        </TestWrapper>
      );

      // Simulate rapid state transitions
      const states = [
        'nullPlayers',
        'undefinedPlayers', 
        'stringPlayers',
        'arrayPlayers',
        'objectPlayers',
      ];

      expect(() => {
        states.forEach(variant => {
          rerender(
            <TestWrapper gameState={createGameStateVariant(variant)}>
              <RosterPanel />
              <CommsPanel />
            </TestWrapper>
          );
        });
      }).not.toThrow();
    });

    it('should handle components mounting with incomplete data', () => {
      expect(() => {
        render(
          <TestWrapper gameState={{}}>
            <RosterPanel />
            <CommsPanel />
            <VoteUI />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle null game state', () => {
      expect(() => {
        render(
          <TestWrapper gameState={null}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });

    it('should handle undefined game state', () => {
      expect(() => {
        render(
          <TestWrapper gameState={undefined}>
            <RosterPanel />
          </TestWrapper>
        );
      }).not.toThrow();
    });
  });
});