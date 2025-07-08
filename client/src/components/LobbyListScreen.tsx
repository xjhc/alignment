import { useState, useEffect, useRef } from 'react';
import { Button } from './ui';
import { FriendsPanel } from './FriendsPanel';
import { PartyPanel } from './PartyPanel';
import { getUserIdForApi } from '../services/guestIdentity';
import { websocketClient } from '../services/websocket';
import { ClientActionType } from '../types/generated';
import { STAGGER_CHILD, applyStaggeredAnimation, SLIDE_IN_UP, FADE_IN } from '../utils/animations';

interface LobbyInfo {
  id: string;
  name: string;
  player_count: number;
  max_players: number;
  min_players: number;
  status: string;
  can_join: boolean;
  created_at: string;
}

interface LobbyListScreenProps {
  playerName: string;
  playerAvatar?: string;
  onJoinLobby: (gameId: string, playerId: string, sessionToken: string) => void;
  onCreateGame: (gameId: string, playerId: string, sessionToken: string) => void;
  onBack: () => void;
}

export function LobbyListScreen({ playerName, playerAvatar, onJoinLobby, onCreateGame, onBack }: LobbyListScreenProps) {
  const [lobbies, setLobbies] = useState<LobbyInfo[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showFriendsPanel, setShowFriendsPanel] = useState(false);
  const [showPartyPanel, setShowPartyPanel] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [pollingInterval, setPollingInterval] = useState(10000); // Start with 10 seconds instead of 5
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const [showSessionConflict, setShowSessionConflict] = useState(false);
  const [joinCooldowns, setJoinCooldowns] = useState<Record<string, number>>({});
  const lobbyListRef = useRef<HTMLDivElement>(null);

  // Clear any existing WebSocket session to reset backend state
  const clearSession = async () => {
    try {
      console.log('Clearing existing session...');
      
      // First, try to send ABANDON_GAME action to the server if we have session data
      const savedSession = sessionStorage.getItem('alignmentGameSession');
      if (savedSession) {
        try {
          const sessionData = JSON.parse(savedSession);
          if (sessionData.gameId && sessionData.playerId) {
            console.log('Preparing to send leave/abandon action to server');
            
            // Send appropriate action based on session state
            if (websocketClient.isValidConnection()) {
              // Use ABANDON_GAME for active games, LEAVE_GAME for lobbies
              const actionType = sessionData.sessionState === 'IN_GAME' 
                ? ClientActionType.AbandonGame 
                : ClientActionType.LeaveGame;
              
              console.log(`Sending ${actionType} action to server`);
              websocketClient.sendAction({
                type: actionType,
                payload: {
                  game_id: sessionData.gameId,
                  player_id: sessionData.playerId
                }
              });
              
              // Give the action a moment to be sent before disconnecting
              await new Promise(resolve => setTimeout(resolve, 100));
            }
          }
        } catch (parseError) {
          console.warn('Failed to parse session data for abandonment:', parseError);
        }
      }
      
      // Disconnect WebSocket connection
      websocketClient.disconnect();
      
      // Clear session storage
      sessionStorage.removeItem('alignmentGameSession');
      
      setShowSessionConflict(false);
      setError(null);
      
      // Refresh lobbies after clearing session
      await fetchLobbies();
    } catch (error) {
      console.error('Error clearing session:', error);
    }
  };

  // Rejoin the existing game session
  const rejoinSession = () => {
    try {
      const savedSession = sessionStorage.getItem('alignmentGameSession');
      if (savedSession) {
        const sessionData = JSON.parse(savedSession);
        if (sessionData.gameId && sessionData.playerId && sessionData.sessionToken) {
          console.log('Rejoining existing session:', sessionData);
          onJoinLobby(sessionData.gameId, sessionData.playerId, sessionData.sessionToken);
          setShowSessionConflict(false);
          return;
        }
      }
      console.error('No valid session data found for rejoin');
      setError('Unable to rejoin: No valid session data found');
      setShowSessionConflict(false);
    } catch (error) {
      console.error('Error rejoining session:', error);
      setError('Failed to rejoin session');
      setShowSessionConflict(false);
    }
  };

  // Restart polling with current interval
  const restartPolling = () => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }
    intervalRef.current = setInterval(() => fetchLobbies(), pollingInterval);
  };

  // Fetch lobby list from REST API with exponential backoff for rate limiting
  const fetchLobbies = async (isRetry = false) => {
    try {
      if (!isRetry) {
        setError(null);
      }
      const response = await fetch('/api/games');
      
      if (response.status === 429) {
        // Rate limited - implement exponential backoff
        const newRetryCount = retryCount + 1;
        const backoffDelay = Math.min(1000 * Math.pow(2, newRetryCount), 60000); // Max 1 minute
        setRetryCount(newRetryCount);
        const newPollingInterval = Math.max(backoffDelay, 30000);
        setPollingInterval(newPollingInterval); // Increase polling interval when rate limited
        console.warn(`Rate limited. Backing off for ${backoffDelay}ms. Next poll in ${newPollingInterval}ms`);
        
        restartPolling();
        setTimeout(() => {
          fetchLobbies(true);
        }, backoffDelay);
        return;
      }
      
      if (!response.ok) {
        throw new Error(`Failed to fetch lobbies: ${response.statusText}`);
      }
      
      const data = await response.json();
      const newLobbies = data.lobbies || [];
      
      // Apply staggered animation when lobbies are updated
      setLobbies(newLobbies);
      
      // Apply animations to lobby items after they're rendered
      setTimeout(() => {
        if (lobbyListRef.current) {
          const lobbyItems = lobbyListRef.current.querySelectorAll('.lobby-item');
          if (lobbyItems.length > 0) {
            applyStaggeredAnimation(lobbyItems, 100, FADE_IN);
          }
        }
      }, 50);
      
      // Reset retry count and polling interval on success
      if (retryCount > 0) {
        setRetryCount(0);
        setPollingInterval(10000); // Reset to normal 10-second interval
        restartPolling();
      }
    } catch (error) {
      console.error('Error fetching lobbies:', error);
      setError(error instanceof Error ? error.message : 'Failed to fetch lobbies');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLobbies();
    restartPolling();
    
    // Subscribe to WebSocket events that should trigger lobby list refreshes
    const handleLobbyChange = () => {
      fetchLobbies(); // Refresh lobby list immediately
    };
    
    // Listen for events that indicate lobby state changes
    websocketClient.on('GAME_CREATED', handleLobbyChange);
    websocketClient.on('PLAYER_JOINED', handleLobbyChange);
    websocketClient.on('PLAYER_LEFT', handleLobbyChange);
    websocketClient.on('GAME_STARTED', handleLobbyChange);
    websocketClient.on('GAME_ENDED', handleLobbyChange);
    websocketClient.on('HOST_TRANSFERRED', handleLobbyChange);
    
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
      
      // Clean up WebSocket event listeners
      websocketClient.off('GAME_CREATED', handleLobbyChange);
      websocketClient.off('PLAYER_JOINED', handleLobbyChange);
      websocketClient.off('PLAYER_LEFT', handleLobbyChange);
      websocketClient.off('GAME_STARTED', handleLobbyChange);
      websocketClient.off('GAME_ENDED', handleLobbyChange);
      websocketClient.off('HOST_TRANSFERRED', handleLobbyChange);
    };
  }, []); // Only run once on mount

  // Clean up expired cooldowns and trigger re-renders for countdown updates
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      setJoinCooldowns(prev => {
        const updated = { ...prev };
        let hasChanges = false;
        
        for (const [gameId, cooldownEnd] of Object.entries(updated)) {
          if (now >= cooldownEnd) {
            delete updated[gameId];
            hasChanges = true;
          }
        }
        
        return hasChanges ? updated : prev;
      });
    }, 1000); // Update every second for countdown

    return () => clearInterval(interval);
  }, []);

  const handleJoinLobby = async (gameId: string) => {
    try {
      setError(null);
      
      // Check if we're in a cooldown period for this specific lobby
      const now = Date.now();
      const cooldownEnd = joinCooldowns[gameId];
      if (cooldownEnd && now < cooldownEnd) {
        const remainingSeconds = Math.ceil((cooldownEnd - now) / 1000);
        setError(`Please wait ${remainingSeconds} seconds before trying to join this lobby again.`);
        return;
      }

      const userId = getUserIdForApi();
      const response = await fetch(`/api/games/${gameId}/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          player_name: playerName,
          player_avatar: playerAvatar || '',
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        if (response.status === 409 && errorText.includes('already in an active game session')) {
          setShowSessionConflict(true);
        } else if (response.status === 429) {
          // Rate limited - set a cooldown for this specific lobby
          setJoinCooldowns(prev => ({
            ...prev,
            [gameId]: now + 5000 // 5 second cooldown
          }));
          setError('Too many join attempts. Please wait a moment before trying again.');
          return;
        }
        throw new Error(errorText || 'Failed to join lobby');
      }

      // Clear any existing cooldown for this lobby on successful join
      setJoinCooldowns(prev => {
        const updated = { ...prev };
        delete updated[gameId];
        return updated;
      });

      const data = await response.json();
      // NEW: We now get playerId and sessionToken from the API
      onJoinLobby(data.game_id, data.player_id, data.session_token);
    } catch (error) {
      console.error('Failed to join lobby:', error);
      setError(error instanceof Error ? error.message : 'Failed to join lobby');
    }
  };

  const handleCreateGame = async () => {
    try {
      setError(null);
      const userId = getUserIdForApi();
      const response = await fetch('/api/games', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          lobby_name: `${playerName} Game`,
          player_name: playerName,
          player_avatar: playerAvatar || '',
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        if (response.status === 409 && errorText.includes('already in an active game session')) {
          setShowSessionConflict(true);
        }
        throw new Error(errorText || 'Failed to create game');
      }

      const data = await response.json();
      
      // NEW: Pass all the necessary info with session-based auth
      onCreateGame(data.game_id, data.player_id, data.session_token);
    } catch (error) {
      console.error('Failed to create game:', error);
      setError(error instanceof Error ? error.message : 'Failed to create game');
    }
  };

  if (isLoading) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <div className="flex flex-col gap-4 items-center w-80">
          <div className="animate-pulse text-2xl">🔍</div>
          <h2 className="text-lg font-medium">Scanning for active emergency sessions...</h2>
          <p className="text-text-secondary text-sm text-center">
            Connecting to Loebian Inc. crisis management network
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
      
      {/* User Profile Header */}
      <div className="flex items-center gap-4 px-6 py-4 bg-background-secondary rounded-lg border border-border">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-full bg-primary flex items-center justify-center text-background-primary text-lg font-semibold">
            {playerAvatar || playerName.charAt(0).toUpperCase()}
          </div>
          <div className="flex flex-col">
            <span className="font-medium text-text-primary">{playerName}</span>
            <span className="text-sm text-text-secondary">Emergency Response Agent</span>
          </div>
        </div>
        <div className="flex-1" />
        <div className="flex items-center gap-4 text-sm text-text-secondary">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-human"></div>
            <span>Human Status: Verified</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-success"></div>
            <span>Network: Connected</span>
          </div>
        </div>
      </div>
      
      <div className="flex flex-col gap-6 max-w-2xl mx-auto">
        <div className="flex justify-between items-center pb-4 border-b border-border">
          <h2>Game Lobbies</h2>
          <div className="flex gap-2">
            <Button
              variant="ghost"
              onClick={() => setShowFriendsPanel(true)}
              className="text-sm font-medium"
            >
              👥 Friends
            </Button>
            <Button
              variant="ghost"
              onClick={() => setShowPartyPanel(true)}
              className="text-sm font-medium"
            >
              🎉 Party
            </Button>
            <Button
              variant="secondary"
              onClick={handleCreateGame}
              className="text-sm font-medium"
            >
              + Create New Game
            </Button>
          </div>
        </div>

        {error && (
          <div className="text-red my-4 p-2 bg-red/10 rounded">
            {error}
          </div>
        )}

        {showSessionConflict && (
          <div className="bg-danger/10 border border-danger/20 rounded-lg p-4 my-4">
            <h3 className="text-danger font-semibold mb-2">🚨 Active Session Detected</h3>
            <p className="text-text-secondary text-sm mb-4">
              You're already in an active game session. You can rejoin your current session or clear it to start a new game.
            </p>
            <div className="flex gap-2">
              <Button
                variant="primary"
                size="sm"
                onClick={rejoinSession}
                className="text-sm font-medium"
              >
                Rejoin Game
              </Button>
              <Button
                variant="danger"
                size="sm"
                onClick={clearSession}
                className="text-sm font-medium"
              >
                Clear Current Session
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowSessionConflict(false)}
                className="text-sm"
              >
                Cancel
              </Button>
            </div>
          </div>
        )}
        
        <div className="flex flex-col gap-2 bg-background-secondary rounded-lg p-4">
          <div className="grid grid-cols-4 items-center gap-4 px-4 py-3 text-text-muted text-xs uppercase bg-transparent">
            <div>Lobby Name</div>
            <div>Players</div>
            <div>Status</div>
            <div>Action</div>
          </div>
          
          <div ref={lobbyListRef}>
            {lobbies.length === 0 ? (
              <div className="col-span-4 flex flex-col items-center justify-center py-12 px-6 bg-background-primary rounded-md border-2 border-dashed border-border animation-fade-in">
                <div className="text-4xl mb-4">🎮</div>
                <h3 className="text-lg font-medium text-text-primary mb-2">No active lobbies found</h3>
                <p className="text-text-secondary text-sm text-center mb-6 max-w-md">
                  Looks like you're the first to arrive! Be the pioneer and start a new emergency response session.
                </p>
                <Button
                  variant="primary"
                  onClick={handleCreateGame}
                  className="font-medium text-sm px-6 py-2 animation-scale-in-feedback"
                >
                  🚀 Create New Game
                </Button>
              </div>
            ) : (
              lobbies.map((lobby, index) => (
                <div 
                  key={lobby.id} 
                  className={`lobby-item grid grid-cols-4 items-center gap-4 px-4 py-3 bg-background-primary rounded-md transition-all duration-200 hover:bg-background-hover ${STAGGER_CHILD}`}
                  style={{ animationDelay: `${index * 100}ms` }}
                >
                  <div className="font-mono text-primary font-semibold">#{lobby.name}</div>
                  <div>{lobby.player_count} / {lobby.max_players}</div>
                  <div>
                    <span className={`px-2 py-1 rounded text-xs font-semibold uppercase ${
                      lobby.status === 'waiting' ? 'bg-human text-background-primary' :
                      lobby.status === 'in_progress' ? 'bg-danger text-background-primary' :
                      'bg-text-muted text-background-primary'
                    }`}>
                      {lobby.status || 'Unknown'}
                    </span>
                  </div>
                  <div>
                    {(() => {
                      const now = Date.now();
                      const cooldownEnd = joinCooldowns[lobby.id];
                      const inCooldown = cooldownEnd && now < cooldownEnd;
                      const remainingSeconds = inCooldown ? Math.ceil((cooldownEnd - now) / 1000) : 0;
                      
                      return (
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => handleJoinLobby(lobby.id)}
                          disabled={!lobby.can_join || inCooldown}
                          className="text-sm font-medium transition-all duration-200 hover:enabled:animation-scale-in-feedback"
                          title={inCooldown ? `Please wait ${remainingSeconds} seconds` : undefined}
                        >
                          {inCooldown ? `Wait ${remainingSeconds}s` : 'Join'}
                        </Button>
                      );
                    })()}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
        
        <Button
          onClick={onBack}
          variant="ghost"
          size="sm"
          className="self-start mt-4 text-text-muted hover:enabled:text-text-primary"
        >
          ← Back
        </Button>
      </div>

      {/* Friends Panel */}
      <FriendsPanel
        isVisible={showFriendsPanel}
        onClose={() => setShowFriendsPanel(false)}
      />

      {/* Party Panel */}
      <PartyPanel
        isVisible={showPartyPanel}
        onClose={() => setShowPartyPanel(false)}
        currentPlayerId="current-player-id" // TODO: Get from session context
      />
    </div>
  );
}