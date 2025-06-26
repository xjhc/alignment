import { useState, useCallback } from 'react';
import { useGameContext } from '../contexts/GameContext';
import { useGameEngineContext } from '../contexts/GameEngineContext';
import { ClientAction, ClientActionType } from '../types';

export function useGameActions() {
  const { gameState, localPlayerId, sendAction, isConnected } = useGameContext();
  const { canPlayerAffordAbility, isValidNightActionTarget } = useGameEngineContext();
  
  const [chatInput, setChatInput] = useState('');
  const [selectedNominee, setSelectedNominee] = useState<string>('');
  const [selectedVote, setSelectedVote] = useState<'GUILTY' | 'INNOCENT' | ''>('');
  const [conversionTarget, setConversionTarget] = useState<string>('');
  const [miningTarget, setMiningTarget] = useState<string>('');
  const [replyingTo, setReplyingTo] = useState<{messageId: string, playerName: string, message: string} | null>(null);

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
      let message = chatInput.trim();
      
      // Check for /status command
      const statusMatch = message.match(/^\/status\s+(.+)$/);
      if (statusMatch) {
        const statusMessage = statusMatch[1].trim();
        const action: ClientAction = {
          type: ClientActionType.SetSlackStatus,
          payload: {
            game_id: gameState.id,
            player_id: localPlayerId,
            status_message: statusMessage,
          },
        };
        sendAction(action);
        setChatInput('');
        setReplyingTo(null);
        return;
      }
      
      // Format message with reply quote if replying
      if (replyingTo) {
        message = `[quote=${replyingTo.playerName}]${replyingTo.message}[/quote]\n${message}`;
      }

      const action: ClientAction = {
        type: ClientActionType.SendMessage,
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          message,
          player_name: localPlayer.name,
        },
      };

      sendAction(action);
      setChatInput('');
      setReplyingTo(null);
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  }, [chatInput, localPlayer, isConnected, gameState.id, localPlayerId, sendAction, replyingTo]);

  const handleMineTokens = useCallback(async () => {
    if (!localPlayer || !miningTarget || !isConnected) return;

    try {
      const action: ClientAction = {
        type: ClientActionType.SubmitNightAction,
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
        type: ClientActionType.SubmitNightAction,
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
        type: ClientActionType.SubmitNightAction,
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
        type: ClientActionType.SubmitNightAction,
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
        type: ClientActionType.SubmitVote,
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
        type: ClientActionType.SubmitVote,
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
        type: ClientActionType.SubmitPulseCheck,
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
    } else if (e.key === 'Escape' && replyingTo) {
      setReplyingTo(null);
    }
  }, [handleSendMessage, replyingTo]);

  const startReply = useCallback((messageId: string, playerName: string, message: string) => {
    setReplyingTo({ messageId, playerName, message });
  }, []);

  const cancelReply = useCallback(() => {
    setReplyingTo(null);
  }, []);

  const handleSkipPhase = useCallback(async () => {
    if (!localPlayer || !isConnected) return;

    try {
      const action: ClientAction = {
        type: ClientActionType.SubmitSkipVote,
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to vote to skip phase:', error);
    }
  }, [localPlayer, isConnected, gameState.id, localPlayerId, sendAction]);

  const handleEmojiReaction = useCallback(async (messageId: string, emoji: string, channelId: string = '#war-room') => {
    if (!localPlayer || !isConnected) {
      return;
    }

    try {
      const action: ClientAction = {
        type: ClientActionType.ReactToMessage,
        payload: {
          game_id: gameState.id,
          player_id: localPlayerId,
          message_id: messageId,
          emoji: emoji,
          channel_id: channelId,
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to send emoji reaction:', error);
    }
  }, [localPlayer, isConnected, gameState.id, localPlayerId, sendAction]);

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
    replyingTo,
    
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
    startReply,
    cancelReply,
    handleSkipPhase,
    handleEmojiReaction,
    
    // Utility functions
    getPhaseDisplayName,
    
    // Game engine functions
    canPlayerAffordAbility,
    isValidNightActionTarget,
  };
}