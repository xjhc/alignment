import React from 'react';
import { Player, RoleType } from '../../types';

interface IdentityCardProps {
  localPlayer: Player;
}

export const IdentityCard: React.FC<IdentityCardProps> = ({ localPlayer }) => {
  const getPlayerAvatar = (player: Player) => {
    if (player.role?.type === RoleType.Ciso) return '👤';
    if (player.role?.type === RoleType.Platforms) return '🧑‍💻';
    if (player.role?.type === RoleType.Ethics) return '🕵️';
    if (player.role?.type === RoleType.Cto) return '🤖';
    if (player.role?.type === RoleType.Coo) return '🧑‍🚀';
    if (player.role?.type === RoleType.Cfo) return '👩‍🔬';
    if (player.role?.type === RoleType.Intern) return '🎓';
    return '👤';
  };

  const getAlignmentDisplay = (player: Player) => {
    if (player.alignment === 'AI') {
      return (
        <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-aligned text-white">
          🤖 ALIGNED
          <span className="text-[8px] opacity-60 cursor-help text-pink-200" title="Only you can see this">🔒</span>
        </span>
      );
    }
    return (
      <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-human text-white">
        👤 HUMAN
        <span className="text-[8px] opacity-60 cursor-help text-pink-200" title="Only you can see this">🔒</span>
      </span>
    );
  };

  const getRoleDisplayName = (player: Player) => {
    if (!player.role) return 'Employee';

    switch (player.role.type) {
      case RoleType.Ciso: return 'Chief Information Security Officer';
      case RoleType.Platforms: return 'Systems Administrator';
      case RoleType.Ethics: return 'VP, Ethics';
      case RoleType.Cto: return 'Chief Technology Officer';
      case RoleType.Coo: return 'Chief Operating Officer';
      case RoleType.Cfo: return 'Chief Financial Officer';
      case RoleType.Intern: return 'Research Intern';
      default: return player.role.name || player.jobTitle;
    }
  };

  return (
    <div className="p-4 border-b border-border bg-background-secondary">
      <div className="flex gap-2.5 items-center">
        <div className="w-12 h-12 rounded-full bg-background-tertiary flex items-center justify-center text-2xl flex-shrink-0 border-2 border-border">
          {getPlayerAvatar(localPlayer)}
        </div>
        <div className="flex-grow">
          <h3 className="text-base font-bold text-text-primary m-0">{localPlayer.name}</h3>
          <p className="text-[11px] text-text-secondary uppercase my-0.5 mt-0.5 mb-1.5">{getRoleDisplayName(localPlayer)}</p>
          <div className="flex gap-1.5 mt-0">
            {getAlignmentDisplay(localPlayer)}
            <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-background-tertiary border border-border text-text-primary">
              🪙 {localPlayer.tokens}
            </span>
            {localPlayer.role?.type === RoleType.Intern && (
              <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-purple-500/20 border border-purple-500/50 text-purple-200">
                📚 {localPlayer.bootcampPoints || 0}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};