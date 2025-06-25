import { useState, useEffect } from 'react';
import { Button } from './ui/Button';
import { SLIDE_IN_UP } from '../utils/animations';

interface SystemShockNotificationProps {
  title: string;
  message: string;
  shockType: string;
  expiresAt: string;
  alertLevel: 'low' | 'medium' | 'high';
  onDismiss: () => void;
}

export function SystemShockNotification({ 
  title, 
  message, 
  shockType, 
  expiresAt, 
  alertLevel,
  onDismiss 
}: SystemShockNotificationProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState('');

  useEffect(() => {
    // Animate in
    setIsVisible(true);

    // Auto-dismiss for low/medium priority after 10 seconds
    if (alertLevel !== 'high') {
      const timer = setTimeout(() => {
        handleDismiss();
      }, 10000);
      return () => clearTimeout(timer);
    }
  }, [alertLevel]);

  useEffect(() => {
    // Update countdown timer
    const updateTimer = () => {
      const now = new Date().getTime();
      const expiry = new Date(expiresAt).getTime();
      const diff = expiry - now;

      if (diff <= 0) {
        setTimeRemaining('Expired');
        return;
      }

      const hours = Math.floor(diff / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
      setTimeRemaining(`${hours}h ${minutes}m remaining`);
    };

    updateTimer();
    const interval = setInterval(updateTimer, 60000); // Update every minute

    return () => clearInterval(interval);
  }, [expiresAt]);

  const handleDismiss = () => {
    setIsVisible(false);
    // Small delay to allow animation before removal
    setTimeout(onDismiss, 200);
  };

  const getAlertStyles = () => {
    switch (alertLevel) {
      case 'high':
        return {
          container: 'bg-danger border-danger',
          icon: 'text-white',
          title: 'text-white',
          message: 'text-white/90',
          timer: 'text-white/70'
        };
      case 'medium':
        return {
          container: 'bg-background-secondary border-primary',
          icon: 'text-primary',
          title: 'text-text-primary',
          message: 'text-text-secondary',
          timer: 'text-text-tertiary'
        };
      default:
        return {
          container: 'bg-background-primary border-background-tertiary',
          icon: 'text-text-secondary',
          title: 'text-text-primary',
          message: 'text-text-secondary',
          timer: 'text-text-tertiary'
        };
    }
  };

  const getShockTypeIcon = () => {
    switch (shockType) {
      case 'MESSAGE_CORRUPTION': return '📡';
      case 'ACTION_LOCK': return '🔒';
      case 'FORCED_SILENCE': return '🤐';
      default: return '⚡';
    }
  };

  const styles = getAlertStyles();

  return (
    <div 
      className={`
        fixed top-4 right-4 z-50 w-80 p-4 rounded-lg border-2 shadow-lg
        transform transition-all duration-300 ease-out
        ${isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'}
        ${styles.container}
        ${SLIDE_IN_UP}
      `}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className={`text-lg ${styles.icon}`}>
            {getShockTypeIcon()}
          </span>
          <h3 className={`font-semibold text-sm ${styles.title}`}>
            {title}
          </h3>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleDismiss}
          className={`
            p-1 h-6 w-6 hover:bg-white/10 rounded
            ${styles.icon}
          `}
          aria-label="Dismiss notification"
        >
          ×
        </Button>
      </div>

      {/* Message */}
      <p className={`text-xs leading-relaxed mb-3 ${styles.message}`}>
        {message}
      </p>

      {/* Timer */}
      <div className={`text-xs font-mono ${styles.timer}`}>
        {timeRemaining}
      </div>

      {/* Shock type indicator */}
      <div className={`text-xs mt-2 opacity-70 ${styles.timer}`}>
        Effect: {shockType.replace('_', ' ').toLowerCase()}
      </div>
    </div>
  );
}