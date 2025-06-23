import React from 'react';
import { Player } from '../../types';

interface IdentityCardProps {
  localPlayer: Player;
}

export const IdentityCard: React.FC<IdentityCardProps> = ({ localPlayer }) => {
  const getPlayerAvatar = (player: Player) => {
    if (player.role?.type === 'CISO') return '👤';
    if (player.role?.type === 'SYSTEMS') return '🧑‍💻';
    if (player.role?.type === 'ETHICS') return '🕵️';
    if (player.role?.type === 'CTO') return '🤖';
    if (player.role?.type === 'COO') return '🧑‍🚀';
    if (player.role?.type === 'CFO') return '👩‍🔬';
    return '👤';
  };

  const getAlignmentDisplay = (player: Player) => {
    if (player.alignment === 'AI') {
      return (
        <span className="stat aligned">
          🤖 ALIGNED 
          <span className="visibility-icon private" title="Only you can see this">🔒</span>
        </span>
      );
    }
    return (
      <span className="stat human">
        👤 HUMAN 
        <span className="visibility-icon private" title="Only you can see this">🔒</span>
      </span>
    );
  };

  const getRoleDisplayName = (player: Player) => {
    if (!player.role) return 'Employee';
    
    switch (player.role.type) {
      case 'CISO': return 'Chief Security Officer';
      case 'SYSTEMS': return 'Systems Administrator';
      case 'ETHICS': return 'Ethics Officer';
      case 'CTO': return 'Chief Technology Officer';
      case 'COO': return 'Chief Operating Officer';
      case 'CFO': return 'Chief Financial Officer';
      default: return player.role.name || player.jobTitle;
    }
  };

  return (
    <div className="hud-header">
      <div className="identity-compact">
        <div className="player-avatar large">
          {getPlayerAvatar(localPlayer)}
        </div>
        <div className="identity-info">
          <h3 className="identity-name">{localPlayer.name}</h3>
          <p className="identity-role">{getRoleDisplayName(localPlayer)}</p>
          <div className="identity-stats">
            {getAlignmentDisplay(localPlayer)}
            <span className="stat tokens">🪙 {localPlayer.tokens}</span>
          </div>
        </div>
      </div>
    </div>
  );
};