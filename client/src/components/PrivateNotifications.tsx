import { useState, useEffect } from 'react';
import { PrivateNotification } from '../types';
import { SystemShockNotification } from './SystemShockNotification';
import { Button } from './ui/Button';
import { SLIDE_IN_RIGHT } from '../utils/animations';

interface PrivateNotificationsProps {
  notifications: PrivateNotification[];
  onMarkAsRead?: (notificationId: string) => void;
}

export function PrivateNotifications({ notifications, onMarkAsRead }: PrivateNotificationsProps) {
  const [visibleNotifications, setVisibleNotifications] = useState<PrivateNotification[]>([]);

  useEffect(() => {
    // Show unread notifications as toasts
    const unreadNotifications = notifications.filter(n => !n.isRead);
    setVisibleNotifications(unreadNotifications);

    // Auto-hide notifications after 8 seconds for low/medium priority
    // High priority notifications stay until manually dismissed
    unreadNotifications.forEach(notification => {
      if (notification.priority !== 'high') {
        setTimeout(() => {
          setVisibleNotifications(prev => prev.filter(n => n.id !== notification.id));
          if (onMarkAsRead) {
            onMarkAsRead(notification.id);
          }
        }, 8000);
      }
    });
  }, [notifications, onMarkAsRead]);

  const handleDismiss = (notificationId: string) => {
    setVisibleNotifications(prev => prev.filter(n => n.id !== notificationId));
    if (onMarkAsRead) {
      onMarkAsRead(notificationId);
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'system_shock': return '⚡';
      case 'kpi_progress': return '📊';
      case 'role_ability': return '✨';
      case 'conversion': return '🤖';
      case 'investigation': return '🔍';
      default: return '📢';
    }
  };

  const getPriorityStyles = (priority: string) => {
    switch (priority) {
      case 'high': 
        return {
          container: 'bg-danger border-danger text-white shadow-lg',
          icon: 'text-white',
          title: 'text-white font-semibold',
          message: 'text-white/90',
          timestamp: 'text-white/70'
        };
      case 'medium': 
        return {
          container: 'bg-background-secondary border-primary text-text-primary shadow-md',
          icon: 'text-primary',
          title: 'text-text-primary font-medium',
          message: 'text-text-secondary',
          timestamp: 'text-text-tertiary'
        };
      default: 
        return {
          container: 'bg-background-primary border-background-tertiary text-text-primary shadow-sm',
          icon: 'text-text-secondary',
          title: 'text-text-primary',
          message: 'text-text-secondary',
          timestamp: 'text-text-tertiary'
        };
    }
  };

  if (visibleNotifications.length === 0) {
    return null;
  }

  return (
    <div className="fixed top-4 right-4 z-40 space-y-3 max-w-sm">
      {visibleNotifications.map((notification, index) => {
        // Use specialized SystemShockNotification for system shock events
        if (notification.type === 'system_shock') {
          return (
            <SystemShockNotification
              key={notification.id}
              title={notification.title}
              message={notification.message}
              shockType={notification.message.includes('corruption') ? 'MESSAGE_CORRUPTION' : 'ACTION_LOCK'}
              expiresAt={new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()} // 24 hours from now
              alertLevel={notification.priority}
              onDismiss={() => handleDismiss(notification.id)}
            />
          );
        }

        const styles = getPriorityStyles(notification.priority);
        
        return (
          <div 
            key={notification.id}
            className={`
              w-80 p-4 rounded-lg border-2 shadow-lg transform transition-all duration-300 ease-out
              animate-slide-in-right opacity-0 animation-delay-${index * 100}
              ${styles.container}
              ${SLIDE_IN_RIGHT}
            `}
            style={{
              animationDelay: `${index * 100}ms`,
              animationFillMode: 'forwards'
            }}
          >
            {/* Header */}
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className={`text-lg ${styles.icon}`}>
                  {getNotificationIcon(notification.type)}
                </span>
                <h3 className={`text-sm ${styles.title}`}>
                  {notification.title}
                </h3>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleDismiss(notification.id)}
                className={`
                  p-1 h-6 w-6 hover:bg-white/10 rounded text-sm font-bold
                  ${styles.icon}
                `}
                aria-label="Dismiss notification"
              >
                ×
              </Button>
            </div>

            {/* Message */}
            <p className={`text-xs leading-relaxed mb-3 ${styles.message}`}>
              {notification.message}
            </p>

            {/* Timestamp */}
            <div className={`text-xs font-mono ${styles.timestamp}`}>
              {new Date(notification.timestamp).toLocaleTimeString()}
            </div>
          </div>
        );
      })}
    </div>
  );
}