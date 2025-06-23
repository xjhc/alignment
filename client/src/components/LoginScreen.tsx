import { useState } from 'react';
import styles from './LoginScreen.module.css';

interface LoginScreenProps {
  onLogin: (playerName: string, avatar: string) => void;
}

const avatarOptions = ['👤', '🧑‍💻', '🕵️', '🤖', '🧑‍🚀'];

export function LoginScreen({ onLogin }: LoginScreenProps) {
  const [selectedAvatar, setSelectedAvatar] = useState(avatarOptions[0]);
  const [playerName, setPlayerName] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (playerName.trim()) {
      onLogin(playerName.trim(), selectedAvatar);
    }
  };

  return (
    <div className={styles.launchScreen}>
      <h1 className={styles.logo}>
        LOEBIAN INC. // <span className={styles.glitch}>EMERGENCY BRIDGE</span>
      </h1>
      
      <form className={styles.launchForm} onSubmit={handleSubmit}>
        <div className={styles.avatarSelector}>
          {avatarOptions.map((avatar) => (
            <button
              key={avatar}
              type="button"
              className={`${styles.avatarOption} ${selectedAvatar === avatar ? styles.selected : ''}`}
              onClick={() => setSelectedAvatar(avatar)}
            >
              {avatar}
            </button>
          ))}
        </div>
        
        <input
          type="text"
          value={playerName}
          onChange={(e) => setPlayerName(e.target.value)}
          placeholder="[ENTER YOUR HANDLE]"
          maxLength={20}
          required
        />
        
        <button type="submit" className={styles.btnPrimary} disabled={!playerName.trim()}>
          [ &gt; BROWSE LOBBIES ]
        </button>
      </form>
    </div>
  );
}