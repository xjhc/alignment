import React from 'react';
import { ChatMessage } from '../../types';

interface IncitingIncidentMessageProps {
  message: ChatMessage;
}

export const IncitingIncidentMessage: React.FC<IncitingIncidentMessageProps> = ({ message }) => {
  return (
    <div className="flex gap-3 p-4 mb-3 border-2 border-danger bg-danger/5 rounded-lg animation-scale-in">
      <div className="w-8 h-8 rounded-full bg-danger text-white flex items-center justify-center text-sm font-bold">
        ⚠️
      </div>
      <div className="flex-1">
        <div className="flex items-center justify-between mb-2">
          <span className="text-danger font-mono font-bold text-sm">SECURITY ALERT</span>
          <span className="text-xs bg-danger text-white px-2 py-1 rounded font-bold tracking-wider">
            SEV-1
          </span>
        </div>
        
        <div className="bg-background-primary border border-danger/30 rounded-lg p-4 font-mono text-sm">
          <div className="grid grid-cols-1 gap-1 mb-3 text-xs">
            <div className="flex">
              <span className="text-text-muted font-bold w-16">From:</span>
              <span className="text-text-primary">{message.metadata?.from || 'security@loebian.com'}</span>
            </div>
            <div className="flex">
              <span className="text-text-muted font-bold w-16">To:</span>
              <span className="text-text-primary">{message.metadata?.to || '#all-senior-staff'}</span>
            </div>
            <div className="flex">
              <span className="text-text-muted font-bold w-16">Subject:</span>
              <span className="text-danger font-bold">{message.metadata?.subject || message.message}</span>
            </div>
          </div>
          
          <hr className="border-border my-3" />
          
          <div className="text-text-primary leading-relaxed whitespace-pre-line">
            {message.metadata?.body || message.message}
          </div>
        </div>
      </div>
    </div>
  );
};