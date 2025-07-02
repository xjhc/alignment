import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Button, Input } from './ui';

interface LoginScreenProps {
  onLogin: (playerName: string, avatar: string) => void;
}

const avatarOptions = ['👤', '🧑‍💻', '🕵️', '🤖', '🧑‍🚀'];

export function LoginScreen({ onLogin }: LoginScreenProps) {
  const [selectedAvatar, setSelectedAvatar] = useState(avatarOptions[0]);
  const [playerName, setPlayerName] = useState('');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (playerName.trim()) {
      onLogin(playerName.trim(), selectedAvatar);
      
      // Check if there's a return URL for invite links
      const returnUrl = searchParams.get('return');
      if (returnUrl) {
        // Small delay to ensure login completes
        setTimeout(() => {
          navigate(decodeURIComponent(returnUrl));
        }, 100);
      }
    }
  };

  const handleDiscordLogin = () => {
    // Store the return URL in localStorage before redirecting to Discord
    const returnUrl = searchParams.get('return');
    if (returnUrl) {
      localStorage.setItem('discord_login_return_url', returnUrl);
    }
    
    // Redirect to Discord OAuth
    window.location.href = '/api/auth/login';
  };

  return (
    <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
      <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
        LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
      </h1>
      
      {/* Discord Login Option */}
      <div className="w-80 text-center">
        <Button
          type="button"
          onClick={handleDiscordLogin}
          variant="primary"
          size="lg"
          fullWidth
          className="text-xl font-semibold mb-6 bg-indigo-600 hover:enabled:bg-indigo-700 text-white"
        >
          🎮 Login with Discord
        </Button>
        
        <div className="flex items-center gap-4 mb-6">
          <div className="flex-1 border-t border-border"></div>
          <span className="text-text-secondary text-sm">OR PLAY AS GUEST</span>
          <div className="flex-1 border-t border-border"></div>
        </div>
      </div>
      
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
          [ &gt; CONTINUE AS GUEST ]
        </Button>
      </form>
    </div>
  );
}