import { useState } from 'react';

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
    <div className="launch-screen">
      <h1 className="logo">
        LOEBIAN INC. // <span className="glitch">EMERGENCY BRIDGE</span>
      </h1>
      
      <form className="launch-form" onSubmit={handleSubmit}>
        <div className="avatar-selector">
          {avatarOptions.map((avatar) => (
            <button
              key={avatar}
              type="button"
              className={`avatar-option ${selectedAvatar === avatar ? 'selected' : ''}`}
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
        
        <button type="submit" className="btn-primary" disabled={!playerName.trim()}>
          [ &gt; BROWSE LOBBIES ]
        </button>
      </form>
    </div>
  );
}