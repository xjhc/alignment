import { describe, it, expect } from 'vitest'
import { convertToClientTypes, convertToCoreTypes, CoreGameState } from '../coreTypes'

describe('coreTypes utility functions', () => {
  describe('convertToClientTypes', () => {
    it('should convert core game state to client format', () => {
      const mockCoreState: CoreGameState = {
        id: 'game-123',
        players: {
          'player-1': {
            id: 'player-1',
            name: 'Alice',
            jobTitle: 'Engineer',
            isAlive: true,
            tokens: 2,
            projectMilestones: 0,
            statusMessage: '',
            joinedAt: '2025-01-01T00:00:00Z'
          },
          'player-2': {
            id: 'player-2',
            name: 'Bob',
            jobTitle: 'Manager',
            isAlive: true,
            tokens: 1,
            projectMilestones: 1,
            statusMessage: '',
            joinedAt: '2025-01-01T00:01:00Z'
          }
        },
        phase: {
          type: 'discussion',
          startTime: '2025-01-01T00:00:00Z',
          duration: 120000
        },
        dayNumber: 1,
        chatMessages: [
          {
            id: 'msg-1',
            playerID: 'player-1',
            playerName: 'Alice',
            message: 'Hello everyone!',
            timestamp: '2025-01-01T00:00:00Z',
            isSystem: false
          }
        ],
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
        settings: {
          maxPlayers: 8,
          minPlayers: 2,
          sitrepDuration: 15000,
          pulseCheckDuration: 30000,
          discussionDuration: 120000,
          extensionDuration: 15000,
          nominationDuration: 30000,
          trialDuration: 30000,
          verdictDuration: 30000,
          nightDuration: 30000,
          startingTokens: 1,
          votingThreshold: 0.5
        }
      }

      const result = convertToClientTypes(mockCoreState)

      expect(result.id).toBe('game-123')
      expect(result.players).toHaveLength(2)
      expect(result.players[0].name).toBe('Alice')
      expect(result.players[1].name).toBe('Bob')
      expect(result.phase.type).toBe('discussion')
      expect(result.chatMessages).toHaveLength(1)
      expect(result.dayNumber).toBe(1)
    })

    it('should handle empty players object', () => {
      const mockCoreState: CoreGameState = {
        id: 'game-empty',
        players: {},
        phase: { type: 'lobby', startTime: '2025-01-01T00:00:00Z', duration: 0 },
        dayNumber: 0,
        chatMessages: [],
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
        settings: {
          maxPlayers: 8,
          minPlayers: 2,
          sitrepDuration: 15000,
          pulseCheckDuration: 30000,
          discussionDuration: 120000,
          extensionDuration: 15000,
          nominationDuration: 30000,
          trialDuration: 30000,
          verdictDuration: 30000,
          nightDuration: 30000,
          startingTokens: 1,
          votingThreshold: 0.5
        }
      }

      const result = convertToClientTypes(mockCoreState)

      expect(result.players).toHaveLength(0)
      expect(result.chatMessages).toEqual([])
    })

    it('should handle null/undefined arrays with defaults', () => {
      const mockCoreState = {
        id: 'game-null',
        players: {},
        phase: { type: 'lobby', startTime: '2025-01-01T00:00:00Z', duration: 0 },
        dayNumber: 0,
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
        settings: {
          maxPlayers: 8,
          minPlayers: 2,
          sitrepDuration: 15000,
          pulseCheckDuration: 30000,
          discussionDuration: 120000,
          extensionDuration: 15000,
          nominationDuration: 30000,
          trialDuration: 30000,
          verdictDuration: 30000,
          nightDuration: 30000,
          startingTokens: 1,
          votingThreshold: 0.5
        }
      } as CoreGameState

      const result = convertToClientTypes(mockCoreState)

      expect(result.chatMessages).toEqual([])
      expect(result.nightActionResults).toEqual([])
      expect(result.privateNotifications).toEqual([])
    })
  })

  describe('convertToCoreTypes', () => {
    it('should convert client state to core format', () => {
      const mockClientState = {
        id: 'game-456',
        players: [
          {
            id: 'player-1',
            name: 'Charlie',
            jobTitle: 'Designer',
            isAlive: true,
            tokens: 3,
            projectMilestones: 2,
            statusMessage: 'Working',
            joinedAt: '2025-01-01T00:00:00Z'
          }
        ],
        phase: {
          type: 'night',
          startTime: '2025-01-01T00:00:00Z',
          duration: 30000
        },
        dayNumber: 2,
        chatMessages: []
      }

      const result = convertToCoreTypes(mockClientState)

      expect(result.id).toBe('game-456')
      expect(result.players['player-1'].name).toBe('Charlie')
      expect(result.players['player-1'].tokens).toBe(3)
      expect(result.phase.type).toBe('night')
      expect(result.dayNumber).toBe(2)
      expect(result.settings.maxPlayers).toBe(8)
      expect(result.createdAt).toBeDefined()
      expect(result.updatedAt).toBeDefined()
    })

    it('should handle empty players array', () => {
      const mockClientState = {
        id: 'game-empty-client',
        players: [],
        phase: { type: 'lobby', startTime: '2025-01-01T00:00:00Z', duration: 0 },
        dayNumber: 0,
        chatMessages: []
      }

      const result = convertToCoreTypes(mockClientState)

      expect(result.players).toEqual({})
      expect(Object.keys(result.players)).toHaveLength(0)
    })

    it('should provide default settings when missing', () => {
      const mockClientState = {
        id: 'game-no-settings',
        players: [],
        phase: { type: 'lobby', startTime: '2025-01-01T00:00:00Z', duration: 0 },
        dayNumber: 0,
        chatMessages: []
      }

      const result = convertToCoreTypes(mockClientState)

      expect(result.settings).toBeDefined()
      expect(result.settings.maxPlayers).toBe(8)
      expect(result.settings.minPlayers).toBe(2)
      expect(result.settings.startingTokens).toBe(1)
      expect(result.settings.votingThreshold).toBe(0.5)
    })

    it('should handle non-array players gracefully', () => {
      const mockClientState = {
        id: 'game-bad-players',
        players: null,
        phase: { type: 'lobby', startTime: '2025-01-01T00:00:00Z', duration: 0 },
        dayNumber: 0,
        chatMessages: []
      }

      const result = convertToCoreTypes(mockClientState)

      expect(result.players).toEqual({})
    })
  })
})