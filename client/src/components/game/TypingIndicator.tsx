import React from 'react';
import { FADE_IN, PULSE } from '../../utils/animations';

interface TypingIndicatorProps {
  typingUsers: string[];
}

export const TypingIndicator: React.FC<TypingIndicatorProps> = ({ typingUsers }) => {
  if (typingUsers.length === 0) {
    return null;
  }

  const formatTypingText = (users: string[]) => {
    if (users.length === 1) {
      return `${users[0]} is typing...`;
    } else if (users.length === 2) {
      return `${users[0]} and ${users[1]} are typing...`;
    } else if (users.length === 3) {
      return `${users[0]}, ${users[1]}, and ${users[2]} are typing...`;
    } else {
      return `${users[0]}, ${users[1]}, and ${users.length - 2} others are typing...`;
    }
  };

  return (
    <div className="px-4 py-2 text-xs text-text-muted flex items-center gap-2">
      <div className="flex gap-1">
        <div className={`w-1 h-1 bg-text-muted rounded-full ${PULSE}`} style={{ animationDelay: '0ms' }} />
        <div className={`w-1 h-1 bg-text-muted rounded-full ${PULSE}`} style={{ animationDelay: '150ms' }} />
        <div className={`w-1 h-1 bg-text-muted rounded-full ${PULSE}`} style={{ animationDelay: '300ms' }} />
      </div>
      <span>{formatTypingText(typingUsers)}</span>
    </div>
  );
};