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
    <div className="chat-message-compact">
      <div className="message-avatar loebmate">🤖</div>
      <div className="message-content">
        <span className="message-author loebmate-name">Loebmate</span>
        <div className="message-body loebmate-message">
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
              • <span className="high-alert">[HIGH ALERT]</span> {crisisEvent.description}<br/>
            </>
          )}
        </div>
      </div>
    </div>
  );
};