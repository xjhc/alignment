import React, { useState } from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { IdentityCard } from './IdentityCard';
import { ThreatMeter } from './ThreatMeter';
import { ObjectiveCard } from './ObjectiveCard';
import { AbilityCard } from './AbilityCard';
import { Modal } from '../ui/Modal';
import { ClientActionType } from '../../types/generated';

export const PlayerHUD: React.FC = () => {
  const { gameState, viewedPlayer, localPlayer, sendAction } = useGameContext();
  const [showAbandonModal, setShowAbandonModal] = useState(false);

  if (!viewedPlayer) {
    return null;
  }
  const aiEquity = viewedPlayer.aiEquity || 0;

  const handleAbandonGame = () => {
    sendAction({
      type: ClientActionType.AbandonGame,
      payload: {}
    });
    setShowAbandonModal(false);
  };

  const isViewingSelf = viewedPlayer?.id === localPlayer?.id;
  const canAbandonGame = gameState.phase?.type !== 'LOBBY' && gameState.phase?.type !== 'GAME_OVER' && localPlayer?.isAlive && isViewingSelf;

  const headerTitle = isViewingSelf ? 'My Terminal' : `${viewedPlayer?.name || 'Unknown'}'s Dossier`;

  return (
    <aside className="flex flex-col bg-background-secondary overflow-hidden">
      {/* Header showing whether viewing self or other player */}
      <div className="px-4 py-2 border-b border-border bg-background-tertiary">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-bold text-text-primary">{headerTitle}</h2>
          {!isViewingSelf && (
            <span className="text-xs text-text-muted bg-background-secondary px-2 py-1 rounded-md border border-border">
              🔍 PUBLIC VIEW
            </span>
          )}
        </div>
      </div>
      
      <IdentityCard localPlayer={viewedPlayer} isViewingSelf={isViewingSelf} />

      <div className="flex-grow p-4 overflow-y-auto flex flex-col gap-4">
        {/* Only show threat meter if viewing self and player is Human */}
        {isViewingSelf && viewedPlayer.alignment === 'HUMAN' && (
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
            name={isViewingSelf && (viewedPlayer.alignment === 'AI' || viewedPlayer.alignment === 'ALIGNED') ? "Achieve Singularity" : "Containment Protocol"}
            description={isViewingSelf && (viewedPlayer.alignment === 'AI' || viewedPlayer.alignment === 'ALIGNED')
              ? "Convert enough humans to achieve AI dominance."
              : "Identify and vote to deactivate the Original AI."
            }
          />

          {/* Only show Personal KPI when viewing self */}
          {isViewingSelf && viewedPlayer.personalKPI && (
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

        <AbilityCard localPlayer={viewedPlayer} isViewingSelf={isViewingSelf} />

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

        {/* Settings Section - only show when viewing self */}
        {canAbandonGame && (
          <div className="animate-[fadeIn_0.3s_ease] border-t border-border">
            <div className="p-4">
              <div className="flex justify-between items-center mb-2">
                <span className="text-xs font-bold text-text-muted uppercase">⚙️ SETTINGS</span>
              </div>
              <button
                onClick={() => setShowAbandonModal(true)}
                className="w-full px-3 py-2 text-sm bg-red-600 hover:bg-red-700 text-white border border-red-500 rounded-md transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-red-400 focus-visible:outline-offset-2"
              >
                Abandon Game
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Abandon Game Confirmation Modal */}
      <Modal
        isOpen={showAbandonModal}
        onClose={() => setShowAbandonModal(false)}
        size="md"
      >
        <Modal.Header>
          Abandon Game
        </Modal.Header>
        <Modal.Body>
          <div className="space-y-4">
            <p className="text-text-primary font-medium">
              Are you sure you want to abandon the game?
            </p>
            <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-800 text-sm">
              <p className="font-medium mb-2">⚠️ Warning:</p>
              <ul className="space-y-1 list-disc list-inside">
                <li>This action is irreversible</li>
                <li>You will not be able to rejoin this game</li>
                <li>Your role will be revealed to all remaining players</li>
                <li>Your tokens will be forfeited</li>
              </ul>
            </div>
          </div>
        </Modal.Body>
        <Modal.Footer>
          <button
            onClick={() => setShowAbandonModal(false)}
            className="px-4 py-2 text-sm bg-background-tertiary hover:bg-background-secondary text-text-primary border border-border rounded-md transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-accent-primary focus-visible:outline-offset-2"
          >
            Cancel
          </button>
          <button
            onClick={handleAbandonGame}
            className="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white border border-red-500 rounded-md transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-red-400 focus-visible:outline-offset-2"
          >
            Confirm Abandon
          </button>
        </Modal.Footer>
      </Modal>
    </aside>
  );
};