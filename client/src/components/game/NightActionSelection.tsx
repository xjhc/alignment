import React, { useState } from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { useGameActions } from '../../hooks/useGameActions';
import { Player } from '../../types';
import styles from './NightActionSelection.module.css';

interface NightActionSelectionProps {
  // No props needed - everything comes from context
}

type ActionType = 'mine' | 'project' | 'ability' | null;

export const NightActionSelection: React.FC<NightActionSelectionProps> = () => {
  const { gameState, localPlayer } = useGameContext();
  const {
    conversionTarget,
    setConversionTarget,
    miningTarget,
    setMiningTarget,
    handleConversionAttempt,
    handleMineTokens,
    handleUseAbility,
    canPlayerAffordAbility,
    isValidNightActionTarget,
  } = useGameActions();
  
  if (!localPlayer) return null;
  const [selectedAction, setSelectedAction] = useState<ActionType>(null);
  const [selectedTarget, setSelectedTarget] = useState<string>('');
  const [isMinimized, setIsMinimized] = useState(false);

  const alivePlayers = gameState.players.filter(p => p.isAlive && p.id !== localPlayer.id);
  const canAffordAbility = canPlayerAffordAbility(localPlayer.id);
  const hasUnlockedAbility = localPlayer.role?.type && localPlayer.projectMilestones >= 3;

  const handleActionSelect = (action: ActionType) => {
    setSelectedAction(action);
    setSelectedTarget('');
    
    if (action === 'project') {
      // Project milestones doesn't need a target
      handleUseAbility();
      setIsMinimized(true);
    }
  };

  const handleTargetSelect = (targetId: string) => {
    setSelectedTarget(targetId);
    
    if (selectedAction === 'mine') {
      setMiningTarget(targetId);
      handleMineTokens();
      setIsMinimized(true);
    }
  };

  const getActionSuccessRate = (action: ActionType) => {
    if (action === 'mine') {
      const totalMiningAttempts = alivePlayers.length;
      const successRate = Math.max(33, Math.min(75, Math.floor((3 / totalMiningAttempts) * 100)));
      return `${successRate}% Success`;
    }
    return null;
  };

  const getPlayerIcon = (player: Player) => {
    switch (player.jobTitle) {
      case 'CISO': return '👤';
      case 'Systems': return '🧑‍💻';
      case 'Ethics': return '🕵️';
      case 'CTO': return '🤖';
      case 'COO': return '🧑‍🚀';
      case 'CFO': return '👩‍🔬';
      default: return '👤';
    }
  };

  if (isMinimized) {
    return (
      <div className={styles.nightActionPrompt}>
        <div className={styles.actionSelectionMinimized}>
          <div className={styles.actionsList}>
            <div className={`${styles.actionItem} ${styles.available} ${selectedAction === 'mine' ? styles.selected : ''}`} onClick={() => setIsMinimized(false)}>
              <span className={styles.actionIcon}>⛏️</span>
              <span className={styles.actionName}>Mine for Player</span>
              <span className={styles.actionTarget}>{selectedAction === 'mine' && selectedTarget ? 'SELECTED' : ''}</span>
            </div>
            <div className={`${styles.actionItem} ${selectedAction === 'project' ? styles.selected : ''}`} onClick={() => handleActionSelect('project')}>
              <span className={styles.actionIcon}>📈</span>
              <span className={styles.actionName}>Project Milestones</span>
              <span className={styles.actionCategory}>Role</span>
            </div>
            <div className={`${styles.actionItem} ${hasUnlockedAbility && canAffordAbility ? '' : styles.disabled}`}>
              <span className={styles.actionIcon}>🔒</span>
              <span className={styles.actionName}>{localPlayer.role?.type || 'Role Ability'}</span>
              <span className={styles.actionDisabled}>{hasUnlockedAbility ? (canAffordAbility ? 'Ready' : 'No Tokens') : 'Locked'}</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.nightActionPrompt}>
      <div className={styles.actionSelectionMenu}>
        <div className={styles.menuHeader}>
          <h3>Choose Night Action</h3>
        </div>

        <div className={styles.actionOptions}>
          <div 
            className={`${styles.actionOption} ${selectedAction === 'mine' ? styles.selected : ''}`}
            onClick={() => handleActionSelect('mine')}
          >
            <div className={styles.optionHeader}>
              <span className={styles.optionIcon}>⛏️</span>
              <span className={styles.optionName}>Mine for Player</span>
              <span className={styles.optionSuccess}>{getActionSuccessRate('mine')}</span>
            </div>
            <div className={styles.optionDescription}>
              Generate 1 Token for another player. Success depends on liquidity pool (currently 3 
              slots for {alivePlayers.length} players attempting). Cannot mine for yourself.
            </div>
            <div className={styles.optionUsage}>
              <strong>Command:</strong> <code>Mine for [Player Name]</code>
            </div>
          </div>

          <div 
            className={`${styles.actionOption} ${selectedAction === 'project' ? styles.selected : ''}`}
            onClick={() => handleActionSelect('project')}
          >
            <div className={styles.optionHeader}>
              <span className={styles.optionIcon}>📈</span>
              <span className={styles.optionName}>Project Milestones</span>
              <span className={styles.optionCategory}>Role Ability</span>
            </div>
            <div className={styles.optionDescription}>
              Advance your role's project by 1 point. At 3 points, unlock your powerful role-specific 
              ability for future nights.
            </div>
            <div className={styles.optionUsage}>
              <strong>Command:</strong> <code>Project Milestones</code>
            </div>
          </div>

          <div 
            className={`${styles.actionOption} ${hasUnlockedAbility && canAffordAbility ? '' : styles.disabled} ${selectedAction === 'ability' ? styles.selected : ''}`}
            onClick={() => hasUnlockedAbility && canAffordAbility && handleActionSelect('ability')}
          >
            <div className={styles.optionHeader}>
              <span className={styles.optionIcon}>🔒</span>
              <span className={styles.optionName}>{localPlayer.role?.type || 'Role Ability'}</span>
              <span className={styles.optionDisabled}>
                {hasUnlockedAbility ? (canAffordAbility ? 'Ready' : 'No Tokens') : 'Locked'}
              </span>
            </div>
            <div className={styles.optionDescription}>
              {hasUnlockedAbility 
                ? `Use your role-specific ability. ${canAffordAbility ? 'Available to use.' : 'Requires more tokens.'}`
                : 'Role ability requires system access currently unavailable.'
              }
            </div>
            <div className={styles.optionUsage}>
              <strong>Status:</strong> {hasUnlockedAbility ? 'System access granted' : 'Network security protocols offline'}
            </div>
          </div>
        </div>

        {selectedAction === 'mine' && (
          <div className={styles.targetSelection}>
            <div className={styles.targetHeader}>
              <h4>Select Mining Target</h4>
            </div>
            <div className={styles.targetGrid}>
              {alivePlayers.map((player) => (
                <button
                  key={player.id}
                  className={`${styles.targetBtn} ${selectedTarget === player.id ? styles.selected : ''} ${player.alignment === 'ALIGNED' ? styles.aligned : ''}`}
                  onClick={() => handleTargetSelect(player.id)}
                >
                  <span className={styles.targetEmoji}>{getPlayerIcon(player)}</span>
                  <span className={`${styles.targetName} ${player.alignment === 'ALIGNED' ? styles.glitched : ''}`}>
                    {player.name}
                  </span>
                  <span className={styles.targetTokens}>🪙 {player.tokens}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        {selectedAction === 'ability' && hasUnlockedAbility && (
          <div className={styles.targetSelection}>
            <div className={styles.targetHeader}>
              <h4>Select Ability Target</h4>
            </div>
            <div className={styles.targetGrid}>
              {alivePlayers.map((player) => (
                <button
                  key={player.id}
                  className={`${styles.targetBtn} ${selectedTarget === player.id ? styles.selected : ''} ${!isValidNightActionTarget(localPlayer.id, player.id, 'ability') ? styles.disabled : ''}`}
                  onClick={() => isValidNightActionTarget(localPlayer.id, player.id, 'ability') && setSelectedTarget(player.id)}
                  disabled={!isValidNightActionTarget(localPlayer.id, player.id, 'ability')}
                >
                  <span className={styles.targetEmoji}>{getPlayerIcon(player)}</span>
                  <span className={styles.targetName}>{player.name}</span>
                  <span className={styles.targetTokens}>🪙 {player.tokens}</span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};