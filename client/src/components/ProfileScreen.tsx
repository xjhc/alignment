import { useState, useEffect } from 'react';
import { useSessionContext } from '../contexts/SessionContext';
import { Button } from './ui';

interface PlayerProfile {
  id: string;
  username: string;
  displayName: string;
  email: string;
  equippedAvatar: string;
  equippedTitle: string;
  totalGamesPlayed: number;
  totalGamesWon: number;
  totalTokensMined: number;
  kudosReceived: number;
  unlockedAchievements: string[];
  unlockedAvatars: string[];
  unlockedTitles: string[];
}

interface Avatar {
  id: string;
  name: string;
  iconEmoji: string;
  description: string;
  rarity: string;
}

interface Title {
  id: string;
  name: string;
  description: string;
  color: string;
  rarity: string;
}

interface Achievement {
  id: string;
  name: string;
  description: string;
  rarity: string;
  iconUrl?: string;
}

interface GameHistory {
  id: string;
  gameId: string;
  role: string;
  alignment: string;
  winnerFaction: string;
  isWinner: boolean;
  tokensMined: number;
  daysSurvived: number;
  createdAt: string;
}

export function ProfileScreen() {
  const { onBackToLogin } = useSessionContext();
  const [profile, setProfile] = useState<PlayerProfile | null>(null);
  const [avatars, setAvatars] = useState<Avatar[]>([]);
  const [titles, setTitles] = useState<Title[]>([]);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [gameHistory, setGameHistory] = useState<GameHistory[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'overview' | 'customization' | 'achievements' | 'history'>('overview');

  useEffect(() => {
    // In a real implementation, this would fetch from the API
    // For now, we'll use mock data
    fetchProfileData();
  }, []);

  const fetchProfileData = async () => {
    try {
      // Mock data for demonstration
      setProfile({
        id: 'user-123',
        username: 'player1',
        displayName: 'Elite Player',
        email: 'player@example.com',
        equippedAvatar: 'default',
        equippedTitle: 'rookie',
        totalGamesPlayed: 25,
        totalGamesWon: 12,
        totalTokensMined: 150,
        kudosReceived: 8,
        unlockedAchievements: ['first_win', 'survive_5_days'],
        unlockedAvatars: ['default', 'scientist'],
        unlockedTitles: ['rookie', 'veteran'],
      });

      setAvatars([
        { id: 'default', name: 'Default Avatar', iconEmoji: '👤', description: 'The standard corporate headshot', rarity: 'common' },
        { id: 'scientist', name: 'Scientist', iconEmoji: '🧑‍🔬', description: 'For the analytically minded', rarity: 'common' },
        { id: 'detective', name: 'Detective', iconEmoji: '🕵️', description: 'Trust but verify', rarity: 'rare' },
        { id: 'robot', name: 'Robot', iconEmoji: '🤖', description: 'Embrace the silicon future', rarity: 'epic' },
      ]);

      setTitles([
        { id: 'rookie', name: 'Rookie', description: 'New to the corporate world', color: '#ffffff', rarity: 'common' },
        { id: 'veteran', name: 'Veteran', description: 'Seasoned corporate warrior', color: '#4ade80', rarity: 'rare' },
        { id: 'mastermind', name: 'Mastermind', description: 'Strategic genius', color: '#8b5cf6', rarity: 'epic' },
      ]);

      setAchievements([
        { id: 'first_win', name: 'First Victory', description: 'Win your first game', rarity: 'common' },
        { id: 'survive_5_days', name: 'Survivor', description: 'Survive 5 days in a single game', rarity: 'rare' },
        { id: 'perfect_detective', name: 'Perfect Detective', description: 'Vote correctly in all elimination rounds', rarity: 'epic' },
      ]);

      setGameHistory([
        {
          id: '1',
          gameId: 'game-123',
          role: 'CISO',
          alignment: 'HUMAN',
          winnerFaction: 'HUMAN',
          isWinner: true,
          tokensMined: 8,
          daysSurvived: 4,
          createdAt: '2024-01-15T10:30:00Z',
        },
        // Add more mock game history...
      ]);

    } catch (error) {
      console.error('Failed to fetch profile data:', error);
    } finally {
      setLoading(false);
    }
  };

  const equipAvatar = async (avatarId: string) => {
    // In a real implementation, this would call the API
    console.log('Equipping avatar:', avatarId);
    if (profile) {
      setProfile({ ...profile, equippedAvatar: avatarId });
    }
  };

  const equipTitle = async (titleId: string) => {
    // In a real implementation, this would call the API
    console.log('Equipping title:', titleId);
    if (profile) {
      setProfile({ ...profile, equippedTitle: titleId });
    }
  };

  const getRarityClass = (rarity: string) => {
    switch (rarity) {
      case 'common': return 'text-text-secondary';
      case 'rare': return 'text-info';
      case 'epic': return 'text-primary';
      case 'legendary': return 'text-warning';
      default: return 'text-text-secondary';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background-primary text-text-primary flex items-center justify-center">
        <div className="text-lg">Loading profile...</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen bg-background-primary text-text-primary flex items-center justify-center">
        <div className="text-center">
          <div className="text-lg mb-4">Failed to load profile</div>
          <Button onClick={onBackToLogin}>Back to Login</Button>
        </div>
      </div>
    );
  }

  const renderOverviewTab = () => (
    <div className="space-y-6">
      <div className="bg-background-secondary border border-border rounded-xl p-6">
        <h3 className="text-xl font-bold text-text-primary mb-4">Player Statistics</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-primary">{profile.totalGamesPlayed}</div>
            <div className="text-sm text-text-secondary">Games Played</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-success">{profile.totalGamesWon}</div>
            <div className="text-sm text-text-secondary">Games Won</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-warning">{profile.totalTokensMined}</div>
            <div className="text-sm text-text-secondary">Tokens Mined</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-info">{profile.kudosReceived}</div>
            <div className="text-sm text-text-secondary">Kudos Received</div>
          </div>
        </div>
        <div className="mt-4 text-center">
          <div className="text-lg text-text-primary">
            Win Rate: {profile.totalGamesPlayed > 0 ? Math.round((profile.totalGamesWon / profile.totalGamesPlayed) * 100) : 0}%
          </div>
        </div>
      </div>

      <div className="bg-background-secondary border border-border rounded-xl p-6">
        <h3 className="text-xl font-bold text-text-primary mb-4">Current Loadout</h3>
        <div className="flex items-center gap-6">
          <div className="text-center">
            <div className="text-4xl mb-2">
              {avatars.find(a => a.id === profile.equippedAvatar)?.iconEmoji || '👤'}
            </div>
            <div className="text-sm text-text-secondary">Avatar</div>
            <div className="font-semibold text-text-primary">
              {avatars.find(a => a.id === profile.equippedAvatar)?.name || 'Default'}
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-bold mb-2" style={{ color: titles.find(t => t.id === profile.equippedTitle)?.color || '#ffffff' }}>
              {titles.find(t => t.id === profile.equippedTitle)?.name || 'No Title'}
            </div>
            <div className="text-sm text-text-secondary">Title</div>
          </div>
        </div>
      </div>
    </div>
  );

  const renderCustomizationTab = () => (
    <div className="space-y-6">
      <div className="bg-background-secondary border border-border rounded-xl p-6">
        <h3 className="text-xl font-bold text-text-primary mb-4">Avatars</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {avatars.map(avatar => {
            const isUnlocked = profile.unlockedAvatars.includes(avatar.id);
            const isEquipped = profile.equippedAvatar === avatar.id;
            
            return (
              <div key={avatar.id} className={`p-4 rounded-lg border-2 transition-colors ${
                isEquipped ? 'border-primary bg-primary/10' : 
                isUnlocked ? 'border-border bg-background-tertiary hover:border-primary/50' : 
                'border-border/50 bg-background-primary opacity-50'
              }`}>
                <div className="text-center">
                  <div className="text-3xl mb-2">{avatar.iconEmoji}</div>
                  <div className="font-semibold text-sm text-text-primary">{avatar.name}</div>
                  <div className={`text-xs ${getRarityClass(avatar.rarity)}`}>{avatar.rarity}</div>
                  <div className="text-xs text-text-secondary mt-1">{avatar.description}</div>
                  {isUnlocked && !isEquipped && (
                    <Button size="sm" className="mt-2" onClick={() => equipAvatar(avatar.id)}>
                      Equip
                    </Button>
                  )}
                  {isEquipped && (
                    <div className="mt-2 text-xs text-primary font-semibold">EQUIPPED</div>
                  )}
                  {!isUnlocked && (
                    <div className="mt-2 text-xs text-text-muted">LOCKED</div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className="bg-background-secondary border border-border rounded-xl p-6">
        <h3 className="text-xl font-bold text-text-primary mb-4">Titles</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {titles.map(title => {
            const isUnlocked = profile.unlockedTitles.includes(title.id);
            const isEquipped = profile.equippedTitle === title.id;
            
            return (
              <div key={title.id} className={`p-4 rounded-lg border-2 transition-colors ${
                isEquipped ? 'border-primary bg-primary/10' : 
                isUnlocked ? 'border-border bg-background-tertiary hover:border-primary/50' : 
                'border-border/50 bg-background-primary opacity-50'
              }`}>
                <div className="text-center">
                  <div className="font-bold text-lg mb-2" style={{ color: title.color }}>{title.name}</div>
                  <div className={`text-xs ${getRarityClass(title.rarity)}`}>{title.rarity}</div>
                  <div className="text-xs text-text-secondary mt-1">{title.description}</div>
                  {isUnlocked && !isEquipped && (
                    <Button size="sm" className="mt-2" onClick={() => equipTitle(title.id)}>
                      Equip
                    </Button>
                  )}
                  {isEquipped && (
                    <div className="mt-2 text-xs text-primary font-semibold">EQUIPPED</div>
                  )}
                  {!isUnlocked && (
                    <div className="mt-2 text-xs text-text-muted">LOCKED</div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );

  const renderAchievementsTab = () => (
    <div className="bg-background-secondary border border-border rounded-xl p-6">
      <h3 className="text-xl font-bold text-text-primary mb-4">Achievements</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {achievements.map(achievement => {
          const isUnlocked = profile.unlockedAchievements.includes(achievement.id);
          
          return (
            <div key={achievement.id} className={`p-4 rounded-lg border transition-colors ${
              isUnlocked ? 'border-success bg-success/10' : 'border-border/50 bg-background-primary opacity-50'
            }`}>
              <div className="flex items-start gap-3">
                <div className="text-2xl">
                  {isUnlocked ? '🏆' : '🔒'}
                </div>
                <div className="flex-1">
                  <div className="font-semibold text-text-primary">{achievement.name}</div>
                  <div className={`text-xs ${getRarityClass(achievement.rarity)}`}>{achievement.rarity}</div>
                  <div className="text-sm text-text-secondary mt-1">{achievement.description}</div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );

  const renderHistoryTab = () => (
    <div className="bg-background-secondary border border-border rounded-xl p-6">
      <h3 className="text-xl font-bold text-text-primary mb-4">Recent Games</h3>
      <div className="space-y-4">
        {gameHistory.map(game => (
          <div key={game.id} className="flex items-center justify-between p-4 bg-background-tertiary rounded-lg">
            <div className="flex items-center gap-4">
              <div className={`w-3 h-3 rounded-full ${game.isWinner ? 'bg-success' : 'bg-danger'}`}></div>
              <div>
                <div className="font-semibold text-text-primary">{game.role}</div>
                <div className="text-sm text-text-secondary">{game.alignment}</div>
              </div>
            </div>
            <div className="text-right">
              <div className="text-sm text-text-primary">
                {game.isWinner ? 'Victory' : 'Defeat'}
              </div>
              <div className="text-xs text-text-secondary">
                {game.daysSurvived} days, {game.tokensMined} tokens
              </div>
            </div>
          </div>
        ))}
        {gameHistory.length === 0 && (
          <div className="text-center text-text-secondary py-8">
            No games played yet
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-background-primary text-text-primary">
      <header className="p-4 border-b border-border bg-background-secondary">
        <div className="max-w-6xl mx-auto flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold">{profile.displayName}</h1>
            <div className="text-text-secondary">@{profile.username}</div>
          </div>
          <Button variant="secondary" onClick={onBackToLogin}>
            Back to Menu
          </Button>
        </div>
      </header>

      <div className="max-w-6xl mx-auto p-6">
        <div className="border-b border-border mb-6">
          <div className="flex space-x-6">
            {[
              { id: 'overview', label: 'Overview' },
              { id: 'customization', label: 'Customization' },
              { id: 'achievements', label: 'Achievements' },
              { id: 'history', label: 'Game History' },
            ].map(tab => (
              <button
                key={tab.id}
                className={`pb-3 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-text-muted hover:text-text-primary'
                }`}
                onClick={() => setActiveTab(tab.id as any)}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        <div className="animate-fade-in">
          {activeTab === 'overview' && renderOverviewTab()}
          {activeTab === 'customization' && renderCustomizationTab()}
          {activeTab === 'achievements' && renderAchievementsTab()}
          {activeTab === 'history' && renderHistoryTab()}
        </div>
      </div>
    </div>
  );
}