/**
 * Haptic Feedback Hook
 * 
 * Provides tactile feedback for mobile devices using the Web Vibration API.
 * Enhances user experience on mobile by providing physical confirmation
 * for critical game actions.
 */

import { useCallback } from 'react';

// Haptic feedback patterns for different interaction types
export const HAPTIC_PATTERNS = {
  // Light tap for regular button presses
  light: [10],
  
  // Medium pulse for important actions
  medium: [30],
  
  // Strong feedback for critical actions
  strong: [50],
  
  // Double tap for confirmations
  confirm: [20, 50, 20],
  
  // Short burst for errors
  error: [100, 50, 100],
  
  // Subtle notification pulse
  notification: [15, 30, 15],
  
  // Selection feedback
  select: [25],
  
  // Vote casting (important game action)
  vote: [40, 60, 40],
  
  // Success feedback
  success: [20, 20, 80],
} as const;

export type HapticPattern = keyof typeof HAPTIC_PATTERNS;

interface UseHapticsReturn {
  triggerHaptic: (pattern: HapticPattern) => void;
  isSupported: boolean;
  vibrate: (pattern: number | number[]) => void;
}

export function useHaptics(): UseHapticsReturn {
  // Check if vibration API is supported
  const isSupported = 'vibrate' in navigator;

  const vibrate = useCallback((pattern: number | number[]) => {
    if (!isSupported) {
      return;
    }

    try {
      navigator.vibrate(pattern);
    } catch (error) {
      console.debug('Vibration failed:', error);
    }
  }, [isSupported]);

  const triggerHaptic = useCallback((pattern: HapticPattern) => {
    if (!isSupported) {
      return;
    }

    const vibrationPattern = HAPTIC_PATTERNS[pattern];
    vibrate(vibrationPattern);
  }, [isSupported, vibrate]);

  return {
    triggerHaptic,
    isSupported,
    vibrate,
  };
}

// Convenience hook for specific haptic effects
export function useHapticFeedback() {
  const { triggerHaptic, isSupported } = useHaptics();

  return {
    light: () => triggerHaptic('light'),
    medium: () => triggerHaptic('medium'),
    strong: () => triggerHaptic('strong'),
    confirm: () => triggerHaptic('confirm'),
    error: () => triggerHaptic('error'),
    notification: () => triggerHaptic('notification'),
    select: () => triggerHaptic('select'),
    vote: () => triggerHaptic('vote'),
    success: () => triggerHaptic('success'),
    isSupported,
  };
}