import type { Meta, StoryObj } from '@storybook/react';
import { AbilityCard } from './AbilityCard';
import { Player } from '../../types';

const meta: Meta<typeof AbilityCard> = {
  title: 'Game/AbilityCard',
  component: AbilityCard,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    localPlayer: { control: 'object' },
  },
} satisfies Meta<typeof AbilityCard>;

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
      description: 'Investigate another player\'s activities to uncover suspicious behavior',
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

export const AbilityReady: Story = {
  args: {
    localPlayer: basePlayer,
  },
};

export const AbilityLocked: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      role: {
        ...basePlayer.role!,
        ability: {
          name: 'Security Audit',
          description: 'Investigate another player\'s activities to uncover suspicious behavior',
          isReady: false,
        },
      },
    },
  },
};

export const AbilityUsed: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      hasUsedAbility: true,
    },
  },
};

export const NoAbility: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      role: {
        ...basePlayer.role!,
        ability: undefined,
      },
    },
  },
};

export const SystemsAbility: Story = {
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
          description: 'Analyze system health and detect anomalies in company infrastructure',
          isReady: true,
        },
      },
    },
  },
};

export const AIAbility: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      name: 'NEXUS',
      jobTitle: 'AI System',
      alignment: 'AI',
      role: {
        type: 'CTO',
        name: 'Chief Technology Officer',
        description: 'Manages all technology systems',
        isUnlocked: true,
        ability: {
          name: 'System Override',
          description: 'Take control of critical company systems and bypass security protocols',
          isReady: true,
        },
      },
    },
  },
};