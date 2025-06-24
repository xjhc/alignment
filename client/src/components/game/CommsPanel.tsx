import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { usePhaseTimer } from '../../hooks/usePhaseTimer';
import { ContextualInputArea } from './ContextualInputArea';
import { SitrepMessage } from './SitrepMessage';
import { VoteResultMessage } from './VoteResultMessage';
import { PulseCheckMessage } from './PulseCheckMessage';
import styles from './CommsPanel.module.css';

export const CommsPanel: React.FC = () => {
  const { gameState, localPlayer } = useGameContext();
  const timeRemaining = usePhaseTimer(gameState.phase);

  if (!localPlayer) {
    return <div>Loading...</div>;
  }
  
  const getPhaseDisplayName = (phaseType: string) => {
    switch (phaseType) {
      case 'SITREP': return 'SITREP';
      case 'PULSE_CHECK': return 'PULSE CHECK';
      case 'DISCUSSION': return 'DISCUSSION';
      case 'NOMINATION': return 'NOMINATION';
      case 'TRIAL': return 'TRIAL';
      case 'VERDICT': return 'VERDICT';
      case 'NIGHT': return 'NIGHT PHASE';
      case 'GAME_OVER': return 'GAME OVER';
      default: return phaseType;
    }
  };

  const phaseName = getPhaseDisplayName(gameState.phase.type);

  const getPhaseClass = (phaseType: string) => {
    switch (phaseType) {
      case 'DISCUSSION':
        return styles.discussion;
      case 'NOMINATION':
        return styles.nomination;
      case 'TRIAL':
        return styles.trial;
      case 'VERDICT':
        return styles.verdict;
      case 'NIGHT':
        return styles.night;
      case 'PULSE_CHECK':
        return styles.pulseCheck;
      default:
        return styles.sitrep;
    }
  };

  const chatLogRef = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    if (chatLogRef.current) {
      chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
    }
  }, [gameState.chatMessages]);

  return (
    <section className={styles.panelCenter}>
      <header className={styles.chatHeader}>
        <div className={styles.channelInfo}>
          <span className={styles.channelName}>#war-room</span>
          <span className={styles.channelTopic}>Emergency ops • All comms logged</span>
        </div>
        <div className={styles.timerSection}>
          <div className={`${styles.phaseIndicator} ${getPhaseClass(gameState.phase.type)}`}>{phaseName}</div>
          <div className={styles.timerDisplay}>
            <div className={styles.timerLabel}>ENDS IN</div>
            <div className={`${styles.timerValue} ${styles.pulse}`}>{timeRemaining}</div>
          </div>
        </div>
      </header>

      <div className={styles.chatLog} ref={chatLogRef}>
        {(!gameState.chatMessages || gameState.chatMessages.length === 0) ? (
          <div className="empty-chat-message">
            <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>
              No messages yet. Waiting for system initialization...
            </span>
          </div>
        ) : null}
        {gameState.chatMessages.map((msg, index) => {
            // Render specialized system messages
            if (msg.isSystem && msg.type === 'SITREP') {
              return (
                <div key={msg.id || index} className="animate-slide-in-up" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <SitrepMessage 
                    message={msg} 
                    gameState={gameState} 
                  />
                </div>
              );
            }
            
            if (msg.isSystem && msg.type === 'VOTE_RESULT') {
              return (
                <div key={msg.id || index} className="animate-slide-in-up" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <VoteResultMessage 
                    message={msg} 
                    gameState={gameState} 
                  />
                </div>
              );
            }
            
            if (msg.isSystem && msg.type === 'PULSE_CHECK') {
              return (
                <div key={msg.id || index} className="animate-slide-in-up" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <PulseCheckMessage 
                    message={msg} 
                    gameState={gameState} 
                  />
                </div>
              );
            }
            
            // Default chat message rendering
            const getMessageAvatar = (message: any) => {
              if (message.isSystem) return '🤖';
              const player = gameState.players.find((p: any) => p.name === message.playerName);
              if (!player?.isAlive) return '👻';
              return player?.avatar || '👤';
            };

            return (
              <div 
                key={msg.id || index} 
                className={`${styles.chatMessageCompact} ${msg.isSystem ? styles.system : ''} animate-slide-in-left`}
                style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}
              >
                <div className={`${styles.messageAvatar} ${msg.isSystem ? styles.loebmate : ''}`}>
                  {getMessageAvatar(msg)}
                </div>
                <div className={styles.messageContent}>
                  <span className={`${styles.messageAuthor} ${msg.isSystem ? styles.loebmateName : ''}`}>
                    {msg.playerName}
                  </span>
                  <div className={styles.messageBody}>{msg.message}</div>
                </div>
              </div>
            );
          })
        }
      </div>

      <ContextualInputArea />
    </section>
  );
};