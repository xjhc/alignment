import React from 'react';
import { getTeamBalance } from '../utils/roleDistribution';

interface RulesetSummaryProps {
  playerCount: number;
  maxPlayers: number;
  gameMode?: 'Classic' | 'Play as AI' | 'Sabotage';
  playAsAI?: boolean;
  initialAlignedCount?: number;
}

export const RulesetSummary: React.FC<RulesetSummaryProps> = ({
  playerCount,
  maxPlayers,
  playAsAI = false,
  initialAlignedCount = 0
}) => {
  // Get team balance for current player count (what would happen if we start now)
  const currentTeamBalance = getTeamBalance(playerCount, { initialAlignedCount });
  
  const getGameMode = () => {
    if (playAsAI) return 'Play as AI';
    if (initialAlignedCount > 0) return 'Sabotage';
    return 'Classic';
  };

  return (
    <div>
      <h3 className="text-sm font-bold text-text-primary mb-3 flex items-center gap-2">
        <span className="text-base">⚙️</span>
        Game Configuration
      </h3>
      
      <div className="space-y-3">
        {/* Game Details */}
        <div className="bg-background-primary border border-border rounded-lg p-3 space-y-2">
          <div className="flex items-center justify-between text-xs">
            <span className="text-text-secondary">Players</span>
            <span className="font-bold text-text-primary">
              {playerCount} / {maxPlayers}
            </span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-text-secondary">Mode</span>
            <span className={`font-bold ${
              getGameMode() === 'Classic' ? 'text-human' :
              getGameMode() === 'Play as AI' ? 'text-ai' :
              'text-aligned'
            }`}>
              {getGameMode()}
            </span>
          </div>
        </div>

        {/* Team Balance - If Started Now */}
        <div className="bg-background-primary border border-border rounded-lg p-3 space-y-1">
          <div className="text-xs font-medium text-text-primary mb-2">
            Team Balance
            <span className="text-text-muted font-normal ml-2">(if started now)</span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-human">👤 Human</span>
            <span className="font-bold text-human">
              {currentTeamBalance.human}
            </span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-ai">🤖 AI</span>
            <span className="font-bold text-ai">
              {currentTeamBalance.ai}
            </span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-aligned">⚖️ Aligned</span>
            <span className="font-bold text-aligned">
              {currentTeamBalance.aligned}
            </span>
          </div>
        </div>

        {/* Game Start Status */}
        {playerCount < 4 && (
          <div className="bg-amber/10 border border-amber/30 rounded-lg p-2">
            <div className="text-xs text-amber">
              ⏳ Need {4 - playerCount} more players to start
            </div>
          </div>
        )}
      </div>
    </div>
  );
};