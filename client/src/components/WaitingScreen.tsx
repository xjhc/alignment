import { useSessionContext } from '../contexts/SessionContext';
import { Button } from './ui';
import { useState } from 'react';
import { InviteFriendsModal } from './InviteFriendsModal';
import { PlayerProfile } from './PlayerProfile';

export function WaitingScreen() {
  const { 
    appState, 
    lobbyState, 
    isConnected, 
    onStartGame, 
    onLeaveLobby,
    onBackToLogin 
  } = useSessionContext();
  const [inviteCopied, setInviteCopied] = useState(false);
  const [showInviteFriends, setShowInviteFriends] = useState(false);
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedPlayerId, setSelectedPlayerId] = useState<string | null>(null);
  // All state is now managed by App.tsx - this is a pure presentation component
  const {
    playerInfos,
    isHost,
    canStart,
    hostId,
    lobbyName,
    maxPlayers,
    connectionError,
    countdown
  } = lobbyState;

  const formatGameId = (id: string) => {
    return id.substring(0, 6); // Show first 6 characters for readability
  };

  const copyInviteLink = async () => {
    if (!appState.gameId) return;
    
    const inviteUrl = `${window.location.origin}/join/${appState.gameId}`;
    
    try {
      await navigator.clipboard.writeText(inviteUrl);
      setInviteCopied(true);
      setTimeout(() => setInviteCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy invite link:', err);
      // Fallback for browsers that don't support clipboard API
      const textArea = document.createElement('textarea');
      textArea.value = inviteUrl;
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand('copy');
        setInviteCopied(true);
        setTimeout(() => setInviteCopied(false), 2000);
      } catch (fallbackErr) {
        console.error('Fallback copy failed:', fallbackErr);
      }
      document.body.removeChild(textArea);
    }
  };

  const togglePrivacy = async () => {
    if (!appState.gameId || !isHost) return;
    
    try {
      const response = await fetch(`/api/games/${appState.gameId}/privacy`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          is_private: !isPrivate,
        }),
      });

      if (response.ok) {
        setIsPrivate(!isPrivate);
      } else {
        console.error('Failed to update lobby privacy');
      }
    } catch (err) {
      console.error('Error updating lobby privacy:', err);
    }
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
          <p className="text-red my-4 text-center">
            {connectionError}
          </p>
          <div className="flex gap-3">
            <Button
              onClick={onLeaveLobby}
              variant="secondary"
              className="text-sm font-medium"
            >
              ← Go Back
            </Button>
            <Button
              onClick={onBackToLogin}
              variant="ghost"
              className="text-sm font-medium text-text-muted hover:enabled:text-text-primary"
            >
              Logout
            </Button>
          </div>
          <p className="text-xs text-text-muted mt-2 text-center">
            If the lobby has ended or expired, try logging out and starting fresh
          </p>
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
          
          <div className="flex gap-3 mt-6">
            <Button
              onClick={onLeaveLobby}
              variant="secondary"
              className="text-sm font-medium"
            >
              Cancel
            </Button>
            <Button
              onClick={onBackToLogin}
              variant="ghost"
              className="text-sm font-medium text-text-muted hover:enabled:text-text-primary"
            >
              Logout
            </Button>
          </div>
          
          <p className="text-xs text-text-muted mt-2 text-center">
            If connection takes too long, try canceling and rejoining
          </p>
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
        <div className="text-text-secondary mb-6">
          <p>
            Lobby: <strong>{lobbyName || 'Loading...'}</strong>
            {isPrivate && (
              <span className="ml-2 text-xs bg-amber text-black px-2 py-0.5 rounded">PRIVATE</span>
            )}
            <br />
            Game ID: <code className="bg-background-secondary px-1.5 py-0.5 rounded font-mono">{formatGameId(appState.gameId || 'unknown')}</code>
          </p>
          {isHost && (
            <div className="mt-2">
              <label className="flex items-center justify-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={isPrivate}
                  onChange={togglePrivacy}
                  className="rounded"
                />
                Private lobby (hidden from public list)
              </label>
            </div>
          )}
          <div className="mt-3 flex gap-2 justify-center">
            <Button
              onClick={copyInviteLink}
              variant="secondary"
              size="sm"
              className="text-xs"
            >
              {inviteCopied ? '✓ Invite Link Copied!' : '📋 Copy Invite Link'}
            </Button>
            <Button
              onClick={() => setShowInviteFriends(true)}
              variant="secondary"
              size="sm"
              className="text-xs"
            >
              👥 Invite Friends
            </Button>
          </div>
          <p className="text-xs mt-2 text-text-muted">
            Share the invite link for easy access
          </p>
        </div>

        <div className="w-full text-left">
          <div className="text-xs font-bold text-text-muted uppercase mb-2 text-center tracking-[0.5px]">
            Personnel Connected - {playerInfos.length} / {maxPlayers}
          </div>

          {[...playerInfos]
            .sort((a, b) => new Date(a.joinedAt).getTime() - new Date(b.joinedAt).getTime())
            .map((playerInfo, index) => (
              <div 
                key={playerInfo.id} 
                className="flex items-start gap-2 p-1.5 px-2 rounded-md mb-0.5 hover:bg-background-tertiary animation-slide-in-left group"
                style={{ animationDelay: `${index * 100}ms` }}
              >
                <div 
                  className="w-7 h-7 rounded-full bg-background-tertiary flex items-center justify-center text-sm flex-shrink-0 border border-border relative cursor-pointer hover:border-primary"
                  onClick={() => setSelectedPlayerId(playerInfo.id)}
                  title="View profile"
                >
                  {playerInfo.avatar || '👤'}
                  {playerInfo.id === hostId && (
                    <div className="absolute -top-1.5 -right-1.5 text-xs bg-amber rounded-full w-4 h-4 flex items-center justify-center border border-background-primary">👑</div>
                  )}
                </div>
                <div className="flex-1 flex items-center justify-between">
                  <div className="flex flex-col gap-0.5 flex-1">
                    <div className="flex items-center justify-between">
                      <span 
                        className="font-semibold text-text-primary text-sm cursor-pointer hover:text-primary"
                        onClick={() => setSelectedPlayerId(playerInfo.id)}
                      >
                        {playerInfo.name}
                        {playerInfo.id === hostId && ' (Host)'}
                        {playerInfo.id === appState.playerId && ' (You)'}
                      </span>
                      {playerInfo.id !== appState.playerId && (
                        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-xs px-2 py-1 h-auto"
                            onClick={async () => {
                              try {
                                const response = await fetch('/api/friends/requests', {
                                  method: 'POST',
                                  headers: {
                                    'Content-Type': 'application/json',
                                  },
                                  body: JSON.stringify({
                                    recipient_username: playerInfo.name,
                                  }),
                                });

                                if (response.ok) {
                                  alert('Friend request sent!');
                                } else {
                                  const errorData = await response.json();
                                  alert(errorData.error || 'Failed to send friend request');
                                }
                              } catch (err) {
                                alert('Error sending friend request');
                              }
                            }}
                            title="Add friend"
                          >
                            👥
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-xs px-2 py-1 h-auto"
                            onClick={async () => {
                              const reason = prompt('Why are you giving kudos to this player?');
                              if (reason) {
                                try {
                                  const response = await fetch('/api/users/me/kudos', {
                                    method: 'POST',
                                    headers: {
                                      'Content-Type': 'application/json',
                                    },
                                    body: JSON.stringify({
                                      target_player_id: playerInfo.id,
                                      reason: reason,
                                    }),
                                  });

                                  if (response.ok) {
                                    alert('Kudos sent!');
                                  } else {
                                    const errorData = await response.json();
                                    alert(errorData.error || 'Failed to give kudos');
                                  }
                                } catch (err) {
                                  alert('Error giving kudos');
                                }
                              }
                            }}
                            title="Give kudos"
                          >
                            👏
                          </Button>
                        </div>
                      )}
                    </div>
                    <span className="text-xs text-text-secondary uppercase font-medium">
                      Personnel
                    </span>
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
            disabled={!canStart || !isConnected || (countdown?.isActive)}
            className="text-base font-semibold text-black bg-amber hover:enabled:bg-amber-light mt-6"
          >
            {countdown?.isActive
              ? '[ INITIATING PROTOCOL... ]'
              : canStart
              ? '[ > INITIATE CONTAINMENT PROTOCOL ]'
              : `[ NEED ${Math.max(0, 4 - playerInfos.length)} MORE PLAYERS ]`
            }
          </Button>
        )}

        {!isHost && (
          <p className="text-text-secondary italic mt-6">
            {countdown?.isActive 
              ? 'Protocol initiating...'
              : 'Waiting for host to start the game...'
            }
          </p>
        )}

        <div className="flex gap-3 self-start mt-4">
          <Button
            onClick={onLeaveLobby}
            variant="ghost"
            size="sm"
            className="text-text-muted hover:enabled:text-text-primary"
          >
            ← Leave Lobby
          </Button>
          <Button
            onClick={onBackToLogin}
            variant="ghost"
            size="sm"
            className="text-text-muted hover:enabled:text-text-primary"
          >
            Logout
          </Button>
        </div>
      </div>

      {/* Countdown Modal */}
      {countdown?.isActive && (
        <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50">
          <div className="bg-background-secondary border border-border rounded-lg p-8 text-center animation-scale-in">
            <div className="text-amber text-6xl font-mono font-bold mb-4 animation-pulse">
              {countdown.remaining > 0 ? countdown.remaining : 'GO!'}
            </div>
            <div className="text-text-primary text-lg mb-2">
              INITIATING CONTAINMENT PROTOCOL
            </div>
            <div className="text-text-secondary text-sm">
              {countdown.remaining > 0 
                ? 'Last chance to leave lobby...'
                : 'Protocol activated!'
              }
            </div>
          </div>
        </div>
      )}

      {/* Invite Friends Modal */}
      <InviteFriendsModal
        isVisible={showInviteFriends}
        onClose={() => setShowInviteFriends(false)}
        lobbyId={appState.gameId}
      />

      {/* Player Profile Modal */}
      <PlayerProfile
        playerId={selectedPlayerId || ''}
        isVisible={!!selectedPlayerId}
        onClose={() => setSelectedPlayerId(null)}
        currentPlayerId={appState.playerId}
      />
    </div>
  );
}