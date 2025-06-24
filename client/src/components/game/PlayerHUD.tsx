import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { IdentityCard } from './IdentityCard';
import { ThreatMeter } from './ThreatMeter';
import { ObjectiveCard } from './ObjectiveCard';
import { AbilityCard } from './AbilityCard';
import styles from './PlayerHUD.module.css';

export const PlayerHUD: React.FC = () => {
  const { gameState, localPlayer } = useGameContext();

  if (!localPlayer) {
    return null;
  }
  const aiEquity = localPlayer.aiEquity || 0;

  return (
    <aside className={styles.panelRight}>
      <IdentityCard localPlayer={localPlayer} />

      <div className={styles.hudContentSingle}>
        {/* Only show threat meter if player is Human */}
        {localPlayer.alignment === 'HUMAN' && (
          <ThreatMeter
            tokens={localPlayer.tokens}
            aiEquity={aiEquity}
          />
        )}

        <div className={styles.hudSection}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionTitle}>📋 OBJECTIVES</span>
          </div>

          <ObjectiveCard
            type="Team Objective"
            name={localPlayer.alignment === 'HUMAN' ? "Containment Protocol" : "Achieve Singularity"}
            description={localPlayer.alignment === 'HUMAN'
              ? "Identify and vote to deactivate the Original AI."
              : "Convert enough humans to achieve AI dominance."
            }
          />

          {localPlayer.personalKPI && (
            <ObjectiveCard
              type="Personal KPI"
              name={localPlayer.personalKPI.type}
              description={localPlayer.personalKPI.description}
              progressText={
                `Progress: ${localPlayer.personalKPI.progress || 0}/${localPlayer.personalKPI.target || 1} ${localPlayer.personalKPI.isCompleted ? '✓' : ''}`
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

        <AbilityCard localPlayer={localPlayer} />

        {localPlayer.lastNightAction && (
          <div className={styles.hudSection}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>🌙 LAST NIGHT'S ACTION</span>
            </div>
            <div className={styles.actionsList}>
              <div className={`${styles.actionItem} ${styles.ability} ${styles.selected}`}>
                <span className={styles.actionIcon}>➡️</span>
                <span className={styles.actionName}>{localPlayer.lastNightAction.type}</span>
                {localPlayer.lastNightAction.targetId && (
                  <span className={styles.actionTarget}>
                    TARGET: {gameState.players.find(p => p.id === localPlayer.lastNightAction?.targetId)?.name || 'Unknown'}
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