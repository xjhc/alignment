import { useState, useEffect, useCallback, useRef } from "react";
import { useWebSocketEvent } from "./useWebSocket";
import { Player } from "../types";

export interface PendingMessage {
  id: string;
  message: string;
  timestamp: number;
  status: "pending" | "sent" | "failed";
}

export function useChatBuffer(localPlayer: Player | null) {
  const [messageBuffer, setMessageBuffer] = useState<string[]>([]);
  const [pendingMessages, setPendingMessages] = useState<PendingMessage[]>([]);
  const [rateLimitError, setRateLimitError] = useState<string | null>(null);
  const bufferTimer = useRef<NodeJS.Timeout | null>(null);
  const errorTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const FLUSH_INTERVAL = 500; // Send batch every 0.5 seconds
  const MAX_BUFFER_SIZE = 5;

  // Listen for rate limit exceeded events from server
  useWebSocketEvent("RATE_LIMIT_EXCEEDED", (payload: { message: string }) => {
    setRateLimitError(payload.message);
    if (errorTimeoutRef.current) clearTimeout(errorTimeoutRef.current);
    errorTimeoutRef.current = setTimeout(() => setRateLimitError(null), 5000);
  });

  // Listen for successful message broadcasts to update pending status
  useWebSocketEvent(
    "CHAT_MESSAGE",
    (payload: {
      id?: string;
      sender_id?: string;
      message?: string;
      player_id?: string;
    }) => {
      if (!localPlayer) return;

      // Check if the message is from the current user
      if ((payload.sender_id || payload.player_id) === localPlayer.id) {
        setPendingMessages((prev) =>
          // Remove the first pending message that matches the content.
          // This handles the optimistic UI confirmation.
          prev.filter((msg, index) => {
            const messageMatches = msg.message === payload.message;
            // Only remove the first match to prevent accidentally removing multiple identical pending messages.
            if (
              messageMatches &&
              index === prev.findIndex((p) => p.message === payload.message)
            ) {
              return false;
            }
            return true;
          })
        );
      }
    }
  );

  const flushBuffer = useCallback(() => {
    if (messageBuffer.length === 0) return;

    // Dispatch the custom event to be handled by useGameActions
    window.dispatchEvent(
      new CustomEvent("flushChatBuffer", {
        detail: { messages: [...messageBuffer] },
      })
    );

    setMessageBuffer([]);
    if (bufferTimer.current) {
      clearTimeout(bufferTimer.current);
      bufferTimer.current = null;
    }
  }, [messageBuffer]);

  useEffect(() => {
    if (messageBuffer.length > 0 && !bufferTimer.current) {
      bufferTimer.current = setTimeout(flushBuffer, FLUSH_INTERVAL);
    }

    return () => {
      if (bufferTimer.current) clearTimeout(bufferTimer.current);
      if (errorTimeoutRef.current) clearTimeout(errorTimeoutRef.current);
    };
  }, [messageBuffer.length, flushBuffer]);

  const addMessageToBuffer = useCallback(
    (message: string): string => {
      const newPendingMessage: PendingMessage = {
        id: `pending_${Date.now()}_${Math.random()}`,
        message,
        timestamp: Date.now(),
        status: "pending",
      };

      setPendingMessages((prev) => [...prev, newPendingMessage]);

      setMessageBuffer((prev) => {
        const newBuffer = [...prev, message];
        if (newBuffer.length >= MAX_BUFFER_SIZE) {
          setTimeout(flushBuffer, 0);
        }
        return newBuffer;
      });

      return newPendingMessage.id;
    },
    [flushBuffer]
  );

  const clearBuffer = useCallback(() => {
    setMessageBuffer([]);
    setPendingMessages([]);
    if (bufferTimer.current) {
      clearTimeout(bufferTimer.current);
      bufferTimer.current = null;
    }
  }, []);

  const getBufferStatus = useCallback(() => {
    return {
      bufferLength: messageBuffer.length,
      hasPendingMessages: pendingMessages.length > 0,
      nextFlushETA:
        messageBuffer.length > 0 ? Math.ceil(FLUSH_INTERVAL / 1000) : 0,
    };
  }, [messageBuffer.length, pendingMessages.length]);

  return {
    addMessageToBuffer,
    flushBuffer,
    clearBuffer,
    pendingMessages,
    rateLimitError,
    getBufferStatus,
    bufferLength: messageBuffer.length,
  };
}
