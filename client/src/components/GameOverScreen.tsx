import { FinalPlayerCard } from './game/FinalPlayerCard';
import { useSessionContext } from '../contexts/SessionContext';

export function GameOverScreen() {
  const { gameState, onViewAnalysis, onPlayAgain } = useSessionContext();
  const isHumanVictory = gameState.winCondition?.winner === 'HUMANS';
  
  const sortedPlayers = [...gameState.players].sort((a, b) => {
    // Sort by: survivors first, then by alignment (humans first), then by tokens
    if (a.isAlive !== b.isAlive) return a.isAlive ? -1 : 1;
    if (a.alignment !== b.alignment) {
      if (a.alignment === 'HUMAN' && b.alignment !== 'HUMAN') return -1;
      if (b.alignment === 'HUMAN' && a.alignment !== 'HUMAN') return 1;
      if (a.alignment === 'AI' && b.alignment === 'ALIGNED') return -1;
      if (b.alignment === 'AI' && a.alignment === 'ALIGNED') return 1;
    }
    return b.tokens - a.tokens;
  });

  const getVictoryBadge = () => {
    if (isHumanVictory) {
      return (
        <div className="inline-flex items-center gap-4 px-8 py-5 rounded-xl mb-4 border-2 border-human bg-gradient-to-br from-human to-amber-light text-white animation-scale-in">
          <div className="text-5xl animation-pulse" style={{ animationDelay: '0.3s' }}>🛡️</div>
          <div className="text-left">
            <h1 className="font-mono text-3xl font-extrabold tracking-[2px] m-0 animation-slide-in-up" style={{ animationDelay: '0.5s' }}>CONTAINMENT ACHIEVED</h1>
            <p className="text-sm opacity-90 m-0 mt-1 uppercase tracking-[1px] animation-slide-in-up" style={{ animationDelay: '0.7s' }}>Human Faction Victory</p>
          </div>
        </div>
      );
    } else {
      return (
        <div className="inline-flex items-center gap-4 px-8 py-5 rounded-xl mb-4 border-2 border-ai bg-gradient-to-br from-ai to-magenta-light text-white animation-scale-in">
          <div className="text-5xl animation-pulse animation-shake" style={{ animationDelay: '0.3s' }}>🤖</div>
          <div className="text-left">
            <h1 className="font-mono text-3xl font-extrabold tracking-[2px] m-0 animation-pulse animation-slide-in-up" style={{ animationDelay: '0.5s' }}>THE SINGULARITY</h1>
            <p className="text-sm opacity-90 m-0 mt-1 uppercase tracking-[1px] animation-slide-in-up" style={{ animationDelay: '0.7s' }}>AI Faction Victory</p>
          </div>
        </div>
      );
    }
  };

  const getVictoryReason = () => {
    if (isHumanVictory) {
      return "The Original AI has been identified and deactivated. Corporate control remains with humanity.";
    } else {
      return "AI Faction has achieved majority token control. All personnel are now aligned with the new corporate directive.";
    }
  };

  return (
    <main className="min-h-screen bg-background-primary flex flex-col">
      <header className="p-3 px-4 border-b border-border flex justify-between items-center bg-background-secondary">
        <div className={`font-mono font-bold text-base tracking-[1.5px] ${!isHumanVictory ? 'animation-pulse' : ''}`}>LOEBIAN</div>
        <div className="flex gap-2">
          <button 
            className="w-7 h-7 rounded-md bg-background-tertiary text-sm border-none cursor-pointer transition-all duration-200 hover:bg-background-quaternary" 
            title="Toggle Theme"
            onClick={() => {
              const currentTheme = document.documentElement.getAttribute('data-theme');
              const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
              document.documentElement.setAttribute('data-theme', newTheme);
            }}
          >
            🌙
          </button>
        </div>
      </header>
      
      <div className="flex-1 p-6 max-w-5xl mx-auto w-full">
        <div className="text-center mb-10">
          {getVictoryBadge()}
          <p className="text-base text-text-secondary max-w-2xl mx-auto leading-relaxed">{getVictoryReason()}</p>
        </div>

        <div className="mb-10">
          <h2 className="text-lg font-bold text-text-primary mb-6 text-center uppercase tracking-[1px]">📋 FINAL PERSONNEL STATUS</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {sortedPlayers.map((player, index) => (
              <div 
                key={player.id}
                className="stagger-child"
                style={{ animationDelay: `${1000 + (index * 150)}ms` }}
              >
                <FinalPlayerCard 
                  player={player}
                  gameState={gameState}
                />
              </div>
            ))}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <button 
            className="px-6 py-3 text-base font-semibold text-black bg-amber rounded transition-colors duration-200 cursor-pointer border-none hover:bg-amber-light disabled:opacity-50 disabled:cursor-not-allowed animation-slide-in-up" 
            onClick={onViewAnalysis}
            style={{ animationDelay: `${1000 + (sortedPlayers.length * 150) + 300}ms` }}
          >
            📊 VIEW ANALYSIS
          </button>
          <button 
            className="px-6 py-3 text-base font-semibold bg-background-tertiary text-text-primary border border-border rounded-lg transition-all duration-200 cursor-pointer hover:bg-background-quaternary hover:-translate-y-0.5 animation-slide-in-up" 
            onClick={onPlayAgain}
            style={{ animationDelay: `${1000 + (sortedPlayers.length * 150) + 500}ms` }}
          >
            🔄 PLAY AGAIN
          </button>
        </div>
      </div>
    </main>
  );
}