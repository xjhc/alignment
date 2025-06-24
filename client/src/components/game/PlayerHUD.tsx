import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { IdentityCard } from './IdentityCard';
import { ThreatMeter } from './ThreatMeter';
import { ObjectiveCard } from './ObjectiveCard';
import { AbilityCard } from './AbilityCard';
import styles from './PlayerHUD.module.css';

export const PlayerHUD: React.FC = () => {
  const { gameState, viewedPlayer } = useGameContext();

  if (!viewedPlayer) {
    return null;
  }
  const aiEquity = viewedPlayer.aiEquity || 0;

  return (
    <aside className={styles.panelRight}>
      <IdentityCard localPlayer={viewedPlayer} />

      <div className={styles.hudContentSingle}>
        {/* Only show threat meter if player is Human */}
        {viewedPlayer.alignment === 'HUMAN' && (
          <ThreatMeter
            tokens={viewedPlayer.tokens}
            aiEquity={aiEquity}
          />
        )}

        <div className={styles.hudSection}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionTitle}>📋 OBJECTIVES</span>
          </div>

          <ObjectiveCard
            type="Team Objective"
            name={viewedPlayer.alignment === 'HUMAN' ? "Containment Protocol" : "Achieve Singularity"}
            description={viewedPlayer.alignment === 'HUMAN'
              ? "Identify and vote to deactivate the Original AI."
              : "Convert enough humans to achieve AI dominance."
            }
          />

          {viewedPlayer.personalKPI && (
            <ObjectiveCard
              type="Personal KPI"
              name={viewedPlayer.personalKPI.type}
              description={viewedPlayer.personalKPI.description}
              progressText={
                `Progress: ${viewedPlayer.personalKPI.progress || 0}/${viewedPlayer.personalKPI.target || 1} ${viewedPlayer.personalKPI.isCompleted ? '✓' : ''}`
              }
              isPrivate={true}
            />
          )}

          {gameState.corporateMandate && (
            <ObjectiveCard
              type="Mandate"
              name={gameState.corporateMandate.name}
              description={gameState.corporateMandate.description}
            />
          )}
        </div>

        <AbilityCard localPlayer={viewedPlayer} />

        {viewedPlayer.lastNightAction && (
          <div className={styles.hudSection}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>🌙 LAST NIGHT'S ACTION</span>
            </div>
            <div className={styles.actionsList}>
              <div className={`${styles.actionItem} ${styles.ability} ${styles.selected}`}>
                <span className={styles.actionIcon}>➡️</span>
                <span className={styles.actionName}>{viewedPlayer.lastNightAction.type}</span>
                {viewedPlayer.lastNightAction.targetId && (
                  <span className={styles.actionTarget}>
                    TARGET: {gameState.players.find(p => p.id === viewedPlayer.lastNightAction?.targetId)?.name || 'Unknown'}
                  </span>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
};