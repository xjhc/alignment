import type { Meta, StoryObj } from '@storybook/react';
import { IdentityCard } from './IdentityCard';
import { Player } from '../../types';

const meta: Meta<typeof IdentityCard> = {
  title: 'Game/IdentityCard',
  component: IdentityCard,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    localPlayer: { control: 'object' },
  },
} satisfies Meta<typeof IdentityCard>;

export default meta;
type Story = StoryObj<typeof meta>;

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
    type: 'CISO',
    name: 'Chief Information Security Officer',
    description: 'Protects the company from security threats',
    isUnlocked: true,
    ability: {
      name: 'Security Audit',
      description: 'Investigate another player\'s activities',
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

export const HumanPlayer: Story = {
  args: {
    localPlayer: basePlayer,
  },
};

export const AIPlayer: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      name: 'NEXUS',
      jobTitle: 'AI System',
      alignment: 'AI',
      statusMessage: '"OPTIMIZING HUMAN RESOURCES"',
      avatar: '🤖',
      role: {
        type: 'CTO',
        name: 'Chief Technology Officer',
        description: 'Manages all technology systems',
        isUnlocked: true,
        ability: {
          name: 'System Override',
          description: 'Take control of company systems',
          isReady: true,
        },
      },
    },
  },
};

export const AlignedPlayer: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      name: 'Eve',
      jobTitle: 'Chief Operating Officer',
      alignment: 'ALIGNED',
      statusMessage: '"Efficiency is key."',
      avatar: '🧑‍🚀',
      role: {
        type: 'COO',
        name: 'Chief Operating Officer',
        description: 'Manages company operations',
        isUnlocked: true,
        ability: {
          name: 'Resource Allocation',
          description: 'Redistribute tokens between players',
          isReady: false,
        },
      },
    },
  },
};

export const SystemsRole: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      name: 'Bob',
      jobTitle: 'Systems Administrator',
      role: {
        type: 'SYSTEMS',
        name: 'Systems Administrator',
        description: 'Maintains critical infrastructure',
        isUnlocked: true,
        ability: {
          name: 'System Diagnostic',
          description: 'Analyze system health and detect anomalies',
          isReady: true,
        },
      },
    },
  },
};

export const EthicsRole: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      name: 'Diana',
      jobTitle: 'Ethics Officer',
      role: {
        type: 'ETHICS',
        name: 'Ethics Officer',
        description: 'Ensures ethical AI deployment',
        isUnlocked: true,
        ability: {
          name: 'Ethical Review',
          description: 'Review and flag unethical behavior',
          isReady: false,
        },
      },
    },
  },
};