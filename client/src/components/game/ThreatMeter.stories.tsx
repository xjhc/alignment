import type { Meta, StoryObj } from '@storybook/react';
import { ThreatMeter } from './ThreatMeter';

// The 'meta' object describes your component
const meta: Meta<typeof ThreatMeter> = {
  title: 'Game/ThreatMeter',
  component: ThreatMeter,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    tokens: { 
      control: { type: 'number', min: 0, max: 20, step: 1 },
      description: 'Number of tokens the player has'
    },
    aiEquity: { 
      control: { type: 'number', min: 0, max: 20, step: 1 },
      description: 'Amount of AI equity the player has accumulated'
    },
  },
} satisfies Meta<typeof ThreatMeter>;

export default meta;
type Story = StoryObj<typeof meta>;

// --- Stories ---

// Safe status - low AI equity relative to tokens
export const Safe: Story = {
  args: {
    tokens: 10,
    aiEquity: 2,
  },
};

// Danger status - moderate AI equity
export const Danger: Story = {
  args: {
    tokens: 10,
    aiEquity: 6,
  },
};

// Critical status - high AI equity
export const Critical: Story = {
  args: {
    tokens: 10,
    aiEquity: 8,
  },
};

// Aligned status - AI equity equals or exceeds tokens
export const Aligned: Story = {
  args: {
    tokens: 10,
    aiEquity: 10,
  },
};

// Edge case: No tokens (immediate alignment)
export const NoTokens: Story = {
  args: {
    tokens: 0,
    aiEquity: 5,
  },
};

// Starting state - no AI equity yet
export const Starting: Story = {
  args: {
    tokens: 5,
    aiEquity: 0,
  },
};

// High tokens, high equity but still safe
export const HighTokensStillSafe: Story = {
  args: {
    tokens: 20,
    aiEquity: 8,
  },
};