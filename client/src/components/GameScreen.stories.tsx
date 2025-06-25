import type { Meta, StoryObj } from '@storybook/react';
import React from 'react';
import { GameScreen } from './GameScreen';
import { GameProvider } from '../contexts/GameContext';
import { GameState, Player, Phase, ChatMessage } from '../types';

// Create a wrapper component that provides contexts
const GameScreenWrapper: React.FC<{ gameState: GameState; playerID: string }> = ({ gameState, playerID }) => (
  <GameProvider gameState={gameState} localPlayerId={playerID}>
    <GameScreen />
  </GameProvider>
);

const meta: Meta<typeof GameScreenWrapper> = {
  title: 'Core/GameScreen',
  component: GameScreenWrapper,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  argTypes: {
    gameState: { control: 'object' },
    playerID: { control: 'text' },
  },
} satisfies Meta<typeof GameScreenWrapper>;

export default meta;
type Story = StoryObj<typeof meta>;

// Base players for our stories
const basePlayers: Player[] = [
  {
    id: 'p-1',
    name: 'Alice',
    jobTitle: 'Chief Security Officer',
    isAlive: true,
    tokens: 8,
    projectMilestones: 3,
    statusMessage: '"Trust the CISO"',
    alignment: 'HUMAN',
    avatar: '👤',
    joinedAt: '2024-01-01T00:00:00Z',
    role: {
      type: 'SECURITY_ANALYST',
      name: 'Security Analyst',
      description: 'Protects the company from threats',
      isUnlocked: true,
      ability: {
        name: 'Security Scan',
        description: 'Detect suspicious activity',
        isReady: true,
      },
    },
    personalKPI: {
      type: 'THREAT_MITIGATION',
      description: 'Identify and neutralize 2 security threats',
      progress: 1,
      target: 2,
      isCompleted: false,
      reward: 'Bonus tokens',
    },
  },
  {
    id: 'p-2',
    name: 'Bob',
    jobTitle: 'Senior Developer',
    isAlive: true,
    tokens: 5,
    projectMilestones: 2,
    statusMessage: '"Coffee first, code second"',
    alignment: 'HUMAN',
    avatar: '👨‍💻',
    joinedAt: '2024-01-01T00:05:00Z',
    role: {
      type: 'SOFTWARE_ENGINEER',
      name: 'Software Engineer',
      description: 'Builds and maintains systems',
      isUnlocked: true,
      ability: {
        name: 'Code Review',
        description: 'Analyze code for vulnerabilities',
        isReady: false,
      },
    },
    personalKPI: {
      type: 'CODE_QUALITY',
      description: 'Complete 3 code reviews',
      progress: 2,
      target: 3,
      isCompleted: false,
      reward: 'Promotion consideration',
    },
  },
  {
    id: 'p-3',
    name: 'Eve',
    jobTitle: 'Chief Operating Officer',
    isAlive: true,
    tokens: 12,
    projectMilestones: 4,
    statusMessage: '"Efficiency is key."',
    alignment: 'ALIGNED',
    avatar: '🧑‍🚀',
    joinedAt: '2024-01-01T00:10:00Z',
    role: {
      type: 'EXECUTIVE',
      name: 'Executive',
      description: 'Manages company operations',
      isUnlocked: true,
      ability: {
        name: 'Resource Allocation',
        description: 'Redistribute tokens between players',
        isReady: true,
      },
    },
    personalKPI: {
      type: 'OPERATIONAL_EFFICIENCY',
      description: 'Optimize 2 company processes',
      progress: 2,
      target: 2,
      isCompleted: true,
      reward: 'Executive bonus',
    },
  },
  {
    id: 'p-4',
    name: 'Charlie',
    jobTitle: 'Former Employee',
    isAlive: false,
    tokens: 0,
    projectMilestones: 1,
    statusMessage: 'Eve is the AI!',
    alignment: 'HUMAN',
    avatar: '👻',
    joinedAt: '2024-01-01T00:15:00Z',
    role: {
      type: 'DATA_ANALYST',
      name: 'Data Analyst',
      description: 'Analyzes company data',
      isUnlocked: false,
      ability: {
        name: 'Data Mining',
        description: 'Extract insights from data',
        isReady: false,
      },
    },
    personalKPI: {
      type: 'DATA_INSIGHTS',
      description: 'Generate 3 data reports',
      progress: 1,
      target: 3,
      isCompleted: false,
      reward: 'Data access privileges',
    },
  },
];

