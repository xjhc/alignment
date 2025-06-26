import React from 'react';
import { ChatMessage } from '../../types';

interface LoebmateMessageProps {
  message: ChatMessage;
}

export const LoebmateMessage: React.FC<LoebmateMessageProps> = ({ message }) => {
  return (
    <div className="flex gap-3 p-3 bg-ai/10 border border-ai/20 rounded-lg mb-2 animation-fade-in">
      <div className="w-8 h-8 rounded-full bg-ai text-white flex items-center justify-center text-sm">
        🤖
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-ai font-mono font-bold text-sm">
            {message.playerName || 'Loebmate'}
          </span>
          <span className="text-xs bg-ai/20 text-ai px-2 py-1 rounded font-bold tracking-wider">
            BOT
          </span>
        </div>
        
        <div className="text-text-primary bg-background-quaternary border border-border/20 rounded-lg p-3 font-mono text-sm leading-relaxed whitespace-pre-line">
          {message.metadata?.body || message.message}
        </div>
      </div>
    </div>
  );
};