import { useCallback } from 'react';
import { soundManager, type SoundEffect } from '../services/soundManager';

interface SoundOptions {
  volume?: number;
}

// Legacy sound mapping for backward compatibility
interface SoundMap {
  vote: SoundEffect;
  message: SoundEffect;
  notification: SoundEffect;
  timerWarning: SoundEffect;
  phaseChange: SoundEffect;
  buttonClick: SoundEffect;
  error: SoundEffect;
  success: SoundEffect;
}

const SOUND_MAP: SoundMap = {
  vote: 'vote',
  message: 'message',
  notification: 'notification',
  timerWarning: 'timerWarning',
  phaseChange: 'phaseChange',
  buttonClick: 'buttonClick',
  error: 'error',
  success: 'success',
};

export const useSound = () => {
  const playSound = useCallback((
    soundKey: keyof SoundMap, 
    options: SoundOptions = {}
  ) => {
    const effectKey = SOUND_MAP[soundKey];
    soundManager.playSound(effectKey, options);
  }, []);

  const toggleSound = useCallback(() => {
    return soundManager.toggleMute();
  }, []);

  const preloadSound = useCallback((soundKey: keyof SoundMap) => {
    // SoundManager handles preloading automatically
    // This method is kept for backward compatibility
  }, []);

  const preloadCriticalSounds = useCallback(() => {
    soundManager.preloadCriticalSounds();
  }, []);

  return {
    playSound,
    preloadSound,
    preloadCriticalSounds,
    toggleSound,
    isEnabled: () => !soundManager.isMutedState(),
  };
};