// Base chat messages
const baseChatMessages: ChatMessage[] = [
  {
    id: 'c-1',
    playerID: 'p-1',
    playerName: 'Alice',
    message: 'Good morning everyone! Ready to start the workday?',
    timestamp: '2024-01-01T09:00:00Z',
    type: 'REGULAR',
    isSystem: false,
  },
  {
    id: 'c-2',
    playerID: 'p-2',
    playerName: 'Bob',
    message: 'Just finished my coffee. Let\'s do this!',
    timestamp: '2024-01-01T09:01:00Z',
    type: 'REGULAR',
    isSystem: false,
  },
  {
    id: 'c-3',
    playerID: 'system',
    playerName: 'NEXUS',
    message: 'SITREP: All systems operational. Day 1 commencing.',
    timestamp: '2024-01-01T09:02:00Z',
    type: 'SITREP',
    isSystem: true,
    metadata: {
      playerHeadcount: {
        humans: 3,
        aligned: 1,
        dead: 0,
      },
    },
  },
];

// Base phase
const discussionPhase: Phase = {
  type: 'DISCUSSION',
  startTime: '2024-01-01T09:00:00Z',
  duration: 300000000000, // 5 minutes in nanoseconds
};

// Base game state
const baseGameState: GameState = {
  id: 'game-1',
  players: basePlayers,
  phase: discussionPhase,
  dayNumber: 1,
  chatMessages: baseChatMessages,
};

// --- Stories ---

export const Discussion: Story = {
  args: {
    gameState: baseGameState,
    playerID: 'p-1',
  },
};

export const NightPhase: Story = {
  args: {
    gameState: {
      ...baseGameState,
      phase: {
        type: 'NIGHT',
        startTime: '2024-01-01T21:00:00Z',
        duration: 120000000000, // 2 minutes in nanoseconds
      },
    },
    playerID: 'p-1',
  },
};

export const NominationPhase: Story = {
  args: {
    gameState: {
      ...baseGameState,
      phase: {
        type: 'NOMINATION',
        startTime: '2024-01-01T12:00:00Z',
        duration: 180000000000, // 3 minutes in nanoseconds
      },
      voteState: {
        type: 'NOMINATION',
        votes: {},
        tokenWeights: {},
        results: {},
        isComplete: false,
      },
    },
    playerID: 'p-1',
  },
};

export const TrialPhase: Story = {
  args: {
    gameState: {
      ...baseGameState,
      phase: {
        type: 'TRIAL',
        startTime: '2024-01-01T12:05:00Z',
        duration: 120000000000, // 2 minutes in nanoseconds
      },
      nominatedPlayer: 'p-3',
      voteState: {
        type: 'TRIAL',
        votes: {},
        tokenWeights: {},
        results: {},
        isComplete: false,
      },
    },
    playerID: 'p-1',
  },
};

export const VerdictPhase: Story = {
  args: {
    gameState: {
      ...baseGameState,
      phase: {
        type: 'VERDICT',
        startTime: '2024-01-01T12:10:00Z',
        duration: 60000000000, // 1 minute in nanoseconds
      },
      nominatedPlayer: 'p-3',
      voteState: {
        type: 'VERDICT',
        votes: {
          'p-1': 'GUILTY',
          'p-2': 'INNOCENT',
        },
        tokenWeights: {
          'p-1': 8,
          'p-2': 5,
        },
        results: {
          GUILTY: 8,
          INNOCENT: 5,
        },
        isComplete: false,
      },
    },
    playerID: 'p-1',
  },
};

