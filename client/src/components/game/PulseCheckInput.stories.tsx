import type { Meta, StoryObj } from '@storybook/react';
import { PulseCheckInput } from './PulseCheckInput';
import { GameProvider } from '../../contexts/GameContext';
import { GameState, Player } from '../../types';

const meta: Meta<typeof PulseCheckInput> = {
  title: 'Game/PulseCheckInput',
  component: PulseCheckInput,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  decorators: [
    (Story, { args }) => (
      <GameProvider gameState={args.gameState as GameState} localPlayerId={args.localPlayerId as string}>
        <div className="bg-gray-900 min-h-screen">
          <Story />
        </div>
      </GameProvider>
    ),
  ],
} satisfies Meta<typeof PulseCheckInput>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper function to create base players
const createPlayer = (id: string, name: string, hasSubmittedPulseCheck = false): Player => ({
  id,
  name,
  jobTitle: 'Software Engineer',
  isAlive: true,
  tokens: 5,
  projectMilestones: 2,
  statusMessage: '"Working on the project"',
  alignment: 'HUMAN',
  avatar: '👤',
  joinedAt: '2024-01-01T00:00:00Z',
  hasSubmittedPulseCheck,
});

const baseGameState: GameState = {
  id: 'game-1',
  players: [
    createPlayer('player-1', 'Alice', false),
    createPlayer('player-2', 'Bob', false),
    createPlayer('player-3', 'Charlie', false),
  ],
  phase: {
    type: 'PULSE_CHECK',
    startTime: '2024-01-01T11:00:00Z',
    duration: 60000000000,
  },
  dayNumber: 1,
  chatMessages: [],
};

const handlePulseCheckMock = async (response: string) => {
  console.log('Pulse check response:', response);
};

export const DefaultState: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-1',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Alice',
    question: 'What is your immediate response to the current crisis?',
  },
};

export const CustomQuestion: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-1',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Alice',
    question: 'How confident are you in the team\'s ability to handle this situation?',
  },
};

export const AlreadySubmitted: Story = {
  args: {
    gameState: {
      ...baseGameState,
      players: [
        createPlayer('player-1', 'Alice', true),
        createPlayer('player-2', 'Bob', false),
        createPlayer('player-3', 'Charlie', false),
      ],
    },
    localPlayerId: 'player-1',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Alice',
    question: 'What is your immediate response to the current crisis?',
  },
};

export const DifferentPlayer: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-2',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Bob',
    question: 'What measures should we take to ensure project security?',
  },
};

export const LongQuestion: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-1',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Alice',
    question: 'Given the recent security incidents and the increasing complexity of our project, what is your assessment of the current threat level and what immediate actions do you recommend we take to ensure the safety and success of our mission?',
  },
};

export const SecurityBreach: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-1',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Alice',
    question: 'A security breach has been detected in our systems. How do you assess the situation?',
  },
};

export const MissionCritical: Story = {
  args: {
    gameState: baseGameState,
    localPlayerId: 'player-2',
    handlePulseCheck: handlePulseCheckMock,
    localPlayerName: 'Bob',
    question: 'The project deadline is approaching and we are behind schedule. What is your stance?',
  },
};