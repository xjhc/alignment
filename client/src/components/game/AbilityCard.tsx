import React from 'react';
import { motion } from 'framer-motion';
import { Player } from '../../types';
import { useGameContext } from '../../contexts/GameContext';

interface AbilityCardProps {
  localPlayer: Player;
  isViewingSelf: boolean;
}

export const AbilityCard: React.FC<AbilityCardProps> = ({ localPlayer, isViewingSelf }) => {
  const { canPlayerAffordAbility } = useGameContext();
  const ability = localPlayer.role?.ability;
  
  // Only show ability details when viewing self
  if (!isViewingSelf) {
    return (
      <div className="animation-fade-in">
        <div className="flex justify-between items-center mb-2">
          <span className="text-[11px] font-bold text-text-muted uppercase">🎯 ABILITY</span>
          <span className="bg-background-tertiary border border-border text-text-muted px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase">
            CLASSIFIED
          </span>
        </div>
        <div className="bg-background-tertiary border border-border rounded-lg p-3 opacity-70">
          <div className="font-bold text-sm mb-1.5">🔍 Classified Information</div>
          <div className="text-text-secondary text-[11px] leading-snug mb-1.5">Role abilities are not visible in other players' dossiers.</div>
        </div>
      </div>
    );
  }
  
  if (!ability || !ability.name || ability.name === "Unknown Ability") {
    return (
      <div className="animation-fade-in">
        <div className="flex justify-between items-center mb-2">
          <span className="text-[11px] font-bold text-text-muted uppercase">üéØ ABILITY</span>
          <span className="bg-text-muted/50 text-text-secondary px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase">
            PASSIVE
          </span>
        </div>
        <div className="bg-background-tertiary border border-border rounded-lg p-3 opacity-70">
          <div className="font-bold text-sm mb-1.5">No Active Ability</div>
          <div className="text-text-secondary text-[11px] leading-snug mb-1.5">
            {localPlayer.role?.name ? 
              `The ${localPlayer.role.name} role has no active abilities. Focus on collaboration and strategic voting.` :
              'This role has no special active abilities. Your contribution comes through teamwork and influence.'
            }
          </div>
        </div>
      </div>
    );
  }

  // Comprehensive ability state calculation
  const canAfford = canPlayerAffordAbility(localPlayer.id);
  const hasNotUsedAbility = !localPlayer.hasUsedAbility;
  const isAbilityUnlocked = localPlayer.role?.isUnlocked ?? false;
  
  const isReady = isAbilityUnlocked && hasNotUsedAbility && canAfford;
  
  // Determine status and reason for locked state
  let status: string;
  let lockedReason: string = '';
  
  if (isReady) {
    status = 'READY';
  } else {
    status = 'LOCKED';
    if (!isAbilityUnlocked) {
      const role = localPlayer.role;
      const requiredMilestones = role?.milestoneRequirement || 3;
      lockedReason = `Requires ${requiredMilestones} Project Milestones to unlock.`;
    } else if (localPlayer.hasUsedAbility) {
      lockedReason = 'Ability used this night. Refreshes at next Night Phase.';
    } else if (!canAfford) {
      const cost = ability.cost || 1;
      const currentTokens = localPlayer.tokens || 0;
      lockedReason = `Requires ${cost} tokens (you have ${currentTokens}).`;
    } else {
      lockedReason = 'Ability currently unavailable. Conditions not met.';
    }
  }

  return (
    <div className="animation-fade-in">
      <div className="flex justify-between items-center mb-2">
        <span className="text-[11px] font-bold text-text-muted uppercase">🎯 ABILITY</span>
        <span className={`px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase text-white ${
          isReady ? 'bg-success' : 'bg-danger'
        }`}>
          {status}
        </span>
      </div>
      <motion.div 
        className={`bg-background-tertiary border rounded-lg p-3 cursor-pointer ${
          isReady ? 'border-success' : 'border-border opacity-70'
        }`}
        whileHover={isReady ? { 
          scale: 1.02, 
          borderColor: 'rgba(34, 197, 94, 0.8)',
          boxShadow: '0 0 12px rgba(34, 197, 94, 0.3)',
          transition: { duration: 0.2 }
        } : {}}
        whileTap={isReady ? { scale: 0.98 } : {}}
      >
        <div className="font-bold text-sm mb-1.5">{ability.name}</div>
        <div className="text-text-secondary text-[11px] leading-snug mb-1.5">{ability.description}</div>
        
        {/* Show locked reason or ready state */}
        {isReady ? (
          <div className="text-[10px] text-success font-medium">
            ✓ Available to use during the Night Phase.
          </div>
        ) : (
          <div className="text-[10px] text-danger font-medium">
            🔒 {lockedReason}
          </div>
        )}
      </motion.div>
    </div>
  );
};