import { useCallback, useRef } from 'react';

interface SoundOptions {
  volume?: number;
  playbackRate?: number;
}

interface SoundMap {
  vote: string;
  message: string;
  notification: string;
  timerWarning: string;
  phaseChange: string;
  buttonClick: string;
  error: string;
  success: string;
}

// Sound effects mapping - these would be actual audio file paths in production
const SOUND_PATHS: SoundMap = {
  vote: '/sounds/vote-cast.mp3',
  message: '/sounds/message-received.mp3',
  notification: '/sounds/notification.mp3',
  timerWarning: '/sounds/timer-warning.mp3',
  phaseChange: '/sounds/phase-change.mp3',
  buttonClick: '/sounds/button-click.mp3',
  error: '/sounds/error.mp3',
  success: '/sounds/success.mp3',
};

export const useSound = () => {
  const audioCache = useRef<Map<string, HTMLAudioElement>>(new Map());
  const isEnabled = useRef(true); // Could be tied to user preferences

  const preloadSound = useCallback((soundKey: keyof SoundMap) => {
    const path = SOUND_PATHS[soundKey];
    if (!audioCache.current.has(path)) {
      const audio = new Audio(path);
      audio.preload = 'auto';
      audioCache.current.set(path, audio);
    }
  }, []);

  const playSound = useCallback((
    soundKey: keyof SoundMap, 
    options: SoundOptions = {}
  ) => {
    if (!isEnabled.current) return;

    const path = SOUND_PATHS[soundKey];
    let audio = audioCache.current.get(path);

    if (!audio) {
      audio = new Audio(path);
      audioCache.current.set(path, audio);
    }

    // Clone the audio for multiple simultaneous plays
    const audioClone = audio.cloneNode() as HTMLAudioElement;
    
    if (options.volume !== undefined) {
      audioClone.volume = Math.max(0, Math.min(1, options.volume));
    } else {
      audioClone.volume = 0.3; // Default volume
    }
    
    if (options.playbackRate !== undefined) {
      audioClone.playbackRate = options.playbackRate;
    }

    audioClone.play().catch(error => {
      // Silently handle autoplay restrictions
      console.debug('Sound play failed:', error);
    });
  }, []);

  const toggleSound = useCallback(() => {
    isEnabled.current = !isEnabled.current;
    return isEnabled.current;
  }, []);

  // Preload critical sounds
  const preloadCriticalSounds = useCallback(() => {
    ['buttonClick', 'notification', 'message'].forEach(sound => 
      preloadSound(sound as keyof SoundMap)
    );
  }, [preloadSound]);

  return {
    playSound,
    preloadSound,
    preloadCriticalSounds,
    toggleSound,
    isEnabled: () => isEnabled.current,
  };
};