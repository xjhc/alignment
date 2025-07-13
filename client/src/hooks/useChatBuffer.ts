import { useState, useEffect, useCallback, useRef } from "react";
import { useWebSocketEvent } from "./useWebSocket";
import { Player, ClientAction, ClientActionType } from "../types";

export interface PendingMessage {
  id: string;
  clientMessageId: string;
  message: string;
  timestamp: number;
  status: "pending" | "sent" | "failed";
  batchId?: string; // Add batchId for tracking
  channelId: string;
  retryCount?: number;
  errorMessage?: string;
}

export function useChatBuffer(
  localPlayer: Player | null,
  gameId: string | undefined,
  sendAction: (action: ClientAction) => void
) {
  const [messageBuffer, setMessageBuffer] = useState<
    { message: string; clientMessageId: string; channelId: string }[]
  >([]);
  const [pendingMessages, setPendingMessages] = useState<
    Record<string, PendingMessage[]>
  >({});
  const [rateLimitError, setRateLimitError] = useState<string | null>(null);
  const isFlushing = useRef(false); // Add a lock to prevent re-entrant calls
  const bufferTimer = useRef<NodeJS.Timeout | null>(null);
  const errorTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pendingTimeouts = useRef<Map<string, NodeJS.Timeout>>(new Map());

  const FLUSH_INTERVAL = 500; // Send batch every 0.5 seconds
  const MAX_BUFFER_SIZE = 5;
  const MESSAGE_TIMEOUT = 3000; // 3 seconds timeout for message confirmation
  const MAX_RETRY_COUNT = 2;

  // Use a ref to hold the latest props to avoid stale closures in callbacks
  const latestProps = useRef({ localPlayer, gameId, sendAction });
  useEffect(() => {
    latestProps.current = { localPlayer, gameId, sendAction };
  });

  // Function to set message timeout
  const setMessageTimeout = useCallback((clientMessageId: string, channelId: string) => {
    const timeoutId = setTimeout(() => {
      setPendingMessages((prev) => {
        const channelPending = prev[channelId] || [];
        const updatedPending = channelPending.map((msg) => {
          if (msg.clientMessageId === clientMessageId && msg.status === "pending") {
            return {
              ...msg,
              status: "failed" as const,
              errorMessage: "Message failed to send (timeout)",
            };
          }
          return msg;
        });
        return { ...prev, [channelId]: updatedPending };
      });
      pendingTimeouts.current.delete(clientMessageId);
    }, MESSAGE_TIMEOUT);

    pendingTimeouts.current.set(clientMessageId, timeoutId);
  }, []);

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
      client_message_id?: string;
      sender_id?: string;
      channel_id?: string;
      message?: string;
    }) => {
      const { localPlayer } = latestProps.current;
      if (
        !localPlayer ||
        (payload.sender_id && payload.sender_id !== localPlayer.id)
      ) {
        return;
      }

      const channelId = payload.channel_id || "#war-room";
      const clientMessageId = payload.client_message_id;

      setPendingMessages((prev) => {
        const channelPending = prev[channelId] || [];
        const updatedPending = clientMessageId
          ? channelPending.filter((msg) => {
              if (msg.clientMessageId === clientMessageId) {
                // Clear timeout for confirmed message
                const timeoutId = pendingTimeouts.current.get(clientMessageId);
                if (timeoutId) {
                  clearTimeout(timeoutId);
                  pendingTimeouts.current.delete(clientMessageId);
                }
                return false; // Remove the message from pending
              }
              return true;
            })
          : channelPending.filter((msg) => msg.message !== payload.message); // Fallback

        return { ...prev, [channelId]: updatedPending };
      });
    }
  );

  const sendBufferedMessages = useCallback(() => {
    // Prevent re-entrant calls which can happen with React StrictMode
    if (isFlushing.current) {
      return;
    }
    const batchId = `${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
    isFlushing.current = true;
    if (bufferTimer.current) {
      clearTimeout(bufferTimer.current);
      bufferTimer.current = null;
    }

    setMessageBuffer((currentBuffer) => {
      if (currentBuffer.length === 0) {
        isFlushing.current = false;
        return currentBuffer;
      }

      const { localPlayer, gameId, sendAction } = latestProps.current;
      if (!localPlayer || !gameId) {
        console.warn(
          "Cannot flush chat buffer: missing player or gameId. Re-queuing messages."
        );
        return currentBuffer;
      }

      const messagesByChannel = currentBuffer.reduce(
        (acc, msg) => {
          if (!acc[msg.channelId]) {
            acc[msg.channelId] = [];
          }
          acc[msg.channelId].push({
            message: msg.message,
            client_message_id: msg.clientMessageId,
          });
          return acc;
        },
        {} as Record<string, { message: string; client_message_id: string }[]>
      );

      Object.entries(messagesByChannel).forEach(
        ([channelId, channelMessages]) => {
          // Set timeouts for each message in the batch
          channelMessages.forEach((msg) => {
            setMessageTimeout(msg.client_message_id, channelId);
          });

          sendAction({
            type: ClientActionType.SendMessage,
            payload: {
              game_id: gameId,
              player_id: localPlayer.id,
              messages: channelMessages,
              player_name: localPlayer.name,
              channel_id: channelId,
              batch_id: batchId,
            },
          });
        }
      );
      isFlushing.current = false;
      return [];
    });
  }, [setMessageTimeout]);

  useEffect(() => {
    if (messageBuffer.length >= MAX_BUFFER_SIZE) {
      sendBufferedMessages();
    } else if (messageBuffer.length > 0 && !bufferTimer.current) {
      bufferTimer.current = setTimeout(sendBufferedMessages, FLUSH_INTERVAL);
    }

    return () => {
      if (bufferTimer.current) {
        clearTimeout(bufferTimer.current);
        bufferTimer.current = null;
      }
    };
  }, [messageBuffer, sendBufferedMessages]);

  const addMessageToBuffer = useCallback(
    (message: string, channelId: string = "#war-room"): string => {
      const clientMessageId = `${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
      const newPendingMessage: PendingMessage = {
        id: `pending_${clientMessageId}`,
        clientMessageId,
        message,
        timestamp: Date.now(),
        status: "pending",
        channelId,
      };

      // Add to pending messages by channel
      setPendingMessages((prev) => {
        const channelPending = prev[channelId] || [];
        return {
          ...prev,
          [channelId]: [...channelPending, newPendingMessage],
        };
      });

      // Add to buffer (let useEffect handle flushing)
      setMessageBuffer((prev) => [
        ...prev,
        { message, clientMessageId, channelId },
      ]);

      return clientMessageId;
    },
    []
  );

  const retryMessage = useCallback((clientMessageId: string, channelId: string) => {
    setPendingMessages((prev) => {
      const channelPending = prev[channelId] || [];
      const messageToRetry = channelPending.find(
        (msg) => msg.clientMessageId === clientMessageId && msg.status === "failed"
      );

      if (!messageToRetry || (messageToRetry.retryCount || 0) >= MAX_RETRY_COUNT) {
        return prev;
      }

      // Generate new client message ID for retry
      const newClientMessageId = `${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
      
      // Update pending message with new ID and status
      const updatedPending = channelPending.map((msg) => {
        if (msg.clientMessageId === clientMessageId) {
          return {
            ...msg,
            clientMessageId: newClientMessageId,
            status: "pending" as const,
            retryCount: (msg.retryCount || 0) + 1,
            errorMessage: undefined,
            timestamp: Date.now(),
          };
        }
        return msg;
      });

      // Add to buffer for sending
      setMessageBuffer((buffer) => [
        ...buffer,
        {
          message: messageToRetry.message,
          clientMessageId: newClientMessageId,
          channelId,
        },
      ]);

      return { ...prev, [channelId]: updatedPending };
    });
  }, []);

  const clearBuffer = useCallback(() => {
    setMessageBuffer([]);
    setPendingMessages({});
    // Clear all pending timeouts
    pendingTimeouts.current.forEach((timeoutId) => clearTimeout(timeoutId));
    pendingTimeouts.current.clear();
    if (bufferTimer.current) {
      clearTimeout(bufferTimer.current);
      bufferTimer.current = null;
    }
  }, []);

  const getBufferStatus = useCallback(
    () => ({
      bufferLength: messageBuffer.length,
      hasPendingMessages: Object.values(pendingMessages).some(
        (p) => p.length > 0
      ),
      nextFlushETA:
        messageBuffer.length > 0 ? Math.ceil(FLUSH_INTERVAL / 1000) : 0,
    }),
    [messageBuffer.length, pendingMessages]
  );

  const getPendingMessagesForChannel = useCallback(
    (channelId: string = "#war-room"): PendingMessage[] =>
      pendingMessages[channelId] || [],
    [pendingMessages]
  );

  // Cleanup timeouts on unmount
  useEffect(() => {
    return () => {
      pendingTimeouts.current.forEach((timeoutId) => clearTimeout(timeoutId));
      pendingTimeouts.current.clear();
      if (bufferTimer.current) {
        clearTimeout(bufferTimer.current);
      }
      if (errorTimeoutRef.current) {
        clearTimeout(errorTimeoutRef.current);
      }
    };
  }, []);

  return {
    addMessageToBuffer,
    clearBuffer,
    retryMessage,
    pendingMessages,
    getPendingMessagesForChannel,
    rateLimitError,
    getBufferStatus,
    bufferLength: messageBuffer.length,
  };
}
