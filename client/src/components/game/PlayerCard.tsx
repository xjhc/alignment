import React, { useEffect, useState } from 'react';
import { Player } from '../../types';
import styles from './PlayerCard.module.css';

interface PlayerCardProps {
  player: Player;
  isSelf: boolean;
}

export const PlayerCard: React.FC<PlayerCardProps> = ({ player, isSelf }) => {
  const [wasAlive, setWasAlive] = useState(player.isAlive);
  const [showEliminationAnimation, setShowEliminationAnimation] = useState(false);

  useEffect(() => {
    // Trigger elimination animation when player becomes not alive
    if (wasAlive && !player.isAlive) {
      setShowEliminationAnimation(true);
      // Reset animation after it completes
      const timer = setTimeout(() => {
        setShowEliminationAnimation(false);
      }, 1500);
      return () => clearTimeout(timer);
    }
    setWasAlive(player.isAlive);
  }, [player.isAlive, wasAlive]);

  const getPlayerAvatar = (p: Player) => {
    // If the player is not alive, always show the ghost.
    if (!p.isAlive) return '👻';

    // Use the avatar from player object if available, otherwise fall back to generic icon
    return p.avatar || '👤';
  };

  const getPlayerClasses = () => {
    const classes = [styles.playerCardUniform];
    if (isSelf) classes.push(styles.isMe);
    if (!player.isAlive) classes.push(styles.deactivated);
    if (player.alignment === 'AI' || player.alignment === 'ALIGNED') classes.push(styles.aligned);
    if (player.systemShocks?.some(shock => shock.isActive)) {
      classes.push(styles.hasShock);
    }
    if (showEliminationAnimation) {
      classes.push('animate-elimination-fade');
    }
    return classes.join(' ');
  };

  const getDataAttributes = () => {
    const attrs: { [key: string]: string } = { 'data-name': player.name };
    if (player.systemShocks?.some(shock => shock.isActive)) {
      attrs['data-status'] = 'shocked';
    }
    return attrs;
  };

  const renderProjectMilestones = () => {
    const maxMilestones = 5; // Assuming max 5 milestones
    const icons = [];
    for (let i = 0; i < maxMilestones; i++) {
      const filled = i < player.projectMilestones;
      const progressClass = filled ? 
        `${styles.progressIcon} ${styles.filled} ${player.alignment === 'AI' || player.alignment === 'ALIGNED' ? styles.aligned : ''}` : 
        styles.progressIcon;
      icons.push(
        <span key={i} className={progressClass}>●</span>
      );
    }
    return icons;
  };

  const displayName = isSelf ? `${player.name} (Me)` : player.name;
  const displayTokens = player.isAlive ? `🪙 ${player.tokens}` : '❌';

  return (
    <div
      className={getPlayerClasses()}
      {...getDataAttributes()}
      title={`View Dossier for ${player.name}`}
    >
      <div className={styles.playerAvatar}>
        {getPlayerAvatar(player)}
      </div>
      <div className={styles.playerContent}>
        <div className={styles.playerMainInfo}>
          <span className={styles.playerName}>{displayName}</span>
          <span className={styles.playerJob}>{player.role?.name || player.jobTitle || 'Employee'}</span>
          <div className={styles.playerTokensProgress}>
            <span className={styles.playerTokens}>{displayTokens}</span>
            <div className={styles.playerProgressInline}>
              {renderProjectMilestones()}
            </div>
          </div>
        </div>
        {player.statusMessage && (
          <div className={`${styles.playerStatusLine} ${!player.isAlive ? styles.parting : ''}`}>
            "{player.statusMessage}"
          </div>
        )}
      </div>
    </div>
  );
};