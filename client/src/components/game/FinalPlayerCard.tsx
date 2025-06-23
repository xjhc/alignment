import { Player, GameState } from '../../types';
import styles from './FinalPlayerCard.module.css';

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
    const classes = [styles.finalPlayerCard];
    
    if (player.isAlive) {
      classes.push(styles.survivor);
    } else {
      classes.push(styles.eliminated);
    }

    // Add alignment classes
    if (player.alignment === 'HUMAN') {
      classes.push(styles.human);
    } else if (player.alignment === 'AI') {
      classes.push(styles.aiOriginal, styles.defeated);
    } else if (player.alignment === 'ALIGNED') {
      classes.push(styles.aligned, styles.defeated);
    }

    return classes.join(' ');
  };

  const getStatusBadge = () => {
    if (player.isAlive) {
      return <div className={`${styles.cardStatusBadge} ${styles.survivorBadge}`}>SURVIVOR</div>;
    } else {
      // For eliminated players, we could show what day they were eliminated
      // For now, using generic "ELIMINATED" or specific badges for AI types
      if (player.alignment === 'AI') {
        return <div className={`${styles.cardStatusBadge} ${styles.aiBadge}`}>DEACTIVATED</div>;
      } else if (player.alignment === 'ALIGNED') {
        return <div className={`${styles.cardStatusBadge} ${styles.alignedBadge}`}>DEACTIVATED</div>;
      } else {
        return <div className={`${styles.cardStatusBadge} ${styles.eliminatedBadge}`}>ELIMINATED</div>;
      }
    }
  };

  const getAlignmentDisplay = () => {
    switch (player.alignment) {
      case 'HUMAN':
        return <span className={`${styles.statItem} ${styles.human}`}>HUMAN</span>;
      case 'AI':
        return <span className={`${styles.statItem} ${styles.ai}`}>ORIGINAL AI</span>;
      case 'ALIGNED':
        return <span className={`${styles.statItem} ${styles.aligned}`}>ALIGNED</span>;
      default:
        return <span className={styles.statItem}>UNKNOWN</span>;
    }
  };

  const getTokensDisplay = () => {
    if (player.isAlive) {
      return <span className={styles.statItem}>🪙 {player.tokens}</span>;
    } else {
      if (player.tokens > 0) {
        return <span className={`${styles.statItem} ${styles.faded}`}>🪙 {player.tokens}</span>;
      } else {
        return <span className={`${styles.statItem} ${styles.faded}`}>❌</span>;
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
    const classes = [styles.partingMessage];
    
    if (player.alignment === 'AI') {
      classes.push(styles.aiMessage);
    } else if (player.alignment === 'ALIGNED') {
      classes.push(styles.alignedMessage);
    } else if (!player.isAlive && player.alignment === 'HUMAN') {
      // Could add logic to detect if their final message was prescient
      // For now, just using default styling
    }
    
    return classes.join(' ');
  };

  return (
    <div className={getCardClasses()}>
      {getStatusBadge()}
      <div className={`${styles.playerAvatar} ${styles.large} ${!player.isAlive ? styles.faded : ''} ${player.alignment === 'AI' ? styles.glitch : ''}`}>
        {getPlayerAvatar(player)}
      </div>
      <div className={styles.playerInfo}>
        <h3 className={`${styles.playerName} ${player.alignment === 'AI' ? styles.glitch : ''}`}>
          {player.name}
        </h3>
        <p className={styles.playerRole}>{player.role?.name || player.jobTitle || 'Employee'}</p>
        <div className={styles.finalStats}>
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