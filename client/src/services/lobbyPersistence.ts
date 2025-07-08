/**
 * Lobby Persistence Service
 * Handles saving and restoring lobby session data across page refreshes
 */

export interface LobbySessionData {
  gameId: string;
  playerId: string;
  sessionToken: string;
  playerName: string;
  lobbyName: string;
  isHost: boolean;
  connectedAt: string;
  lastActivity: string;
  readyStatus: boolean;
  sessionState: 'LOBBY' | 'IN_GAME' | 'ENDED';
}

export interface LobbyPreferences {
  showRulesOnJoin: boolean;
  showRolePreviewOnJoin: boolean;
  autoExpandChat: boolean;
  soundEnabled: boolean;
  showConnectionQuality: boolean;
}

const STORAGE_KEYS = {
  SESSION: 'alignmentLobbySession',
  PREFERENCES: 'alignmentLobbyPreferences',
  BACKUP_SESSION: 'alignmentLobbyBackup'
} as const;

export class LobbyPersistenceService {
  private static instance: LobbyPersistenceService;
  private sessionCheckInterval: number | null = null;

  private constructor() {
    // Start session validity monitoring
    this.startSessionMonitoring();
  }

  public static getInstance(): LobbyPersistenceService {
    if (!LobbyPersistenceService.instance) {
      LobbyPersistenceService.instance = new LobbyPersistenceService();
    }
    return LobbyPersistenceService.instance;
  }

  /**
   * Save lobby session data
   */
  saveLobbySession(data: LobbySessionData): void {
    try {
      const sessionData = {
        ...data,
        lastActivity: new Date().toISOString()
      };

      // Save to both primary and backup storage
      localStorage.setItem(STORAGE_KEYS.SESSION, JSON.stringify(sessionData));
      sessionStorage.setItem(STORAGE_KEYS.BACKUP_SESSION, JSON.stringify(sessionData));

      console.log('Lobby session saved:', sessionData.gameId);
    } catch (error) {
      console.error('Failed to save lobby session:', error);
    }
  }

  /**
   * Load lobby session data
   */
  loadLobbySession(): LobbySessionData | null {
    try {
      // Try primary storage first
      let sessionData = localStorage.getItem(STORAGE_KEYS.SESSION);
      
      // Fallback to backup storage
      if (!sessionData) {
        sessionData = sessionStorage.getItem(STORAGE_KEYS.BACKUP_SESSION);
      }

      if (!sessionData) {
        return null;
      }

      const parsed = JSON.parse(sessionData) as LobbySessionData;
      
      // Validate session data
      if (!this.isValidSessionData(parsed)) {
        console.warn('Invalid session data found, clearing...');
        this.clearLobbySession();
        return null;
      }

      // Check if session is expired (24 hours)
      if (this.isSessionExpired(parsed)) {
        console.log('Session expired, clearing...');
        this.clearLobbySession();
        return null;
      }

      console.log('Lobby session loaded:', parsed.gameId);
      return parsed;
    } catch (error) {
      console.error('Failed to load lobby session:', error);
      return null;
    }
  }

  /**
   * Update session activity timestamp
   */
  updateSessionActivity(): void {
    const session = this.loadLobbySession();
    if (session) {
      this.saveLobbySession(session);
    }
  }

  /**
   * Update session state
   */
  updateSessionState(state: LobbySessionData['sessionState']): void {
    const session = this.loadLobbySession();
    if (session) {
      session.sessionState = state;
      this.saveLobbySession(session);
    }
  }

  /**
   * Update ready status
   */
  updateReadyStatus(isReady: boolean): void {
    const session = this.loadLobbySession();
    if (session) {
      session.readyStatus = isReady;
      this.saveLobbySession(session);
    }
  }

  /**
   * Clear lobby session data
   */
  clearLobbySession(): void {
    try {
      localStorage.removeItem(STORAGE_KEYS.SESSION);
      sessionStorage.removeItem(STORAGE_KEYS.BACKUP_SESSION);
      console.log('Lobby session cleared');
    } catch (error) {
      console.error('Failed to clear lobby session:', error);
    }
  }

  /**
   * Save lobby preferences
   */
  saveLobbyPreferences(preferences: LobbyPreferences): void {
    try {
      localStorage.setItem(STORAGE_KEYS.PREFERENCES, JSON.stringify(preferences));
      console.log('Lobby preferences saved');
    } catch (error) {
      console.error('Failed to save lobby preferences:', error);
    }
  }

  /**
   * Load lobby preferences
   */
  loadLobbyPreferences(): LobbyPreferences {
    try {
      const data = localStorage.getItem(STORAGE_KEYS.PREFERENCES);
      if (data) {
        return JSON.parse(data) as LobbyPreferences;
      }
    } catch (error) {
      console.error('Failed to load lobby preferences:', error);
    }

    // Return default preferences
    return {
      showRulesOnJoin: true,
      showRolePreviewOnJoin: true,
      autoExpandChat: false,
      soundEnabled: true,
      showConnectionQuality: true
    };
  }

  /**
   * Check if session data is valid
   */
  private isValidSessionData(data: any): data is LobbySessionData {
    return data &&
      typeof data.gameId === 'string' &&
      typeof data.playerId === 'string' &&
      typeof data.sessionToken === 'string' &&
      typeof data.playerName === 'string' &&
      typeof data.connectedAt === 'string' &&
      typeof data.lastActivity === 'string' &&
      ['LOBBY', 'IN_GAME', 'ENDED'].includes(data.sessionState);
  }

  /**
   * Check if session is expired
   */
  private isSessionExpired(data: LobbySessionData): boolean {
    try {
      const lastActivity = new Date(data.lastActivity);
      const now = new Date();
      const timeDiff = now.getTime() - lastActivity.getTime();
      
      // Sessions expire after 24 hours of inactivity
      return timeDiff > 24 * 60 * 60 * 1000;
    } catch (error) {
      console.error('Failed to check session expiry:', error);
      return true;
    }
  }

  /**
   * Start monitoring session validity
   */
  private startSessionMonitoring(): void {
    // Update activity every 5 minutes
    this.sessionCheckInterval = window.setInterval(() => {
      this.updateSessionActivity();
    }, 5 * 60 * 1000);
  }

  /**
   * Stop session monitoring
   */
  stopSessionMonitoring(): void {
    if (this.sessionCheckInterval) {
      clearInterval(this.sessionCheckInterval);
      this.sessionCheckInterval = null;
    }
  }

  /**
   * Get session age in milliseconds
   */
  getSessionAge(): number | null {
    const session = this.loadLobbySession();
    if (!session) return null;

    try {
      const connectedAt = new Date(session.connectedAt);
      return Date.now() - connectedAt.getTime();
    } catch (error) {
      console.error('Failed to calculate session age:', error);
      return null;
    }
  }

  /**
   * Check if user has an active session
   */
  hasActiveSession(): boolean {
    const session = this.loadLobbySession();
    return session !== null && session.sessionState !== 'ENDED';
  }

  /**
   * Export session data for debugging
   */
  exportSessionData(): {
    session: LobbySessionData | null;
    preferences: LobbyPreferences;
    hasBackup: boolean;
  } {
    return {
      session: this.loadLobbySession(),
      preferences: this.loadLobbyPreferences(),
      hasBackup: sessionStorage.getItem(STORAGE_KEYS.BACKUP_SESSION) !== null
    };
  }
}

// Export singleton instance
export const lobbyPersistence = LobbyPersistenceService.getInstance();