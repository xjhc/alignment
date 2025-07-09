import React, { useState, useEffect } from 'react';
import { LoebmateNotification, useNotifications } from '../contexts/NotificationContext';
import { Button } from './ui/Button';
import { MarkdownRenderer } from './game/MarkdownRenderer';
import { SLIDE_IN_UP } from '../utils/animations';

interface LoebmateWhisperProps {
  message: LoebmateNotification;
  index: number;
}

export const LoebmateWhisper: React.FC<LoebmateWhisperProps> = ({
  message,
  index,
}) => {
  const { dismissLoebmateMessage } = useNotifications();
  const [isVisible, setIsVisible] = useState(false);
  
  useEffect(() => {
    // Animate in with stagger
    const timer = setTimeout(() => {
      setIsVisible(true);
    }, index * 200);
    
    return () => clearTimeout(timer);
  }, [index]);
  
  useEffect(() => {
    // Auto-dismiss after 12 seconds
    const timer = setTimeout(() => {
      handleDismiss();
    }, 12000);
    
    return () => clearTimeout(timer);
  }, []);
  
  const handleDismiss = () => {
    setIsVisible(false);
    setTimeout(() => {
      dismissLoebmateMessage(message.id);
    }, 300);
  };
  
  const getTriggerIcon = (trigger: LoebmateNotification['trigger']) => {
    switch (trigger) {
      case 'nomination_phase':
        return '🗳️';
      case 'night_phase':
        return '🌙';
      case 'first_tokens':
        return '🪙';
      case 'role_unlocked':
        return '✨';
      case 'system_shock':
        return '⚡';
      default:
        return '💡';
    }
  };
  
  return (
    <div
      className={`
        w-96 p-4 rounded-lg border-2 shadow-lg bg-gradient-to-br from-blue-500/10 to-purple-500/10
        border-blue-400/50 backdrop-blur-sm
        transform transition-all duration-300 ease-out
        ${isVisible ? 'translate-y-0 opacity-100' : 'translate-y-full opacity-0'}
        ${SLIDE_IN_UP}
      `}
      style={{
        animationDelay: `${index * 200}ms`,
        animationFillMode: 'forwards',
      }}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-full bg-blue-500/20 flex items-center justify-center text-sm">
            🤖
          </div>
          <div className="flex items-center gap-2">
            <span className="text-blue-400 font-bold text-sm">
              Loebmate
            </span>
            <span className="text-xs bg-blue-500/20 text-blue-400 px-2 py-1 rounded font-bold tracking-wider">
              ASSISTANT
            </span>
            {message.isPrivate && (
              <span className="text-xs text-text-muted" title="Private message">
                🔒
              </span>
            )}
          </div>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleDismiss}
          className="p-1 h-6 w-6 rounded font-bold text-sm text-text-muted hover:text-text-primary hover:bg-background-tertiary"
          aria-label="Dismiss Loebmate message"
        >
          ×
        </Button>
      </div>
      
      {/* Trigger icon and category */}
      <div className="flex items-center gap-2 mb-3">
        <span className="text-lg">
          {getTriggerIcon(message.trigger)}
        </span>
        <span className="text-xs text-blue-400 font-medium uppercase tracking-wide">
          {message.trigger.replace('_', ' ')}
        </span>
      </div>
      
      {/* Message content */}
      <div className="text-text-primary text-sm leading-relaxed mb-3 bg-background-secondary/30 border border-border/20 rounded-lg p-3">
        <MarkdownRenderer
          content={message.content}
          localPlayerName="" // Not relevant for Loebmate messages
        />
      </div>
      
      {/* Timestamp */}
      <div className="text-xs font-mono text-text-tertiary opacity-70">
        {new Date(message.timestamp).toLocaleTimeString()}
      </div>
    </div>
  );
};