import React, { useState } from 'react';
import { GameState, Player } from '../../types';

interface NightActionSelectionProps {
  gameState: GameState;
  localPlayer: Player;
  conversionTarget: string;
  setConversionTarget: (value: string) => void;
  miningTarget: string;
  setMiningTarget: (value: string) => void;
  handleConversionAttempt: () => Promise<void>;
  handleMineTokens: () => Promise<void>;
  handleUseAbility: () => Promise<void>;
  canPlayerAffordAbility: (playerId: string) => boolean;
  isValidNightActionTarget: (playerId: string, targetId: string, actionType: string) => boolean;
}

type ActionType = 'mine' | 'project' | 'ability' | null;

export const NightActionSelection: React.FC<NightActionSelectionProps> = ({
  gameState,
  localPlayer,
  setMiningTarget,
  handleMineTokens,
  handleUseAbility,
  canPlayerAffordAbility,
  isValidNightActionTarget,
}) => {
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
      <div className="night-action-prompt">
        <div className="action-selection-minimized">
          <div className="actions-list">
            <div className={`action-item available ${selectedAction === 'mine' ? 'selected' : ''}`} onClick={() => setIsMinimized(false)}>
              <span className="action-icon">⛏️</span>
              <span className="action-name">Mine for Player</span>
              <span className="action-target">{selectedAction === 'mine' && selectedTarget ? 'SELECTED' : ''}</span>
            </div>
            <div className={`action-item ${selectedAction === 'project' ? 'selected' : ''}`} onClick={() => handleActionSelect('project')}>
              <span className="action-icon">📈</span>
              <span className="action-name">Project Milestones</span>
              <span className="action-category">Role</span>
            </div>
            <div className={`action-item ${hasUnlockedAbility && canAffordAbility ? '' : 'disabled'}`}>
              <span className="action-icon">🔒</span>
              <span className="action-name">{localPlayer.role?.type || 'Role Ability'}</span>
              <span className="action-disabled">{hasUnlockedAbility ? (canAffordAbility ? 'Ready' : 'No Tokens') : 'Locked'}</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="night-action-prompt">
      <div className="action-selection-menu">
        <div className="menu-header">
          <h3>Choose Night Action</h3>
        </div>

        <div className="action-options">
          <div 
            className={`action-option ${selectedAction === 'mine' ? 'selected' : ''}`}
            onClick={() => handleActionSelect('mine')}
          >
            <div className="option-header">
              <span className="option-icon">⛏️</span>
              <span className="option-name">Mine for Player</span>
              <span className="option-success">{getActionSuccessRate('mine')}</span>
            </div>
            <div className="option-description">
              Generate 1 Token for another player. Success depends on liquidity pool (currently 3 
              slots for {alivePlayers.length} players attempting). Cannot mine for yourself.
            </div>
            <div className="option-usage">
              <strong>Command:</strong> <code>Mine for [Player Name]</code>
            </div>
          </div>

          <div 
            className={`action-option ${selectedAction === 'project' ? 'selected' : ''}`}
            onClick={() => handleActionSelect('project')}
          >
            <div className="option-header">
              <span className="option-icon">📈</span>
              <span className="option-name">Project Milestones</span>
              <span className="option-category">Role Ability</span>
            </div>
            <div className="option-description">
              Advance your role's project by 1 point. At 3 points, unlock your powerful role-specific 
              ability for future nights.
            </div>
            <div className="option-usage">
              <strong>Command:</strong> <code>Project Milestones</code>
            </div>
          </div>

          <div 
            className={`action-option ${hasUnlockedAbility && canAffordAbility ? '' : 'disabled'} ${selectedAction === 'ability' ? 'selected' : ''}`}
            onClick={() => hasUnlockedAbility && canAffordAbility && handleActionSelect('ability')}
          >
            <div className="option-header">
              <span className="option-icon">🔒</span>
              <span className="option-name">{localPlayer.role?.type || 'Role Ability'}</span>
              <span className="option-disabled">
                {hasUnlockedAbility ? (canAffordAbility ? 'Ready' : 'No Tokens') : 'Locked'}
              </span>
            </div>
            <div className="option-description">
              {hasUnlockedAbility 
                ? `Use your role-specific ability. ${canAffordAbility ? 'Available to use.' : 'Requires more tokens.'}`
                : 'Role ability requires system access currently unavailable.'
              }
            </div>
            <div className="option-usage">
              <strong>Status:</strong> {hasUnlockedAbility ? 'System access granted' : 'Network security protocols offline'}
            </div>
          </div>
        </div>

        {selectedAction === 'mine' && (
          <div className="target-selection">
            <div className="target-header">
              <h4>Select Mining Target</h4>
            </div>
            <div className="target-grid">
              {alivePlayers.map((player) => (
                <button
                  key={player.id}
                  className={`target-btn ${selectedTarget === player.id ? 'selected' : ''} ${player.alignment === 'ALIGNED' ? 'aligned' : ''}`}
                  onClick={() => handleTargetSelect(player.id)}
                >
                  <span className="target-emoji">{getPlayerIcon(player)}</span>
                  <span className={`target-name ${player.alignment === 'ALIGNED' ? 'glitched' : ''}`}>
                    {player.name}
                  </span>
                  <span className="target-tokens">🪙 {player.tokens}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        {selectedAction === 'ability' && hasUnlockedAbility && (
          <div className="target-selection">
            <div className="target-header">
              <h4>Select Ability Target</h4>
            </div>
            <div className="target-grid">
              {alivePlayers.map((player) => (
                <button
                  key={player.id}
                  className={`target-btn ${selectedTarget === player.id ? 'selected' : ''} ${!isValidNightActionTarget(localPlayer.id, player.id, 'ability') ? 'disabled' : ''}`}
                  onClick={() => isValidNightActionTarget(localPlayer.id, player.id, 'ability') && setSelectedTarget(player.id)}
                  disabled={!isValidNightActionTarget(localPlayer.id, player.id, 'ability')}
                >
                  <span className="target-emoji">{getPlayerIcon(player)}</span>
                  <span className="target-name">{player.name}</span>
                  <span className="target-tokens">🪙 {player.tokens}</span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};