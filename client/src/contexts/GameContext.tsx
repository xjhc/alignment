import { createContext, useContext, ReactNode, KeyboardEvent } from "react";
import { GameState, Player, ClientAction } from "../types";
import { PendingMessage } from "../hooks/useChatBuffer";
import { SkipVoteState } from "../state/appReducer";

export interface GameContextType {
  // Game State
  gameState: GameState;
  localPlayerId: string;
  viewedPlayerId: string;
  localPlayer: Player | null;
  viewedPlayer: Player | null;
  isConnected: boolean;
  activeChannel: string;
  skipVoteState: SkipVoteState | null;

  // Action Dispatchers
  sendAction: (action: ClientAction) => void;
  setViewedPlayer: (playerId: string) => void;
  setActiveChannel: (channelId: string) => void;

  // Local UI State & Handlers
  chatInput: string;
  setChatInput: (value: string) => void;
  selectedNominee: string;
  setSelectedNominee: (id: string) => void;
  selectedVote: "GUILTY" | "INNOCENT" | "";
  setSelectedVote: (vote: "GUILTY" | "INNOCENT" | "") => void;
  conversionTarget: string;
  setConversionTarget: (id: string) => void;
  miningTarget: string;
  setMiningTarget: (id: string) => void;
  replyingTo: { messageId: string; playerName: string; message: string } | null;

  // Action Functions
  handleSendMessage: () => Promise<void>;
  handleMineTokens: () => Promise<void>;
  handleUseAbility: (targetId?: string) => Promise<void>;
  handleProjectMilestones: () => Promise<void>;
  handleConversionAttempt: () => Promise<void>;
  handleNominate: () => Promise<void>;
  handleVote: () => Promise<void>;
  handleExtensionVote: (choice: "EXTEND" | "NOMINATE") => Promise<void>;
  handlePulseCheck: (response: string) => Promise<void>;
  handleKeyDown: (
    e: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => void;
  startReply: (messageId: string, playerName: string, message: string) => void;
  cancelReply: () => void;
  handleSkipPhase: () => Promise<void>;
  handleEmojiReaction: (
    messageId: string,
    emoji: string,
    channelId?: string
  ) => Promise<void>;
  handleSubmitPartingShot: (partingShot: string) => Promise<void>;
  submitWhistleblowerVote: (crisisChoice: string) => Promise<void>;

  // Utility & State from other hooks
  getPhaseDisplayName: (phaseType: string) => string;
  canPlayerAffordAbility: (playerId: string) => boolean;
  isValidNightActionTarget: (
    actorId: string,
    targetId: string,
    actionType: string
  ) => boolean;

  // Chat Buffering State
  pendingMessages: Record<string, PendingMessage[]>;
  getPendingMessagesForChannel: (channelId?: string) => PendingMessage[];
  retryMessage: (clientMessageId: string, channelId: string) => void;
  rateLimitError: string | null;
  getBufferStatus: () => {
    bufferLength: number;
    hasPendingMessages: boolean;
    nextFlushETA: number;
  };
}

const GameContext = createContext<GameContextType | undefined>(undefined);

interface GameProviderProps {
  children: ReactNode;
  value: GameContextType;
}

export function GameProvider({ children, value }: GameProviderProps) {
  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}

export function useGameContext() {
  const context = useContext(GameContext);
  if (context === undefined) {
    throw new Error("useGameContext must be used within a GameProvider");
  }
  return context;
}
