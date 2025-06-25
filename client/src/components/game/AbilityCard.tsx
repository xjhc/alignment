import React from 'react';
import { Player } from '../../types';

interface AbilityCardProps {
  localPlayer: Player;
}

export const AbilityCard: React.FC<AbilityCardProps> = ({ localPlayer }) => {
  const ability = localPlayer.role?.ability;
  
  if (!ability) {
    return (
      <div className="animate-fade-in">
        <div className="flex justify-between items-center mb-2">
          <span className="text-[11px] font-bold text-text-muted uppercase">🎯 ABILITY</span>
          <span className="bg-text-muted text-white px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase">
            NO ABILITY
          </span>
        </div>
        <div className="bg-background-tertiary border border-border rounded-lg p-3 opacity-70">
          <div className="font-bold text-sm mb-1.5">No Active Ability</div>
          <div className="text-text-secondary text-[11px] leading-snug mb-1.5">This role has no special abilities.</div>
        </div>
      </div>
    );
  }

  const isReady = ability.isReady && !localPlayer.hasUsedAbility;
  const status = isReady ? 'READY' : 'LOCKED';

  return (
    <div className="animate-fade-in">
      <div className="flex justify-between items-center mb-2">
        <span className="text-[11px] font-bold text-text-muted uppercase">🎯 ABILITY</span>
        <span className={`px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase text-white ${
          isReady ? 'bg-success' : 'bg-text-muted'
        }`}>
          {status}
        </span>
      </div>
      <div className={`bg-background-tertiary border rounded-lg p-3 ${
        isReady ? 'border-success' : 'border-border opacity-70'
      }`}>
        <div className="font-bold text-sm mb-1.5">{ability.name}</div>
        <div className="text-text-secondary text-[11px] leading-snug mb-1.5">{ability.description}</div>
        <div className="text-[10px] text-text-muted italic">Used during Night Phase (30s window)</div>
      </div>
    </div>
  );
};