import React from 'react';
import { GameState, Player } from '../../types';

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
}) => {
  const phaseName = getPhaseDisplayName(gameState.phase.type);
  const timeRemaining = formatTimeRemaining(gameState.phase);

  const getPhaseClass = (phaseType: string) => {
    switch (phaseType) {
      case 'DISCUSSION':
      case 'NOMINATION':
      case 'VERDICT':
        return 'discussion';
      case 'NIGHT':
        return 'night';
      default:
        return 'sitrep';
    }
  };

  const chatLogRef = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    if (chatLogRef.current) {
      chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
    }
  }, [gameState.chatMessages]);

  return (
    <section className="panel-center">
      <header className="chat-header">
        <div className="channel-info">
          <span className="channel-name">#war-room</span>
          <span className="channel-topic">Emergency ops • All comms logged</span>
        </div>
        <div className="timer-section">
          <div className={`phase-indicator ${getPhaseClass(gameState.phase.type)}`}>{phaseName}</div>
          <div className="timer-display">
            <div className="timer-label">ENDS IN</div>
            <div className="timer-value pulse">{timeRemaining}</div>
          </div>
        </div>
      </header>

      <div className="chat-log" ref={chatLogRef}>
        {isChatHistoryLoading ? (
          <div className="loading">Loading chat history...</div>
        ) : (
          gameState.chatMessages.map((msg, index) => (
            <div key={msg.id || index} className={`chat-message-compact ${msg.isSystem ? 'system' : ''}`}>
              <div className={`message-avatar ${msg.isSystem ? 'loebmate' : ''}`}>
                {msg.isSystem ? '🤖' : '👤'}
              </div>
              <div className="message-content">
                <span className={`message-author ${msg.isSystem ? 'loebmate-name' : ''}`}>
                  {msg.playerName}
                </span>
                <div className="message-body">{msg.message}</div>
              </div>
            </div>
          ))
        )}
      </div>

      <div className="chat-input-area">
        <input
          className="chat-input"
          type="text"
          placeholder={
            !isConnected
              ? 'Reconnecting...'
              : gameState.phase.type === 'DISCUSSION'
                ? `Message #war-room as ${localPlayer.name}`
                : `Channel locked during ${phaseName}`
          }
          value={chatInput}
          onChange={(e) => setChatInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={!isConnected || gameState.phase.type !== 'DISCUSSION'}
        />
        <button onClick={handleSendMessage} style={{ display: 'none' }}>Send</button>
      </div>
    </section>
  );
};