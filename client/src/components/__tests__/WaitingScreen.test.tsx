import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WaitingScreen } from '../WaitingScreen'

const mockLobbyState = {
  playerInfos: [
    { id: 'player1', name: 'Alice', avatar: '👤' },
    { id: 'player2', name: 'Bob', avatar: '🧑‍💻' }
  ],
  isHost: true,
  canStart: false,
  hostId: 'player1',
  lobbyName: 'Test Lobby',
  maxPlayers: 8,
  connectionError: null
}

describe('WaitingScreen', () => {
  it('renders lobby information correctly', () => {
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    expect(screen.getByText(/WAITING IN LOBBY/)).toBeInTheDocument()
    expect(screen.getByText('Test Lobby')).toBeInTheDocument()
    expect(screen.getByText(/test-g/)).toBeInTheDocument() // Shortened game ID
    expect(screen.getByText(/Personnel Connected - 2 \/ 8/)).toBeInTheDocument()
  })

  it('displays players with correct information', () => {
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    expect(screen.getByText('Alice (Host) (You)')).toBeInTheDocument()
    expect(screen.getByText('Bob')).toBeInTheDocument()
    expect(screen.getByText('👤')).toBeInTheDocument()
    expect(screen.getByText('🧑‍💻')).toBeInTheDocument()
  })

  it('shows host controls when user is host', () => {
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    const startButton = screen.getByRole('button', { name: /NEED.*MORE PLAYERS/ })
    expect(startButton).toBeInTheDocument()
    expect(startButton).toBeDisabled()
  })

  it('shows waiting message when user is not host', () => {
    const nonHostLobbyState = { ...mockLobbyState, isHost: false }
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player2"
        lobbyState={nonHostLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    expect(screen.getByText(/Waiting for host to start the game/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /INITIATE CONTAINMENT/ })).not.toBeInTheDocument()
  })

  it('enables start button when game can start', () => {
    const canStartLobbyState = { ...mockLobbyState, canStart: true }
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={canStartLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    const startButton = screen.getByRole('button', { name: /INITIATE CONTAINMENT PROTOCOL/ })
    expect(startButton).not.toBeDisabled()
  })

  it('calls onStartGame when start button is clicked', async () => {
    const user = userEvent.setup()
    const canStartLobbyState = { ...mockLobbyState, canStart: true }
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={canStartLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    const startButton = screen.getByRole('button', { name: /INITIATE CONTAINMENT PROTOCOL/ })
    await user.click(startButton)

    expect(mockOnStartGame).toHaveBeenCalledOnce()
  })

  it('calls onLeaveLobby when leave button is clicked', async () => {
    const user = userEvent.setup()
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    const leaveButton = screen.getByRole('button', { name: /Leave Lobby/ })
    await user.click(leaveButton)

    expect(mockOnLeaveLobby).toHaveBeenCalledOnce()
  })

  it('shows empty slots for remaining players', () => {
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    const waitingSlots = screen.getAllByText('Waiting for player...')
    expect(waitingSlots).toHaveLength(6) // 8 max - 2 current = 6 empty slots
  })

  it('shows loading screen when not connected', () => {
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={false}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    expect(screen.getByText(/CONNECTING TO LOBBY/)).toBeInTheDocument()
    expect(screen.getByText(/Establishing secure connection/)).toBeInTheDocument()
  })

  it('shows connection error when present', () => {
    const errorLobbyState = { ...mockLobbyState, connectionError: 'Failed to connect to server' }
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={errorLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    expect(screen.getByText(/CONNECTION ERROR/)).toBeInTheDocument()  
    expect(screen.getByText('Failed to connect to server')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Go Back/ })).toBeInTheDocument()
  })

  it('formats game ID correctly', () => {
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="very-long-game-id-123456789"
        playerId="player1"
        lobbyState={mockLobbyState}
        isConnected={true}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    expect(screen.getByText('very-l')).toBeInTheDocument() // First 6 characters
  })

  it('disables start button when disconnected', () => {
    const canStartLobbyState = { ...mockLobbyState, canStart: true }
    const mockOnStartGame = vi.fn()
    const mockOnLeaveLobby = vi.fn()

    render(
      <WaitingScreen
        gameId="test-game-id-123456"
        playerId="player1"
        lobbyState={canStartLobbyState}
        isConnected={false}
        onStartGame={mockOnStartGame}
        onLeaveLobby={mockOnLeaveLobby}
      />
    )

    // Should show loading screen when disconnected, so start button shouldn't be visible
    expect(screen.queryByRole('button', { name: /INITIATE CONTAINMENT/ })).not.toBeInTheDocument()
  })
})