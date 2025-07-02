import React, { useState, useEffect } from 'react';
import { Button } from './ui/Button';

interface PlayerStats {
  gamesPlayed: number;
  gamesWon: number;
  winRate: number;
  averageGameLength: number;
  favoriteRole: string;
  totalPlayTime: number;
  achievementsUnlocked: number;
  kudosReceived: number;
  friendsCount: number;
}

interface Achievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  unlockedAt?: string;
  isRare: boolean;
}

interface GameHistory {
  id: string;
  date: string;
  result: 'won' | 'lost';
  role: string;
  alignment: string;
  duration: number;
  playersCount: number;
}

interface PlayerData {
  id: string;
  username: string;
  displayName: string;
  bio?: string;
  avatar: string;
  title?: string;
  joinDate: string;
  lastSeen: string;
  isOnline: boolean;
  stats: PlayerStats;
  achievements: Achievement[];
  recentGames: GameHistory[];
  isFriend: boolean;
  isBlocked: boolean;
}

interface PlayerProfileProps {
  playerId: string;
  isVisible: boolean;
  onClose: () => void;
  currentPlayerId?: string;
}

export const PlayerProfile: React.FC<PlayerProfileProps> = ({ 
  playerId, 
  isVisible, 
  onClose, 
  currentPlayerId 
}) => {
  const [playerData, setPlayerData] = useState<PlayerData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'achievements' | 'history'>('overview');

  useEffect(() => {
    if (isVisible && playerId) {
      loadPlayerProfile();
    }
  }, [isVisible, playerId]);

  const loadPlayerProfile = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const response = await fetch(`/api/players/${playerId}/profile`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setPlayerData(data.player);
      } else {
        setError('Failed to load player profile');
      }
    } catch (err) {
      setError('Error loading player profile');
      console.error('Error loading player profile:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const sendFriendRequest = async () => {
    if (!playerData) return;

    try {
      const response = await fetch('/api/friends/requests', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          recipient_username: playerData.username,
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
      console.error('Error sending friend request:', err);
    }
  };

  const blockPlayer = async () => {
    if (!playerData || !confirm(`Are you sure you want to block ${playerData.displayName}?`)) {
      return;
    }

    try {
      const response = await fetch('/api/users/me/block', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          player_id: playerData.id,
        }),
      });

      if (response.ok) {
        setPlayerData(prev => prev ? { ...prev, isBlocked: true } : null);
        alert('Player blocked successfully');
      } else {
        alert('Failed to block player');
      }
    } catch (err) {
      alert('Error blocking player');
      console.error('Error blocking player:', err);
    }
  };

  const giveKudos = async () => {
    if (!playerData) return;

    const reason = prompt('Why are you giving kudos to this player?');
    if (!reason) return;

    try {
      const response = await fetch('/api/users/me/kudos', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          target_player_id: playerData.id,
          reason: reason,
        }),
      });

      if (response.ok) {
        alert('Kudos sent!');
        loadPlayerProfile(); // Refresh to show updated kudos count
      } else {
        const errorData = await response.json();
        alert(errorData.error || 'Failed to give kudos');
      }
    } catch (err) {
      alert('Error giving kudos');
      console.error('Error giving kudos:', err);
    }
  };

  const formatDuration = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return hours > 0 ? `${hours}h ${mins}m` : `${mins}m`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusColor = (isOnline: boolean) => {
    return isOnline ? 'text-green' : 'text-text-muted';
  };

  const isOwnProfile = currentPlayerId === playerId;

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-background-primary border border-border rounded-lg w-full max-w-2xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-text-primary">Player Profile</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-text-secondary hover:text-text-primary"
          >
            ✕
          </Button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto">
          {isLoading ? (
            <div className="text-center p-8 text-text-secondary">Loading profile...</div>
          ) : error ? (
            <div className="text-center p-8 text-danger">{error}</div>
          ) : !playerData ? (
            <div className="text-center p-8 text-text-secondary">Player not found</div>
          ) : (
            <>
              {/* Profile Header */}
              <div className="p-6 border-b border-border">
                <div className="flex items-start gap-4">
                  <div className="w-16 h-16 bg-primary rounded-full flex items-center justify-center text-white text-2xl font-bold flex-shrink-0">
                    {playerData.avatar || playerData.displayName[0].toUpperCase()}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-xl font-semibold text-text-primary">
                        {playerData.displayName}
                      </h3>
                      <span className={`text-sm ${getStatusColor(playerData.isOnline)}`}>
                        {playerData.isOnline ? '● Online' : '○ Offline'}
                      </span>
                    </div>
                    <p className="text-text-secondary text-sm">@{playerData.username}</p>
                    {playerData.title && (
                      <p className="text-amber text-sm font-medium mt-1">{playerData.title}</p>
                    )}
                    {playerData.bio && (
                      <p className="text-text-primary text-sm mt-2">{playerData.bio}</p>
                    )}
                    <p className="text-text-muted text-xs mt-2">
                      Joined {formatDate(playerData.joinDate)} • Last seen {formatDate(playerData.lastSeen)}
                    </p>
                  </div>
                </div>

                {/* Action Buttons */}
                {!isOwnProfile && (
                  <div className="flex gap-2 mt-4">
                    {!playerData.isFriend && !playerData.isBlocked && (
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={sendFriendRequest}
                      >
                        Add Friend
                      </Button>
                    )}
                    {playerData.isFriend && (
                      <span className="text-green text-sm">✓ Friends</span>
                    )}
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={giveKudos}
                      disabled={playerData.isBlocked}
                    >
                      👏 Give Kudos
                    </Button>
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={blockPlayer}
                      disabled={playerData.isBlocked}
                    >
                      {playerData.isBlocked ? 'Blocked' : 'Block'}
                    </Button>
                  </div>
                )}
              </div>

              {/* Tabs */}
              <div className="flex border-b border-border">
                <button
                  className={`flex-1 px-4 py-3 text-sm font-medium ${
                    activeTab === 'overview'
                      ? 'text-primary border-b-2 border-primary'
                      : 'text-text-secondary hover:text-text-primary'
                  }`}
                  onClick={() => setActiveTab('overview')}
                >
                  Overview
                </button>
                <button
                  className={`flex-1 px-4 py-3 text-sm font-medium ${
                    activeTab === 'achievements'
                      ? 'text-primary border-b-2 border-primary'
                      : 'text-text-secondary hover:text-text-primary'
                  }`}
                  onClick={() => setActiveTab('achievements')}
                >
                  Achievements ({playerData.achievements.filter(a => a.unlockedAt).length})
                </button>
                <button
                  className={`flex-1 px-4 py-3 text-sm font-medium ${
                    activeTab === 'history'
                      ? 'text-primary border-b-2 border-primary'
                      : 'text-text-secondary hover:text-text-primary'
                  }`}
                  onClick={() => setActiveTab('history')}
                >
                  Game History
                </button>
              </div>

              {/* Tab Content */}
              <div className="p-6">
                {activeTab === 'overview' && (
                  <div className="space-y-6">
                    {/* Stats Grid */}
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div className="text-center p-3 bg-background-secondary rounded">
                        <div className="text-xl font-bold text-primary">{playerData.stats.gamesPlayed}</div>
                        <div className="text-xs text-text-muted">Games Played</div>
                      </div>
                      <div className="text-center p-3 bg-background-secondary rounded">
                        <div className="text-xl font-bold text-green">{playerData.stats.winRate}%</div>
                        <div className="text-xs text-text-muted">Win Rate</div>
                      </div>
                      <div className="text-center p-3 bg-background-secondary rounded">
                        <div className="text-xl font-bold text-amber">{playerData.stats.kudosReceived}</div>
                        <div className="text-xs text-text-muted">Kudos Received</div>
                      </div>
                      <div className="text-center p-3 bg-background-secondary rounded">
                        <div className="text-xl font-bold text-blue">{playerData.stats.friendsCount}</div>
                        <div className="text-xs text-text-muted">Friends</div>
                      </div>
                    </div>

                    {/* Additional Stats */}
                    <div className="space-y-3">
                      <div className="flex justify-between items-center p-3 bg-background-secondary rounded">
                        <span className="text-text-primary">Favorite Role</span>
                        <span className="font-medium text-primary">{playerData.stats.favoriteRole}</span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-background-secondary rounded">
                        <span className="text-text-primary">Average Game Length</span>
                        <span className="font-medium">{formatDuration(playerData.stats.averageGameLength)}</span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-background-secondary rounded">
                        <span className="text-text-primary">Total Play Time</span>
                        <span className="font-medium">{formatDuration(playerData.stats.totalPlayTime)}</span>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'achievements' && (
                  <div className="space-y-3">
                    {playerData.achievements.map((achievement) => (
                      <div
                        key={achievement.id}
                        className={`flex items-center gap-3 p-3 rounded border ${
                          achievement.unlockedAt
                            ? 'bg-background-secondary border-border'
                            : 'bg-background-tertiary border-border opacity-50'
                        }`}
                      >
                        <div className="text-2xl">{achievement.icon}</div>
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <h4 className="font-medium text-text-primary">{achievement.name}</h4>
                            {achievement.isRare && (
                              <span className="text-xs bg-amber text-black px-2 py-0.5 rounded">RARE</span>
                            )}
                          </div>
                          <p className="text-sm text-text-secondary">{achievement.description}</p>
                          {achievement.unlockedAt && (
                            <p className="text-xs text-text-muted mt-1">
                              Unlocked on {formatDate(achievement.unlockedAt)}
                            </p>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {activeTab === 'history' && (
                  <div className="space-y-2">
                    {playerData.recentGames.map((game) => (
                      <div
                        key={game.id}
                        className="flex items-center justify-between p-3 bg-background-secondary rounded border"
                      >
                        <div className="flex items-center gap-3">
                          <div
                            className={`w-3 h-3 rounded-full ${
                              game.result === 'won' ? 'bg-green' : 'bg-danger'
                            }`}
                          />
                          <div>
                            <div className="font-medium text-text-primary">
                              {game.role} ({game.alignment})
                            </div>
                            <div className="text-sm text-text-secondary">
                              {game.playersCount} players • {formatDuration(game.duration)}
                            </div>
                          </div>
                        </div>
                        <div className="text-xs text-text-muted">
                          {formatDate(game.date)}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};