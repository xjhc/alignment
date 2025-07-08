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
  
  if (!ability) {
    return (
      <div className="animation-fade-in">
        <div className="flex justify-between items-center mb-2">
          <span className="text-[11px] font-bold text-text-muted uppercase">🎯 ABILITY</span>
          <span className="bg-text-muted text-white px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase">
            NO ABILITY
          </span>
        </div>
        <div className="bg-background-tertiary border border-border rounded-lg p-3 opacity-70">
          <div className="font-bold text-sm mb-1.5">No Active Ability</div>
          <div className="text-text-secondary text-[11px] leading-snug mb-1.5">
            {localPlayer.role?.name ? 
              `The ${localPlayer.role.name} role has no special abilities. Focus on collaboration and discussion.` :
              'This role has no special abilities. Your contribution comes through strategic thinking and teamwork.'
            }
          </div>
          <div className="text-[10px] text-text-muted italic">Some roles gain power through other means - tokens, milestones, or voting influence.</div>
        </div>
      </div>
    );
  }

  // Comprehensive ability state calculation
  const canAfford = canPlayerAffordAbility(localPlayer.id);
  const hasNotUsedAbility = !localPlayer.hasUsedAbility;
  const isAbilityUnlocked = ability.isReady;
  
  const isReady = isAbilityUnlocked && hasNotUsedAbility && canAfford;
  
  // Determine status and reason for locked state
  let status: string;
  let lockedReason: string = '';
  
  if (isReady) {
    status = 'READY';
  } else {
    status = 'LOCKED';
    if (!isAbilityUnlocked) {
      lockedReason = 'System access required. Complete more objectives to unlock.';
    } else if (localPlayer.hasUsedAbility) {
      lockedReason = 'Already used this phase. Abilities refresh each day.';
    } else if (!canAfford) {
      lockedReason = 'Insufficient tokens. Abilities require token payment.';
    } else {
      lockedReason = 'Ability currently unavailable.';
    }
  }

  return (
    <div className="animation-fade-in">
      <div className="flex justify-between items-center mb-2">
        <span className="text-[11px] font-bold text-text-muted uppercase">🎯 ABILITY</span>
        <span className={`px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase text-white ${
          isReady ? 'bg-success' : 'bg-text-muted'
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
            ✓ Ready to use during Night Phase (30s window)
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