import { useState, useCallback, useEffect, useRef } from 'react';

interface TypingState {
  [userId: string]: {
    name: string;
    timestamp: number;
  };
}

export function useTypingIndicator() {
  const [typingUsers, setTypingUsers] = useState<TypingState>({});
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Clean up expired typing indicators
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      setTypingUsers(prev => {
        const filtered = Object.fromEntries(
          Object.entries(prev).filter(([, state]) => 
            now - state.timestamp < 3000 // Remove after 3 seconds
          )
        );
        return Object.keys(filtered).length !== Object.keys(prev).length ? filtered : prev;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const startTyping = useCallback((userId: string, userName: string) => {
    setTypingUsers(prev => ({
      ...prev,
      [userId]: {
        name: userName,
        timestamp: Date.now()
      }
    }));
  }, []);

  const stopTyping = useCallback((userId: string) => {
    setTypingUsers(prev => {
      const { [userId]: removed, ...rest } = prev;
      return rest;
    });
  }, []);

  // For local typing detection
  const handleTypingStart = useCallback(() => {
    // In the future, this could send START_TYPING action to server
    // For now, we'll simulate local typing
    console.log('User started typing');
  }, []);

  const handleTypingStop = useCallback(() => {
    // Clear typing timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    // Set a timeout to stop typing after 1 second of inactivity
    typingTimeoutRef.current = setTimeout(() => {
      // In the future, this could send STOP_TYPING action to server
      console.log('User stopped typing');
    }, 1000);
  }, []);

  const getTypingUserNames = useCallback(() => {
    return Object.values(typingUsers)
      .sort((a, b) => a.timestamp - b.timestamp) // Sort by when they started typing
      .map(state => state.name);
  }, [typingUsers]);

  return {
    typingUsers: getTypingUserNames(),
    startTyping,
    stopTyping,
    handleTypingStart,
    handleTypingStop
  };
}