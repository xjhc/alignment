# Audio Assets Directory

This directory contains audio assets for the Alignment game's dynamic music and ambiance system.

## Music Tracks (Background/Looping)

### `ambiance-lobby.mp3`
- **Type**: Ambient background loop
- **Description**: Quiet corporate office ambiance with subtle synth pads
- **Usage**: Played during lobby/waiting screens
- **Volume**: 40%
- **Loop**: Yes

### `music-day.mp3`
- **Type**: Background music
- **Description**: Low-tempo, minimalist electronic track for day phases
- **Usage**: Played during DAY, DISCUSSION, and VOTING phases
- **Volume**: 30%
- **Loop**: Yes

### `music-night.mp3`
- **Type**: Background music
- **Description**: Suspenseful track with subtle pulsing bassline for night phases
- **Usage**: Played during NIGHT and NIGHT_ACTIONS phases
- **Volume**: 40%
- **Loop**: Yes

## Sound Effects (One-shot)

### UI Sound Effects
- `vote-cast.mp3` - Sound when a player casts a vote
- `message-received.mp3` - Sound for incoming chat messages
- `notification.mp3` - General notification sound
- `timer-warning.mp3` - Warning sound for phase timer
- `phase-change.mp3` - Sound for phase transitions
- `button-click.mp3` - UI button click sound
- `error.mp3` - Error notification sound
- `success.mp3` - Success notification sound

### Game Event Stingers
- `stinger-victory.mp3` - Short, resolved chord for victory
- `stinger-defeat.mp3` - Short, dissonant sound for defeat

## Implementation Notes

All audio files are managed by the `SoundManager` service in `/src/services/soundManager.ts` using Howler.js for:

- Smooth cross-fading between music tracks (1-2 second transitions)
- Global mute/unmute functionality with localStorage persistence
- Automatic preloading of critical sounds
- Volume management per sound type
- Support for multiple simultaneous sound effects

## Audio Requirements

For production, audio files should be:
- Format: MP3 (for broad browser compatibility)
- Bitrate: 128kbps (balance of quality vs file size)
- Channels: Stereo
- Music tracks: 2-5 minutes duration for seamless looping
- Sound effects: Under 2 seconds duration
- Royalty-free or properly licensed

## Accessibility

The mute/unmute button in the RosterPanel provides user control over all audio, with preference persistence across sessions.