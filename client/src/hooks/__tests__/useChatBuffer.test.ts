import { renderHook, act } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { useChatBuffer } from '../useChatBuffer';

// Mock the useWebSocketEvent hook
vi.mock('../useWebSocket', () => ({
  useWebSocketEvent: vi.fn()
}));

describe('useChatBuffer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.clearAllTimers();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('should buffer messages and flush automatically after interval', () => {
    const { result } = renderHook(() => useChatBuffer());

    // Add messages to buffer
    act(() => {
      result.current.addMessageToBuffer('Message 1');
      result.current.addMessageToBuffer('Message 2');
    });

    // Check buffer status
    expect(result.current.bufferLength).toBe(2);
    expect(result.current.getBufferStatus().bufferLength).toBe(2);
    expect(result.current.getBufferStatus().nextFlushETA).toBe(2); // 2 seconds

    // Mock the custom event listener
    const eventSpy = vi.fn();
    window.addEventListener('flushChatBuffer', eventSpy);

    // Advance time to trigger flush
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    // Buffer should be flushed
    expect(result.current.bufferLength).toBe(0);
    expect(eventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        detail: { messages: ['Message 1', 'Message 2'] }
      })
    );

    window.removeEventListener('flushChatBuffer', eventSpy);
  });

  it('should flush immediately when buffer reaches max size', () => {
    const { result } = renderHook(() => useChatBuffer());

    // Mock the custom event listener
    const eventSpy = vi.fn();
    window.addEventListener('flushChatBuffer', eventSpy);

    // Add messages up to max buffer size (5)
    act(() => {
      result.current.addMessageToBuffer('Message 1');
      result.current.addMessageToBuffer('Message 2');
      result.current.addMessageToBuffer('Message 3');
      result.current.addMessageToBuffer('Message 4');
      result.current.addMessageToBuffer('Message 5'); // This should trigger immediate flush
    });

    // Process any pending timers (the immediate flush uses setTimeout)
    act(() => {
      vi.runAllTimers();
    });

    // Buffer should be flushed immediately
    expect(result.current.bufferLength).toBe(0);
    expect(eventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        detail: { messages: ['Message 1', 'Message 2', 'Message 3', 'Message 4', 'Message 5'] }
      })
    );

    window.removeEventListener('flushChatBuffer', eventSpy);
  });

  it('should clear buffer when requested', () => {
    const { result } = renderHook(() => useChatBuffer());

    // Add messages to buffer
    act(() => {
      result.current.addMessageToBuffer('Message 1');
      result.current.addMessageToBuffer('Message 2');
    });

    expect(result.current.bufferLength).toBe(2);

    // Clear the buffer
    act(() => {
      result.current.clearBuffer();
    });

    expect(result.current.bufferLength).toBe(0);
    expect(result.current.pendingMessages.length).toBe(0);
  });

  it('should provide correct buffer status information', () => {
    const { result } = renderHook(() => useChatBuffer());

    // Initially no messages
    expect(result.current.getBufferStatus()).toEqual({
      bufferLength: 0,
      hasPendingMessages: false,
      nextFlushETA: 0
    });

    // Add messages to buffer
    act(() => {
      result.current.addMessageToBuffer('Message 1');
      result.current.addMessageToBuffer('Message 2');
    });

    const status = result.current.getBufferStatus();
    expect(status.bufferLength).toBe(2);
    expect(status.hasPendingMessages).toBe(false); // No pending messages yet
    expect(status.nextFlushETA).toBe(2); // 2 seconds until next flush
  });

  it('should generate unique message IDs', () => {
    const { result } = renderHook(() => useChatBuffer());

    let id1: string;
    let id2: string;

    act(() => {
      id1 = result.current.addMessageToBuffer('Message 1');
      id2 = result.current.addMessageToBuffer('Message 2');
    });

    expect(id1).toBeDefined();
    expect(id2).toBeDefined();
    expect(id1).not.toBe(id2);
  });

  it('should track pending messages and update their status', async () => {
    const mockUseWebSocketEvent = vi.mocked(await import('../useWebSocket')).useWebSocketEvent;
    let chatMessageHandler: (payload: any) => void;

    // Capture the chat message handler
    mockUseWebSocketEvent.mockImplementation((eventType: string, handler: any) => {
      if (eventType === 'CHAT_MESSAGE') {
        chatMessageHandler = handler;
      }
    });

    const { result } = renderHook(() => useChatBuffer());

    // Mock flush to create pending messages
    act(() => {
      result.current.addMessageToBuffer('Test message');
      result.current.flushBuffer();
    });

    expect(result.current.pendingMessages).toHaveLength(1);
    expect(result.current.pendingMessages[0].status).toBe('pending');

    // Simulate receiving the message back from server
    act(() => {
      chatMessageHandler({
        sender_id: 'test-player',
        message: 'Test message'
      });
    });

    expect(result.current.pendingMessages[0].status).toBe('sent');

    // Messages should be cleaned up after delay
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    expect(result.current.pendingMessages).toHaveLength(0);
  });

  it('should handle manual flush correctly', () => {
    const { result } = renderHook(() => useChatBuffer());

    // Mock the custom event listener
    const eventSpy = vi.fn();
    window.addEventListener('flushChatBuffer', eventSpy);

    // Add messages to buffer
    act(() => {
      result.current.addMessageToBuffer('Message 1');
      result.current.addMessageToBuffer('Message 2');
    });

    expect(result.current.bufferLength).toBe(2);

    // Manually flush
    act(() => {
      result.current.flushBuffer();
    });

    expect(result.current.bufferLength).toBe(0);
    expect(eventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        detail: { messages: ['Message 1', 'Message 2'] }
      })
    );

    window.removeEventListener('flushChatBuffer', eventSpy);
  });
});