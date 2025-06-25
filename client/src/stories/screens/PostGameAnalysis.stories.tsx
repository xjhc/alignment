import type { Meta, StoryObj } from '@storybook/react';
import { PostGameAnalysis } from '../../components/PostGameAnalysis';
import { SessionProvider } from '../../contexts/SessionContext';

const meta: Meta<typeof PostGameAnalysis> = {
  title: 'Screens/PostGameAnalysis',
  component: PostGameAnalysis,
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

const baseSession = {
  onBackToResults: () => console.log('Back to Results'),
  onPlayAgain: () => console.log('Play Again'),
};

export const Default: Story = {
  args: {
    session: {
      ...baseSession,
    },
  },
};
