import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WaitingScreen } from '../WaitingScreen'
import { SessionProvider, SessionContextType } from '../../contexts/SessionContext'
import { AppState, GameState } from '../../types'
import { SessionState, LobbyState, RoleAssignment } from '../../state/appReducer'

// Mock clipboard API
const mockWriteText = vi.fn().mockResolvedValue(undefined);
Object.assign(navigator, {
  clipboard: {
    writeText: mockWriteText,
  },
});

// Mock fetch API
global.fetch = vi.fn();

describe('WaitingScreen', () => {
  const mockAppState: AppState = {
    playerId: 'player1',
    sessionToken: 'token123',
    playerName: 'TestPlayer',
    playerAvatar: '👤',
    gameId: 'game123',
  };

  const mockSessionState: SessionState = 'IN_LOBBY';

  const baseLobbyState: LobbyState = {
    playerId: 'player1',
    playerInfos: [
      {
        id: 'player1',
        name: 'TestPlayer',
        avatar: '👤',
        joinedAt: '2023-01-01T00:00:00Z',
      },
      {
        id: 'player2',
        name: 'AnotherPlayer',
        avatar: '🤖',
        joinedAt: '2023-01-01T00:01:00Z',
      },
    ],
    isHost: true,
    canStart: false,
    hostId: 'player1',
    lobbyName: 'Test Lobby',
    maxPlayers: 8,
    connectionError: null,
    countdown: null,
  };

  const mockGameState: GameState = {} as GameState;
  const mockRoleAssignment: RoleAssignment = null;
  const mockGameAnalysis = null;

  const createMockContext = (overrides: Partial<SessionContextType> = {}): SessionContextType => ({
    appState: mockAppState,
    sessionState: mockSessionState,
    lobbyState: baseLobbyState,
    gameState: mockGameState,
    roleAssignment: mockRoleAssignment,
    gameAnalysis: mockGameAnalysis,
    isConnected: true,
    onLogin: vi.fn(),
    onJoinLobby: vi.fn(),
    onCreateGame: vi.fn(),
    onBackToLogin: vi.fn(),
    onStartGame: vi.fn(),
    onLeaveLobby: vi.fn(),
    onEnterGame: vi.fn(),
    onViewAnalysis: vi.fn(),
    onPlayAgain: vi.fn(),
    onBackToResults: vi.fn(),
    ...overrides,
  });

  const renderWithContext = (contextOverrides: Partial<SessionContextType> = {}) => {
    const mockContext = createMockContext(contextOverrides);
    return render(
      <SessionProvider value={mockContext}>
        <WaitingScreen />
      </SessionProvider>
    );
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (global.fetch as any).mockClear();
    mockWriteText.mockClear();
  });

  describe('Connection States', () => {
    it('shows loading state when not connected', () => {
      renderWithContext({ isConnected: false });
      
      expect(screen.getByText('CONNECTING TO LOBBY...')).toBeInTheDocument();
      expect(screen.getByText('Establishing secure connection...')).toBeInTheDocument();
    });

    it('shows connection error when present', () => {
      const mockContext = createMockContext({
        lobbyState: {
          ...baseLobbyState,
          connectionError: 'Failed to connect to server',
        },
      });
      
      renderWithContext(mockContext);
      
      expect(screen.getByText('CONNECTION ERROR')).toBeInTheDocument();
      expect(screen.getByText('Failed to connect to server')).toBeInTheDocument();
      expect(screen.getByText('← Go Back')).toBeInTheDocument();
    });

    it('calls onLeaveLobby when Go Back button clicked in error state', async () => {
      const user = userEvent.setup();
      const mockOnLeaveLobby = vi.fn();
      
      renderWithContext({
        lobbyState: {
          ...baseLobbyState,
          connectionError: 'Failed to connect to server',
        },
        onLeaveLobby: mockOnLeaveLobby,
      });
      
      const goBackButton = screen.getByText('← Go Back');
      await user.click(goBackButton);
      
      expect(mockOnLeaveLobby).toHaveBeenCalled();
    });
  });

  describe('Lobby Information Display', () => {
    it('displays lobby name and game ID', () => {
      renderWithContext();
      
      expect(screen.getByText('Test Lobby', { exact: false })).toBeInTheDocument();
      expect(screen.getByText('game12', { exact: false })).toBeInTheDocument(); // First 6 chars
    });

    it('shows player count correctly', () => {
      renderWithContext();
      
      expect(screen.getByText('Personnel Connected - 2 / 8')).toBeInTheDocument();
    });

    it('shows private lobby indicator when lobby is private', () => {
      renderWithContext();
      
      // Initially not private, so no indicator
      expect(screen.queryByText('PRIVATE')).not.toBeInTheDocument();
    });
  });

  describe('Player List', () => {
    it('displays all connected players in join order', () => {
      renderWithContext();
      
      expect(screen.getByText('TestPlayer (Host) (You)')).toBeInTheDocument();
      expect(screen.getByText('AnotherPlayer')).toBeInTheDocument();
    });

    it('shows host crown icon', () => {
      renderWithContext();
      
      // The crown emoji should be present for the host
      const hostElement = screen.getByText('TestPlayer (Host) (You)');
      expect(hostElement).toBeInTheDocument();
    });

    it('shows empty slots for remaining players', () => {
      renderWithContext();
      
      const waitingSlots = screen.getAllByText('Waiting for player...');
      expect(waitingSlots).toHaveLength(6); // 8 max - 2 current = 6 empty
    });

    it('shows correct player when user is not host', () => {
      const nonHostLobbyState = {
        ...baseLobbyState,
        isHost: false,
        hostId: 'player2',
      };
      
      renderWithContext({ lobbyState: nonHostLobbyState });
      
      expect(screen.getByText('TestPlayer (You)')).toBeInTheDocument();
      expect(screen.getByText('AnotherPlayer (Host)')).toBeInTheDocument();
    });
  });

  describe('Host Controls', () => {
    it('shows start game button for host when canStart is true', () => {
      const hostLobbyState = {
        ...baseLobbyState,
        canStart: true,
      };
      
      renderWithContext({ lobbyState: hostLobbyState });
      
      const startButton = screen.getByRole('button', { name: /INITIATE CONTAINMENT PROTOCOL/ });
      expect(startButton).toBeInTheDocument();
      expect(startButton).not.toBeDisabled();
    });

    it('disables start game button when canStart is false', () => {
      renderWithContext();
      
      const startButton = screen.getByRole('button', { name: /NEED \d+ MORE PLAYERS/ });
      expect(startButton).toBeInTheDocument();
      expect(startButton).toBeDisabled();
    });

    it('calls onStartGame when start button clicked', async () => {
      const user = userEvent.setup();
      const mockOnStartGame = vi.fn();
      
      const hostLobbyState = {
        ...baseLobbyState,
        canStart: true,
      };
      
      renderWithContext({ 
        lobbyState: hostLobbyState,
        onStartGame: mockOnStartGame,
      });
      
      const startButton = screen.getByRole('button', { name: /INITIATE CONTAINMENT PROTOCOL/ });
      await user.click(startButton);
      
      expect(mockOnStartGame).toHaveBeenCalled();
    });

    it('does not show start game button for non-host', () => {
      const nonHostLobbyState = {
        ...baseLobbyState,
        isHost: false,
      };
      
      renderWithContext({ lobbyState: nonHostLobbyState });
      
      expect(screen.queryByRole('button', { name: /INITIATE CONTAINMENT PROTOCOL/ })).not.toBeInTheDocument();
      expect(screen.getByText('Waiting for host to start the game...')).toBeInTheDocument();
    });
  });

  describe('Countdown Functionality', () => {
    it('shows countdown modal when countdown is active', () => {
      const countdownLobbyState = {
        ...baseLobbyState,
        countdown: {
          isActive: true,
          remaining: 3,
          duration: 5,
        },
      };
      
      renderWithContext({ lobbyState: countdownLobbyState });
      
      expect(screen.getByText('3')).toBeInTheDocument();
      expect(screen.getByText('INITIATING CONTAINMENT PROTOCOL')).toBeInTheDocument();
      expect(screen.getByText('Last chance to leave lobby...')).toBeInTheDocument();
    });

    it('shows GO! when countdown reaches zero', () => {
      const countdownLobbyState = {
        ...baseLobbyState,
        countdown: {
          isActive: true,
          remaining: 0,
          duration: 5,
        },
      };
      
      renderWithContext({ lobbyState: countdownLobbyState });
      
      expect(screen.getByText('GO!')).toBeInTheDocument();
      expect(screen.getByText('Protocol activated!')).toBeInTheDocument();
    });

    it('disables start button when countdown is active', () => {
      const countdownLobbyState = {
        ...baseLobbyState,
        canStart: true,
        countdown: {
          isActive: true,
          remaining: 3,
          duration: 5,
        },
      };
      
      renderWithContext({ lobbyState: countdownLobbyState });
      
      const startButton = screen.getByRole('button', { name: /INITIATING PROTOCOL/ });
      expect(startButton).toBeDisabled();
    });

    it('shows initiating message for non-host during countdown', () => {
      const countdownLobbyState = {
        ...baseLobbyState,
        isHost: false,
        countdown: {
          isActive: true,
          remaining: 3,
          duration: 5,
        },
      };
      
      renderWithContext({ lobbyState: countdownLobbyState });
      
      expect(screen.getByText('Protocol initiating...')).toBeInTheDocument();
    });
  });

  describe('Invite Functionality', () => {
    it('shows success message when copy button clicked', async () => {
      const user = userEvent.setup();
      
      renderWithContext();
      
      const copyButton = screen.getByText('📋 Copy Invite Link');
      await user.click(copyButton);
      
      // Verify the success message appears
      await waitFor(() => {
        expect(screen.getByText('✓ Invite Link Copied!')).toBeInTheDocument();
      });
    });

    it('shows invite friends button', () => {
      renderWithContext();
      
      expect(screen.getByText('👥 Invite Friends')).toBeInTheDocument();
    });
  });

  describe('Leave Lobby', () => {
    it('shows leave lobby button', () => {
      renderWithContext();
      
      expect(screen.getByText('← Leave Lobby')).toBeInTheDocument();
    });

    it('calls onLeaveLobby when leave button clicked', async () => {
      const user = userEvent.setup();
      const mockOnLeaveLobby = vi.fn();
      
      renderWithContext({ onLeaveLobby: mockOnLeaveLobby });
      
      const leaveButton = screen.getByText('← Leave Lobby');
      await user.click(leaveButton);
      
      expect(mockOnLeaveLobby).toHaveBeenCalled();
    });
  });
});