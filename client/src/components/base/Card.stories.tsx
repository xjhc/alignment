import type { Meta, StoryObj } from '@storybook/react';
import { Card } from './Card';

const meta: Meta<typeof Card> = {
  title: 'Base/Card',
  component: Card,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'outlined', 'elevated'],
    },
    padding: {
      control: 'select',
      options: ['none', 'sm', 'md', 'lg'],
    },
    hoverable: { control: 'boolean' },
  },
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: (
      <div>
        <h3>Card Title</h3>
        <p>This is the default card with standard styling and medium padding.</p>
      </div>
    ),
  },
};

export const Outlined: Story = {
  args: {
    variant: 'outlined',
    children: (
      <div>
        <h3>Outlined Card</h3>
        <p>This card has a transparent background with a border outline.</p>
      </div>
    ),
  },
};

export const Elevated: Story = {
  args: {
    variant: 'elevated',
    children: (
      <div>
        <h3>Elevated Card</h3>
        <p>This card has a shadow effect to create visual elevation.</p>
      </div>
    ),
  },
};

export const SmallPadding: Story = {
  args: {
    padding: 'sm',
    children: (
      <div>
        <h3>Small Padding</h3>
        <p>This card uses small padding (12px).</p>
      </div>
    ),
  },
};

export const LargePadding: Story = {
  args: {
    padding: 'lg',
    children: (
      <div>
        <h3>Large Padding</h3>
        <p>This card uses large padding (24px).</p>
      </div>
    ),
  },
};

export const NoPadding: Story = {
  args: {
    padding: 'none',
    children: (
      <div className="p-4">
        <h3>No Padding</h3>
        <p>This card has no built-in padding, but content has custom padding.</p>
      </div>
    ),
  },
};

export const Hoverable: Story = {
  args: {
    hoverable: true,
    children: (
      <div>
        <h3>Hoverable Card</h3>
        <p>Hover over this card to see the hover effect.</p>
      </div>
    ),
  },
};

export const Clickable: Story = {
  args: {
    onClick: () => alert('Card clicked!'),
    children: (
      <div>
        <h3>Clickable Card</h3>
        <p>This card is clickable and will show an alert when clicked.</p>
      </div>
    ),
  },
};

export const WithSubComponents: Story = {
  render: () => (
    <Card>
      <Card.Header>
        <h2>Card with Sub-components</h2>
      </Card.Header>
      <Card.Body>
        <p>This demonstrates using Card.Header, Card.Body, and Card.Footer sub-components.</p>
        <p>The body contains the main content of the card.</p>
      </Card.Body>
      <Card.Footer>
        <button className="btn-primary">Action Button</button>
      </Card.Footer>
    </Card>
  ),
};

export const ComplexLayout: Story = {
  render: () => (
    <div className="space-y-4">
      <Card variant="elevated" hoverable>
        <Card.Header>
          <h3>Feature Card</h3>
        </Card.Header>
        <Card.Body>
          <p>A more complex card layout with multiple sections and interactive elements.</p>
          <ul className="list-disc ml-4 mt-2">
            <li>Elevated variant for visual prominence</li>
            <li>Hoverable for interactive feedback</li>
            <li>Structured with header, body, and footer</li>
          </ul>
        </Card.Body>
        <Card.Footer>
          <div className="flex gap-2">
            <button className="btn-primary">Primary Action</button>
            <button className="btn-secondary">Secondary</button>
          </div>
        </Card.Footer>
      </Card>
    </div>
  ),
};