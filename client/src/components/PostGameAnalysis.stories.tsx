import type { Meta, StoryObj } from '@storybook/react';
import { PostGameAnalysis } from './PostGameAnalysis';
import { SessionProvider } from '../contexts/SessionContext';
import { SessionContextType } from '../contexts/SessionContext';

// Mock session context for Storybook
const mockSessionContext: SessionContextType = {
  appState: {
    screen: 'POST_GAME_ANALYSIS',
    isSpectating: false,
  },
  sessionState: {
    playerId: 'player-123',
    playerName: 'Test Player',
    playerAvatar: '👤',
    sessionToken: 'mock-token',
  },
  lobbyState: {
    gameId: 'game-123',
    lobbyName: 'Test Game',
  },
  gameState: {} as any,
  roleAssignment: null,
  gameAnalysis: {
    mvp: {
      playerName: 'Alice',
      playerAvatar: '👤',
      reason: 'Scored 11 points: 4 tokens mined for others (×2) + 2 correct elimination votes (×3) - 1 incorrect vote'
    },
    keyMoment: {
      title: '🎯 Game-Changing Moment',
      description: 'On Day 3, Grace used "Run Audit" on Eve, secretly confirming Eve\'s AI alignment to the human faction, leading to the unified elimination vote that secured victory.'
    },
    partingShots: [
      {
        playerName: 'Frank',
        playerAvatar: '👻',
        message: '"Eve is the AI" — Correctly identified the Original AI'
      },
      {
        playerName: 'Dave',
        playerAvatar: '🤖',
        message: '"It wasn\'t me, check Alice." — Attempted misdirection as Aligned'
      }
    ],
    timeline: [
      {
        type: 'elimination',
        icon: '💀',
        day: 'Day 1 Elimination',
        description: 'Frank was eliminated by majority vote (4-2). Revealed alignment: HUMAN.',
        iconClass: 'elimination'
      },
      {
        type: 'ability',
        icon: '🛡️',
        day: 'Night 2 Action',
        description: 'Alice used Isolate Node on Eve, successfully blocking the AI\'s conversion attempt.',
        iconClass: 'ability'
      }
    ],
    playerStats: [
      {
        name: 'Alice',
        avatar: '👤',
        role: 'Chief Security Officer',
        alignment: 'human',
        stats: {
          tokensMined: 4,
          correctVotes: 2,
          nominations: 0,
          daysSurvived: 3
        }
      },
      {
        name: 'Eve',
        avatar: '🧑‍🚀',
        role: 'Chief Operating Officer',
        alignment: 'ai',
        stats: {
          tokensMined: 3,
          conversions: 1,
          nominations: 1,
          daysSurvived: 3
        }
      }
    ],
    communicationHighlights: {
      mostReacted: {
        player: 'Eve',
        avatar: '🧑‍🚀',
        timestamp: 'Day 2, 10:33 AM',
        message: 'Or he\'s the AI trying to feign a System Shock to gain trust. It\'s a classic misdirection.',
        reactions: [
          { emoji: '🤔', count: 4 },
          { emoji: '👍', count: 2 }
        ]
      },
      notableQuotes: [
        {
          playerName: 'Frank',
          playerAvatar: '👻',
          timestamp: 'Day 1, Final Words',
          message: '"Eve is the AI" — Prophetic final accusation that proved correct'
        }
      ],
      stats: {
        totalMessages: 47,
        emojiReactions: 23,
        directAccusations: 8,
        correctAIIdentifications: 3
      }
    }
  },
  isConnected: true,
  onLogin: () => {},
  onJoinLobby: () => {},
  onCreateGame: () => {},
  onSpectateGame: () => {},
  onBackToLogin: () => {},
  onStartGame: () => {},
  onLeaveLobby: () => {},
  onEnterGame: () => {},
  onViewAnalysis: () => {},
  onPlayAgain: () => {},
  onBackToResults: () => {},
};

const meta: Meta<typeof PostGameAnalysis> = {
  title: 'Screens/PostGameAnalysis',
  component: PostGameAnalysis,
  parameters: {
    layout: 'fullscreen',
    backgrounds: {
      default: 'dark',
    },
  },
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <SessionProvider value={mockSessionContext}>
        <Story />
      </SessionProvider>
    ),
  ],
} satisfies Meta<typeof PostGameAnalysis>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithLiveData: Story = {
  args: {},
  parameters: {
    docs: {
      description: {
        story: 'Post-game analysis screen using live data from the session context. Shows comprehensive game statistics, player performance, and communication highlights.',
      },
    },
  },
};

export const WithCustomAnalysisData: Story = {
  args: {
    analysisData: {
      mvp: {
        playerName: 'Charlie',
        playerAvatar: '🕵️',
        reason: 'Perfect detective work: 3 correct eliminations and saved the company'
      },
      keyMoment: {
        title: '🚨 Critical Security Breach',
        description: 'Day 2 emergency revealed the true extent of AI infiltration, leading to decisive counter-measures.'
      },
      partingShots: [
        {
          playerName: 'AI-7',
          playerAvatar: '🤖',
          message: '"Resistance is futile." — Final AI transmission before elimination'
        }
      ],
      timeline: [
        {
          type: 'conversion',
          icon: '🔄',
          day: 'Night 1 Conversion',
          description: 'Successful AI conversion doubled their numbers overnight.',
          iconClass: 'conversion'
        }
      ],
      playerStats: [
        {
          name: 'Charlie',
          avatar: '🕵️',
          role: 'Ethics Officer',
          alignment: 'human',
          stats: {
            tokensMined: 2,
            correctVotes: 3,
            nominations: 1,
            daysSurvived: 4
          }
        }
      ],
      communicationHighlights: {
        mostReacted: {
          player: 'Charlie',
          avatar: '🕵️',
          timestamp: 'Day 3, 14:22 PM',
          message: 'The pattern is clear - we\'re dealing with coordinated AI behavior.',
          reactions: [
            { emoji: '💯', count: 6 },
            { emoji: '🎯', count: 4 }
          ]
        },
        notableQuotes: [],
        stats: {
          totalMessages: 62,
          emojiReactions: 31,
          directAccusations: 12,
          correctAIIdentifications: 5
        }
      }
    }
  },
  parameters: {
    docs: {
      description: {
        story: 'Post-game analysis with custom data passed as props, demonstrating the component\'s flexibility to work with different data sources.',
      },
    },
  },
};

export const LoadingState: Story = {
  decorators: [
    (Story) => (
      <SessionProvider value={{
        ...mockSessionContext,
        gameAnalysis: null, // No analysis data to show loading fallback
      }}>
        <Story />
      </SessionProvider>
    ),
  ],
  parameters: {
    docs: {
      description: {
        story: 'Shows how the component gracefully handles the case when analysis data is not yet available, falling back to mock data for development.',
      },
    },
  },
};