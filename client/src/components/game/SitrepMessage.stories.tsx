import type { Meta, StoryObj } from '@storybook/react';
import { SitrepMessage } from './SitrepMessage';
import { ChatMessage, GameState, Player } from '../../types';

const meta: Meta<typeof SitrepMessage> = {
  title: 'Game/SitrepMessage',
  component: SitrepMessage,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    message: { control: 'object' },
    gameState: { control: 'object' },
  },
} satisfies Meta<typeof SitrepMessage>;

export default meta;
type Story = StoryObj<typeof meta>;

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
    isAlive: false,
    tokens: 0,
    projectMilestones: 4,
    statusMessage: 'Deactivated during night phase',
    alignment: 'ALIGNED',
    avatar: '👻',
    joinedAt: '2024-01-01T00:10:00Z',
    role: {
      type: 'EXECUTIVE',
      name: 'Executive',
      description: 'Manages company operations',
      isUnlocked: true,
      ability: {
        name: 'Resource Allocation',
        description: 'Redistribute tokens between players',
        isReady: false,
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
];

const baseGameState: GameState = {
  id: 'game-1',
  players: basePlayers,
  phase: {
    type: 'DISCUSSION',
    startTime: '2024-01-01T09:00:00Z',
    duration: 300000000000,
  },
  dayNumber: 2,
  chatMessages: [],
};

const baseSitrepMessage: ChatMessage = {
  id: 'sitrep-1',
  playerID: 'system',
  playerName: 'NEXUS',
  message: 'Good morning, team. Here\'s the SITREP.',
  timestamp: '2024-01-01T09:00:00Z',
  type: 'SITREP',
  isSystem: true,
  metadata: {
    playerHeadcount: {
      humans: 2,
      aligned: 0,
      dead: 1,
    },
    nightActions: [
      'Alice used Security Scan on Bob - No suspicious activity detected',
      'Eve was deactivated by AI takeover attempt'
    ],
  },
};

export const MorningBriefing: Story = {
  args: {
    message: baseSitrepMessage,
    gameState: baseGameState,
  },
};

export const CrisisActive: Story = {
  args: {
    message: {
      ...baseSitrepMessage,
      metadata: {
        ...baseSitrepMessage.metadata,
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
    },
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
  },
};

export const QuietNight: Story = {
  args: {
    message: {
      ...baseSitrepMessage,
      metadata: {
        playerHeadcount: {
          humans: 3,
          aligned: 1,
          dead: 0,
        },
        nightActions: [],
      },
    },
    gameState: {
      ...baseGameState,
      players: baseGameState.players.map(p => ({ ...p, isAlive: true })),
      dayNumber: 1,
    },
  },
};

export const IntenseNight: Story = {
  args: {
    message: {
      ...baseSitrepMessage,
      metadata: {
        playerHeadcount: {
          humans: 1,
          aligned: 0,
          dead: 3,
        },
        nightActions: [
          'Alice used Security Scan on Bob - SUSPICIOUS ACTIVITY DETECTED',
          'Charlie used System Diagnostic on Network - Multiple intrusion attempts blocked',
          'Bob attempted to access restricted files - BLOCKED by security protocols',
          'Eve was deactivated due to confirmed AI alignment',
          'Diana was deactivated during AI takeover attempt',
        ],
      },
    },
    gameState: {
      ...baseGameState,
      dayNumber: 4,
    },
  },
};