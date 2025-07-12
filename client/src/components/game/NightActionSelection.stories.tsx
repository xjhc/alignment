import type { Meta, StoryObj } from '@storybook/react';
import { NightActionSelection } from './NightActionSelection';
import { GameProvider } from '../../contexts/GameContext';
import { Player, RoleType, KPIType, GameState } from '../../types';

// Mock game context that would normally be provided by GameProvider
const mockGameState: GameState = {
  phase: 'NIGHT',
  dayNumber: 1,
  timeRemaining: 300,
  players: [
    {
      id: 'p-1',
      name: 'Alice',
      jobTitle: 'Chief Information Security Officer',
      controlType: 'HUMAN',
      isAlive: true,
      tokens: 5,
      projectMilestones: 3,
      statusMessage: '"Leading security initiatives"',
      alignment: 'HUMAN',
      avatar: '👤',
      joinedAt: '2024-01-01T00:00:00Z',
      role: {
        type: RoleType.Ciso,
        name: 'CISO',
        description: 'Leads cybersecurity initiatives',
        isUnlocked: true,
        ability: {
          name: 'Isolate Node',
          description: 'Block another player from taking any night actions',
          isReady: true,
        },
      },
      personalKPI: {
        type: KPIType.Capitalist,
        description: 'Complete strategic objectives',
        progress: 3,
        target: 3,
        isCompleted: true,
        reward: 'Role ability unlock',
      },
    },
    {
      id: 'p-2',
      name: 'Bob',
      jobTitle: 'Chief Technology Officer',
      controlType: 'HUMAN',
      isAlive: true,
      tokens: 3,
      projectMilestones: 2,
      statusMessage: '"Optimizing systems"',
      alignment: 'HUMAN',
      avatar: '🤖',
      joinedAt: '2024-01-01T00:00:00Z',
      role: {
        type: RoleType.Cto,
        name: 'CTO',
        description: 'Leads technology initiatives',
        isUnlocked: false,
      },
    },
    {
      id: 'p-3',
      name: 'Charlie',
      jobTitle: 'Intern',
      controlType: 'HUMAN',
      isAlive: true,
      tokens: 2,
      projectMilestones: 1,
      bootcampPoints: 1,
      statusMessage: '"Learning the ropes"',
      alignment: 'HUMAN',
      avatar: '👨‍💼',
      joinedAt: '2024-01-01T00:00:00Z',
    },
  ],
  corporateMandate: {
    isActive: false,
  },
};

const mockGameContext = {
  gameState: mockGameState,
  localPlayer: mockGameState.players[0], // Alice with unlocked CISO ability
  setMiningTarget: () => {},
  handleMineTokens: () => {},
  handleUseAbility: () => {},
  handleProjectMilestones: () => {},
  canPlayerAffordAbility: () => true,
  isValidNightActionTarget: () => true,
};

const mockGameContextIntern = {
  ...mockGameContext,
  localPlayer: mockGameState.players[2], // Charlie the intern
};

const meta: Meta<typeof NightActionSelection> = {
  title: 'Game/NightActionSelection',
  component: NightActionSelection,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  decorators: [
    (Story, context) => {
      const isIntern = context.args?.isIntern;
      const contextValue = isIntern ? mockGameContextIntern : mockGameContext;
      
      return (
        <div style={{ maxWidth: '400px', backgroundColor: '#111827' }}>
          <GameProvider value={contextValue}>
            <Story />
          </GameProvider>
        </div>
      );
    },
  ],
} satisfies Meta<typeof NightActionSelection>;

export default meta;
type Story = StoryObj<typeof meta>;

// Default state - no action selected
export const Default: Story = {
  args: {},
};

// Mine action selected
export const MineActionSelected: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    // Simulate clicking the mine action
    const mineButton = canvasElement.querySelector('[data-testid="mine-action"], div:has(span:contains("⛏️"))') as HTMLElement;
    if (mineButton) {
      mineButton.click();
    }
  },
};

// Project action selected
export const ProjectActionSelected: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    // Simulate clicking the project action
    const projectButton = canvasElement.querySelector('[data-testid="project-action"], div:has(span:contains("📈"))') as HTMLElement;
    if (projectButton) {
      projectButton.click();
    }
  },
};

// Intern with bootcamp action selected
export const InternBootcampSelected: Story = {
  args: {
    isIntern: true,
  },
  play: async ({ canvasElement }) => {
    // Simulate clicking the bootcamp action
    const bootcampButton = canvasElement.querySelector('[data-testid="bootcamp-action"], div:has(span:contains("📚"))') as HTMLElement;
    if (bootcampButton) {
      bootcampButton.click();
    }
  },
};

// Intern with shadow action selected
export const InternShadowSelected: Story = {
  args: {
    isIntern: true,
  },
  play: async ({ canvasElement }) => {
    // Simulate clicking the shadow action
    const shadowButton = canvasElement.querySelector('[data-testid="shadow-action"], div:has(span:contains("👤"))') as HTMLElement;
    if (shadowButton) {
      shadowButton.click();
    }
  },
};

// Role ability action selected
export const AbilityActionSelected: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    // Simulate clicking the ability action
    const abilityButton = canvasElement.querySelector('[data-testid="ability-action"], div:has(span:contains("🔒"))') as HTMLElement;
    if (abilityButton) {
      abilityButton.click();
    }
  },
};

// Intern view (shows different action types)
export const InternView: Story = {
  args: {
    isIntern: true,
  },
};