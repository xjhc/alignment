import type { Meta, StoryObj } from '@storybook/react';
import { NightActionSelection } from './NightActionSelection';
import { RoleType, KPIType, GameState } from '../../types';
import { PhaseType } from '../../types/generated';

// Mock game state data for Storybook
const mockGameStateData: Partial<GameState> = {
  id: 'night-action-game',
  phase: {
    type: PhaseType.Night,
    startTime: new Date().toISOString(),
    duration: 300000000000, // 5 minutes
  },
  dayNumber: 1,
  players: [
    {
      id: 'p-1',
      name: 'Alice',
      jobTitle: 'Chief Information Security Officer',
      controlType: 'HUMAN',
      status: 'ACTIVE',
      isAlive: true,
      connectionStatus: 'CONNECTED',
      isRolePubliclyRevealed: false,
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
      status: 'ACTIVE',
      isAlive: true,
      connectionStatus: 'CONNECTED',
      isRolePubliclyRevealed: false,
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
      status: 'ACTIVE',
      isAlive: true,
      connectionStatus: 'CONNECTED',
      isRolePubliclyRevealed: false,
      tokens: 2,
      projectMilestones: 1,
      bootcampPoints: 1,
      statusMessage: '"Learning the ropes"',
      alignment: 'HUMAN',
      avatar: '👨‍💼',
      joinedAt: '2024-01-01T00:00:00Z',
    },
  ],
  chatMessages: [],
};


const meta: Meta<typeof NightActionSelection> = {
  title: 'Game/NightActionSelection',
  component: NightActionSelection,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    storyGameState: { control: 'object' },
    localPlayerId: { control: 'text' },
    isIntern: { control: 'boolean' },
  },
  decorators: [
    (Story) => (
      <div style={{ maxWidth: '400px', backgroundColor: '#111827' }}>
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof NightActionSelection>;

export default meta;
type Story = StoryObj<typeof meta>;

// Default state - no action selected
export const Default: Story = {
  args: {
    storyGameState: mockGameStateData,
    localPlayerId: 'p-1', // Alice with CISO ability
  },
};

// Mine action selected
export const MineActionSelected: Story = {
  args: {
    storyGameState: mockGameStateData,
    localPlayerId: 'p-1',
  },
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
  args: {
    storyGameState: mockGameStateData,
    localPlayerId: 'p-1',
  },
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
    storyGameState: mockGameStateData,
    localPlayerId: 'p-3', // Charlie the intern
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
    storyGameState: mockGameStateData,
    localPlayerId: 'p-3', // Charlie the intern
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
  args: {
    storyGameState: mockGameStateData,
    localPlayerId: 'p-1',
  },
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
    storyGameState: mockGameStateData,
    localPlayerId: 'p-3', // Charlie the intern
    isIntern: true,
  },
};