import React, { useEffect, useCallback } from 'react';
import { useNotifications, createToastNotification, createLoebmateMessage } from '../contexts/NotificationContext';
import { useGameContext } from '../contexts/GameContext';
import { useWebSocketEvent } from './useWebSocket';

/**
 * Bridge between game events and the notification system
 * Translates game events into appropriate notifications across all tiers
 */
export const useNotificationBridge = () => {
  const { showToast, showLoebmateMessage, setSubtleIndicator, clearSubtleIndicator } = useNotifications();
  const { gameState, localPlayer } = useGameContext();
  
  // Track previous game state for change detection
  const previousGameState = React.useRef(gameState);
  
  // Handle WebSocket private notifications
  useWebSocketEvent('PRIVATE_NOTIFICATION', useCallback((payload: any) => {
    switch (payload.type) {
      case 'SYSTEM_SHOCK':
        showToast(createToastNotification.conversion(payload.message));
        break;
        
      case 'ACTION_BLOCKED':
        showToast(createToastNotification.actionBlocked(payload.message));
        break;
        
      case 'ROLE_ABILITY_UNLOCKED':
        showToast(createToastNotification.abilityUnlocked(payload.abilityName || 'Unknown Ability'));
        break;
        
      case 'KPI_COMPLETE':
        showToast(createToastNotification.kpiComplete(payload.objective || 'Unknown Objective'));
        break;
        
      default:
        // Generic notification for other types
        showToast({
          type: 'toast',
          category: 'generic',
          icon: '📢',
          color: 'info',
          title: payload.title || 'Notification',
          message: payload.message,
          priority: payload.priority || 'medium',
        });
    }
  }, [showToast]));

  // Handle existing private notifications from game state
  useEffect(() => {
    if (!gameState?.privateNotifications) return;

    // Convert existing private notifications to unified toast notifications
    gameState.privateNotifications.forEach((notification: any) => {
      if (notification.isRead) return; // Skip already read notifications

      switch (notification.type) {
        case 'system_shock':
          showToast(createToastNotification.conversion(notification.message));
          break;
          
        case 'kpi_progress':
          showToast({
            type: 'toast',
            category: 'kpi_complete',
            icon: '📊',
            color: 'info',
            title: 'KPI Progress',
            message: notification.message,
            priority: 'medium',
          });
          break;
          
        case 'role_ability':
          showToast(createToastNotification.abilityUnlocked(notification.title.replace('Ability Unlocked: ', '')));
          break;
          
        case 'loebmate_hint':
          showLoebmateMessage({
            type: 'loebmate',
            trigger: 'night_phase', // Default trigger
            content: notification.message,
            isPrivate: true,
          });
          break;
          
        case 'help_response':
          showLoebmateMessage({
            type: 'loebmate',
            trigger: 'first_tokens', // Default trigger
            content: notification.message,
            isPrivate: true,
          });
          break;
          
        default:
          showToast({
            type: 'toast',
            category: 'generic',
            icon: '📢',
            color: 'info',
            title: notification.title || 'Notification',
            message: notification.message,
            priority: notification.priority || 'medium',
          });
      }
    });
  }, [gameState?.privateNotifications, showToast, showLoebmateMessage]);
  
  // Handle game phase changes for Loebmate messages
  useEffect(() => {
    if (!gameState || !localPlayer) return;
    
    const currentPhase = gameState.phase?.type;
    const previousPhase = previousGameState.current?.phase?.type;
    
    if (currentPhase !== previousPhase) {
      switch (currentPhase) {
        case 'NOMINATION':
          showLoebmateMessage(createLoebmateMessage.nominationPhase());
          break;
          
        case 'NIGHT':
          showLoebmateMessage(createLoebmateMessage.nightPhase());
          break;
      }
    }
    
    previousGameState.current = gameState;
  }, [gameState?.phase?.type, showLoebmateMessage, localPlayer]);
  
  // Handle token changes for first tokens Loebmate message
  useEffect(() => {
    if (!gameState || !localPlayer) return;
    
    const currentTokens = localPlayer.tokens || 0;
    const previousTokens = previousGameState.current?.players?.find(p => p.id === localPlayer.id)?.tokens || 0;
    
    if (currentTokens > previousTokens && previousTokens === 0) {
      // First time receiving tokens
      showLoebmateMessage(createLoebmateMessage.firstTokens());
    }
  }, [localPlayer?.tokens, showLoebmateMessage, gameState, localPlayer]);
  
  // Handle subtle indicators for unread messages
  useEffect(() => {
    if (!gameState?.messages) return;
    
    // Count unread messages in each channel
    const channelUnreadCounts = gameState.messages.reduce((acc: { [key: string]: number }, message: any) => {
      if (message.channel && message.isUnread) {
        acc[message.channel] = (acc[message.channel] || 0) + 1;
      }
      return acc;
    }, {});
    
    // Update channel indicators
    Object.entries(channelUnreadCounts).forEach(([channelId, count]) => {
      if (count > 0) {
        setSubtleIndicator({
          id: `channel-${channelId}`,
          type: 'subtle',
          element: 'channel',
          state: 'unread',
          value: count,
          tooltip: `${count} unread message${count !== 1 ? 's' : ''}`,
        });
      } else {
        clearSubtleIndicator(`channel-${channelId}`);
      }
    });
    
    // Update browser tab indicator
    const totalUnread = Object.values(channelUnreadCounts).reduce((sum: number, count: number) => sum + count, 0);
    if (totalUnread > 0) {
      setSubtleIndicator({
        id: 'browser-tab-unread',
        type: 'subtle',
        element: 'browser_tab',
        state: 'new_message',
        value: totalUnread,
      });
    } else {
      clearSubtleIndicator('browser-tab-unread');
    }
  }, [gameState?.messages, setSubtleIndicator, clearSubtleIndicator]);
  
  // Handle vote button indicator
  useEffect(() => {
    if (!gameState || !localPlayer) return;
    
    const hasVoted = gameState.voteState?.votes && localPlayer.id in gameState.voteState.votes;
    const votedFor = hasVoted ? gameState.voteState.votes[localPlayer.id] : null;
    
    if (hasVoted && votedFor) {
      setSubtleIndicator({
        id: 'vote-button-voted',
        type: 'subtle',
        element: 'vote_button',
        state: 'voted',
        value: votedFor,
        tooltip: `You have voted for ${votedFor}`,
      });
    } else {
      clearSubtleIndicator('vote-button-voted');
    }
  }, [gameState?.voteState, localPlayer, setSubtleIndicator, clearSubtleIndicator]);
  
  // Handle typing indicator
  useEffect(() => {
    // This would be connected to the typing indicator system
    // For now, we'll just clear it as it's handled by the existing TypingIndicator component
    clearSubtleIndicator('typing-indicator');
  }, [clearSubtleIndicator]);
  
  return {
    // Expose manual trigger functions for specific scenarios
    triggerActionBlocked: (message: string) => {
      showToast(createToastNotification.actionBlocked(message));
    },
    
    triggerAbilityUnlocked: (abilityName: string) => {
      showToast(createToastNotification.abilityUnlocked(abilityName));
    },
    
    triggerKpiComplete: (objective: string) => {
      showToast(createToastNotification.kpiComplete(objective));
    },
    
    triggerLoebmateHelp: (trigger: any, content: string) => {
      showLoebmateMessage({
        type: 'loebmate',
        trigger,
        content,
        isPrivate: true,
      });
    },
  };
};