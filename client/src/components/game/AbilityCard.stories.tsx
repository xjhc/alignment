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

// Helper function to create players with specific roles
const createPlayerWithRole = (roleType: string, roleName: string, abilityName: string, abilityDesc: string, milestones = 3, tokens = 5): Player => ({
  id: 'p-1',
  name: 'Alice',
  jobTitle: roleName,
  isAlive: true,
  tokens,
  projectMilestones: milestones,
  statusMessage: `"Leading as ${roleName}"`,
  alignment: 'HUMAN',
  avatar: '👤',
  joinedAt: '2024-01-01T00:00:00Z',
  role: {
    type: roleType as any,
    name: roleName,
    description: `Leads the company's ${roleType.toLowerCase()} initiatives`,
    isUnlocked: milestones >= 3,
    ability: milestones >= 3 ? {
      name: abilityName,
      description: abilityDesc,
      isReady: true,
    } : undefined,
  },
  personalKPI: {
    type: 'PRODUCTIVITY',
    description: 'Complete strategic objectives',
    progress: milestones,
    target: 3,
    isCompleted: milestones >= 3,
    reward: 'Role ability unlock',
  },
});

const basePlayer: Player = createPlayerWithRole(
  'CISO',
  'Chief Information Security Officer',
  'Isolate Node',
  'Block another player from taking any night actions by isolating their network access'
);

// CISO - Isolate Node (Blocking ability)
export const CISOAbility: Story = {
  args: {
    localPlayer: basePlayer,
  },
};

// CEO - Performance Review (Force action)
export const CEOAbility: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'CEO',
      'Chief Executive Officer',
      'Performance Review',
      'Force another player to work on Project Milestones tonight instead of their chosen action'
    ),
  },
};

// CTO - Overclock Servers (Enhanced mining)
export const CTOAbility: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'CTO',
      'Chief Technology Officer', 
      'Overclock Servers',
      'Mine tokens for yourself and a target player with guaranteed success'
    ),
  },
};

// CFO - Reallocate Budget (Token transfer)
export const CFOAbility: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'CFO',
      'Chief Financial Officer',
      'Reallocate Budget',
      'Transfer 1 token from one player to another player'
    ),
  },
};

// COO - Pivot (Crisis selection)
export const COOAbility: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'COO',
      'Chief Operating Officer',
      'Pivot',
      'Choose the crisis event that will occur during the next day phase'
    ),
  },
};

// VP Ethics - Run Audit (Investigation)
export const VPEthicsAbility: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'ETHICS',
      'VP, Ethics & Alignment',
      'Run Audit',
      'Investigate a player\'s true alignment, with results visible only to the AI faction'
    ),
  },
};

// VP Platforms - Deploy Hotfix (Information control)
export const VPPlatformsAbility: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'PLATFORMS',
      'VP, Platforms',
      'Deploy Hotfix',
      'Redact one section of tomorrow\'s SITREP to hide information from other players'
    ),
  },
};

// Locked ability (insufficient milestones)
export const AbilityLocked: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'CISO',
      'Chief Information Security Officer',
      'Isolate Node',
      'Block another player from taking any night actions by isolating their network access',
      2  // Only 2 milestones, needs 3
    ),
  },
};

// Insufficient tokens
export const InsufficientTokens: Story = {
  args: {
    localPlayer: createPlayerWithRole(
      'CISO',
      'Chief Information Security Officer',
      'Isolate Node',
      'Block another player from taking any night actions by isolating their network access',
      3,  // Milestones sufficient
      0   // No tokens
    ),
  },
};

// Ability already used
export const AbilityUsed: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      hasUsedAbility: true,
    },
  },
};

// No role/ability (Intern)
export const NoAbility: Story = {
  args: {
    localPlayer: {
      ...basePlayer,
      role: undefined,
      jobTitle: 'Intern',
    },
  },
};

// AI player with ability
export const AIPlayerAbility: Story = {
  args: {
    localPlayer: {
      ...createPlayerWithRole(
        'CTO',
        'Chief Technology Officer',
        'System Override',
        'Execute advanced system commands with AI-enhanced capabilities'
      ),
      name: 'NEXUS',
      alignment: 'ALIGNED',
    },
  },
};