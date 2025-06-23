import React from 'react';
import { Player } from '../../types';

interface AbilityCardProps {
  localPlayer: Player;
}

export const AbilityCard: React.FC<AbilityCardProps> = ({ localPlayer }) => {
  const ability = localPlayer.role?.ability;
  
  if (!ability) {
    return (
      <div className="hud-section">
        <div className="section-header">
          <span className="section-title">🎯 ABILITY</span>
          <span className="ability-status">NO ABILITY</span>
        </div>
        <div className="ability-card locked">
          <div className="ability-name">No Active Ability</div>
          <div className="ability-description">This role has no special abilities.</div>
        </div>
      </div>
    );
  }

  const isReady = ability.isReady && !localPlayer.hasUsedAbility;
  const status = isReady ? 'READY' : 'LOCKED';
  const cardClass = isReady ? 'ability-card ready' : 'ability-card locked';

  return (
    <div className="hud-section">
      <div className="section-header">
        <span className="section-title">🎯 ABILITY</span>
        <span className="ability-status">{status}</span>
      </div>
      <div className={cardClass}>
        <div className="ability-name">{ability.name}</div>
        <div className="ability-description">{ability.description}</div>
        <div className="ability-usage">Used during Night Phase (30s window)</div>
      </div>
    </div>
  );
};