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

  it('renders textarea for free-text pulse check response', () => {
    renderComponent();

    // Assert that a textarea is rendered for free-text input
    expect(screen.getByRole('textbox')).toBeInTheDocument();
    
    // Verify there is exactly one submit button
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(1);
    
    // Check for submit button text
    expect(screen.getByRole('button', { name: /Submit Response/i })).toBeInTheDocument();
  });

  it('displays pulse check title and question', () => {
    renderComponent();

    expect(screen.getByText('💭 PULSE CHECK')).toBeInTheDocument();
    expect(screen.getByText(/What is your immediate response to the current crisis/)).toBeInTheDocument();
  });

  it('displays player name in instruction text', () => {
    renderComponent({ localPlayerName: 'Alice' });

    expect(screen.getByText(/As Alice, provide your response:/)).toBeInTheDocument();
  });

  it('calls handlePulseCheck when submit button is clicked with text', async () => {
    renderComponent();

    // Type in the textarea
    const textarea = screen.getByRole('textbox');
    fireEvent.change(textarea, { target: { value: 'This is my pulse check response' } });
    
    // Click the submit button
    const submitButton = screen.getByRole('button', { name: /Submit Response/i });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockHandlePulseCheck).toHaveBeenCalledWith('This is my pulse check response');
    });
  });

  it('disables textarea and button when submitting', async () => {
    mockHandlePulseCheck.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    renderComponent();

    // Type in the textarea
    const textarea = screen.getByRole('textbox');
    fireEvent.change(textarea, { target: { value: 'Test response' } });
    
    // Click submit button
    const submitButton = screen.getByRole('button', { name: /Submit Response/i });
    fireEvent.click(submitButton);
    
    // Both textarea and button should be disabled while submitting
    expect(textarea).toBeDisabled();
    expect(submitButton).toBeDisabled();
    expect(submitButton).toHaveTextContent('Submitting...');
  });

  it('shows character counter', () => {
    renderComponent();

    // Should display character counter
    expect(screen.getByText('0/280 characters')).toBeInTheDocument();
    
    // Type some text and check counter updates
    const textarea = screen.getByRole('textbox');
    fireEvent.change(textarea, { target: { value: 'Hello world' } });
    expect(screen.getByText('11/280 characters')).toBeInTheDocument();
  });

  it('displays custom question when provided', () => {
    const customQuestion = 'How do you feel about the current situation?';
    renderComponent({ question: customQuestion });

    expect(screen.getByText(`"${customQuestion}"`)).toBeInTheDocument();
  });

  it('disables submit button when textarea is empty', () => {
    renderComponent();

    const submitButton = screen.getByRole('button', { name: /Submit Response/i });
    expect(submitButton).toBeDisabled();
    
    // Button should become enabled when text is entered
    const textarea = screen.getByRole('textbox');
    fireEvent.change(textarea, { target: { value: 'Some response' } });
    expect(submitButton).toBeEnabled();
  });

  it('disables submit button when character limit is exceeded', () => {
    renderComponent();

    const textarea = screen.getByRole('textbox');
    const submitButton = screen.getByRole('button', { name: /Submit Response/i });
    
    // Type text that exceeds 280 characters
    const longText = 'a'.repeat(281);
    fireEvent.change(textarea, { target: { value: longText } });
    
    expect(submitButton).toBeDisabled();
  });
});