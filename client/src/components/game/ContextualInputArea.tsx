import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { useGameActions } from '../../hooks/useGameActions';
import { VoteUI } from './VoteUI';
import { NightActionSelection } from './NightActionSelection';
import { PulseCheckInput } from './PulseCheckInput';
// import styles from './CommsPanel.module.css';

interface ContextualInputAreaProps {
  // No props needed - everything comes from context
}

export const ContextualInputArea: React.FC<ContextualInputAreaProps> = () => {
  const { gameState, localPlayer, isConnected } = useGameContext();
  const gameActions = useGameActions();
  
  if (!localPlayer) return null;
  
  switch (gameState.phase.type) {
    case 'NOMINATION':
    case 'VERDICT':
      return (
        <VoteUI />
      );

    case 'NIGHT':
      return (
        <NightActionSelection />
      );

    case 'PULSE_CHECK':
      return (
        <PulseCheckInput
          handlePulseCheck={gameActions.handlePulseCheck}
          localPlayerName={localPlayer.name}
        />
      );

    case 'DISCUSSION':
    default:
      return (
        <div className="flex-1 flex items-center gap-3 p-3">
          <input
            className="flex-1 bg-background-secondary border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 disabled:opacity-50 disabled:cursor-not-allowed"
            type="text"
            placeholder={
              !isConnected
                ? 'Reconnecting...'
                : gameState.phase.type.toUpperCase() === 'DISCUSSION'
                  ? `Message #war-room`
                  : `Channel locked during ${gameState.phase.type}`
            }
            value={gameActions.chatInput}
            onChange={(e) => gameActions.setChatInput(e.target.value)}
            onKeyDown={gameActions.handleKeyDown}
            disabled={!isConnected || gameState.phase.type.toUpperCase() !== 'DISCUSSION'}
          />
        </div>
      );
  }
};

