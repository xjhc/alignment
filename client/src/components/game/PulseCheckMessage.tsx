import React from 'react';
import { ChatMessage, GameState } from '../../types';

interface PulseCheckMessageProps {
  message: ChatMessage;
  gameState: GameState;
}

export const PulseCheckMessage: React.FC<PulseCheckMessageProps> = ({ message }) => {
  const pulseCheckResponses = message.metadata?.pulseCheckResponses || {};
  const question = message.message;

  return (
    <div className="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4 mb-3">
      <div className="text-blue-400 font-mono font-bold text-sm mb-2">💭 PULSE CHECK</div>
      <div className="text-text-primary font-medium mb-3 italic">"{question}"</div>
      <div className="space-y-2">
        {Object.entries(pulseCheckResponses).map(([playerName, response]) => (
          <div key={playerName} className="bg-background-secondary/50 border border-border/30 rounded px-3 py-2 text-sm">
            <strong className="text-primary">{playerName}:</strong> <span className="text-text-secondary">"{response}"</span>
          </div>
        ))}
      </div>
    </div>
  );
};