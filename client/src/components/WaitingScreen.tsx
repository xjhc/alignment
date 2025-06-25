import { useSessionContext } from '../contexts/SessionContext';
import { Button } from './base';

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
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
          LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
        </h1>

        <div className="flex flex-col gap-4 items-center w-80">
          <h2>CONNECTION ERROR</h2>
          <p className="text-red my-4">
            {connectionError}
          </p>
          <Button
            onClick={onLeaveLobby}
            variant="secondary"
            className="text-sm font-medium"
          >
            ← Go Back
          </Button>
        </div>
      </div>
    );
  }

  // Show loading while connecting
  if (!isConnected) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
          LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
        </h1>

        <div className="flex flex-col gap-4 items-center w-80">
          <h2>CONNECTING TO LOBBY...</h2>
          <p>Establishing secure connection...</p>
          <div className="loading-spinner large mt-4"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
      <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
        LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
      </h1>

      <div className="flex flex-col gap-4 items-center w-96 text-center">
        <h2>WAITING IN LOBBY...</h2>
        <p className="text-text-secondary mb-6">
          Lobby: <strong>{lobbyName || 'Loading...'}</strong><br />
          Game ID: <code className="bg-background-secondary px-1.5 py-0.5 rounded font-mono">{formatGameId(appState.gameId || 'unknown')}</code><br />
          Share this ID with other personnel.
        </p>

        <div className="w-full text-left">
          <div className="text-xs font-bold text-text-muted uppercase mb-2 text-center tracking-[0.5px]">
            Personnel Connected - {playerInfos.length} / {maxPlayers}
          </div>

          {playerInfos.map((playerInfo, index) => (
            <div 
              key={playerInfo.id} 
              className="flex items-start gap-2 p-1.5 px-2 rounded-md cursor-pointer mb-0.5 hover:bg-background-tertiary animation-slide-in-left"
              style={{ animationDelay: `${index * 100}ms` }}
            >
              <div className="w-7 h-7 rounded-full bg-background-tertiary flex items-center justify-center text-sm flex-shrink-0 border border-border relative">
                {playerInfo.avatar || '👤'}
                {playerInfo.id === hostId && (
                  <div className="absolute -top-1.5 -right-1.5 text-xs bg-amber rounded-full w-4 h-4 flex items-center justify-center border border-background-primary">👑</div>
                )}
              </div>
              <div className="flex-1 flex items-center justify-between">
                <div className="flex flex-col gap-0.5">
                  <span className="font-semibold text-text-primary text-sm">
                    {playerInfo.name}
                    {playerInfo.id === hostId && ' (Host)'}
                    {playerInfo.id === appState.playerId && ' (You)'}
                  </span>
                  <span className="text-xs text-text-secondary uppercase font-medium">
                    Personnel
                  </span>
                </div>
                <div className="text-amber font-semibold text-sm">
                  🪙0
                </div>
              </div>
            </div>
          ))}

          {/* Show empty slots */}
          {Array.from({ length: Math.max(0, maxPlayers - playerInfos.length) }).map((_, index) => (
            <div 
              key={`empty-${index}`} 
              className="flex items-start gap-2 p-1.5 px-2 rounded-md mb-0.5 opacity-50 animate-pulse"
            >
              <div className="w-7 h-7 rounded-full bg-background-tertiary flex items-center justify-center text-sm flex-shrink-0 border border-border">⏳</div>
              <div className="flex-1 flex items-center justify-between">
                <div className="flex flex-col gap-0.5">
                  <span className="font-semibold text-text-primary text-sm">Waiting for player...</span>
                  <span className="text-xs text-text-secondary uppercase font-medium">-</span>
                </div>
                <div className="text-amber font-semibold text-sm">🪙-</div>
              </div>
            </div>
          ))}
        </div>

        {isHost && (
          <Button
            variant="primary"
            size="lg"
            fullWidth
            onClick={onStartGame}
            disabled={!canStart || !isConnected}
            className="text-base font-semibold text-black bg-amber hover:enabled:bg-amber-light mt-6"
          >
            {canStart
              ? '[ > INITIATE CONTAINMENT PROTOCOL ]'
              : `[ NEED ${Math.max(0, 4 - playerInfos.length)} MORE PLAYERS ]`
            }
          </Button>
        )}

        {!isHost && (
          <p className="text-text-secondary italic mt-6">
            Waiting for host to start the game...
          </p>
        )}

        <Button
          onClick={onLeaveLobby}
          variant="ghost"
          size="sm"
          className="self-start mt-4 text-text-muted hover:enabled:text-text-primary"
        >
          ← Leave Lobby
        </Button>
      </div>
    </div>
  );
}