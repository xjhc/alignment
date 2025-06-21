interface PlayerLobbyInfo {
  id: string;
  name: string;
  avatar: string;
}

interface LobbyState {
  playerInfos: PlayerLobbyInfo[];
  isHost: boolean;
  canStart: boolean;
  hostId: string;
  lobbyName: string;
  maxPlayers: number;
  connectionError: string | null;
}

interface WaitingScreenProps {
  gameId: string;
  playerId?: string;
  lobbyState: LobbyState;
  isConnected: boolean;
  onStartGame: () => void;
  onLeaveLobby: () => void;
}

export function WaitingScreen({
  gameId,
  playerId,
  lobbyState,
  isConnected,
  onStartGame,
  onLeaveLobby
}: WaitingScreenProps) {
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
      <div className="launch-screen">
        <h1 className="logo">
          LOEBIAN INC. // <span className="glitch">EMERGENCY BRIDGE</span>
        </h1>

        <div className="launch-form">
          <h2>CONNECTION ERROR</h2>
          <p style={{ color: 'var(--error)', margin: '1rem 0' }}>
            {connectionError}
          </p>
          <button onClick={onLeaveLobby} className="btn-secondary">
            ← Go Back
          </button>
        </div>
      </div>
    );
  }

  // Show loading while connecting
  if (!isConnected) {
    return (
      <div className="launch-screen">
        <h1 className="logo">
          LOEBIAN INC. // <span className="glitch">EMERGENCY BRIDGE</span>
        </h1>

        <div className="launch-form">
          <h2>CONNECTING TO LOBBY...</h2>
          <p>Establishing secure connection...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="launch-screen">
      <h1 className="logo">
        LOEBIAN INC. // <span className="glitch">EMERGENCY BRIDGE</span>
      </h1>

      <div className="launch-form waiting-screen">
        <h2>WAITING IN LOBBY...</h2>
        <p className="game-id-info">
          Lobby: <strong>{lobbyName || 'Loading...'}</strong><br />
          Game ID: <code>{formatGameId(gameId)}</code><br />
          Share this ID with other personnel.
        </p>

        <div className="player-roster">
          <div className="list-header">
            Personnel Connected - {playerInfos.length} / {maxPlayers}
          </div>

          {playerInfos.map((playerInfo) => (
            <div key={playerInfo.id} className="player-card">
              <div className="player-avatar">
                {playerInfo.avatar || '👤'}
                {playerInfo.id === hostId && (
                  <div className="host-crown">👑</div>
                )}
              </div>
              <div className="player-content">
                <div className="player-main-info">
                  <span className="player-name">
                    {playerInfo.name}
                    {playerInfo.id === hostId && ' (Host)'}
                    {playerInfo.id === playerId && ' (You)'}
                  </span>
                  <span className="player-job">
                    Personnel
                  </span>
                </div>
                <div className="player-tokens">
                  🪙0
                </div>
              </div>
            </div>
          ))}

          {/* Show empty slots */}
          {Array.from({ length: Math.max(0, maxPlayers - playerInfos.length) }).map((_, index) => (
            <div key={`empty-${index}`} className="player-card empty">
              <div className="player-avatar">⏳</div>
              <div className="player-content">
                <div className="player-main-info">
                  <span className="player-name">Waiting for player...</span>
                  <span className="player-job">-</span>
                </div>
                <div className="player-tokens">🪙-</div>
              </div>
            </div>
          ))}
        </div>

        {isHost && (
          <button
            className="btn-primary"
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
          <p className="waiting-message">
            Waiting for host to start the game...
          </p>
        )}

        <button onClick={onLeaveLobby} className="back-button">
          ← Leave Lobby
        </button>
      </div>
    </div>
  );
}