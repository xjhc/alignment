import React, { useState } from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { useGameActions } from '../../hooks/useGameActions';
import { Player } from '../../types';
import { Button } from '../ui';

interface NightActionSelectionProps {
  // No props needed - everything comes from context
}

type ActionType = 'mine' | 'project' | 'ability' | null;

export const NightActionSelection: React.FC<NightActionSelectionProps> = () => {
  const { gameState, localPlayer } = useGameContext();
  const {
    setMiningTarget,
    handleMineTokens,
    handleUseAbility,
    handleProjectMilestones,
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
      handleProjectMilestones();
      setIsMinimized(true);
    }
  };

  const handleTargetSelect = (targetId: string) => {
    setSelectedTarget(targetId);
    
    if (selectedAction === 'mine') {
      setMiningTarget(targetId);
      handleMineTokens();
      setIsMinimized(true);
    } else if (selectedAction === 'ability') {
      handleUseAbility(targetId);
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
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
        <div className="p-2">
          <div className="flex flex-col gap-1">
            <div className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md cursor-pointer transition-all duration-150 hover:bg-gray-600 border-amber-500 ${
              selectedAction === 'mine' ? 'bg-amber-500/10 border-amber-500' : ''
            }`} onClick={() => setIsMinimized(false)}>
              <span className="text-xs w-4 text-center">⛏️</span>
              <span className="font-medium text-xs text-gray-100 flex-grow">Mine for Player</span>
              <span className="ml-auto text-xs font-bold text-blue-500 font-mono">{selectedAction === 'mine' && selectedTarget ? 'SELECTED' : ''}</span>
            </div>
            <div className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md cursor-pointer transition-all duration-150 hover:bg-gray-600 ${
              selectedAction === 'project' ? 'bg-amber-500/10 border-amber-500' : ''
            }`} onClick={() => handleActionSelect('project')}>
              <span className="text-xs w-4 text-center">📈</span>
              <span className="font-medium text-xs text-gray-100 flex-grow">Project Milestones</span>
              <span className="text-xs text-gray-500 bg-gray-900 px-1 py-0.5 rounded-lg font-medium uppercase">Role</span>
            </div>
            <div className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md transition-all duration-150 ${
              hasUnlockedAbility && canAffordAbility ? 'cursor-pointer hover:bg-gray-600' : 'opacity-60 cursor-not-allowed grayscale-30'
            }`}>
              <span className="text-xs w-4 text-center">🔒</span>
              <span className="font-medium text-xs text-gray-100 flex-grow">{localPlayer.role?.type || 'Role Ability'}</span>
              <span className="text-xs text-red-500 font-bold">{hasUnlockedAbility ? (canAffordAbility ? 'Ready' : 'No Tokens') : 'Locked'}</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
      <div className="flex flex-col gap-4">
        <div className="m-0">
          <h3 className="text-sm font-bold text-gray-100 m-0 mb-3">Choose Night Action</h3>
        </div>

        <div className="flex flex-col gap-2">
          <div 
            className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 cursor-pointer transition-all duration-150 hover:bg-gray-700 hover:-translate-y-0.5 ${
              selectedAction === 'mine' ? 'border-amber-500 bg-amber-500/10' : ''
            }`}
            onClick={() => handleActionSelect('mine')}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-base w-5 text-center">⛏️</span>
              <span className="font-semibold text-gray-100 text-sm flex-grow">Mine for Player</span>
              <span className="text-xs text-green-500 bg-green-500/10 px-1.5 py-0.5 rounded-2xl font-bold">{getActionSuccessRate('mine')}</span>
            </div>
            <div className="text-xs text-gray-400 leading-snug mb-1.5">
              Generate 1 Token for another player. Success depends on liquidity pool (currently 3 
              slots for {alivePlayers.length} players attempting). Cannot mine for yourself.
            </div>
            <div className="text-xs text-gray-500">
              <strong className="text-gray-100 font-semibold">Command:</strong> <code className="bg-gray-700 px-1 py-0.5 rounded font-mono text-xs">Mine for [Player Name]</code>
            </div>
          </div>

          <div 
            className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 cursor-pointer transition-all duration-150 hover:bg-gray-700 hover:-translate-y-0.5 ${
              selectedAction === 'project' ? 'border-amber-500 bg-amber-500/10' : ''
            }`}
            onClick={() => handleActionSelect('project')}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-base w-5 text-center">📈</span>
              <span className="font-semibold text-gray-100 text-sm flex-grow">Project Milestones</span>
              <span className="text-xs text-gray-500 bg-gray-700 px-1.5 py-0.5 rounded-2xl font-medium uppercase">Role Ability</span>
            </div>
            <div className="text-xs text-gray-400 leading-snug mb-1.5">
              Advance your role's project by 1 point. At 3 points, unlock your powerful role-specific 
              ability for future nights.
            </div>
            <div className="text-xs text-gray-500">
              <strong className="text-gray-100 font-semibold">Command:</strong> <code className="bg-gray-700 px-1 py-0.5 rounded font-mono text-xs">Project Milestones</code>
            </div>
          </div>

          <div 
            className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 transition-all duration-150 ${
              hasUnlockedAbility && canAffordAbility 
                ? `cursor-pointer hover:bg-gray-700 hover:-translate-y-0.5 ${selectedAction === 'ability' ? 'border-amber-500 bg-amber-500/10' : ''}` 
                : 'opacity-60 cursor-not-allowed grayscale-30'
            }`}
            onClick={() => hasUnlockedAbility && canAffordAbility && handleActionSelect('ability')}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-base w-5 text-center">🔒</span>
              <span className="font-semibold text-gray-100 text-sm flex-grow">{localPlayer.role?.type || 'Role Ability'}</span>
              <span className="text-xs text-red-500 bg-red-500/10 px-1.5 py-0.5 rounded-2xl font-bold">
                {hasUnlockedAbility ? (canAffordAbility ? 'Ready' : 'No Tokens') : 'Locked'}
              </span>
            </div>
            <div className="text-xs text-gray-400 leading-snug mb-1.5">
              {hasUnlockedAbility 
                ? `Use your role-specific ability. ${canAffordAbility ? 'Available to use.' : 'Requires more tokens.'}`
                : 'Role ability requires system access currently unavailable.'
              }
            </div>
            <div className="text-xs text-gray-500">
              <strong className="text-gray-100 font-semibold">Status:</strong> {hasUnlockedAbility ? 'System access granted' : 'Network security protocols offline'}
            </div>
          </div>
        </div>

        {selectedAction === 'mine' && (
          <div className="pt-3 border-t border-gray-600">
            <div className="mb-2">
              <h4 className="text-xs font-semibold text-gray-100 m-0">Select Mining Target</h4>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {alivePlayers.map((player) => (
                <Button
                  key={player.id}
                  variant={selectedTarget === player.id ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => handleTargetSelect(player.id)}
                  className={`flex items-center gap-1.5 px-2.5 py-1.5 ${
                    selectedTarget === player.id ? 'border-amber-500 bg-amber-500/10' : ''
                  } ${
                    player.alignment === 'ALIGNED' ? 'bg-cyan-500/5 border-cyan-600' : ''
                  }`}
                >
                  <span className="text-sm leading-none flex-shrink-0">{getPlayerIcon(player)}</span>
                  <span className={`font-medium text-xs text-gray-100 flex-shrink-0 ${
                    player.alignment === 'ALIGNED' ? 'text-cyan-600 animate-[glitch_1.5s_infinite]' : ''
                  }`}>
                    {player.name}
                  </span>
                  <span className="font-mono font-bold text-gray-400 text-xs ml-auto">🪙 {player.tokens}</span>
                </Button>
              ))}
            </div>
          </div>
        )}

        {selectedAction === 'ability' && hasUnlockedAbility && (
          <div className="pt-3 border-t border-gray-600">
            <div className="mb-2">
              <h4 className="text-xs font-semibold text-gray-100 m-0">Select Ability Target</h4>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {alivePlayers.map((player) => (
                <Button
                  key={player.id}
                  variant={selectedTarget === player.id ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => isValidNightActionTarget(localPlayer.id, player.id, 'ability') && handleTargetSelect(player.id)}
                  disabled={!isValidNightActionTarget(localPlayer.id, player.id, 'ability')}
                  className={`flex items-center gap-1.5 px-2.5 py-1.5 ${
                    selectedTarget === player.id ? 'border-amber-500 bg-amber-500/10' : ''
                  } ${
                    !isValidNightActionTarget(localPlayer.id, player.id, 'ability') ? 'grayscale-50' : ''
                  }`}
                >
                  <span className="text-sm leading-none flex-shrink-0">{getPlayerIcon(player)}</span>
                  <span className="font-medium text-xs text-gray-100 flex-shrink-0">{player.name}</span>
                  <span className="font-mono font-bold text-gray-400 text-xs ml-auto">🪙 {player.tokens}</span>
                </Button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};