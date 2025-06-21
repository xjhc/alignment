import { useState, useEffect } from 'react';
import { PrivateNotification } from '../types';

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

  const getPriorityClass = (priority: string) => {
    switch (priority) {
      case 'high': return 'notification-high';
      case 'medium': return 'notification-medium';
      case 'low': return 'notification-low';
      default: return 'notification-medium';
    }
  };

  if (visibleNotifications.length === 0) {
    return null;
  }

  return (
    <div className="private-notifications">
      {visibleNotifications.map(notification => (
        <div 
          key={notification.id} 
          className={`notification-toast ${getPriorityClass(notification.priority)}`}
        >
          <div className="notification-header">
            <span className="notification-icon">
              {getNotificationIcon(notification.type)}
            </span>
            <span className="notification-title">{notification.title}</span>
            <button 
              className="notification-close"
              onClick={() => handleDismiss(notification.id)}
              aria-label="Dismiss notification"
            >
              ×
            </button>
          </div>
          <div className="notification-message">
            {notification.message}
          </div>
          <div className="notification-timestamp">
            {new Date(notification.timestamp).toLocaleTimeString()}
          </div>
        </div>
      ))}
    </div>
  );
}