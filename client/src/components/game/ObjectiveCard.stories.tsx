import type { Meta, StoryObj } from '@storybook/react';
import { ObjectiveCard } from './ObjectiveCard';

const meta: Meta<typeof ObjectiveCard> = {
  title: 'Game/ObjectiveCard',
  component: ObjectiveCard,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: 'select',
      options: ['Team Objective', 'Personal KPI', 'Mandate'],
    },
    name: { control: 'text' },
    description: { control: 'text' },
    progressText: { control: 'text' },
    isPrivate: { control: 'boolean' },
  },
} satisfies Meta<typeof ObjectiveCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TeamObjective: Story = {
  args: {
    type: 'Team Objective',
    name: 'Secure the Network',
    description: 'Work together to identify and eliminate all security threats to the company network.',
    progressText: '2 of 3 threats eliminated',
    isPrivate: false,
  },
};

export const PersonalKPI: Story = {
  args: {
    type: 'Personal KPI',
    name: 'Threat Analysis',
    description: 'Identify and neutralize 2 security threats using your security expertise.',
    progressText: '1 of 2 completed',
    isPrivate: true,
  },
};

export const PersonalKPICompleted: Story = {
  args: {
    type: 'Personal KPI',
    name: 'Code Review Excellence',
    description: 'Complete 3 thorough code reviews to ensure system security.',
    progressText: '3 of 3 completed ✓',
    isPrivate: true,
  },
};

export const Mandate: Story = {
  args: {
    type: 'Mandate',
    name: 'AI Compliance',
    description: 'Ensure all AI systems operate within ethical guidelines and company policies.',
    progressText: 'Ongoing assessment',
    isPrivate: false,
  },
};

export const MandatePrivate: Story = {
  args: {
    type: 'Mandate',
    name: 'System Integration',
    description: 'Integrate new AI capabilities while maintaining operational security.',
    progressText: 'Phase 2 of 3 active',
    isPrivate: true,
  },
};

export const NoProgress: Story = {
  args: {
    type: 'Personal KPI',
    name: 'Network Monitoring',
    description: 'Monitor network traffic for anomalous patterns and potential security breaches.',
    isPrivate: true,
  },
};

export const TeamObjectiveCritical: Story = {
  args: {
    type: 'Team Objective',
    name: 'Crisis Response',
    description: 'Respond to the active security breach and restore system integrity before data loss occurs.',
    progressText: 'CRITICAL - 12 minutes remaining',
    isPrivate: false,
  },
};