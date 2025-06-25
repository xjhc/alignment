import React from 'react';
import { ChatMessage, GameState } from '../../types';

interface PulseCheckMessageProps {
  message: ChatMessage;
  gameState: GameState;
}

export const PulseCheckMessage: React.FC<PulseCheckMessageProps> = ({ message }) => {
  // Handle both old and new formats
  const pulseCheckResponses = message.metadata?.pulseCheckResponses || message.metadata?.player_responses || {};
  const question = message.metadata?.question || message.message;
  const totalResponses = message.metadata?.total_responses || Object.keys(pulseCheckResponses).length;

  return (
    <div className="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4 mb-3">
      <div className="text-blue-400 font-mono font-bold text-sm mb-2">💭 PULSE CHECK RESULTS</div>
      {question && (
        <div className="text-text-primary font-medium mb-3 italic">"{question}"</div>
      )}
      <div className="text-sm text-gray-400 mb-3">
        {totalResponses} response{totalResponses !== 1 ? 's' : ''} received:
      </div>
      <div className="space-y-2">
        {Object.entries(pulseCheckResponses).map(([playerName, response]) => (
          <div key={playerName} className="bg-background-secondary/50 border border-border/30 rounded px-3 py-2 text-sm">
            <strong className="text-primary">{playerName}:</strong> 
            <span className="text-text-secondary ml-2">"{String(response)}"</span>
          </div>
        ))}
      </div>
      {Object.keys(pulseCheckResponses).length === 0 && (
        <div className="text-gray-500 text-sm italic">No responses received during the pulse check period.</div>
      )}
    </div>
  );
};