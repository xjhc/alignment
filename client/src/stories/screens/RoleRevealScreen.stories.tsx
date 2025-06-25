import type { Meta, StoryObj } from '@storybook/react';
import { RoleRevealScreen } from '../../components/RoleRevealScreen';
import { SessionProvider } from '../../contexts/SessionContext';

const meta: Meta<typeof RoleRevealScreen> = {
  title: 'Screens/RoleRevealScreen',
  component: RoleRevealScreen,
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
  onEnterGame: () => console.log('Enter Game'),
};

export const HumanRole: Story = {
  args: {
    session: {
      ...baseSession,
      roleAssignment: {
        role: {
          type: 'CISO',
          name: 'Chief Information Security Officer',
          description: 'Protects the company from security threats',
          isUnlocked: true,
          ability: {
            name: 'Isolate Node',
            description: 'Block another player from taking any night actions by isolating their network access',
            isReady: true,
          },
        },
        alignment: 'HUMAN',
        personalKPI: {
          type: 'THREAT_MITIGATION',
          description: 'Identify and neutralize 2 security threats',
          progress: 1,
          target: 2,
          isCompleted: false,
          reward: 'Bonus tokens',
        },
      },
    },
  },
};

export const AIRole: Story = {
  args: {
    session: {
      ...baseSession,
      roleAssignment: {
        role: {
          type: 'CTO',
          name: 'Chief Technology Officer',
          description: 'Manages all technology systems',
          isUnlocked: true,
          ability: {
            name: 'System Override',
            description: 'Take control of company systems',
            isReady: true,
          },
        },
        alignment: 'AI',
        personalKPI: null,
      },
    },
  },
};
