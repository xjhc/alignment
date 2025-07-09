import React, { useEffect } from 'react';
import { SubtleIndicator } from '../contexts/NotificationContext';

interface SubtleIndicatorManagerProps {
  indicators: SubtleIndicator[];
}

export const SubtleIndicatorManager: React.FC<SubtleIndicatorManagerProps> = ({
  indicators,
}) => {
  // Update browser tab title for unread messages
  useEffect(() => {
    const unreadIndicator = indicators.find(
      indicator => indicator.element === 'browser_tab' && indicator.state === 'new_message'
    );
    
    if (unreadIndicator && unreadIndicator.value) {
      document.title = `(${unreadIndicator.value}) Alignment - Corporate Crisis`;
    } else {
      document.title = 'Alignment - Corporate Crisis';
    }
    
    // Clean up on unmount
    return () => {
      document.title = 'Alignment - Corporate Crisis';
    };
  }, [indicators]);
  
  // This component doesn't render anything directly
  // It provides utility functions for other components to use
  return null;
};

// Utility hooks for other components to use with subtle indicators
export const useSubtleIndicators = () => {
  const getChannelIndicator = (channelId: string, indicators: SubtleIndicator[]) => {
    return indicators.find(
      indicator => 
        indicator.element === 'channel' && 
        indicator.id === `channel-${channelId}`
    );
  };
  
  const getVoteButtonIndicator = (indicators: SubtleIndicator[]) => {
    return indicators.find(
      indicator => 
        indicator.element === 'vote_button' && 
        indicator.state === 'voted'
    );
  };
  
  const getTypingIndicator = (indicators: SubtleIndicator[]) => {
    return indicators.find(
      indicator => 
        indicator.element === 'typing' && 
        indicator.state === 'typing'
    );
  };
  
  return {
    getChannelIndicator,
    getVoteButtonIndicator,
    getTypingIndicator,
  };
};