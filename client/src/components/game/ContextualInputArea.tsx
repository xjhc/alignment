import React from 'react';
import { GameState, Player } from '../../types';
import { VoteUI } from './VoteUI';
import { NightActionSelection } from './NightActionSelection';
import { PulseCheckInput } from './PulseCheckInput';

interface ContextualInputAreaProps {
  gameState: GameState;
  localPlayer: Player;
  chatInput: string;
  setChatInput: (value: string) => void;
  handleSendMessage: () => void;
  handleKeyDown: (e: React.KeyboardEvent) => void;
  isConnected: boolean;

  // Voting props
  selectedNominee: string;
  setSelectedNominee: (value: string) => void;
  selectedVote: 'GUILTY' | 'INNOCENT' | '';
  setSelectedVote: (value: 'GUILTY' | 'INNOCENT' | '') => void;
  handleNominate: () => Promise<void>;
  handleVote: () => Promise<void>;
  handlePulseCheck: (response: string) => Promise<void>;

  // Night action props
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

export const ContextualInputArea: React.FC<ContextualInputAreaProps> = ({
  gameState,
  localPlayer,
  chatInput,
  setChatInput,
  handleSendMessage,
  handleKeyDown,
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
  switch (gameState.phase.type) {
    case 'NOMINATION':
    case 'VERDICT':
      return (
        <VoteUI
          gameState={gameState}
          localPlayer={localPlayer}
          selectedNominee={selectedNominee}
          setSelectedNominee={setSelectedNominee}
          selectedVote={selectedVote}
          setSelectedVote={setSelectedVote}
          handleNominate={handleNominate}
          handleVote={handleVote}
        />
      );

    case 'NIGHT':
      return (
        <NightActionSelection
          gameState={gameState}
          localPlayer={localPlayer}
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
      );

    case 'PULSE_CHECK':
      return (
        <PulseCheckInput
          handlePulseCheck={handlePulseCheck}
          localPlayerName={localPlayer.name}
        />
      );

    case 'DISCUSSION':
    default:
      return (
        <div className="chat-input-area">
          <input
            className="chat-input"
            type="text"
            placeholder={
              !isConnected
                ? 'Reconnecting...'
                : gameState.phase.type === 'DISCUSSION'
                  ? `Message #war-room as ${localPlayer.name}`
                  : `Channel locked during ${gameState.phase.type}`
            }
            value={chatInput}
            onChange={(e) => setChatInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={!isConnected || gameState.phase.type !== 'DISCUSSION'}
          />
          <button onClick={handleSendMessage} style={{ display: 'none' }}>Send</button>
        </div>
      );
  }
};

