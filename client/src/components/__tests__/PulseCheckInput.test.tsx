import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GameProvider } from '../../contexts/GameContext';
import { PulseCheckInput } from '../game/PulseCheckInput';
import type { Player, GameState } from '../../types';

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
    type: 'PRODUCTIVITY' as const,
    description: 'Complete objectives',
    progress: 3,
    target: 3,
    isCompleted: true,
    reward: 'Role unlock',
  },
  hasSubmittedPulseCheck: false,
  nightActions: [],
  voteTarget: null,
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

  it('renders a text input for freeform pulse check responses', () => {
    renderComponent();

    // Assert that a text input is rendered for freeform responses
    const textInput = screen.getByRole('textbox');
    expect(textInput).toBeInTheDocument();
    expect(textInput).toHaveAttribute('placeholder', 'Enter your response (max 200 characters)...');
    expect(textInput).toHaveAttribute('maxLength', '200');

    // Verify there's a submit button
    const submitButton = screen.getByRole('button', { name: /submit/i });
    expect(submitButton).toBeInTheDocument();

    // Ensure the old preset buttons are not present
    expect(screen.queryByRole('button', { name: /nominal/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /elevated/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /critical/i })).not.toBeInTheDocument();
  });

  it('displays pulse check title and question', () => {
    renderComponent();

    expect(screen.getByText('💭 PULSE CHECK')).toBeInTheDocument();
    expect(screen.getByText(/What is your immediate response to the current crisis/)).toBeInTheDocument();
  });

  it('displays player name in instruction text', () => {
    renderComponent({ localPlayerName: 'Alice' });

    expect(screen.getByText(/As Alice, your response:/)).toBeInTheDocument();
  });

  it('calls handlePulseCheck when submit button is clicked with text', async () => {
    renderComponent();

    const textInput = screen.getByRole('textbox');
    const submitButton = screen.getByRole('button', { name: /submit/i });

    // Type some text
    fireEvent.change(textInput, { target: { value: 'This is concerning' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockHandlePulseCheck).toHaveBeenCalledWith('This is concerning');
    });
  });

  it('disables input and button when submitting', async () => {
    mockHandlePulseCheck.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    renderComponent();

    const textInput = screen.getByRole('textbox');
    const submitButton = screen.getByRole('button', { name: /submit/i });

    // Type some text and submit
    fireEvent.change(textInput, { target: { value: 'Test response' } });
    fireEvent.click(submitButton);

    // Input and button should be disabled while submitting
    expect(textInput).toBeDisabled();
    expect(screen.getByRole('button', { name: /submitting/i })).toBeDisabled();
  });

  it('shows character counter', () => {
    renderComponent();

    const textInput = screen.getByRole('textbox');
    
    // Initially should show 0/200
    expect(screen.getByText('0/200 characters • Press Enter to submit')).toBeInTheDocument();

    // Type some text
    fireEvent.change(textInput, { target: { value: 'Hello world' } });
    expect(screen.getByText('11/200 characters • Press Enter to submit')).toBeInTheDocument();
  });

  it('displays custom question when provided', () => {
    const customQuestion = 'How do you feel about the current situation?';
    renderComponent({ question: customQuestion });

    expect(screen.getByText(`"${customQuestion}"`)).toBeInTheDocument();
  });

  it('submits when Enter key is pressed', async () => {
    renderComponent();

    const textInput = screen.getByRole('textbox');
    
    // Type some text
    fireEvent.change(textInput, { target: { value: 'Enter key response' } });
    
    // Press Enter
    fireEvent.keyDown(textInput, { key: 'Enter', shiftKey: false });

    await waitFor(() => {
      expect(mockHandlePulseCheck).toHaveBeenCalledWith('Enter key response');
    });
  });

  it('does not submit when Enter is pressed with Shift', async () => {
    renderComponent();

    const textInput = screen.getByRole('textbox');
    
    // Type some text
    fireEvent.change(textInput, { target: { value: 'Shift+Enter response' } });
    
    // Press Shift+Enter (should not submit)
    fireEvent.keyDown(textInput, { key: 'Enter', shiftKey: true });

    // Wait a bit to ensure it doesn't submit
    await new Promise(resolve => setTimeout(resolve, 50));
    
    expect(mockHandlePulseCheck).not.toHaveBeenCalled();
  });
});