import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GameProvider } from '../../contexts/GameContext';
import { PulseCheckInput } from '../game/PulseCheckInput';
import type { Player, GameState } from '../../types';
import { KPIType } from '../../types/generated';

// Mock the WebSocket context
vi.mock('../../contexts/WebSocketContext', () => ({
  useWebSocketContext: () => ({
    isConnected: true,
    sendAction: vi.fn(),
  }),
}));

const createTestPlayer = (overrides: Partial<Player> = {}): Player => ({
  id: 'test-player-1',
  name: 'TestPlayer',
  jobTitle: 'Developer',
  controlType: 'HUMAN',
  isAlive: true,
  tokens: 5,
  projectMilestones: 3,
  statusMessage: 'Working hard',
  alignment: 'HUMAN' as const,
  avatar: '👤',
  joinedAt: '2024-01-01T00:00:00Z',
  role: {
    type: 'DEVELOPER' as any,
    name: 'Developer',
    description: 'Developer role',
    isUnlocked: true,
  },
  personalKPI: {
    type: KPIType.Capitalist,
    description: 'Complete objectives',
    progress: 3,
    target: 3,
    isCompleted: true,
    reward: 'Role unlock',
  },
  hasSubmittedPulseCheck: false,
  ...overrides,
});

const createTestGameState = (players: Player[] = []): GameState => ({
  id: 'test-game',
  players,
  phase: 'DAY' as any,
  dayNumber: 1,
  chatMessages: [],
});

describe('PulseCheckInput', () => {
  let mockHandlePulseCheck: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockHandlePulseCheck = vi.fn().mockResolvedValue(undefined);
    vi.clearAllMocks();
  });

  const renderComponent = (props = {}, playerOverrides: Partial<Player> = {}) => {
    const testPlayer = createTestPlayer(playerOverrides);
    const gameState = createTestGameState([testPlayer]);
    
    const defaultProps = {
      handlePulseCheck: mockHandlePulseCheck,
      localPlayerName: testPlayer.name,
      ...props,
    };

    return render(
      <GameProvider gameState={gameState} localPlayerId={testPlayer.id}>
        <PulseCheckInput {...defaultProps} />
      </GameProvider>
    );
  };

  // TODO: Update this test when PulseCheckInput is changed to use three buttons instead of text input
  it('renders three buttons for pulse check responses instead of text input', () => {
    renderComponent();

    // TODO: When the component is updated, uncomment these assertions:
    // Assert that no text input is rendered - we now use three buttons instead
    // expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    // 
    // // Verify there are exactly three pulse check response buttons
    // const buttons = screen.getAllByRole('button');
    // expect(buttons).toHaveLength(3);
    // 
    // // Check for expected button texts (these may vary based on implementation)
    // const buttonTexts = buttons.map(button => button.textContent?.toLowerCase());
    // expect(buttonTexts).toContain('confident');
    // expect(buttonTexts).toContain('concerned'); 
    // expect(buttonTexts).toContain('uncertain');
    // 
    // // All buttons should be enabled (no text input to validate)
    // buttons.forEach(button => {
    //   expect(button).toBeEnabled();
    // });

    // Assert that no text input is rendered - we now use three buttons instead
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    
    // Verify there are exactly three pulse check response buttons
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(3);
    
    // Check for expected button texts 
    const buttonTexts = buttons.map(button => button.textContent?.toLowerCase());
    expect(buttonTexts.some(text => text?.includes('confident'))).toBe(true);
    expect(buttonTexts.some(text => text?.includes('concerned'))).toBe(true); 
    expect(buttonTexts.some(text => text?.includes('suspicious'))).toBe(true);
  });

  it('displays pulse check title and question', () => {
    renderComponent();

    expect(screen.getByText('💭 PULSE CHECK')).toBeInTheDocument();
    expect(screen.getByText(/What is your immediate response to the current crisis/)).toBeInTheDocument();
  });

  it('displays player name in instruction text', () => {
    renderComponent({ localPlayerName: 'Alice' });

    expect(screen.getByText(/As Alice, choose your response:/)).toBeInTheDocument();
  });

  // TODO: Update this test when component uses buttons instead of text input
  it('calls handlePulseCheck when a response button is clicked', async () => {
    renderComponent();

    // Click on the "Concerned" button
    const concernedButton = screen.getByRole('button', { name: /concerned/i });
    fireEvent.click(concernedButton);
    
    await waitFor(() => {
      expect(mockHandlePulseCheck).toHaveBeenCalledWith('I have concerns about how things are progressing right now.');
    });
  });

  // TODO: Update this test when component uses buttons instead of text input
  it('disables all buttons when submitting', async () => {
    mockHandlePulseCheck.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    renderComponent();

    // TODO: When component is updated, use this approach:
    // const confidentButton = screen.getByRole('button', { name: /confident/i });
    // 
    // // Click a button to submit
    // fireEvent.click(confidentButton);
    // 
    // // All buttons should be disabled while submitting
    // const buttons = screen.getAllByRole('button');
    // buttons.forEach(button => {
    //   expect(button).toBeDisabled();
    // });

    const confidentButton = screen.getByRole('button', { name: /confident/i });
    
    // Click a button to submit
    fireEvent.click(confidentButton);
    
    // All buttons should be disabled while submitting
    const buttons = screen.getAllByRole('button');
    buttons.forEach(button => {
      expect(button).toBeDisabled();
    });
  });

  // TODO: Update this test when component uses buttons instead of text input
  it('shows response options clearly', () => {
    renderComponent();

    // Should display all three response options
    expect(screen.getByRole('button', { name: /confident/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /concerned/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /suspicious/i })).toBeInTheDocument();
  });

  it('displays custom question when provided', () => {
    const customQuestion = 'How do you feel about the current situation?';
    renderComponent({ question: customQuestion });

    expect(screen.getByText(`"${customQuestion}"`)).toBeInTheDocument();
  });

  // TODO: Update this test when component uses buttons instead of text input
  it('allows clicking different response buttons', async () => {
    renderComponent();

    // Test clicking confident button
    const confidentButton = screen.getByRole('button', { name: /confident/i });
    fireEvent.click(confidentButton);
    
    await waitFor(() => {
      expect(mockHandlePulseCheck).toHaveBeenCalledWith('I feel confident about our current situation and next steps.');
    });
  });
});