export const VerdictComplete: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: basePlayers.map(p => p.id === 'p-3' ? { ...p, isAlive: false } : p),
      phase: {
        type: 'DISCUSSION',
        startTime: '2024-01-01T12:11:00Z',
        duration: 300000000000,
      },
      chatMessages: [
        ...baseChatMessages,
        {
          id: 'c-4',
          playerID: 'system',
          playerName: 'NEXUS',
          message: 'Vote result for deactivating Eve',
          timestamp: '2024-01-01T12:11:00Z',
          type: 'VOTE_RESULT',
          isSystem: true,
          metadata: {
            voteResult: {
              question: 'Deactivate Eve?',
              outcome: 'Eve has been deactivated.',
              votes: {
                'p-1': 'GUILTY',
                'p-2': 'GUILTY',
              },
              tokenWeights: {
                'p-1': 8,
                'p-2': 5,
              },
              results: {
                GUILTY: 13,
                INNOCENT: 0,
              },
              eliminatedPlayer: {
                id: 'p-3',
                name: 'Eve',
                role: 'Chief Operating Officer',
                alignment: 'AI',
              },
            },
          },
        },
      ],
    },
    playerID: 'p-1',
  },
};

export const PulseCheckPhase: Story = {
  args: {
    gameState: {
      ...baseGameState,
      phase: {
        type: 'PULSE_CHECK',
        startTime: '2024-01-01T11:00:00Z',
        duration: 60000000000, // 1 minute in nanoseconds
      },
      chatMessages: [
        ...baseChatMessages,
        {
          id: 'c-4',
          playerID: 'system',
          playerName: 'NEXUS',
          message: 'PULSE CHECK: Respond with your current status.',
          timestamp: '2024-01-01T11:00:00Z',
          type: 'PULSE_CHECK',
          isSystem: true,
        },
      ],
    },
    playerID: 'p-1',
  },
};

export const AsAIPlayer: Story = {
  args: {
    gameState: baseGameState,
    playerID: 'p-3', // Eve (AI-aligned player)
  },
};

export const AsDeadPlayer: Story = {
  args: {
    gameState: baseGameState,
    playerID: 'p-4', // Charlie (dead player)
  },
};

export const WithCrisisEvent: Story = {
  args: {
    gameState: {
      ...baseGameState,
      crisisEvent: {
        type: 'SYSTEM_BREACH',
        title: 'Security Breach Detected',
        description: 'Unauthorized access detected in the main database.',
        effects: {
          tokenLoss: 2,
          communicationDisruption: true,
        },
      },
    },
    playerID: 'p-1',
  },
};

export const WithPrivateNotifications: Story = {
  args: {
    gameState: {
      ...baseGameState,
      privateNotifications: [
        {
          id: 'notif-1',
          type: 'role_ability',
          title: 'Ability Ready',
          message: 'Your Security Scan ability is now ready to use.',
          timestamp: '2024-01-01T09:30:00Z',
          isRead: false,
          priority: 'medium',
        },
        {
          id: 'notif-2',
          type: 'kpi_progress',
          title: 'KPI Progress',
          message: 'You are 50% complete with your personal objective.',
          timestamp: '2024-01-01T09:15:00Z',
          isRead: false,
          priority: 'low',
        },
      ],
    },
    playerID: 'p-1',
  },
};

export const ChatHistoryLoading: Story = {
  args: {
    gameState: {
      ...baseGameState,
      chatMessages: [],
    },
    playerID: 'p-1',
  },
};

export const GameOver: Story = {
  args: {
    gameState: {
      ...baseGameState,
      phase: {
        type: 'GAME_OVER',
        startTime: '2024-01-01T15:00:00Z',
        duration: 0,
      },
      winCondition: {
        winner: 'AI',
        condition: 'AI_TAKEOVER',
        description: 'The AI has successfully converted enough humans to take control.',
      },
    },
    playerID: 'p-1',
  },
};
