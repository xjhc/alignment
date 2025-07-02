import { useState, useEffect } from 'react';
import { useWebSocketEvent } from '../hooks/useWebSocket';

interface Achievement {
  id: string;
  name: string;
  description: string;
  rarity: string;
  iconUrl?: string;
  avatarReward?: string;
  titleReward?: string;
}

interface AchievementNotificationProps {
  achievement: Achievement;
  onClose: () => void;
}

function AchievementNotification({ achievement, onClose }: AchievementNotificationProps) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Trigger animation on mount
    setIsVisible(true);

    // Auto-close after 5 seconds
    const timer = setTimeout(() => {
      handleClose();
    }, 5000);

    return () => clearTimeout(timer);
  }, []);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(onClose, 300); // Wait for animation to complete
  };

  const getRarityClass = (rarity: string) => {
    switch (rarity) {
      case 'common': return 'border-text-secondary bg-background-secondary';
      case 'rare': return 'border-info bg-info/10';
      case 'epic': return 'border-primary bg-primary/10';
      case 'legendary': return 'border-warning bg-warning/10';
      default: return 'border-text-secondary bg-background-secondary';
    }
  };

  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case 'common': return '#9ca3af';
      case 'rare': return '#3b82f6';
      case 'epic': return '#8b5cf6';
      case 'legendary': return '#f59e0b';
      default: return '#9ca3af';
    }
  };

  return (
    <div
      className={`fixed top-4 right-4 z-50 max-w-sm p-4 rounded-xl border-2 shadow-lg transition-all duration-300 ${
        isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'
      } ${getRarityClass(achievement.rarity)}`}
    >
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          <div className="w-12 h-12 rounded-full bg-warning/20 flex items-center justify-center text-2xl">
            🏆
          </div>
        </div>
        
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="font-bold text-text-primary text-sm">Achievement Unlocked!</h3>
            <button
              onClick={handleClose}
              className="text-text-muted hover:text-text-primary text-lg leading-none"
            >
              ×
            </button>
          </div>
          
          <div className="mb-2">
            <div className="font-semibold text-text-primary">{achievement.name}</div>
            <div 
              className="text-xs font-medium uppercase tracking-wide"
              style={{ color: getRarityColor(achievement.rarity) }}
            >
              {achievement.rarity}
            </div>
          </div>
          
          <div className="text-sm text-text-secondary mb-3">
            {achievement.description}
          </div>
          
          {(achievement.avatarReward || achievement.titleReward) && (
            <div className="text-xs text-success">
              <div className="font-medium">Rewards:</div>
              {achievement.avatarReward && <div>• New Avatar Unlocked</div>}
              {achievement.titleReward && <div>• New Title Unlocked</div>}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Achievement notification manager component
export function AchievementNotificationManager() {
  const [notifications, setNotifications] = useState<Achievement[]>([]);

  // Listen for achievement unlocked events
  useWebSocketEvent('PRIVATE_NOTIFICATION', (payload: any) => {
    if (payload.type === 'ACHIEVEMENT_UNLOCKED' && payload.achievement) {
      setNotifications(prev => [...prev, payload.achievement]);
    }
  });

  const removeNotification = (index: number) => {
    setNotifications(prev => prev.filter((_, i) => i !== index));
  };

  return (
    <div className="fixed top-0 right-0 z-50 pointer-events-none">
      {notifications.map((achievement, index) => (
        <div
          key={`${achievement.id}-${index}`}
          className="pointer-events-auto"
          style={{ marginTop: `${index * 120}px` }}
        >
          <AchievementNotification
            achievement={achievement}
            onClose={() => removeNotification(index)}
          />
        </div>
      ))}
    </div>
  );
}