import type { Meta, StoryObj } from '@storybook/react';
import { GameScreen } from './GameScreen';
import { GameState, Player, Phase, ChatMessage } from '../types';

const meta: Meta<typeof GameScreen> = {
  title: 'Core/GameScreen',
  component: GameScreen,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  argTypes: {
    gameState: { control: 'object' },
    playerId: { control: 'text' },
    isChatHistoryLoading: { control: 'boolean' },
  },
} satisfies Meta<typeof GameScreen>;

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
    playerId: 'p-1',
    playerName: 'Alice',
    message: 'Good morning everyone! Ready to start the workday?',
    timestamp: '2024-01-01T09:00:00Z',
    type: 'REGULAR',
  },
  {
    id: 'c-2',
    playerId: 'p-2',
    playerName: 'Bob',
    message: 'Just finished my coffee. Let\'s do this!',
    timestamp: '2024-01-01T09:01:00Z',
    type: 'REGULAR',
  },
  {
    id: 'c-3',
    playerId: 'system',
    playerName: 'NEXUS',
    message: 'SITREP: All systems operational. Day 1 commencing.',
    timestamp: '2024-01-01T09:02:00Z',
    type: 'SITREP',
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
    playerId: 'p-1',
    isChatHistoryLoading: false,
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
    playerId: 'p-1',
    isChatHistoryLoading: false,
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
        phase: 'NOMINATION',
        nominees: [],
        votes: {},
        deadline: '2024-01-01T12:03:00Z',
      },
    },
    playerId: 'p-1',
    isChatHistoryLoading: false,
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
        phase: 'TRIAL',
        nominees: ['p-3'],
        votes: {},
        deadline: '2024-01-01T12:07:00Z',
      },
    },
    playerId: 'p-1',
    isChatHistoryLoading: false,
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
          playerId: 'system',
          playerName: 'NEXUS',
          message: 'PULSE CHECK: Respond with your current status.',
          timestamp: '2024-01-01T11:00:00Z',
          type: 'PULSE_CHECK',
        },
      ],
    },
    playerId: 'p-1',
    isChatHistoryLoading: false,
  },
};

export const AsAIPlayer: Story = {
  args: {
    gameState: baseGameState,
    playerId: 'p-3', // Eve (AI-aligned player)
    isChatHistoryLoading: false,
  },
};

export const AsDeadPlayer: Story = {
  args: {
    gameState: baseGameState,
    playerId: 'p-4', // Charlie (dead player)
    isChatHistoryLoading: false,
  },
};

export const WithCrisisEvent: Story = {
  args: {
    gameState: {
      ...baseGameState,
      crisisEvent: {
        id: 'crisis-1',
        type: 'SYSTEM_BREACH',
        name: 'Security Breach Detected',
        description: 'Unauthorized access detected in the main database.',
        severity: 'HIGH',
        isActive: true,
        effects: {
          tokenLoss: 2,
          communicationDisruption: true,
        },
        activatedAt: '2024-01-01T10:30:00Z',
        expiresAt: '2024-01-01T11:30:00Z',
      },
    },
    playerId: 'p-1',
    isChatHistoryLoading: false,
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
    playerId: 'p-1',
    isChatHistoryLoading: false,
  },
};

export const ChatHistoryLoading: Story = {
  args: {
    gameState: {
      ...baseGameState,
      chatMessages: [],
    },
    playerId: 'p-1',
    isChatHistoryLoading: true,
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
        type: 'AI_VICTORY',
        description: 'The AI has successfully converted enough humans to take control.',
        winningPlayers: ['p-3'],
        gameEndReason: 'AI_TAKEOVER',
      },
    },
    playerId: 'p-1',
    isChatHistoryLoading: false,
  },
};