import { renderHook, act } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { useChatRateLimit } from '../useChatRateLimit';

// Mock the useWebSocketEvent hook
vi.mock('../useWebSocket', () => ({
  useWebSocketEvent: vi.fn()
}));

describe('useChatRateLimit', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.clearAllTimers();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('should queue messages and process them with rate limiting', () => {
    const { result } = renderHook(() => useChatRateLimit());

    // Queue multiple messages
    act(() => {
      result.current.queueMessage('Message 1');
      result.current.queueMessage('Message 2');
      result.current.queueMessage('Message 3');
    });

    // Check queue status
    expect(result.current.queueLength).toBe(3);
    expect(result.current.getQueueStatus().queueLength).toBe(3);
    expect(result.current.getQueueStatus().nextMessageETA).toBe(7); // Math.ceil(3 * 2.1) = 7

    // Advance time to trigger first message processing
    act(() => {
      vi.advanceTimersByTime(2100);
    });

    // First message should be processed, queue length reduced
    expect(result.current.queueLength).toBe(2);
  });

  it('should clear the queue when requested', () => {
    const { result } = renderHook(() => useChatRateLimit());

    // Queue some messages
    act(() => {
      result.current.queueMessage('Message 1');
      result.current.queueMessage('Message 2');
    });

    expect(result.current.queueLength).toBe(2);

    // Clear the queue
    act(() => {
      result.current.clearQueue();
    });

    expect(result.current.queueLength).toBe(0);
    expect(result.current.pendingMessages.length).toBe(0);
  });

  it('should provide correct queue status information', () => {
    const { result } = renderHook(() => useChatRateLimit());

    // Initially no messages
    expect(result.current.getQueueStatus()).toEqual({
      queueLength: 0,
      hasPendingMessages: false,
      nextMessageETA: 0
    });

    // Queue some messages
    act(() => {
      result.current.queueMessage('Message 1');
      result.current.queueMessage('Message 2');
    });

    const status = result.current.getQueueStatus();
    expect(status.queueLength).toBe(2);
    expect(status.hasPendingMessages).toBe(false); // No pending messages yet
    expect(status.nextMessageETA).toBe(5); // Math.ceil(2 * 2.1) = 5
  });

  it('should generate unique message IDs', () => {
    const { result } = renderHook(() => useChatRateLimit());

    let id1: string;
    let id2: string;

    act(() => {
      id1 = result.current.queueMessage('Message 1');
      id2 = result.current.queueMessage('Message 2');
    });

    expect(id1).toBeDefined();
    expect(id2).toBeDefined();
    expect(id1).not.toBe(id2);
  });
});