import type { Meta, StoryObj } from '@storybook/react';
import { GameOverScreen } from '../../components/GameOverScreen';
import { SessionProvider } from '../../contexts/SessionContext';
import { GameState } from '../../types';

const meta: Meta<typeof GameOverScreen> = {
  title: 'Screens/GameOverScreen',
  component: GameOverScreen,
  parameters: {
    layout: 'fullscreen',
  },
  decorators: [
    (Story, { args }) => (
      <SessionProvider value={args.session as any}>
        <Story />
      </SessionProvider>
    ),
  ],
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof meta>;

const baseGameState: GameState = {
  id: 'game-1',
  players: [
    { id: 'p-1', name: 'Alice', isAlive: true, alignment: 'HUMAN', tokens: 10, avatar: '👤', role: { name: 'CISO' } },
    { id: 'p-2', name: 'Bob', isAlive: false, alignment: 'HUMAN', tokens: 2, avatar: '🧑‍💻', role: { name: 'Engineer' } },
    { id: 'p-3', name: 'Eve', isAlive: false, alignment: 'AI', tokens: 8, avatar: '🤖', role: { name: 'CTO' } },
  ],
  phase: { type: 'GAME_OVER', startTime: new Date().toISOString(), duration: 0 },
  dayNumber: 5,
  chatMessages: [],
};

const baseSession = {
  onViewAnalysis: () => console.log('View Analysis'),
  onPlayAgain: () => console.log('Play Again'),
};

export const HumanVictory: Story = {
  args: {
    session: {
      ...baseSession,
      gameState: {
        ...baseGameState,
        winCondition: { winner: 'HUMANS', condition: 'AI_ELIMINATED', description: '' },
      },
    },
  },
};

export const AIVictory: Story = {
  args: {
    session: {
      ...baseSession,
      gameState: {
        ...baseGameState,
        winCondition: { winner: 'AI', condition: 'TOKEN_MAJORITY', description: '' },
      },
    },
  },
};
