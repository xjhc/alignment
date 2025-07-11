import { useCallback, KeyboardEvent } from "react";
import { useWebSocketContext } from "../contexts/WebSocketContext";
import { useGameEngineContext } from "../contexts/GameEngineContext";
import { useSessionContext } from "../contexts/SessionContext";
import { useSound } from "./useSound";
import { ClientActionType } from "../types";

interface UseGameActionsProps {
  gameId: string | null;
  localPlayer: any;
  chatInput: string;
  setChatInput: (input: string) => void;
  activeChannel: string;
  miningTarget: string;
  setMiningTarget: (target: string) => void;
  conversionTarget: string;
  setConversionTarget: (target: string) => void;
  selectedNominee: string;
  setSelectedNominee: (nominee: string) => void;
  selectedVote: "GUILTY" | "INNOCENT" | "";
  setSelectedVote: (vote: "GUILTY" | "INNOCENT" | "") => void;
  replyingTo: { messageId: string; playerName: string; message: string } | null;
  setReplyingTo: (reply: { messageId: string; playerName: string; message: string } | null) => void;
  addMessageToBuffer: (message: string, channel: string) => void;
}

export function useGameActions({
  gameId,
  localPlayer,
  chatInput,
  setChatInput,
  activeChannel,
  miningTarget,
  setMiningTarget,
  conversionTarget,
  setConversionTarget,
  selectedNominee,
  setSelectedNominee,
  selectedVote,
  setSelectedVote,
  replyingTo,
  setReplyingTo,
  addMessageToBuffer,
}: UseGameActionsProps) {
  const { sendAction, isConnected } = useWebSocketContext();
  const { canPlayerAffordAbility, isValidNightActionTarget } = useGameEngineContext();
  const { appState } = useSessionContext();
  const { playSound } = useSound();
  const isSpectating = appState.isSpectating;

  const handleSendMessage = useCallback(async () => {
    // For spectators, we don't need a localPlayer. For regular players, we do.
    if (!chatInput.trim() || (!isSpectating && !localPlayer) || !isConnected || !gameId) return;
    try {
      let message = chatInput.trim();
      
      // For spectators, handle different action type
      if (isSpectating) {
        if (activeChannel === '#spectators') {
          // Send spectator message action
          sendAction({
            type: "POST_SPECTATOR_MESSAGE", // Custom action type for spectators
            payload: {
              game_id: gameId,
              messages: [{ message, client_message_id: `spectator_${Date.now()}` }],
              channel_id: '#spectators',
            },
          });
          setChatInput("");
          setReplyingTo(null);
          playSound("message");
          return;
        }
        // Spectators can only send messages in #spectators channel
        return;
      }
      
      // Regular player message handling
      const statusMatch = message.match(/^\/status\s+(.+)$/);
      if (statusMatch) {
        const statusMessage = statusMatch[1].trim();
        sendAction({
          type: ClientActionType.SetSlackStatus,
          payload: {
            game_id: gameId,
            player_id: localPlayer.id,
            status_message: statusMessage,
          },
        });
        setChatInput("");
        setReplyingTo(null);
        return;
      }
      if (replyingTo) {
        message = `[quote=${replyingTo.playerName}]${replyingTo.message}[/quote]\n${message}`;
      }
      
      // Apply system shock effects
      const isCorrupted = localPlayer.systemShocks?.some(shock => 
        shock.type === 'MESSAGE_CORRUPTION' && shock.isActive
      );
      
      if (isCorrupted && Math.random() < 0.25) {
        // Override message content with corruption effect
        message = "lol";
      }
      
      addMessageToBuffer(message, activeChannel);
      setChatInput("");
      setReplyingTo(null);
      playSound("message");
    } catch (error) {
      console.error("Failed to buffer message:", error);
      playSound("error");
    }
  }, [
    chatInput,
    localPlayer,
    isSpectating,
    isConnected,
    addMessageToBuffer,
    replyingTo,
    playSound,
    sendAction,
    gameId,
    activeChannel,
    setChatInput,
    setReplyingTo,
  ]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        handleSendMessage();
      } else if (e.key === "Escape" && replyingTo) {
        setReplyingTo(null);
      }
    },
    [handleSendMessage, replyingTo, setReplyingTo]
  );

  const startReply = useCallback(
    (messageId: string, playerName: string, message: string) => {
      setReplyingTo({ messageId, playerName, message });
    },
    [setReplyingTo]
  );

  const cancelReply = useCallback(() => {
    setReplyingTo(null);
  }, [setReplyingTo]);

  const handleMineTokens = useCallback(async () => {
    if (!localPlayer || !miningTarget || !isConnected || !gameId) return;
    sendAction({
      type: ClientActionType.SubmitNightAction,
      payload: {
        game_id: gameId,
        player_id: localPlayer.id,
        action_type: "MINE_TOKENS",
        target_player_id: miningTarget,
      },
    });
    setMiningTarget("");
  }, [localPlayer, miningTarget, isConnected, gameId, sendAction, setMiningTarget]);

  const handleUseAbility = useCallback(
    async (targetId?: string) => {
      if (
        !localPlayer ||
        !canPlayerAffordAbility(localPlayer.id) ||
        !isConnected ||
        !gameId
      ) return;
      sendAction({
        type: ClientActionType.SubmitNightAction,
        payload: {
          game_id: gameId,
          player_id: localPlayer.id,
          action_type: "USE_ABILITY",
          ability_type: localPlayer.role?.type || "UNKNOWN",
          target_id: targetId,
        },
      });
    },
    [localPlayer, canPlayerAffordAbility, isConnected, gameId, sendAction]
  );

  const handleProjectMilestones = useCallback(async () => {
    if (!localPlayer || !isConnected || !gameId) return;
    sendAction({
      type: ClientActionType.SubmitNightAction,
      payload: {
        game_id: gameId,
        player_id: localPlayer.id,
        action_type: "PROJECT_MILESTONES",
      },
    });
  }, [localPlayer, isConnected, gameId, sendAction]);

  const handleConversionAttempt = useCallback(async () => {
    if (!localPlayer || !conversionTarget || !isConnected || !gameId) return;
    if (
      !isValidNightActionTarget(
        localPlayer.id,
        conversionTarget,
        "ATTEMPT_CONVERSION"
      )
    ) return;
    sendAction({
      type: ClientActionType.SubmitNightAction,
      payload: {
        game_id: gameId,
        player_id: localPlayer.id,
        action_type: "ATTEMPT_CONVERSION",
        target_player_id: conversionTarget,
      },
    });
    setConversionTarget("");
  }, [
    localPlayer,
    conversionTarget,
    isConnected,
    isValidNightActionTarget,
    gameId,
    sendAction,
    setConversionTarget,
  ]);

  const handleNominate = useCallback(async () => {
    if (!selectedNominee || !isConnected || !gameId) return;
    sendAction({
      type: ClientActionType.SubmitVote,
      payload: {
        game_id: gameId,
        player_id: localPlayer?.id,
        target_id: selectedNominee,
        vote_type: "NOMINATION",
      },
    });
    setSelectedNominee("");
  }, [selectedNominee, isConnected, gameId, localPlayer?.id, sendAction, setSelectedNominee]);

  const handleVote = useCallback(async () => {
    if (!selectedVote || !isConnected || !gameId) return;
    sendAction({
      type: ClientActionType.SubmitVote,
      payload: {
        game_id: gameId,
        player_id: localPlayer?.id,
        target_id: selectedVote,
        vote_type: "VERDICT",
      },
    });
    setSelectedVote("");
    playSound("vote");
  }, [selectedVote, isConnected, gameId, localPlayer?.id, sendAction, playSound, setSelectedVote]);

  const handleExtensionVote = useCallback(
    async (choice: "EXTEND" | "NOMINATE") => {
      if (!isConnected || !gameId) return;
      sendAction({
        type: ClientActionType.SubmitVote,
        payload: {
          game_id: gameId,
          player_id: localPlayer?.id,
          target_id: choice,
          vote_type: "EXTENSION",
        },
      });
    },
    [isConnected, gameId, localPlayer?.id, sendAction]
  );

  const handlePulseCheck = useCallback(
    async (response: string) => {
      if (!isConnected || !gameId) return;
      sendAction({
        type: ClientActionType.SubmitPulseCheck,
        payload: {
          game_id: gameId,
          player_id: localPlayer?.id,
          response,
        },
      });
    },
    [isConnected, gameId, localPlayer?.id, sendAction]
  );

  const handleSkipPhase = useCallback(async () => {
    if (!localPlayer || !isConnected || !gameId) return;
    sendAction({
      type: ClientActionType.SubmitSkipVote,
      payload: { game_id: gameId, player_id: localPlayer.id },
    });
  }, [localPlayer, isConnected, gameId, sendAction]);

  const handleEmojiReaction = useCallback(
    async (
      messageId: string,
      emoji: string,
      channelId: string = "#war-room"
    ) => {
      if (!localPlayer || !isConnected || !gameId) return;
      sendAction({
        type: ClientActionType.ReactToMessage,
        payload: {
          game_id: gameId,
          player_id: localPlayer.id,
          message_id: messageId,
          emoji,
          channel_id: channelId,
        },
      });
    },
    [localPlayer, isConnected, gameId, sendAction]
  );

  const handleSubmitPartingShot = useCallback(
    async (partingShot: string) => {
      if (!isConnected || !partingShot.trim() || !gameId) return;
      sendAction({
        type: ClientActionType.SubmitExitInterview,
        payload: {
          game_id: gameId,
          player_id: localPlayer?.id,
          parting_shot: partingShot.trim(),
        },
      });
    },
    [isConnected, gameId, localPlayer?.id, sendAction]
  );

  const submitWhistleblowerVote = useCallback(
    async (crisisChoice: string) => {
      if (!isConnected || !crisisChoice.trim() || !gameId) return;
      sendAction({
        type: ClientActionType.SubmitWhistleblowerVote,
        payload: {
          game_id: gameId,
          player_id: localPlayer?.id,
          crisis_choice: crisisChoice,
        },
      });
    },
    [isConnected, gameId, localPlayer?.id, sendAction]
  );

  const handleAbandonGame = useCallback(async () => {
    if (!localPlayer || !isConnected || !gameId) return;
    sendAction({
      type: ClientActionType.AbandonGame,
      payload: {
        game_id: gameId,
        player_id: localPlayer.id,
      },
    });
  }, [localPlayer, isConnected, gameId, sendAction]);

  const getPhaseDisplayName = useCallback((phaseType: string) => {
    const PHASE_DISPLAY_NAMES: Record<string, string> = {
      SITREP: "SITREP",
      PULSE_CHECK: "PULSE CHECK",
      DISCUSSION: "DISCUSSION",
      NOMINATION: "NOMINATION",
      TRIAL: "TRIAL",
      VERDICT: "VERDICT",
      NIGHT: "NIGHT PHASE",
      GAME_OVER: "GAME OVER",
    };
    return PHASE_DISPLAY_NAMES[phaseType] || phaseType;
  }, []);

  return {
    handleSendMessage,
    handleKeyDown: handleKeyDown as (e: KeyboardEvent<any>) => void,
    startReply,
    cancelReply,
    handleMineTokens,
    handleUseAbility,
    handleProjectMilestones,
    handleConversionAttempt,
    handleNominate,
    handleVote,
    handleExtensionVote,
    handlePulseCheck,
    handleSkipPhase,
    handleEmojiReaction,
    handleSubmitPartingShot,
    submitWhistleblowerVote,
    handleAbandonGame,
    getPhaseDisplayName,
  };
}