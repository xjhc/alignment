import type { Meta, StoryObj } from "@storybook/react";
import { VoteBlock } from "./VoteBlock";
import { Player } from "../../types/generated";

const meta: Meta<typeof VoteBlock> = {
  title: "Game/VoteBlock",
  component: VoteBlock,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A blockchain-style vote block representing a single player's vote with their token weight. Used in the voting UI to create a visual 'chain' of votes.",
      },
    },
  },
  argTypes: {
    player: {
      description: "The player who cast this vote",
    },
    tokenCount: {
      description: "The number of tokens this vote is worth",
      control: { type: "number", min: 0, max: 50 },
    },
    isSelf: {
      description: "Whether this vote belongs to the current player",
      control: { type: "boolean" },
    },
    isAnimating: {
      description: "Whether to show entrance animation",
      control: { type: "boolean" },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof VoteBlock>;

// Mock player data
const mockPlayer: Player = {
  id: "player1",
  name: "Alice Chen",
  jobTitle: "CTO",
  controlType: "HUMAN",
  status: "ALIVE",
  isAlive: true,
  connectionStatus: "CONNECTED",
  tokens: 5,
  projectMilestones: 2,
  statusMessage: "",
  joinedAt: new Date(),
};

const mockPlayerCEO: Player = {
  ...mockPlayer,
  id: "player2",
  name: "Bob Wilson",
  jobTitle: "CEO",
  tokens: 12,
};

const mockPlayerIntern: Player = {
  ...mockPlayer,
  id: "player3",
  name: "Charlie Rodriguez",
  jobTitle: "Intern",
  tokens: 1,
};

const mockPlayerLongName: Player = {
  ...mockPlayer,
  id: "player4",
  name: "Alexander Thompson III",
  jobTitle: "CFO",
  tokens: 8,
};

export const Default: Story = {
  args: {
    player: mockPlayer,
    tokenCount: 5,
    isSelf: false,
    isAnimating: false,
  },
};

export const MyVote: Story = {
  args: {
    player: mockPlayer,
    tokenCount: 5,
    isSelf: true,
    isAnimating: false,
  },
};

export const HighTokenCount: Story = {
  args: {
    player: mockPlayerCEO,
    tokenCount: 12,
    isSelf: false,
    isAnimating: false,
  },
};

export const LowTokenCount: Story = {
  args: {
    player: mockPlayerIntern,
    tokenCount: 1,
    isSelf: false,
    isAnimating: false,
  },
};

export const LongPlayerName: Story = {
  args: {
    player: mockPlayerLongName,
    tokenCount: 8,
    isSelf: false,
    isAnimating: false,
  },
};

export const WithAnimation: Story = {
  args: {
    player: mockPlayer,
    tokenCount: 5,
    isSelf: false,
    isAnimating: true,
  },
};

export const MyVoteHighTokens: Story = {
  args: {
    player: mockPlayerCEO,
    tokenCount: 12,
    isSelf: true,
    isAnimating: false,
  },
};

// Interactive playground
export const Playground: Story = {
  args: {
    player: mockPlayer,
    tokenCount: 5,
    isSelf: false,
    isAnimating: false,
  },
};

// Multiple blocks for chain visualization
export const VoteChain: Story = {
  render: () => (
    <div style={{ 
      display: "flex", 
      alignItems: "center", 
      gap: "4px",
      padding: "16px",
      background: "#1a1a1a",
      borderRadius: "8px",
      overflow: "auto",
      maxWidth: "600px"
    }}>
      <VoteBlock player={mockPlayerCEO} tokenCount={12} isSelf={false} />
      <VoteBlock player={mockPlayer} tokenCount={5} isSelf={true} />
      <VoteBlock player={mockPlayerIntern} tokenCount={1} isSelf={false} />
      <VoteBlock player={mockPlayerLongName} tokenCount={8} isSelf={false} />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Multiple vote blocks chained together as they would appear in the voting UI",
      },
    },
  },
};