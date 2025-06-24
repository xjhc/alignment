import styles from './WaitingScreen.module.css';
import { useSessionContext } from '../contexts/SessionContext';

export function WaitingScreen() {
  const { 
    appState, 
    lobbyState, 
    isConnected, 
    onStartGame, 
    onLeaveLobby 
  } = useSessionContext();
  // All state is now managed by App.tsx - this is a pure presentation component
  const {
    playerInfos,
    isHost,
    canStart,
    hostId,
    lobbyName,
    maxPlayers,
    connectionError
  } = lobbyState;

  const formatGameId = (id: string) => {
    return id.substring(0, 6); // Show first 6 characters for readability
  };

  // Show connection error if unable to connect
  if (connectionError) {
    return (
      <div className={styles.launchScreen}>
        <h1 className={styles.logo}>
          LOEBIAN INC. // <span className={styles.glitch}>EMERGENCY BRIDGE</span>
        </h1>

        <div className={styles.launchForm}>
          <h2>CONNECTION ERROR</h2>
          <p style={{ color: 'var(--accent-red)', margin: '1rem 0' }}>
            {connectionError}
          </p>
          <button onClick={onLeaveLobby} className={styles.btnSecondary}>
            ← Go Back
          </button>
        </div>
      </div>
    );
  }

  // Show loading while connecting
  if (!isConnected) {
    return (
      <div className={styles.launchScreen}>
        <h1 className={styles.logo}>
          LOEBIAN INC. // <span className={styles.glitch}>EMERGENCY BRIDGE</span>
        </h1>

        <div className={styles.launchForm}>
          <h2>CONNECTING TO LOBBY...</h2>
          <p>Establishing secure connection...</p>
          <div className="loading-spinner large" style={{ marginTop: '16px' }}></div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.launchScreen}>
      <h1 className={styles.logo}>
        LOEBIAN INC. // <span className={styles.glitch}>EMERGENCY BRIDGE</span>
      </h1>

      <div className={`${styles.launchForm} ${styles.waitingScreen}`}>
        <h2>WAITING IN LOBBY...</h2>
        <p className={styles.gameIdInfo}>
          Lobby: <strong>{lobbyName || 'Loading...'}</strong><br />
          Game ID: <code>{formatGameId(appState.gameId || 'unknown')}</code><br />
          Share this ID with other personnel.
        </p>

        <div className={styles.playerRoster}>
          <div className={styles.listHeader}>
            Personnel Connected - {playerInfos.length} / {maxPlayers}
          </div>

          {playerInfos.map((playerInfo, index) => (
            <div 
              key={playerInfo.id} 
              className={`${styles.playerCard} animate-slide-in-left`}
              style={{ animationDelay: `${index * 100}ms` }}
            >
              <div className={styles.playerAvatar}>
                {playerInfo.avatar || '👤'}
                {playerInfo.id === hostId && (
                  <div className={styles.hostCrown}>👑</div>
                )}
              </div>
              <div className={styles.playerContent}>
                <div className={styles.playerMainInfo}>
                  <span className={styles.playerName}>
                    {playerInfo.name}
                    {playerInfo.id === hostId && ' (Host)'}
                    {playerInfo.id === appState.playerId && ' (You)'}
                  </span>
                  <span className={styles.playerJob}>
                    Personnel
                  </span>
                </div>
                <div className={styles.playerTokens}>
                  🪙0
                </div>
              </div>
            </div>
          ))}

          {/* Show empty slots */}
          {Array.from({ length: Math.max(0, maxPlayers - playerInfos.length) }).map((_, index) => (
            <div 
              key={`empty-${index}`} 
              className={`${styles.playerCard} ${styles.empty} animate-pulse`}
            >
              <div className={styles.playerAvatar}>⏳</div>
              <div className={styles.playerContent}>
                <div className={styles.playerMainInfo}>
                  <span className={styles.playerName}>Waiting for player...</span>
                  <span className={styles.playerJob}>-</span>
                </div>
                <div className={styles.playerTokens}>🪙-</div>
              </div>
            </div>
          ))}
        </div>

        {isHost && (
          <button
            className={styles.btnPrimary}
            onClick={onStartGame}
            disabled={!canStart || !isConnected}
            style={{ marginTop: '24px' }}
          >
            {canStart
              ? '[ > INITIATE CONTAINMENT PROTOCOL ]'
              : `[ NEED ${Math.max(0, 4 - playerInfos.length)} MORE PLAYERS ]`
            }
          </button>
        )}

        {!isHost && (
          <p className={styles.waitingMessage}>
            Waiting for host to start the game...
          </p>
        )}

        <button onClick={onLeaveLobby} className={styles.backButton}>
          ← Leave Lobby
        </button>
      </div>
    </div>
  );
}