import React from 'react';
import { Player } from '../../types';
import styles from './IdentityCard.module.css';

interface IdentityCardProps {
  localPlayer: Player;
}

export const IdentityCard: React.FC<IdentityCardProps> = ({ localPlayer }) => {
  const getPlayerAvatar = (player: Player) => {
    if (player.role?.type === 'CISO') return '👤';
    if (player.role?.type === 'SYSTEMS') return '🧑‍💻';
    if (player.role?.type === 'ETHICS') return '🕵️';
    if (player.role?.type === 'CTO') return '🤖';
    if (player.role?.type === 'COO') return '🧑‍🚀';
    if (player.role?.type === 'CFO') return '👩‍🔬';
    return '👤';
  };

  const getAlignmentDisplay = (player: Player) => {
    if (player.alignment === 'AI') {
      return (
        <span className={`${styles.stat} ${styles.aligned}`}>
          🤖 ALIGNED 
          <span className={`${styles.visibilityIcon} ${styles.private}`} title="Only you can see this">🔒</span>
        </span>
      );
    }
    return (
      <span className={`${styles.stat} ${styles.human}`}>
        👤 HUMAN 
        <span className={`${styles.visibilityIcon} ${styles.private}`} title="Only you can see this">🔒</span>
      </span>
    );
  };

  const getRoleDisplayName = (player: Player) => {
    if (!player.role) return 'Employee';
    
    switch (player.role.type) {
      case 'CISO': return 'Chief Security Officer';
      case 'SYSTEMS': return 'Systems Administrator';
      case 'ETHICS': return 'Ethics Officer';
      case 'CTO': return 'Chief Technology Officer';
      case 'COO': return 'Chief Operating Officer';
      case 'CFO': return 'Chief Financial Officer';
      default: return player.role.name || player.jobTitle;
    }
  };

  return (
    <div className={styles.hudHeader}>
      <div className={styles.identityCompact}>
        <div className={`${styles.playerAvatar} ${styles.large}`}>
          {getPlayerAvatar(localPlayer)}
        </div>
        <div className={styles.identityInfo}>
          <h3 className={styles.identityName}>{localPlayer.name}</h3>
          <p className={styles.identityRole}>{getRoleDisplayName(localPlayer)}</p>
          <div className={styles.identityStats}>
            {getAlignmentDisplay(localPlayer)}
            <span className={`${styles.stat} ${styles.tokens}`}>🪙 {localPlayer.tokens}</span>
          </div>
        </div>
      </div>
    </div>
  );
};