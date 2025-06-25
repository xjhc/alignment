import type { Meta, StoryObj } from '@storybook/react';
import { LobbyListScreen } from '../../components/LobbyListScreen';

const meta: Meta<typeof LobbyListScreen> = {
  title: 'Screens/LobbyListScreen',
  component: LobbyListScreen,
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [],
    },
  },
  tags: ['autodocs'],
  argTypes: {
    playerName: { control: 'text' },
    playerAvatar: { control: 'text' },
    onJoinLobby: { action: 'onJoinLobby' },
    onCreateGame: { action: 'onCreateGame' },
    onBack: { action: 'onBack' },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    playerName: 'Alice',
    playerAvatar: '👤',
  },
};
