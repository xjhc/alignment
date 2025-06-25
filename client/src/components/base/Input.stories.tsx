import type { Meta, StoryObj } from '@storybook/react';
import { Input } from './Input';

const meta: Meta<typeof Input> = {
  title: 'Base/Input',
  component: Input,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    variant: {
      control: 'select',
      options: ['default', 'filled'],
    },
    fullWidth: { control: 'boolean' },
    disabled: { control: 'boolean' },
    hasError: { control: 'boolean' },
    label: { control: 'text' },
    error: { control: 'text' },
    helperText: { control: 'text' },
    placeholder: { control: 'text' },
  },
} satisfies Meta<typeof Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    placeholder: 'Enter text...',
  },
};

export const WithLabel: Story = {
  args: {
    label: 'Username',
    placeholder: 'Enter your username',
  },
};

export const WithHelperText: Story = {
  args: {
    label: 'Password',
    placeholder: 'Enter your password',
    helperText: 'Password must be at least 8 characters long',
    type: 'password',
  },
};

export const WithError: Story = {
  args: {
    label: 'Email',
    placeholder: 'Enter your email',
    error: 'Please enter a valid email address',
    value: 'invalid-email',
  },
};

export const HasErrorState: Story = {
  args: {
    label: 'Email',
    placeholder: 'Enter your email',
    hasError: true,
    value: 'invalid-email',
  },
  parameters: {
    docs: {
      description: {
        story: 'Using the hasError prop to show error state without error text.',
      },
    },
  },
};

export const SmallSize: Story = {
  args: {
    size: 'sm',
    label: 'Small Input',
    placeholder: 'Small size input',
  },
};

export const LargeSize: Story = {
  args: {
    size: 'lg',
    label: 'Large Input',
    placeholder: 'Large size input',
  },
};

export const FilledVariant: Story = {
  args: {
    variant: 'filled',
    label: 'Filled Input',
    placeholder: 'Filled background',
  },
};

export const FullWidth: Story = {
  args: {
    fullWidth: true,
    label: 'Full Width Input',
    placeholder: 'This input spans full width',
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    label: 'Disabled Input',
    placeholder: 'This input is disabled',
    value: 'Some value',
  },
};

export const WithLeftIcon: Story = {
  args: {
    label: 'Search',
    placeholder: 'Search...',
    leftIcon: <span>🔍</span>,
  },
};

export const WithRightIcon: Story = {
  args: {
    label: 'Secure Input',
    placeholder: 'Enter secure data',
    rightIcon: <span>🔒</span>,
  },
};

export const WithBothIcons: Story = {
  args: {
    label: 'Username',
    placeholder: 'Enter username',
    leftIcon: <span>👤</span>,
    rightIcon: <span>✓</span>,
  },
};