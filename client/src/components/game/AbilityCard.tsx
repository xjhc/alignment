import React from 'react';
import { Player } from '../../types';
import styles from './AbilityCard.module.css';

interface AbilityCardProps {
  localPlayer: Player;
}

export const AbilityCard: React.FC<AbilityCardProps> = ({ localPlayer }) => {
  const ability = localPlayer.role?.ability;
  
  if (!ability) {
    return (
      <div className={styles.hudSection}>
        <div className={styles.sectionHeader}>
          <span className={styles.sectionTitle}>🎯 ABILITY</span>
          <span className={styles.abilityStatus}>NO ABILITY</span>
        </div>
        <div className={`${styles.abilityCard} ${styles.locked}`}>
          <div className={styles.abilityName}>No Active Ability</div>
          <div className={styles.abilityDescription}>This role has no special abilities.</div>
        </div>
      </div>
    );
  }

  const isReady = ability.isReady && !localPlayer.hasUsedAbility;
  const status = isReady ? 'READY' : 'LOCKED';
  const cardClass = isReady ? `${styles.abilityCard} ${styles.ready}` : `${styles.abilityCard} ${styles.locked}`;

  return (
    <div className={styles.hudSection}>
      <div className={styles.sectionHeader}>
        <span className={styles.sectionTitle}>🎯 ABILITY</span>
        <span className={styles.abilityStatus}>{status}</span>
      </div>
      <div className={cardClass}>
        <div className={styles.abilityName}>{ability.name}</div>
        <div className={styles.abilityDescription}>{ability.description}</div>
        <div className={styles.abilityUsage}>Used during Night Phase (30s window)</div>
      </div>
    </div>
  );
};