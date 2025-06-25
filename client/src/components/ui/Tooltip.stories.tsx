import type { Meta, StoryObj } from '@storybook/react';
import { Tooltip } from './Tooltip';

const meta: Meta<typeof Tooltip> = {
  title: 'UI/Tooltip',
  component: Tooltip,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    position: {
      control: 'select',
      options: ['top', 'bottom', 'left', 'right'],
    },
    delay: { control: 'number' },
    content: { control: 'text' },
  },
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    content: 'This is a helpful tooltip',
    children: <button className="btn-primary">Hover me</button>,
  },
};

export const TopPosition: Story = {
  args: {
    content: 'Tooltip appears above',
    position: 'top',
    children: <button className="btn-secondary">Top Tooltip</button>,
  },
};

export const BottomPosition: Story = {
  args: {
    content: 'Tooltip appears below',
    position: 'bottom',
    children: <button className="btn-secondary">Bottom Tooltip</button>,
  },
};

export const LeftPosition: Story = {
  args: {
    content: 'Tooltip appears to the left',
    position: 'left',
    children: <button className="btn-secondary">Left Tooltip</button>,
  },
};

export const RightPosition: Story = {
  args: {
    content: 'Tooltip appears to the right',
    position: 'right',
    children: <button className="btn-secondary">Right Tooltip</button>,
  },
};

export const NoDelay: Story = {
  args: {
    content: 'Instant tooltip',
    delay: 0,
    children: <button className="btn-primary">No Delay</button>,
  },
};

export const LongDelay: Story = {
  args: {
    content: 'This tooltip takes 2 seconds to appear',
    delay: 2000,
    children: <button className="btn-primary">Long Delay</button>,
  },
};

export const LongContent: Story = {
  args: {
    content: 'This is a much longer tooltip that demonstrates how the component handles text wrapping and longer content.',
    children: <button className="btn-primary">Long Content</button>,
  },
};

export const OnIcon: Story = {
  args: {
    content: 'Help information about this feature',
    children: <span className="cursor-help text-lg">ℹ️</span>,
  },
};

export const OnDisabledButton: Story = {
  args: {
    content: 'This action is not available right now',
    children: <button className="btn-primary" disabled>Disabled Button</button>,
  },
};

export const MultipleTooltips: Story = {
  render: () => (
    <div className="flex gap-4 items-center">
      <Tooltip content="First tooltip" position="top">
        <button className="btn-primary">Button 1</button>
      </Tooltip>
      <Tooltip content="Second tooltip" position="bottom">
        <button className="btn-secondary">Button 2</button>
      </Tooltip>
      <Tooltip content="Third tooltip" position="left">
        <button className="btn-outline">Button 3</button>
      </Tooltip>
      <Tooltip content="Fourth tooltip" position="right">
        <button className="btn-ghost">Button 4</button>
      </Tooltip>
    </div>
  ),
};