import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { NightActionSelection } from '../game/NightActionSelection'
import { GameProvider } from '../../contexts/GameContext'
import { Player, GameState } from '../../types'

// Mock the WebSocket context
vi.mock('../../contexts/WebSocketContext', () => ({
  useWebSocketContext: () => ({
    isConnected: true,
    sendAction: vi.fn(),
  }),
}))

// Mock the useGameActions hook
vi.mock('../../hooks/useGameActions', () => ({
  useGameActions: () => ({
    setMiningTarget: vi.fn(),
    handleMineTokens: vi.fn(),
    handleUseAbility: vi.fn(),
    handleProjectMilestones: vi.fn(),
    canPlayerAffordAbility: vi.fn(() => true),
    isValidNightActionTarget: vi.fn(() => true),
  }),
}))

// Helper function to create test players
const createTestPlayer = (id: string, name: string, role: string, milestones = 3, tokens = 5): Player => ({
  id,
  name,
  jobTitle: role,
  isAlive: true,
  tokens,
  projectMilestones: milestones,
  statusMessage: `"Working as ${role}"`,
  alignment: 'HUMAN',
  avatar: '👤',
  joinedAt: '2024-01-01T00:00:00Z',
  role: {
    type: role as any,
    name: role,
    description: `${role} role`,
    isUnlocked: milestones >= 3,
    ability: milestones >= 3 ? {
      name: `${role} Ability`,
      description: `${role} specific ability`,
      isReady: true,
    } : undefined,
  },
  personalKPI: {
    type: 'PRODUCTIVITY',
    description: 'Complete objectives',
    progress: milestones,
    target: 3,
    isCompleted: milestones >= 3,
    reward: 'Role unlock',
  },
})

const createTestGameState = (): GameState => ({
  id: 'test-game',
  players: [
    createTestPlayer('player-1', 'Alice', 'CISO', 3, 8),
    createTestPlayer('player-2', 'Bob', 'CEO', 2, 5),
    createTestPlayer('player-3', 'Charlie', 'CTO', 3, 6),
  ],
  phase: 'NIGHT' as any,
  dayNumber: 2,
  chatMessages: [],
})

const renderWithGameContext = (gameState: GameState, localPlayerId: string) => {
  return render(
    <GameProvider gameState={gameState} localPlayerId={localPlayerId}>
      <NightActionSelection />
    </GameProvider>
  )
}

describe('NightActionSelection', () => {
  describe('Basic Rendering', () => {
    it('renders without crashing', () => {
      const gameState = createTestGameState()
      renderWithGameContext(gameState, 'player-1')
      
      // Just check that the component renders
      expect(screen.getByText('Choose Night Action')).toBeInTheDocument()
    })

    it('shows main action options', () => {
      const gameState = createTestGameState()
      renderWithGameContext(gameState, 'player-1')

      expect(screen.getByText('Mine for Player')).toBeInTheDocument()
      expect(screen.getAllByText('Project Milestones').length).toBeGreaterThan(0)
    })

    it('shows role ability when unlocked', () => {
      const gameState = createTestGameState()
      renderWithGameContext(gameState, 'player-1') // Alice with CISO role, 3 milestones

      expect(screen.getByText('CISO')).toBeInTheDocument()
    })

    it('shows locked state for insufficient milestones', () => {
      const gameState = createTestGameState()
      renderWithGameContext(gameState, 'player-2') // Bob with CEO role, 2 milestones

      const abilitySection = screen.getByText('CEO').closest('div')
      expect(abilitySection).toHaveTextContent('Locked')
    })
  })

  describe('Role Abilities', () => {
    it('displays different role types correctly', () => {
      const gameState = createTestGameState()
      
      // Test CISO
      renderWithGameContext(gameState, 'player-1')
      expect(screen.getByText('CISO')).toBeInTheDocument()
    })

    it('shows ready state for unlocked abilities', () => {
      const gameState = createTestGameState()
      renderWithGameContext(gameState, 'player-1') // Alice with unlocked CISO ability

      const abilitySection = screen.getByText('CISO').closest('div')
      expect(abilitySection).toHaveTextContent('Ready')
    })
  })

  describe('Player Interaction', () => {
    it('handles players without valid local player gracefully', () => {
      const gameState = createTestGameState()
      renderWithGameContext(gameState, 'nonexistent-player')
      
      // Component should handle this gracefully and not crash
      expect(document.body).toBeInTheDocument()
    })
  })
})