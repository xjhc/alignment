import type { Meta, StoryObj } from '@storybook/react';
import { NightActionSelection } from './NightActionSelection';
import { GameProvider } from '../../contexts/GameContext';
import { Player, GameState } from '../../types';

const meta: Meta<typeof NightActionSelection> = {
  title: 'Game/NightActionSelection',
  component: NightActionSelection,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  decorators: [
    (Story, { args }) => (
      <GameProvider initialGameState={args.gameState as GameState} localPlayerId={args.localPlayerId as string}>
        <div className="bg-gray-900 min-h-screen">
          <Story />
        </div>
      </GameProvider>
    ),
  ],
} satisfies Meta<typeof NightActionSelection>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper function to create base players
const createPlayer = (id: string, name: string, role: string, milestones = 2, tokens = 5): Player => ({
  id,
  name,
  jobTitle: role,
  isAlive: true,
  tokens,
  projectMilestones: milestones,
  statusMessage: `"Working hard at ${role}"`,
  alignment: 'HUMAN',
  avatar: '👤',
  joinedAt: '2024-01-01T00:00:00Z',
  role: {
    type: role as any,
    name: role,
    description: `${role} role description`,
    isUnlocked: milestones >= 3,
    ability: milestones >= 3 ? {
      name: `${role} Ability`,
      description: `${role} specific ability`,
      isReady: true,
    } : undefined,
  },
  personalKPI: {
    type: 'PRODUCTIVITY',
    description: 'Complete 3 project milestones',
    progress: milestones,
    target: 3,
    isCompleted: milestones >= 3,
    reward: 'Role ability unlock',
  },
});

const baseGameState: GameState = {
  players: {
    'player-1': createPlayer('player-1', 'Alice', 'CISO', 3, 8),
    'player-2': createPlayer('player-2', 'Bob', 'CEO', 2, 5),
    'player-3': createPlayer('player-3', 'Charlie', 'CTO', 3, 6),
    'player-4': createPlayer('player-4', 'Diana', 'CFO', 1, 3),
    'player-5': createPlayer('player-5', 'Eve', 'COO', 3, 7),
  },
  phase: 'NIGHT',
  dayNumber: 2,
  gameStartTime: '2024-01-01T00:00:00Z',
  chat: [],
  votes: {},
  crisis: null,
  nightActions: {},
};

// CISO with unlocked Isolate Node ability
export const CISOWithAbility: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-1',
  },
};

// CEO with unlocked Performance Review ability
export const CEOWithAbility: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-2': createPlayer('player-2', 'Bob', 'CEO', 3, 5),
      },
    },
    localPlayerId: 'player-2',
  },
};

// CTO with unlocked Overclock Servers ability
export const CTOWithAbility: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-3',
  },
};

// CFO with unlocked Reallocate Budget ability (needs special handling for dual targets)
export const CFOWithAbility: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-4': createPlayer('player-4', 'Diana', 'CFO', 3, 3),
      },
    },
    localPlayerId: 'player-4',
  },
};

// COO with unlocked Pivot ability
export const COOWithAbility: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-5',
  },
};

// VP Ethics with unlocked Run Audit ability
export const VPEthicsWithAbility: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-1': createPlayer('player-1', 'Alice', 'ETHICS', 3, 8),
      },
    },
    localPlayerId: 'player-1',
  },
};

// VP Platforms with unlocked Deploy Hotfix ability
export const VPPlatformsWithAbility: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-1': createPlayer('player-1', 'Alice', 'PLATFORMS', 3, 8),
      },
    },
    localPlayerId: 'player-1',
  },
};

// Player with locked ability (insufficient milestones)
export const LockedAbility: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-2': createPlayer('player-2', 'Bob', 'CEO', 2, 5),
      },
    },
    localPlayerId: 'player-2',
  },
};

// Player with unlocked ability but insufficient tokens
export const InsufficientTokens: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-1': createPlayer('player-1', 'Alice', 'CISO', 3, 0),
      },
    },
    localPlayerId: 'player-1',
  },
};

// Player who has already used their ability
export const AbilityAlreadyUsed: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-1': {
          ...createPlayer('player-1', 'Alice', 'CISO', 3, 8),
          hasUsedAbility: true,
        },
      },
    },
    localPlayerId: 'player-1',
  },
};

// Mining action selected
export const MiningActionSelected: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-1',
  },
};

// Project milestones action (immediate execution)
export const ProjectMilestones: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-2': createPlayer('player-2', 'Bob', 'CEO', 2, 5),
      },
    },
    localPlayerId: 'player-2',
  },
};

// Small game with few players (high mining success rate)
export const SmallGame: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        'player-1': createPlayer('player-1', 'Alice', 'CISO', 3, 8),
        'player-2': createPlayer('player-2', 'Bob', 'CEO', 2, 5),
      },
    },
    localPlayerId: 'player-1',
  },
};

// Large game with many players (low mining success rate)
export const LargeGame: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...Object.fromEntries(
          Array.from({ length: 8 }, (_, i) => [
            `player-${i + 1}`,
            createPlayer(`player-${i + 1}`, `Player ${i + 1}`, ['CISO', 'CEO', 'CTO', 'CFO', 'COO', 'ETHICS', 'PLATFORMS', 'INTERN'][i], 3, 5)
          ])
        ),
      },
    },
    localPlayerId: 'player-1',
  },
};

// AI player with abilities (aligned players)
export const AIPlayerWithAbility: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: {
        ...baseGameState.players,
        'player-1': {
          ...createPlayer('player-1', 'NEXUS', 'CTO', 3, 8),
          alignment: 'ALIGNED',
        },
      },
    },
    localPlayerId: 'player-1',
  },
};