import React, { useState, useEffect } from 'react';
import { ToastNotification, useNotifications } from '../contexts/NotificationContext';
import { Button } from './ui/Button';
import { SLIDE_IN_RIGHT } from '../utils/animations';

interface ToastNotificationDisplayProps {
  notification: ToastNotification;
  index: number;
}

export const ToastNotificationDisplay: React.FC<ToastNotificationDisplayProps> = ({
  notification,
  index,
}) => {
  const { dismissToast } = useNotifications();
  const [isVisible, setIsVisible] = useState(false);
  
  useEffect(() => {
    // Animate in with stagger
    const timer = setTimeout(() => {
      setIsVisible(true);
    }, index * 100);
    
    return () => clearTimeout(timer);
  }, [index]);
  
  const handleDismiss = () => {
    setIsVisible(false);
    setTimeout(() => {
      dismissToast(notification.id);
    }, 200);
  };
  
  const getColorStyles = (color: ToastNotification['color']) => {
    switch (color) {
      case 'danger':
        return {
          container: 'bg-red-500 border-red-600 text-white shadow-red-500/30',
          icon: 'text-white',
          title: 'text-white',
          message: 'text-white/90',
          button: 'text-white hover:bg-white/10',
        };
      case 'magenta':
        return {
          container: 'bg-magenta-500 border-magenta-600 text-white shadow-magenta-500/30',
          icon: 'text-white',
          title: 'text-white',
          message: 'text-white/90',
          button: 'text-white hover:bg-white/10',
        };
      case 'info':
        return {
          container: 'bg-blue-500 border-blue-600 text-white shadow-blue-500/30',
          icon: 'text-white',
          title: 'text-white',
          message: 'text-white/90',
          button: 'text-white hover:bg-white/10',
        };
      case 'success':
        return {
          container: 'bg-green-500 border-green-600 text-white shadow-green-500/30',
          icon: 'text-white',
          title: 'text-white',
          message: 'text-white/90',
          button: 'text-white hover:bg-white/10',
        };
      case 'warning':
        return {
          container: 'bg-yellow-500 border-yellow-600 text-black shadow-yellow-500/30',
          icon: 'text-black',
          title: 'text-black',
          message: 'text-black/80',
          button: 'text-black hover:bg-black/10',
        };
      default:
        return {
          container: 'bg-background-secondary border-border text-text-primary shadow-lg',
          icon: 'text-text-primary',
          title: 'text-text-primary',
          message: 'text-text-secondary',
          button: 'text-text-primary hover:bg-background-tertiary',
        };
    }
  };
  
  const getPriorityIndicator = (priority: ToastNotification['priority']) => {
    switch (priority) {
      case 'high':
        return '🔥';
      case 'medium':
        return '⚡';
      case 'low':
        return '💡';
      default:
        return '';
    }
  };
  
  const styles = getColorStyles(notification.color);
  
  return (
    <div
      className={`
        w-80 p-4 rounded-lg border-2 shadow-lg
        transform transition-all duration-300 ease-out
        ${isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'}
        ${styles.container}
        ${SLIDE_IN_RIGHT}
      `}
      style={{
        animationDelay: `${index * 100}ms`,
        animationFillMode: 'forwards',
      }}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className={`text-lg ${styles.icon}`}>
            {notification.icon}
          </span>
          <div className="flex items-center gap-1">
            <h3 className={`font-bold text-sm ${styles.title}`}>
              {notification.title}
            </h3>
            {notification.priority === 'high' && (
              <span className={`text-xs ${styles.icon}`}>
                {getPriorityIndicator(notification.priority)}
              </span>
            )}
          </div>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleDismiss}
          className={`
            p-1 h-6 w-6 rounded font-bold text-sm
            ${styles.button}
          `}
          aria-label="Dismiss notification"
        >
          ×
        </Button>
      </div>
      
      {/* Message */}
      <p className={`text-sm leading-relaxed mb-3 ${styles.message}`}>
        {notification.message}
      </p>
      
      {/* Timestamp */}
      <div className={`text-xs font-mono opacity-70 ${styles.message}`}>
        {new Date(notification.timestamp).toLocaleTimeString()}
      </div>
      
      {/* Category indicator */}
      {notification.category !== 'generic' && (
        <div className={`text-xs mt-2 opacity-60 ${styles.message}`}>
          {notification.category.replace('_', ' ').toUpperCase()}
        </div>
      )}
    </div>
  );
};