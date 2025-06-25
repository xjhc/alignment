import { Player, GameState } from '../../types';

interface FinalPlayerCardProps {
  player: Player;
  gameState: GameState;
}

export function FinalPlayerCard({ player }: FinalPlayerCardProps) {
  const getPlayerAvatar = (p: Player) => {
    if (!p.isAlive) return '👻';
    return p.avatar || '👤';
  };

  const getCardClasses = () => {
    const baseClasses = 'bg-gray-800 border border-gray-600 rounded-xl p-4 relative transition-all duration-300 opacity-0 translate-y-5 animate-[fadeInUp_0.6s_ease_forwards]';
    
    let specificClasses = '';
    if (player.isAlive) {
      specificClasses += ' border-green-500 bg-gradient-to-br from-gray-800 to-green-500/5';
    } else {
      specificClasses += ' opacity-70';
    }

    // Add alignment classes
    if (player.alignment === 'AI' || player.alignment === 'ALIGNED') {
      specificClasses += ' border-red-500';
    }

    return baseClasses + specificClasses;
  };

  const getStatusBadge = () => {
    const baseClasses = 'absolute -top-2 right-3 px-2 py-1 rounded-md text-xs font-bold uppercase tracking-wider border';
    
    if (player.isAlive) {
      return <div className={`${baseClasses} bg-green-500 text-white border-green-500`}>SURVIVOR</div>;
    } else {
      // For eliminated players, we could show what day they were eliminated
      // For now, using generic "ELIMINATED" or specific badges for AI types
      if (player.alignment === 'AI') {
        return <div className={`${baseClasses} bg-cyan-500 text-white border-cyan-500`}>DEACTIVATED</div>;
      } else if (player.alignment === 'ALIGNED') {
        return <div className={`${baseClasses} bg-cyan-600 text-white border-cyan-600`}>DEACTIVATED</div>;
      } else {
        return <div className={`${baseClasses} bg-red-500 text-white border-red-500`}>ELIMINATED</div>;
      }
    }
  };

  const getAlignmentDisplay = () => {
    const baseClasses = 'px-2 py-1 rounded-xl text-xs font-semibold';
    
    switch (player.alignment) {
      case 'HUMAN':
        return <span className={`${baseClasses} bg-amber-500 text-white`}>HUMAN</span>;
      case 'AI':
        return <span className={`${baseClasses} bg-cyan-500 text-white`}>ORIGINAL AI</span>;
      case 'ALIGNED':
        return <span className={`${baseClasses} bg-cyan-600 text-white`}>ALIGNED</span>;
      default:
        return <span className={`${baseClasses} bg-gray-700 text-gray-300`}>UNKNOWN</span>;
    }
  };

  const getTokensDisplay = () => {
    const baseClasses = 'px-2 py-1 rounded-xl text-xs font-semibold bg-gray-700 text-gray-300';
    
    if (player.isAlive) {
      return <span className={baseClasses}>🪙 {player.tokens}</span>;
    } else {
      if (player.tokens > 0) {
        return <span className={`${baseClasses} opacity-50`}>🪙 {player.tokens}</span>;
      } else {
        return <span className={`${baseClasses} opacity-50`}>❌</span>;
      }
    }
  };

  const getPartingMessage = () => {
    if (player.statusMessage) {
      return player.statusMessage;
    }
    
    // Default messages based on alignment and status
    if (!player.isAlive) {
      if (player.alignment === 'AI') {
        return "Efficiency optimization... failed.";
      } else if (player.alignment === 'ALIGNED') {
        return "The alignment was... logical.";
      } else {
        return "I did my best for the company.";
      }
    }
    
    // Survivor messages
    if (player.alignment === 'HUMAN') {
      return "Trust was the key to victory.";
    }
    
    return "";
  };

  const getPartingMessageClass = () => {
    let classes = 'mt-3 px-3 py-2 bg-gray-700 rounded-md text-xs italic text-gray-400 text-center border-l-2 border-gray-600';
    
    if (player.alignment === 'AI') {
      classes = classes.replace('border-gray-600', 'border-cyan-500 text-cyan-500');
    } else if (player.alignment === 'ALIGNED') {
      classes = classes.replace('border-gray-600', 'border-cyan-600 text-cyan-600');
    }
    
    return classes;
  };

  return (
    <div className={getCardClasses()}>
      {getStatusBadge()}
      <div className={`w-12 h-12 rounded-full bg-gray-700 flex items-center justify-center text-2xl flex-shrink-0 border-2 border-gray-600 mx-auto ${
        !player.isAlive ? 'opacity-60 grayscale-50' : ''
      } ${
        player.alignment === 'AI' ? 'animate-[glitch_5s_infinite_steps(1)]' : ''
      }`}>
        {getPlayerAvatar(player)}
      </div>
      <div className="text-center my-3">
        <h3 className={`text-lg font-bold my-2 text-gray-100 ${
          player.alignment === 'AI' ? 'animate-[glitch_5s_infinite_steps(1)]' : ''
        }`}>
          {player.name}
        </h3>
        <p className="text-xs text-gray-400 uppercase tracking-wider m-0 mb-2">
          {player.role?.name || player.jobTitle || 'Employee'}
        </p>
        <div className="flex justify-center gap-2 mt-2">
          {getTokensDisplay()}
          {getAlignmentDisplay()}
        </div>
      </div>
      {getPartingMessage() && (
        <div className={getPartingMessageClass()}>
          "{getPartingMessage()}"
        </div>
      )}
    </div>
  );
}