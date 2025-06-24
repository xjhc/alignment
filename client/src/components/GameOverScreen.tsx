import { GameState } from '../types';
import { FinalPlayerCard } from './game/FinalPlayerCard';
import styles from './GameOverScreen.module.css';

interface GameOverScreenProps {
  gameState: GameState;
  onViewAnalysis: () => void;
  onPlayAgain: () => void;
}

export function GameOverScreen({ gameState, onViewAnalysis, onPlayAgain }: GameOverScreenProps) {
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
        <div className={`${styles.victoryBadge} ${styles.humanVictory} animate-scale-in`}>
          <div className={`${styles.victoryIcon} animate-bounce`} style={{ animationDelay: '0.3s' }}>🛡️</div>
          <div className={styles.victoryText}>
            <h1 className={`${styles.victoryTitle} animate-slide-in-up`} style={{ animationDelay: '0.5s' }}>CONTAINMENT ACHIEVED</h1>
            <p className={`${styles.victorySubtitle} animate-slide-in-up`} style={{ animationDelay: '0.7s' }}>Human Faction Victory</p>
          </div>
        </div>
      );
    } else {
      return (
        <div className={`${styles.victoryBadge} ${styles.aiVictory} animate-scale-in`}>
          <div className={`${styles.victoryIcon} ${styles.glitch} animate-shake`} style={{ animationDelay: '0.3s' }}>🤖</div>
          <div className={styles.victoryText}>
            <h1 className={`${styles.victoryTitle} ${styles.glitch} animate-slide-in-up`} style={{ animationDelay: '0.5s' }}>THE SINGULARITY</h1>
            <p className={`${styles.victorySubtitle} animate-slide-in-up`} style={{ animationDelay: '0.7s' }}>AI Faction Victory</p>
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
    <main className={styles.endgameScreen}>
      <header className={styles.endgameHeader}>
        <div className={`${styles.companyLogo} ${!isHumanVictory ? styles.glitch : ''}`}>LOEBIAN</div>
        <div className={styles.headerControls}>
          <button 
            className={styles.headerBtn} 
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
      
      <div className={styles.endgameContent}>
        <div className={styles.victoryAnnouncement}>
          {getVictoryBadge()}
          <p className={styles.victoryReason}>{getVictoryReason()}</p>
        </div>

        <div className={styles.finalStatusSection}>
          <h2 className={styles.sectionTitle}>📋 FINAL PERSONNEL STATUS</h2>
          
          <div className={styles.personnelGrid}>
            {sortedPlayers.map((player, index) => (
              <div 
                key={player.id}
                className="animate-stagger-reveal"
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

        <div className={styles.victoryActions}>
          <button 
            className={`${styles.btnPrimary} ${styles.analysisBtn} animate-slide-in-up`} 
            onClick={onViewAnalysis}
            style={{ animationDelay: `${1000 + (sortedPlayers.length * 150) + 300}ms` }}
          >
            📊 VIEW ANALYSIS
          </button>
          <button 
            className={`${styles.btnSecondary} ${styles.playAgainBtn} animate-slide-in-up`} 
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