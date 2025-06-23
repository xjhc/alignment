import React from 'react';
import { GameState, Player } from '../../types';
import { ContextualInputArea } from './ContextualInputArea';
import { SitrepMessage } from './SitrepMessage';
import { VoteResultMessage } from './VoteResultMessage';
import { PulseCheckMessage } from './PulseCheckMessage';
import styles from './CommsPanel.module.css';

interface CommsPanelProps {
  gameState: GameState;
  localPlayer: Player;
  chatInput: string;
  setChatInput: (value: string) => void;
  handleSendMessage: () => void;
  handleKeyDown: (e: React.KeyboardEvent) => void;
  getPhaseDisplayName: (phaseType: string) => string;
  formatTimeRemaining: (phase: GameState['phase']) => string;
  isChatHistoryLoading: boolean;
  isConnected: boolean;

  // Props for Part 3 voting UI.
  selectedNominee: string;
  setSelectedNominee: (value: string) => void;
  selectedVote: 'GUILTY' | 'INNOCENT' | '';
  setSelectedVote: (value: 'GUILTY' | 'INNOCENT' | '') => void;
  handleNominate: () => Promise<void>;
  handleVote: () => Promise<void>;
  handlePulseCheck: (response: string) => Promise<void>;

  // Props for Part 3 night action UI.
  conversionTarget: string;
  setConversionTarget: (value: string) => void;
  miningTarget: string;
  setMiningTarget: (value: string) => void;
  handleConversionAttempt: () => Promise<void>;
  handleMineTokens: () => Promise<void>;
  handleUseAbility: () => Promise<void>;
  canPlayerAffordAbility: (playerId: string) => boolean;
  isValidNightActionTarget: (playerId: string, targetId: string, actionType: string) => boolean;
}

export const CommsPanel: React.FC<CommsPanelProps> = ({
  gameState,
  localPlayer,
  chatInput,
  setChatInput,
  handleSendMessage,
  handleKeyDown,
  getPhaseDisplayName,
  formatTimeRemaining,
  isChatHistoryLoading,
  isConnected,
  selectedNominee,
  setSelectedNominee,
  selectedVote,
  setSelectedVote,
  handleNominate,
  handleVote,
  handlePulseCheck,
  conversionTarget,
  setConversionTarget,
  miningTarget,
  setMiningTarget,
  handleConversionAttempt,
  handleMineTokens,
  handleUseAbility,
  canPlayerAffordAbility,
  isValidNightActionTarget,
}) => {
  const phaseName = getPhaseDisplayName(gameState.phase.type);
  const timeRemaining = formatTimeRemaining(gameState.phase);

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
        {isChatHistoryLoading ? (
          <div className={styles.loading}>Loading chat history...</div>
        ) : (
          gameState.chatMessages.map((msg, index) => {
            // Render specialized system messages
            if (msg.isSystem && msg.type === 'SITREP') {
              return (
                <SitrepMessage 
                  key={msg.id || index} 
                  message={msg} 
                  gameState={gameState} 
                />
              );
            }
            
            if (msg.isSystem && msg.type === 'VOTE_RESULT') {
              return (
                <VoteResultMessage 
                  key={msg.id || index} 
                  message={msg} 
                  gameState={gameState} 
                />
              );
            }
            
            if (msg.isSystem && msg.type === 'PULSE_CHECK') {
              return (
                <PulseCheckMessage 
                  key={msg.id || index} 
                  message={msg} 
                  gameState={gameState} 
                />
              );
            }
            
            // Default chat message rendering
            return (
              <div key={msg.id || index} className={`${styles.chatMessageCompact} ${msg.isSystem ? styles.system : ''}`}>
                <div className={`${styles.messageAvatar} ${msg.isSystem ? styles.loebmate : ''}`}>
                  {msg.isSystem ? '🤖' : '👤'}
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
        )}
      </div>

      <ContextualInputArea
        gameState={gameState}
        localPlayer={localPlayer}
        chatInput={chatInput}
        setChatInput={setChatInput}
        handleSendMessage={handleSendMessage}
        handleKeyDown={handleKeyDown}
        isConnected={isConnected}
        selectedNominee={selectedNominee}
        setSelectedNominee={setSelectedNominee}
        selectedVote={selectedVote}
        setSelectedVote={setSelectedVote}
        handleNominate={handleNominate}
        handleVote={handleVote}
        handlePulseCheck={handlePulseCheck}
        conversionTarget={conversionTarget}
        setConversionTarget={setConversionTarget}
        miningTarget={miningTarget}
        setMiningTarget={setMiningTarget}
        handleConversionAttempt={handleConversionAttempt}
        handleMineTokens={handleMineTokens}
        handleUseAbility={handleUseAbility}
        canPlayerAffordAbility={canPlayerAffordAbility}
        isValidNightActionTarget={isValidNightActionTarget}
      />
    </section>
  );
};