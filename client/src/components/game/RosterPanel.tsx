import React from 'react';
import { Player } from '../../types';
import { PlayerCard } from './PlayerCard';
import styles from './RosterPanel.module.css';

interface RosterPanelProps {
  players: Player[];
  localPlayerId: string;
}

export const RosterPanel: React.FC<RosterPanelProps> = ({ players, localPlayerId }) => {
  const localPlayer = players.find(p => p.id === localPlayerId);

  const getPlayerCounts = () => {
    const humanCount = players.filter(p => p.isAlive && p.alignment !== 'AI' && p.alignment !== 'ALIGNED').length;
    const alignedCount = players.filter(p => p.isAlive && (p.alignment === 'AI' || p.alignment === 'ALIGNED')).length;
    const deactivatedCount = players.filter(p => !p.isAlive).length;
    return { humanCount, alignedCount, deactivatedCount };
  };

  const { humanCount, alignedCount, deactivatedCount } = getPlayerCounts();
  const isAI = localPlayer?.alignment === 'AI' || localPlayer?.alignment === 'ALIGNED';

  return (
    <aside className={styles.panelLeft}>
      <header className={styles.header}>
        <div className="header-left">
          <span className={styles.companyLogo}>LOEBIAN</span>
        </div>
        <div className={styles.headerControls}>
          <button className={styles.headerBtn} title="Settings">⚙️</button>
          <button className={styles.headerBtn} title="Toggle Theme">🌙</button>
        </div>
      </header>

      <div className={styles.channelList}>
        <div className={styles.listHeader}>Text Channels</div>
        <div className={`${styles.channel} ${styles.active}`}>
          <span>#</span>
          <span className="channel-name">war-room</span>
        </div>
        <div className={`${styles.channel} ${styles.aiChannel} ${isAI ? styles.unlocked : ''}`}>
          <span>#</span>
          <span className="channel-name">aligned</span>
          {!isAI && <span className={styles.channelBlocked}>❌</span>}
        </div>
        {deactivatedCount > 0 && (
          <div className={styles.channel}>
            <span>#</span>
            <span className="channel-name">off-boarding</span>
          </div>
        )}
      </div>

      <div className={styles.playerRoster}>
        <div className={styles.listHeader}>
          <span>Personnel</span>
          <div className={styles.rosterStats}>
            <span className={`${styles.statItem} ${styles.human}`}>👤 {humanCount}</span>
            <span className={`${styles.statItem} ${styles.aligned}`}>🤖 {alignedCount}</span>
            <span className={`${styles.statItem} ${styles.dead}`}>👻 {deactivatedCount}</span>
          </div>
        </div>

        {players.sort((a, b) => {
          // Sort self to top, then by alive status, then by name
          if (a.id === localPlayerId) return -1;
          if (b.id === localPlayerId) return 1;
          if (a.isAlive !== b.isAlive) return a.isAlive ? -1 : 1;
          return a.name.localeCompare(b.name);
        }).map((player) => (
          <PlayerCard
            key={player.id}
            player={player}
            isSelf={player.id === localPlayerId}
          />
        ))}
      </div>
    </aside>
  );
};