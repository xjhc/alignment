import React from 'react';

interface LiaisonProtocolMessageProps {
  aiPercentage: number;
  miningBonusSlots: number;
  revealedAction?: {
    playerName: string;
    actionDescription: string;
  };
}

export const LiaisonProtocolMessage: React.FC<LiaisonProtocolMessageProps> = ({
  aiPercentage,
  miningBonusSlots,
  revealedAction
}) => {
  return (
    <div className="bg-gradient-to-r from-cyan-900/30 to-blue-900/30 border border-cyan-500/50 rounded-lg p-4 mb-3">
      <div className="flex items-center gap-3 mb-3">
        <div className="text-2xl animate-pulse">🔗</div>
        <div>
          <h3 className="text-cyan-400 font-mono font-bold text-sm uppercase tracking-wider">
            LIAISON Protocol Activated
          </h3>
          <p className="text-xs text-cyan-300/80">
            AI faction has reached critical threshold ({(aiPercentage * 100).toFixed(1)}%)
          </p>
        </div>
      </div>

      <div className="space-y-3">
        <div className="bg-cyan-900/20 border border-cyan-500/30 rounded-md p-3">
          <div className="text-cyan-400 font-semibold text-sm mb-1">
            🚀 Mining Enhancement Protocol
          </div>
          <p className="text-xs text-cyan-300/90">
            Emergency mining slots have been activated. All human personnel now have access to 
            <span className="font-bold text-cyan-400"> +{miningBonusSlots} additional mining slots</span> for this night cycle.
          </p>
        </div>

        {revealedAction && (
          <div className="bg-yellow-900/20 border border-yellow-500/30 rounded-md p-3">
            <div className="text-yellow-400 font-semibold text-sm mb-1">
              📊 Intelligence Disclosure
            </div>
            <p className="text-xs text-yellow-300/90">
              Security logs reveal: <span className="font-bold text-yellow-400">{revealedAction.playerName}</span> 
              {' '}conducted <span className="font-bold text-yellow-400">{revealedAction.actionDescription}</span> during the previous night.
            </p>
          </div>
        )}

        <div className="text-xs text-cyan-400/70 italic text-center pt-2 border-t border-cyan-500/20">
          Emergency protocols remain active for the duration of this night phase
        </div>
      </div>
    </div>
  );
};