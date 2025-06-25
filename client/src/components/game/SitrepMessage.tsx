import React from 'react';
import { ChatMessage, GameState } from '../../types';

interface SitrepMessageProps {
  message: ChatMessage;
  gameState: GameState;
}

export const SitrepMessage: React.FC<SitrepMessageProps> = ({ message, gameState }) => {
  const metadata = message.metadata;
  const nightActions = metadata?.nightActions || [];
  const headcount = metadata?.playerHeadcount || {
    humans: gameState.players.filter(p => p.isAlive && p.alignment !== 'ALIGNED').length,
    aligned: gameState.players.filter(p => p.isAlive && p.alignment === 'ALIGNED').length,
    dead: gameState.players.filter(p => !p.isAlive).length
  };
  const crisisEvent = metadata?.crisisEvent || gameState.crisisEvent;

  return (
    <div className="flex gap-3 p-3 bg-background-tertiary border border-border/30 rounded-lg mb-2">
      <div className="w-8 h-8 rounded-full bg-ai text-white flex items-center justify-center text-sm">🤖</div>
      <div className="flex-1">
        <span className="text-ai font-mono font-bold text-sm">Loebmate</span>
        <div className="text-text-primary bg-background-quaternary border border-border/20 rounded-lg p-3 mt-2 font-mono text-sm">
          <strong>Good morning, team. Here's the SITREP.</strong><br/><br/>
          
          <strong>NIGHT {gameState.dayNumber - 1} ACTIVITY LOG:</strong><br/>
          {nightActions.length > 0 ? (
            nightActions.map((action, index) => (
              <span key={index}>
                • {action.description}<br/>
              </span>
            ))
          ) : (
            <span>• No significant activity detected.<br/></span>
          )}
          <br/>
          
          <strong>HR HEADCOUNT:</strong><br/>
          • <strong>{headcount.humans} Human Life-signs Detected</strong><br/>
          • <strong>{headcount.aligned} Aligned Agents Active</strong> 🤖<br/>
          {headcount.dead > 0 && (
            <>• <strong>{headcount.dead} Personnel Deactivated</strong> 👻<br/></>
          )}
          <br/>
          
          {crisisEvent && (
            <>
              <strong>INCIDENT: {crisisEvent.title}</strong><br/>
              • <span className="text-danger font-bold">[HIGH ALERT]</span> {crisisEvent.description}<br/>
            </>
          )}
        </div>
      </div>
    </div>
  );
};