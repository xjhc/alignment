import { useState, useEffect } from 'react';

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
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Loading lobbies...</h2>
        </div>
      </div>
    );
  }

  return (
    <div className="launch-screen">
      <h1 className="logo">
        LOEBIAN INC. // <span className="glitch">EMERGENCY BRIDGE</span>
      </h1>
      
      <div className="lobby-list-container">
        <div className="lobby-list-header">
          <h2>Game Lobbies</h2>
          <button className="btn-secondary" onClick={handleCreateGame}>
            + Create New Game
          </button>
        </div>

        {error && (
          <div className="error-message" style={{ color: 'var(--error)', margin: '1rem 0', padding: '0.5rem', background: 'rgba(255, 0, 0, 0.1)', borderRadius: '4px' }}>
            {error}
          </div>
        )}
        
        <div className="lobby-list">
          <div className="lobby-item lobby-header">
            <div>Lobby Name</div>
            <div>Players</div>
            <div>Status</div>
            <div>Action</div>
          </div>
          
          {lobbies.length === 0 ? (
            <div className="lobby-item" style={{ gridColumn: '1 / -1', textAlign: 'center', color: 'var(--text-secondary)' }}>
              No active lobbies. Create one to get started!
            </div>
          ) : (
            lobbies.map((lobby) => (
              <div key={lobby.id} className="lobby-item">
                <div className="lobby-name">#{lobby.name}</div>
                <div>{lobby.player_count} / {lobby.max_players}</div>
                <div>
                  <span className={`lobby-status ${lobby.status ? lobby.status.toLowerCase().replace('_', '-') : 'unknown'}`}>
                    {lobby.status || 'Unknown'}
                  </span>
                </div>
                <div>
                  <button
                    className="btn-secondary"
                    onClick={() => handleJoinLobby(lobby.id)}
                    disabled={!lobby.can_join}
                  >
                    Join
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
        
        <button onClick={onBack} className="back-button">
          ← Back
        </button>
      </div>
    </div>
  );
}