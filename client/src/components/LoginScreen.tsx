import { useState } from 'react';
import { Button, Input } from './ui';

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
    <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
      <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
        LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
      </h1>
      
      <form className="flex flex-col gap-4 items-center w-80" onSubmit={handleSubmit}>
        <div className="flex gap-2 mb-4 justify-center">
          {avatarOptions.map((avatar) => (
            <button
              key={avatar}
              type="button"
              className={`w-12 h-12 border-2 bg-background-secondary rounded-lg text-2xl cursor-pointer transition-all duration-200 flex items-center justify-center hover:border-primary hover:scale-105 ${
                selectedAvatar === avatar 
                  ? 'border-primary bg-primary shadow-lg shadow-primary/30' 
                  : 'border-border'
              }`}
              onClick={() => setSelectedAvatar(avatar)}
            >
              {avatar}
            </button>
          ))}
        </div>
        
        <Input
          type="text"
          value={playerName}
          onChange={(e) => setPlayerName(e.target.value)}
          placeholder="[ENTER YOUR HANDLE]"
          maxLength={20}
          required
          size="lg"
          variant="filled"
          fullWidth
          className="text-xl text-center"
        />
        
        <Button 
          type="submit" 
          variant="primary"
          size="lg"
          fullWidth
          disabled={!playerName.trim()}
          className="text-xl font-semibold text-black bg-amber hover:enabled:bg-amber-light"
        >
          [ &gt; BROWSE LOBBIES ]
        </Button>
      </form>
    </div>
  );
}