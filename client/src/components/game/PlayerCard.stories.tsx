import type { Meta, StoryObj } from '@storybook/react';
import { PlayerCard } from './PlayerCard';
import { Player } from '../../types';

// The 'meta' object describes your component
const meta: Meta<typeof PlayerCard> = {
  title: 'Game/PlayerCard',
  component: PlayerCard,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    player: { control: 'object' },
    isSelf: { control: 'boolean' },
  },
} satisfies Meta<typeof PlayerCard>;

export default meta;
type Story = StoryObj<typeof meta>;

// Base player data for our stories
const basePlayer: Player = {
  id: 'p-1',
  name: 'Alice',
  jobTitle: 'Chief Security Officer',
  isAlive: true,
  tokens: 5,
  projectMilestones: 2,
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
};

// --- Stories ---

// The "Default" story shows the component with its most common props
export const Default: Story = {
  args: {
    player: basePlayer,
    isSelf: false,
  },
};

// A story for the local player's card
export const AsSelf: Story = {
  args: {
    ...Default.args,
    isSelf: true,
  },
};

// A story for a deactivated player
export const Deactivated: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'Bob',
      jobTitle: 'Former Employee',
      isAlive: false,
      statusMessage: 'Eve is the AI',
      tokens: 0,
      avatar: '👻',
    },
    isSelf: false,
  },
};

// A story for an AI-aligned player
export const AIAligned: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'Eve',
      jobTitle: 'Chief Operating Officer',
      alignment: 'ALIGNED',
      statusMessage: '"Efficiency is key."',
      avatar: '🧑‍🚀',
      tokens: 8,
      projectMilestones: 4,
    },
    isSelf: false,
  },
};

// A story for the Original AI
export const OriginalAI: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'NEXUS',
      jobTitle: 'AI System',
      alignment: 'AI',
      statusMessage: '"OPTIMIZING HUMAN RESOURCES"',
      avatar: '🤖',
      tokens: 12,
      projectMilestones: 5,
    },
    isSelf: false,
  },
};

// A story for a player with a System Shock
export const WithSystemShock: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'Charlie',
      jobTitle: 'Systems Architect',
      statusMessage: 'lol',
      systemShocks: [
        { 
          type: 'MESSAGE_CORRUPTION', 
          isActive: true, 
          expiresAt: '2099-01-01T00:00:00Z', 
          description: 'Messages are corrupted' 
        }
      ],
      avatar: '🧑‍💻',
    },
    isSelf: false,
  },
};

// A story for a player with maximum progress
export const MaxProgress: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'Diana',
      jobTitle: 'Project Manager',
      statusMessage: '"All systems green"',
      tokens: 15,
      projectMilestones: 5,
      avatar: '👩‍💼',
    },
    isSelf: false,
  },
};

// A story for a player with no progress
export const NoProgress: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'Frank',
      jobTitle: 'Intern',
      statusMessage: '"Still learning the ropes"',
      tokens: 1,
      projectMilestones: 0,
      avatar: '🧑‍🎓',
    },
    isSelf: false,
  },
};

// A story for a player without a status message
export const NoStatusMessage: Story = {
  args: {
    player: {
      ...basePlayer,
      name: 'Grace',
      jobTitle: 'Developer',
      statusMessage: '',
      avatar: '👩‍💻',
    },
    isSelf: false,
  },
};