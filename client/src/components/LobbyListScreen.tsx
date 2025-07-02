import { useState, useEffect } from 'react';
import { Button } from './ui';
import { FriendsPanel } from './FriendsPanel';
import { PartyPanel } from './PartyPanel';
import { getUserIdForApi } from '../services/guestIdentity';

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

  // Fetch lobby list from REST API
  const fetchLobbies = async () => {
    try {
      setError(null);
      const response = await fetch('/api/games');
      if (!response.ok) {
        throw new Error(`Failed to fetch lobbies: ${response.statusText}`);
      }
      const data = await response.json();
      setLobbies(data.lobbies || []);
    } catch (error) {
      console.error('Error fetching lobbies:', error);
      setError(error instanceof Error ? error.message : 'Failed to fetch lobbies');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLobbies();
    // Poll for updates every 5 seconds
    const interval = setInterval(fetchLobbies, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleJoinLobby = async (gameId: string) => {
    try {
      setError(null);
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
        throw new Error(errorText || 'Failed to join lobby');
      }

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
        
        <div className="flex flex-col gap-2 bg-background-secondary rounded-lg p-4">
          <div className="grid grid-cols-4 items-center gap-4 px-4 py-3 text-text-muted text-xs uppercase bg-transparent">
            <div>Lobby Name</div>
            <div>Players</div>
            <div>Status</div>
            <div>Action</div>
          </div>
          
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
            lobbies.map((lobby) => (
              <div key={lobby.id} className="grid grid-cols-4 items-center gap-4 px-4 py-3 bg-background-primary rounded-md transition-all duration-200 hover:bg-background-hover">
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
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => handleJoinLobby(lobby.id)}
                    disabled={!lobby.can_join}
                    className="text-sm font-medium"
                  >
                    Join
                  </Button>
                </div>
              </div>
            ))
          )}
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