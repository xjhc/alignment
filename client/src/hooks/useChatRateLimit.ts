import { useState, useEffect, useCallback, useRef } from 'react';
import { useWebSocketEvent } from './useWebSocket';

export interface PendingMessage {
  id: string;
  message: string;
  timestamp: number;
  status: 'pending' | 'sent' | 'failed';
}

export function useChatRateLimit() {
  const [messageQueue, setMessageQueue] = useState<PendingMessage[]>([]);
  const [pendingMessages, setPendingMessages] = useState<PendingMessage[]>([]);
  const [rateLimitError, setRateLimitError] = useState<string | null>(null);
  const sendIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const errorTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Listen for rate limit exceeded events from server
  useWebSocketEvent('RATE_LIMIT_EXCEEDED', (payload: { message: string }) => {
    setRateLimitError(payload.message);
    
    // Clear the error after 5 seconds
    if (errorTimeoutRef.current) {
      clearTimeout(errorTimeoutRef.current);
    }
    errorTimeoutRef.current = setTimeout(() => {
      setRateLimitError(null);
    }, 5000);
  });

  // Listen for successful message broadcasts to update pending status
  useWebSocketEvent('CHAT_MESSAGE', (payload: { id: string }) => {
    if (payload.id) {
      setPendingMessages(prev => 
        prev.map(msg => 
          msg.id === payload.id ? { ...msg, status: 'sent' } : msg
        )
      );
      
      // Remove sent messages after a delay
      setTimeout(() => {
        setPendingMessages(prev => prev.filter(msg => msg.id !== payload.id));
      }, 2000);
    }
  });

  // Process message queue - send one message every 2.1 seconds
  useEffect(() => {
    if (sendIntervalRef.current) {
      clearInterval(sendIntervalRef.current);
    }

    sendIntervalRef.current = setInterval(() => {
      setMessageQueue(prev => {
        if (prev.length === 0) return prev;

        const [nextMessage, ...rest] = prev;
        
        // Move message to pending state
        setPendingMessages(pendingPrev => [...pendingPrev, nextMessage]);
        
        // Send the message (this will be handled by the caller)
        window.dispatchEvent(new CustomEvent('sendQueuedMessage', { 
          detail: nextMessage 
        }));

        return rest;
      });
    }, 2100); // 2.1 seconds - slightly slower than server limit

    return () => {
      if (sendIntervalRef.current) {
        clearInterval(sendIntervalRef.current);
      }
      if (errorTimeoutRef.current) {
        clearTimeout(errorTimeoutRef.current);
      }
    };
  }, []);

  const queueMessage = useCallback((message: string): string => {
    const messageId = `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    const pendingMessage: PendingMessage = {
      id: messageId,
      message,
      timestamp: Date.now(),
      status: 'pending'
    };

    setMessageQueue(prev => [...prev, pendingMessage]);
    return messageId;
  }, []);

  const clearQueue = useCallback(() => {
    setMessageQueue([]);
    setPendingMessages([]);
  }, []);

  const getQueueStatus = useCallback(() => {
    return {
      queueLength: messageQueue.length,
      hasPendingMessages: pendingMessages.length > 0,
      nextMessageETA: messageQueue.length > 0 ? Math.ceil(messageQueue.length * 2.1) : 0
    };
  }, [messageQueue.length, pendingMessages.length]);

  return {
    queueMessage,
    clearQueue,
    pendingMessages,
    rateLimitError,
    getQueueStatus,
    queueLength: messageQueue.length
  };
}