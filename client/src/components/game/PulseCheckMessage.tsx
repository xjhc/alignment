import React from 'react';
import { ChatMessage, GameState } from '../../types';
import styles from './CommsPanel.module.css';

interface PulseCheckMessageProps {
  message: ChatMessage;
  gameState: GameState;
}

export const PulseCheckMessage: React.FC<PulseCheckMessageProps> = ({ message }) => {
  const pulseCheckResponses = message.metadata?.pulseCheckResponses || {};
  const question = message.message;

  return (
    <div className={styles.pulseCheck}>
      <div className={styles.pulseHeader}>💭 PULSE CHECK</div>
      <div className={styles.pulseQuestion}>"{question}"</div>
      <div className={styles.pulseResponses}>
        {Object.entries(pulseCheckResponses).map(([playerName, response]) => (
          <div key={playerName} className={styles.pulseResponse}>
            <strong>{playerName}:</strong> "{response}"
          </div>
        ))}
      </div>
    </div>
  );
};