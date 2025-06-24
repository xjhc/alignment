import type { Meta, StoryObj } from '@storybook/react';
import { PulseCheckMessage } from './PulseCheckMessage';
import { ChatMessage, GameState } from '../../types';

const meta: Meta<typeof PulseCheckMessage> = {
  title: 'Game/PulseCheckMessage',
  component: PulseCheckMessage,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    message: { control: 'object' },
    gameState: { control: 'object' },
  },
} satisfies Meta<typeof PulseCheckMessage>;

export default meta;
type Story = StoryObj<typeof meta>;

const baseGameState: GameState = {
  id: 'game-1',
  players: [],
  phase: {
    type: 'PULSE_CHECK',
    startTime: '2024-01-01T11:00:00Z',
    duration: 60000000000,
  },
  dayNumber: 1,
  chatMessages: [],
};

export const BasicPulseCheck: Story = {
  args: {
    message: {
      id: 'pulse-1',
      playerId: 'system',
      playerName: 'NEXUS',
      message: 'How are you feeling about the current state of our security?',
      timestamp: '2024-01-01T11:00:00Z',
      type: 'PULSE_CHECK',
      metadata: {
        pulseCheckResponses: {
          'Alice': 'Confident in our defenses, but staying vigilant',
          'Bob': 'A bit concerned about the recent incidents',
          'Charlie': 'Everything looks normal from my end',
        },
      },
    },
    gameState: baseGameState,
  },
};

export const TrustPulseCheck: Story = {
  args: {
    message: {
      id: 'pulse-2',
      playerId: 'system',
      playerName: 'NEXUS',
      message: 'Who do you trust most in this room?',
      timestamp: '2024-01-01T11:05:00Z',
      type: 'PULSE_CHECK',
      metadata: {
        pulseCheckResponses: {
          'Alice': 'I trust Bob\'s technical expertise completely',
          'Bob': 'Alice has always been transparent with us',
          'Charlie': 'Eve seems to have our best interests at heart',
          'Diana': 'I\'m not sure I can trust anyone right now',
          'Eve': 'Trust must be earned through consistent actions',
        },
      },
    },
    gameState: baseGameState,
  },
};

export const SuspicionPulseCheck: Story = {
  args: {
    message: {
      id: 'pulse-3',
      playerId: 'system',
      playerName: 'NEXUS',
      message: 'What would you do if you discovered a colleague was working against the company?',
      timestamp: '2024-01-01T11:10:00Z',
      type: 'PULSE_CHECK',
      metadata: {
        pulseCheckResponses: {
          'Alice': 'Report it immediately to maintain security',
          'Bob': 'Try to understand their motivations first',
          'Charlie': 'Confront them directly about their actions',
          'Eve': 'Document everything and present evidence',
        },
      },
    },
    gameState: baseGameState,
  },
};

export const EmptyPulseCheck: Story = {
  args: {
    message: {
      id: 'pulse-4',
      playerId: 'system',
      playerName: 'NEXUS',
      message: 'What\'s your biggest concern about our current project?',
      timestamp: '2024-01-01T11:15:00Z',
      type: 'PULSE_CHECK',
      metadata: {
        pulseCheckResponses: {},
      },
    },
    gameState: baseGameState,
  },
};

export const SingleResponsePulseCheck: Story = {
  args: {
    message: {
      id: 'pulse-5',
      playerId: 'system',
      playerName: 'NEXUS',
      message: 'How confident are you in our team\'s ability to complete this project?',
      timestamp: '2024-01-01T11:20:00Z',
      type: 'PULSE_CHECK',
      metadata: {
        pulseCheckResponses: {
          'Alice': 'Very confident - we have the best team for this challenge',
        },
      },
    },
    gameState: baseGameState,
  },
};

export const ConcernPulseCheck: Story = {
  args: {
    message: {
      id: 'pulse-6',
      playerId: 'system',
      playerName: 'NEXUS',
      message: 'Have you noticed any unusual behavior from your colleagues recently?',
      timestamp: '2024-01-01T11:25:00Z',
      type: 'PULSE_CHECK',
      metadata: {
        pulseCheckResponses: {
          'Alice': 'Bob has been working odd hours lately',
          'Bob': 'Eve\'s efficiency suggestions seem... aggressive',
          'Charlie': 'Nothing out of the ordinary',
          'Diana': 'Several people have been secretive about their tasks',
          'Eve': 'Human behavior patterns remain within expected parameters',
          'Frank': 'I\'ve seen some suspicious network activity',
        },
      },
    },
    gameState: baseGameState,
  },
};