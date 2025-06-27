import type { Meta, StoryObj } from '@storybook/react';
import { WaitingScreen } from '../../components/WaitingScreen';
import { SessionProvider } from '../../contexts/SessionContext';

const meta: Meta<typeof WaitingScreen> = {
  title: 'Screens/WaitingScreen',
  component: WaitingScreen,
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
  appState: {
    gameId: 'g-f4b1',
    playerId: 'p-alice',
  },
  isConnected: true,
  onStartGame: () => console.log('Start Game'),
  onLeaveLobby: () => console.log('Leave Lobby'),
};

export const HostView: Story = {
  args: {
    session: {
      ...baseSession,
      lobbyState: {
        playerInfos: [
          { id: 'p-alice', name: 'Alice', avatar: '👤' },
          { id: 'p-bob', name: 'Bob', avatar: '🧑‍💻' },
        ],
        isHost: true,
        canStart: true,
        hostId: 'p-alice',
        lobbyName: '#lobby-glorious-gerbil',
        maxPlayers: 8,
        connectionError: null,
      },
    },
  },
};

export const NonHostView: Story = {
  args: {
    session: {
      ...baseSession,
      lobbyState: {
        ...HostView.args.session.lobbyState,
        isHost: false,
        playerId: 'p-bob',
      },
    },
  },
};

export const CountdownActive: Story = {
  args: {
    session: {
      ...baseSession,
      lobbyState: {
        playerInfos: [
          { id: 'p-alice', name: 'Alice', avatar: '👤', joinedAt: '2024-01-01T00:00:00Z' },
          { id: 'p-bob', name: 'Bob', avatar: '🧑‍💻', joinedAt: '2024-01-01T00:01:00Z' },
          { id: 'p-charlie', name: 'Charlie', avatar: '👨‍💼', joinedAt: '2024-01-01T00:02:00Z' },
          { id: 'p-diana', name: 'Diana', avatar: '👩‍💻', joinedAt: '2024-01-01T00:03:00Z' },
        ],
        isHost: true,
        canStart: true,
        hostId: 'p-alice',
        lobbyName: '#lobby-countdown-demo',
        maxPlayers: 8,
        connectionError: null,
        countdown: {
          isActive: true,
          remaining: 3,
        },
      },
    },
  },
};

export const CountdownGo: Story = {
  args: {
    session: {
      ...baseSession,
      lobbyState: {
        playerInfos: [
          { id: 'p-alice', name: 'Alice', avatar: '👤', joinedAt: '2024-01-01T00:00:00Z' },
          { id: 'p-bob', name: 'Bob', avatar: '🧑‍💻', joinedAt: '2024-01-01T00:01:00Z' },
          { id: 'p-charlie', name: 'Charlie', avatar: '👨‍💼', joinedAt: '2024-01-01T00:02:00Z' },
          { id: 'p-diana', name: 'Diana', avatar: '👩‍💻', joinedAt: '2024-01-01T00:03:00Z' },
        ],
        isHost: true,
        canStart: true,
        hostId: 'p-alice',
        lobbyName: '#lobby-countdown-demo',
        maxPlayers: 8,
        connectionError: null,
        countdown: {
          isActive: true,
          remaining: 0,
        },
      },
    },
  },
};
