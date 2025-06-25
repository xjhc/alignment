import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { IdentityCard } from './IdentityCard';
import { ThreatMeter } from './ThreatMeter';
import { ObjectiveCard } from './ObjectiveCard';
import { AbilityCard } from './AbilityCard';

export const PlayerHUD: React.FC = () => {
  const { gameState, viewedPlayer } = useGameContext();

  if (!viewedPlayer) {
    return null;
  }
  const aiEquity = viewedPlayer.aiEquity || 0;

  return (
    <aside className="flex flex-col bg-background-secondary overflow-hidden">
      <IdentityCard localPlayer={viewedPlayer} />

      <div className="flex-grow p-4 overflow-y-auto flex flex-col gap-4">
        {/* Only show threat meter if player is Human */}
        {viewedPlayer.alignment === 'HUMAN' && (
          <ThreatMeter
            tokens={viewedPlayer.tokens}
            aiEquity={aiEquity}
          />
        )}

        <div className="animate-[fadeIn_0.3s_ease]">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs font-bold text-text-muted uppercase">📋 OBJECTIVES</span>
          </div>

          <ObjectiveCard
            type="Team Objective"
            name={viewedPlayer.alignment === 'HUMAN' ? "Containment Protocol" : "Achieve Singularity"}
            description={viewedPlayer.alignment === 'HUMAN'
              ? "Identify and vote to deactivate the Original AI."
              : "Convert enough humans to achieve AI dominance."
            }
          />

          {viewedPlayer.personalKPI && (
            <ObjectiveCard
              type="Personal KPI"
              name={viewedPlayer.personalKPI.type}
              description={viewedPlayer.personalKPI.description}
              progressText={
                `Progress: ${viewedPlayer.personalKPI.progress || 0}/${viewedPlayer.personalKPI.target || 1} ${viewedPlayer.personalKPI.isCompleted ? '✓' : ''}`
              }
              isPrivate={true}
            />
          )}

          {gameState.corporateMandate && (
            <ObjectiveCard
              type="Mandate"
              name={gameState.corporateMandate.name}
              description={gameState.corporateMandate.description}
            />
          )}
        </div>

        <AbilityCard localPlayer={viewedPlayer} />

        {viewedPlayer.lastNightAction && (
          <div className="animate-[fadeIn_0.3s_ease]">
            <div className="flex justify-between items-center mb-2">
              <span className="text-xs font-bold text-text-muted uppercase">🌙 LAST NIGHT'S ACTION</span>
            </div>
            <div className="flex flex-col gap-1">
              <div className="flex items-center gap-1.5 px-2 py-1.5 bg-background-tertiary border border-amber-500 rounded-md bg-amber-500/10">
                <span className="text-xs w-4 text-center">➡️</span>
                <span className="font-medium text-xs text-text-primary flex-grow">{viewedPlayer.lastNightAction.type}</span>
                {viewedPlayer.lastNightAction.targetId && (
                  <span className="ml-auto text-xs font-bold text-blue-500 font-mono">
                    TARGET: {gameState.players.find(p => p.id === viewedPlayer.lastNightAction?.targetId)?.name || 'Unknown'}
                  </span>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
};