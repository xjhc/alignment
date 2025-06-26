import React from 'react';
import { ChatMessage } from '../../types';

interface PulseCheckSubmissionMessageProps {
  message: ChatMessage;
}

export const PulseCheckSubmissionMessage: React.FC<PulseCheckSubmissionMessageProps> = ({ message }) => {
  const playerName = message.playerName;
  const response = message.message;

  return (
    <div className="bg-cyan-900/20 border border-cyan-500/30 rounded-lg p-3 mb-2">
      <div className="flex items-center gap-2 mb-2">
        <div className="text-cyan-400 font-mono font-bold text-xs">💭 PULSE CHECK RESPONSE</div>
        <div className="text-cyan-300 text-xs font-medium">{playerName}</div>
      </div>
      <div className="text-text-secondary text-sm italic">
        "{response}"
      </div>
    </div>
  );
};