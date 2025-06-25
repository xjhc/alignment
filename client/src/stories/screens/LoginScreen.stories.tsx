import type { Meta, StoryObj } from '@storybook/react';
import { LoginScreen } from '../../components/LoginScreen';

const meta: Meta<typeof LoginScreen> = {
  title: 'Screens/LoginScreen',
  component: LoginScreen,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  argTypes: {
    onLogin: { action: 'onLogin' },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    onLogin: (playerName, avatar) => {
      console.log('Logged in as:', playerName, 'with avatar:', avatar);
    },
  },
};

export const EmptyInput: Story = {
  args: {
    ...Default.args,
  },
  // You can add interactions to test the form behavior
  play: async ({ canvasElement }) => {
    // const canvas = within(canvasElement);
    // const loginButton = await canvas.getByRole('button', { name: /browse lobbies/i });
    // expect(loginButton).toBeDisabled();
  },
};
