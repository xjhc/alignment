import { Howl, Howler } from 'howler';

type MusicTrack = 'lobby' | 'day' | 'night';
type SoundEffect = 'vote' | 'message' | 'notification' | 'timerWarning' | 'phaseChange' | 'buttonClick' | 'error' | 'success' | 'victory' | 'defeat';

interface SoundConfig {
  volume: number;
  loop?: boolean;
  preload?: boolean;
}

interface MusicConfig extends SoundConfig {
  loop: true;
}

interface EffectConfig extends SoundConfig {
  loop: false;
}

class SoundManager {
  private static instance: SoundManager;
  private musicTracks: Map<MusicTrack, Howl> = new Map();
  private soundEffects: Map<SoundEffect, Howl> = new Map();
  private currentMusic: Howl | null = null;
  private isMuted: boolean = false;

  // Audio file paths
  private readonly MUSIC_PATHS: Record<MusicTrack, string> = {
    lobby: '/sounds/ambiance-lobby.mp3',
    day: '/sounds/music-day.mp3',
    night: '/sounds/music-night.mp3',
  };

  private readonly EFFECT_PATHS: Record<SoundEffect, string> = {
    vote: '/sounds/vote-cast.mp3',
    message: '/sounds/message-received.mp3',
    notification: '/sounds/notification.mp3',
    timerWarning: '/sounds/timer-warning.mp3',
    phaseChange: '/sounds/phase-change.mp3',
    buttonClick: '/sounds/button-click.mp3',
    error: '/sounds/error.mp3',
    success: '/sounds/success.mp3',
    victory: '/sounds/stinger-victory.mp3',
    defeat: '/sounds/stinger-defeat.mp3',
  };

  private readonly MUSIC_CONFIG: Record<MusicTrack, MusicConfig> = {
    lobby: { volume: 0.4, loop: true, preload: true },
    day: { volume: 0.3, loop: true, preload: true },
    night: { volume: 0.4, loop: true, preload: true },
  };

  private readonly EFFECT_CONFIG: Record<SoundEffect, EffectConfig> = {
    vote: { volume: 0.5, loop: false, preload: true },
    message: { volume: 0.4, loop: false, preload: true },
    notification: { volume: 0.6, loop: false, preload: true },
    timerWarning: { volume: 0.7, loop: false, preload: true },
    phaseChange: { volume: 0.6, loop: false, preload: true },
    buttonClick: { volume: 0.3, loop: false, preload: true },
    error: { volume: 0.5, loop: false, preload: true },
    success: { volume: 0.5, loop: false, preload: true },
    victory: { volume: 0.8, loop: false, preload: false },
    defeat: { volume: 0.8, loop: false, preload: false },
  };

  private constructor() {
    this.loadMutePreference();
    this.initializeAudio();
  }

  public static getInstance(): SoundManager {
    if (!SoundManager.instance) {
      SoundManager.instance = new SoundManager();
    }
    return SoundManager.instance;
  }

  private loadMutePreference(): void {
    try {
      const savedMuteState = localStorage.getItem('alignment-audio-muted');
      this.isMuted = savedMuteState === 'true';
      Howler.mute(this.isMuted);
    } catch (error) {
      console.debug('Failed to load mute preference:', error);
      this.isMuted = false;
    }
  }

  private saveMutePreference(): void {
    try {
      localStorage.setItem('alignment-audio-muted', this.isMuted.toString());
    } catch (error) {
      console.debug('Failed to save mute preference:', error);
    }
  }

  private initializeAudio(): void {
    // Initialize music tracks
    Object.entries(this.MUSIC_PATHS).forEach(([track, path]) => {
      const config = this.MUSIC_CONFIG[track as MusicTrack];
      const howl = new Howl({
        src: [path],
        volume: config.volume,
        loop: config.loop,
        preload: config.preload,
        onloaderror: (id, error) => {
          console.debug(`Failed to load music track ${track}:`, error);
        },
      });
      this.musicTracks.set(track as MusicTrack, howl);
    });

    // Initialize sound effects
    Object.entries(this.EFFECT_PATHS).forEach(([effect, path]) => {
      const config = this.EFFECT_CONFIG[effect as SoundEffect];
      const howl = new Howl({
        src: [path],
        volume: config.volume,
        loop: config.loop,
        preload: config.preload,
        onloaderror: (id, error) => {
          console.debug(`Failed to load sound effect ${effect}:`, error);
        },
      });
      this.soundEffects.set(effect as SoundEffect, howl);
    });
  }

  public playMusic(track: MusicTrack): void {
    const musicHowl = this.musicTracks.get(track);
    if (!musicHowl) {
      console.debug(`Music track ${track} not found`);
      return;
    }

    // If same track is already playing, don't restart
    if (this.currentMusic === musicHowl && musicHowl.playing()) {
      return;
    }

    // Fade out current music if playing
    if (this.currentMusic && this.currentMusic.playing()) {
      this.currentMusic.fade(this.currentMusic.volume(), 0, 1500);
      this.currentMusic.once('fade', () => {
        this.currentMusic?.stop();
      });
    }

    // Start new music with fade in
    this.currentMusic = musicHowl;
    musicHowl.volume(0);
    musicHowl.play();
    const targetVolume = this.MUSIC_CONFIG[track].volume;
    musicHowl.fade(0, targetVolume, 1500);
  }

  public stopMusic(): void {
    if (this.currentMusic && this.currentMusic.playing()) {
      this.currentMusic.fade(this.currentMusic.volume(), 0, 1000);
      this.currentMusic.once('fade', () => {
        this.currentMusic?.stop();
        this.currentMusic = null;
      });
    }
  }

  public playSound(effect: SoundEffect, options: { volume?: number } = {}): void {
    const effectHowl = this.soundEffects.get(effect);
    if (!effectHowl) {
      console.debug(`Sound effect ${effect} not found`);
      return;
    }

    // Clone the sound for multiple simultaneous plays
    const volume = options.volume ?? this.EFFECT_CONFIG[effect].volume;
    effectHowl.volume(volume);
    effectHowl.play();
  }

  public mute(muted: boolean): void {
    this.isMuted = muted;
    Howler.mute(muted);
    this.saveMutePreference();
  }

  public toggleMute(): boolean {
    this.mute(!this.isMuted);
    return this.isMuted;
  }

  public isMutedState(): boolean {
    return this.isMuted;
  }

  public preloadCriticalSounds(): void {
    // Critical sounds are already preloaded in the config
    // This method exists for compatibility with the useSound hook
    const criticalEffects: SoundEffect[] = ['buttonClick', 'notification', 'message'];
    criticalEffects.forEach(effect => {
      const howl = this.soundEffects.get(effect);
      if (howl && howl.state() === 'unloaded') {
        howl.load();
      }
    });
  }

  public destroy(): void {
    // Clean up all audio instances
    this.musicTracks.forEach(howl => howl.unload());
    this.soundEffects.forEach(howl => howl.unload());
    this.musicTracks.clear();
    this.soundEffects.clear();
    this.currentMusic = null;
  }
}

// Export singleton instance
export const soundManager = SoundManager.getInstance();

// Export types for external use
export type { MusicTrack, SoundEffect };