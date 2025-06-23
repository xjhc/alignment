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
    <div className="pulse-check">
      <div className="pulse-header">💭 PULSE CHECK</div>
      <div className="pulse-question">"{question}"</div>
      <div className="pulse-responses">
        {Object.entries(pulseCheckResponses).map(([playerName, response]) => (
          <div key={playerName} className="pulse-response">
            <strong>{playerName}:</strong> "{response}"
          </div>
        ))}
      </div>
    </div>
  );
};