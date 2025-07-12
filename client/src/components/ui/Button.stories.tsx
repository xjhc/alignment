import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'UI/Button',
  component: Button,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'danger', 'ghost', 'outline', 'success'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    fullWidth: { control: 'boolean' },
    disabled: { control: 'boolean' },
    isLoading: { control: 'boolean' },
    disableSound: { control: 'boolean' },
    hapticFeedback: {
      control: 'select',
      options: ['light', 'medium', 'strong', 'confirm', 'vote', false],
    },
    children: { control: 'text' },
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    variant: 'primary',
    children: 'Primary Button',
  },
};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    children: 'Secondary Button',
  },
};

export const Danger: Story = {
  args: {
    variant: 'danger',
    children: 'Danger Button',
  },
};

export const Ghost: Story = {
  args: {
    variant: 'ghost',
    children: 'Ghost Button',
  },
};

export const Outline: Story = {
  args: {
    variant: 'outline',
    children: 'Outline Button',
  },
};

export const Success: Story = {
  args: {
    variant: 'success',
    children: 'Success Button',
  },
};

export const SmallSize: Story = {
  args: {
    size: 'sm',
    children: 'Small Button',
  },
};

export const LargeSize: Story = {
  args: {
    size: 'lg',
    children: 'Large Button',
  },
};

export const FullWidth: Story = {
  args: {
    fullWidth: true,
    children: 'Full Width Button',
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    children: 'Disabled Button',
  },
};

export const IsLoading: Story = {
  args: {
    isLoading: true,
    children: 'Loading Button',
  },
  parameters: {
    docs: {
      description: {
        story: 'When isLoading is true, the button shows a spinner and is automatically disabled.',
      },
    },
  },
};

export const WithLeftIcon: Story = {
  args: {
    leftIcon: <span>🔒</span>,
    children: 'Login',
  },
};

export const WithRightIcon: Story = {
  args: {
    rightIcon: <span>→</span>,
    children: 'Continue',
  },
};

export const LoadingInteractive: Story = {
  args: {
    children: 'Submit',
  },
  render: (args) => {
    const [isLoading, setIsLoading] = React.useState(false);

    const handleClick = () => {
      setIsLoading(true);
      setTimeout(() => setIsLoading(false), 2000);
    };

    return (
      <Button
        {...args}
        isLoading={isLoading}
        onClick={handleClick}
      >
        {isLoading ? 'Submitting...' : 'Submit'}
      </Button>
    );
  },
  parameters: {
    docs: {
      description: {
        story: 'Click to see the loading state in action. The loading will automatically stop after 2 seconds.',
      },
    },
  },
};

export const HapticFeedbackVariations: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <Button hapticFeedback="light">Light Haptic</Button>
        <Button hapticFeedback="medium">Medium Haptic</Button>
        <Button hapticFeedback="strong">Strong Haptic</Button>
        <Button hapticFeedback="confirm" variant="success">Confirm Haptic</Button>
        <Button hapticFeedback="vote" variant="primary">Vote Haptic</Button>
        <Button hapticFeedback={false}>No Haptic</Button>
      </div>
      <p className="text-sm text-text-secondary">
        Click buttons on mobile devices to feel different haptic feedback patterns.
        On desktop, these have no effect but demonstrate the different haptic intensity options.
      </p>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different haptic feedback intensities for mobile devices. Each button provides unique tactile feedback when pressed.',
      },
    },
  },
};

export const SoundAndHapticDisabled: Story = {
  args: {
    children: 'Silent Button',
    disableSound: true,
    hapticFeedback: false,
  },
  parameters: {
    docs: {
      description: {
        story: 'Button with both sound and haptic feedback disabled for situations requiring silent interaction.',
      },
    },
  },
};