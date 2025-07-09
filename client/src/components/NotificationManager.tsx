import React, { useEffect } from 'react';
import { useNotifications } from '../contexts/NotificationContext';
import { useSound } from '../hooks/useSound';
import { ToastNotificationDisplay } from './ToastNotificationDisplay';
import { LoebmateWhisper } from './LoebmateWhisper';
import { SubtleIndicatorManager } from './SubtleIndicatorManager';

/**
 * Unified notification manager that handles all notification tiers
 * - Tier 1: Toast notifications (high priority, ephemeral)
 * - Tier 2: In-line system messages (handled by chat components)
 * - Tier 3: Loebmate whispers (contextual help)
 * - Tier 4: Subtle indicators (ambient feedback)
 */
export const NotificationManager: React.FC = () => {
  const { state } = useNotifications();
  const { playSound } = useSound();
  
  // Play sounds for new notifications
  useEffect(() => {
    if (state.settings.soundEnabled) {
      const latestToast = state.toasts[state.toasts.length - 1];
      if (latestToast) {
        // Different sounds for different priority levels
        switch (latestToast.priority) {
          case 'high':
            playSound('alert');
            break;
          case 'medium':
            playSound('notification');
            break;
          case 'low':
            playSound('soft-notification');
            break;
        }
      }
    }
  }, [state.toasts.length, state.settings.soundEnabled, playSound]);
  
  // Play sound for Loebmate messages
  useEffect(() => {
    if (state.settings.soundEnabled && state.loebmateMessages.length > 0) {
      const latestMessage = state.loebmateMessages[state.loebmateMessages.length - 1];
      if (latestMessage) {
        playSound('loebmate-whisper');
      }
    }
  }, [state.loebmateMessages.length, state.settings.soundEnabled, playSound]);
  
  return (
    <>
      {/* Tier 1: Toast Notifications */}
      <div className="fixed top-4 right-4 z-50 space-y-3 max-w-sm pointer-events-none">
        {state.toasts.map((toast, index) => (
          <div
            key={toast.id}
            className="pointer-events-auto"
            style={{
              transform: `translateY(${index * 10}px)`,
              zIndex: 50 - index,
            }}
          >
            <ToastNotificationDisplay
              notification={toast}
              index={index}
            />
          </div>
        ))}
      </div>
      
      {/* Tier 3: Loebmate Whispers */}
      <div className="fixed bottom-4 right-4 z-40 space-y-2 max-w-md pointer-events-none">
        {state.loebmateMessages
          .filter(msg => msg.isPrivate)
          .slice(-2) // Only show last 2 messages
          .map((message, index) => (
            <div
              key={message.id}
              className="pointer-events-auto"
              style={{
                transform: `translateY(${index * 10}px)`,
                zIndex: 40 - index,
              }}
            >
              <LoebmateWhisper
                message={message}
                index={index}
              />
            </div>
          ))}
      </div>
      
      {/* Tier 4: Subtle Indicators */}
      <SubtleIndicatorManager indicators={state.subtleIndicators} />
    </>
  );
};