import { useState, useCallback } from 'react';
import { useGameContext } from '../contexts/GameContext';
import { useGameEngineContext } from '../contexts/GameEngineContext';
import { ClientAction } from '../types';

export function useGameActions() {
  const { gameState, localPlayerId, sendAction, isConnected } = useGameContext();
  const { canPlayerAffordAbility, isValidNightActionTarget } = useGameEngineContext();
  
  const [chatInput, setChatInput] = useState('');
  const [selectedNominee, setSelectedNominee] = useState<string>('');
  const [selectedVote, setSelectedVote] = useState<'GUILTY' | 'INNOCENT' | ''>('');
  const [conversionTarget, setConversionTarget] = useState<string>('');
  const [miningTarget, setMiningTarget] = useState<string>('');

  const localPlayer = gameState.players.find(p => p.id === localPlayerId);

  const getPhaseDisplayName = useCallback((phaseType: string) => {
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
  }, []);


  const handleSendMessage = useCallback(async () => {
    if (!chatInput.trim() || !localPlayer || !isConnected) {
      return;
    }

    try {
      const action: ClientAction = {
        type: 'POST_CHAT_MESSAGE',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          message: chatInput.trim(),
          player_name: localPlayer.name,
        },
      };

      sendAction(action);
      setChatInput('');
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  }, [chatInput, localPlayer, isConnected, gameState.id, localPlayerId, sendAction]);

  const handleMineTokens = useCallback(async () => {
    if (!localPlayer || !miningTarget || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          action_type: 'MINE_TOKENS',
          target_player_id: miningTarget,
        },
      };

      sendAction(action);
      setMiningTarget('');
    } catch (error) {
      console.error('Failed to mine tokens:', error);
    }
  }, [localPlayer, miningTarget, isConnected, gameState.id, localPlayerId, sendAction]);

  const handleUseAbility = useCallback(async (targetId?: string) => {
    if (!localPlayer || !canPlayerAffordAbility(localPlayerId) || !isConnected) {
      return;
    }

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          action_type: 'USE_ABILITY',
          ability_type: localPlayer.role?.type || 'UNKNOWN',
          target_id: targetId,
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to use ability:', error);
    }
  }, [localPlayer, canPlayerAffordAbility, localPlayerId, isConnected, gameState.id, sendAction]);

  const handleProjectMilestones = useCallback(async () => {
    if (!localPlayer || !isConnected) {
      return;
    }

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          action_type: 'PROJECT_MILESTONES',
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to submit project milestones:', error);
    }
  }, [localPlayer, isConnected, gameState.id, localPlayerId, sendAction]);

  const handleConversionAttempt = useCallback(async () => {
    if (!localPlayer || !conversionTarget || !isConnected) return;

    // Validate that this is a valid target for conversion
    if (!isValidNightActionTarget(localPlayerId, conversionTarget, 'ATTEMPT_CONVERSION')) {
      console.warn('Invalid conversion target');
      return;
    }

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          action_type: 'ATTEMPT_CONVERSION',
          target_player_id: conversionTarget,
        },
      };

      sendAction(action);
      setConversionTarget('');
    } catch (error) {
      console.error('Failed to attempt conversion:', error);
    }
  }, [localPlayer, conversionTarget, isConnected, isValidNightActionTarget, localPlayerId, gameState.id, sendAction]);

  const handleNominate = useCallback(async () => {
    if (!selectedNominee || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_VOTE',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          target_id: selectedNominee,
          vote_type: 'NOMINATION',
        },
      };

      sendAction(action);
      setSelectedNominee('');
    } catch (error) {
      console.error('Failed to nominate player:', error);
    }
  }, [selectedNominee, isConnected, gameState.id, localPlayerId, sendAction]);

  const handleVote = useCallback(async () => {
    if (!selectedVote || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_VOTE',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          target_id: selectedVote, // For GUILTY/INNOCENT votes, the target is the vote itself
          vote_type: 'VERDICT',
        },
      };

      sendAction(action);
      setSelectedVote('');
    } catch (error) {
      console.error('Failed to cast vote:', error);
    }
  }, [selectedVote, isConnected, gameState.id, localPlayerId, sendAction]);

  const handlePulseCheck = useCallback(async (response: string) => {
    if (!isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_PULSE_CHECK',
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          response,
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to submit pulse check:', error);
    }
  }, [isConnected, gameState.id, localPlayerId, sendAction]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSendMessage();
    }
  }, [handleSendMessage]);

  return {
    // State
    chatInput,
    setChatInput,
    selectedNominee,
    setSelectedNominee,
    selectedVote,
    setSelectedVote,
    conversionTarget,
    setConversionTarget,
    miningTarget,
    setMiningTarget,
    
    // Actions
    handleSendMessage,
    handleMineTokens,
    handleUseAbility,
    handleProjectMilestones,
    handleConversionAttempt,
    handleNominate,
    handleVote,
    handlePulseCheck,
    handleKeyDown,
    
    // Utility functions
    getPhaseDisplayName,
    
    // Game engine functions
    canPlayerAffordAbility,
    isValidNightActionTarget,
  };
}