import { useState, useEffect } from 'react';
import { Button } from './ui';

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
      const response = await fetch(`/api/games/${gameId}/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
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
      const response = await fetch('/api/games', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          lobby_name: `${playerName}'s Game`,
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
          <h2>Loading lobbies...</h2>
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
          <Button
            variant="secondary"
            onClick={handleCreateGame}
            className="text-sm font-medium"
          >
            + Create New Game
          </Button>
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
            <div className="grid grid-cols-4 items-center gap-4 px-4 py-3 bg-background-primary rounded-md transition-all duration-200 hover:bg-background-hover col-span-4 text-center text-text-secondary">
              No active lobbies. Create one to get started!
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
    </div>
  );
}