import React, { KeyboardEvent } from "react";
import { useGameContext } from "../../contexts/GameContext";
import { VoteUI } from "./VoteUI";
import { NightActionSelection } from "./NightActionSelection";
import { PulseCheckInput } from "./PulseCheckInput";
import { Button } from "../ui/Button";
import { CrisisEvent } from "../../types";

const generateCrisisQuestion = (crisis?: CrisisEvent): string => {
  if (!crisis) return "What is your immediate response to the current crisis?";
  switch (crisis.type) {
    case "Database Index Corruption":
      return "A critical role has been exposed. How does this change your immediate priority?";
    case "Cascading Server Failure":
      return "With limited bandwidth, what is the one piece of information everyone needs to hear from you?";
    case "Emergency Board Meeting":
      return "The Board demands accountability. Which two roles do you believe are most responsible for this situation?";
    case "Tainted Training Data":
      return "We've learned the AI was trained on compromised data. What 'unshakeable truth' do you now question?";
    case "Nightmare Scenario":
      return "Emergency protocols are in effect. What is your immediate action to protect the company?";
    case "Press Leak":
      return "Sensitive information has leaked. What is your strategy to control the narrative?";
    case "Incident Response Drill":
      return "All communications are monitored. What would you say if you knew everyone was listening?";
    case "Major Service Outage":
      return "Critical services are down. What is your highest priority for recovery efforts?";
    case "Phishing Attack":
      return "Security has been compromised. Who do you trust most in this room and why?";
    case "Data Privacy Audit":
      return "External auditors are reviewing everything. What would concern you most if discovered?";
    case "Vendor Security Breach":
      return "A trusted partner has been compromised. How do you verify who you can still trust?";
    case "Regulatory Review":
      return "Government oversight is imminent. What would you want leadership to know before they arrive?";
    default:
      return "Given the current crisis, what is your immediate concern for the company?";
  }
};

interface ContextualInputAreaProps {}

export const ContextualInputArea: React.FC<ContextualInputAreaProps> = () => {
  const {
    gameState,
    localPlayer,
    isConnected,
    replyingTo,
    cancelReply,
    rateLimitError,
    getBufferStatus,
    chatInput,
    setChatInput,
    handleKeyDown,
    handleSendMessage,
    handlePulseCheck,
  } = useGameContext();

  if (!localPlayer) return null;

  switch (gameState.phase.type) {
    case "NOMINATION":
    case "VERDICT":
      return <VoteUI />;

    case "NIGHT":
      return <NightActionSelection />;

    case "PULSE_CHECK":
      if (!localPlayer?.hasSubmittedPulseCheck) {
        const pulseCheckQuestion = generateCrisisQuestion(
          gameState.crisisEvent
        );
        return (
          <PulseCheckInput
            handlePulseCheck={handlePulseCheck}
            localPlayerName={localPlayer.name}
            question={pulseCheckQuestion}
          />
        );
      }
    // After submission, fall through to the default case to show the chat input
    // Fallthrough is intentional here.

    case "SITREP":
    case "DISCUSSION":
    case "TRIAL":
    default:
      const isChatEnabled = () => {
        if (!isConnected) return false;

        switch (gameState.phase.type) {
          case "SITREP":
          case "DISCUSSION":
          case "TRIAL":
            return true;
          case "PULSE_CHECK":
            return localPlayer?.hasSubmittedPulseCheck === true;
          default:
            return false;
        }
      };

      const getPlaceholder = () => {
        if (!isConnected) return "Reconnecting...";
        if (replyingTo) return `Reply to ${replyingTo.playerName}...`;

        switch (gameState.phase.type) {
          case "SITREP":
          case "DISCUSSION":
            return "Message #war-room";
          case "PULSE_CHECK":
            if (!localPlayer?.hasSubmittedPulseCheck) {
              return "Submit your pulse check response to enable chat";
            }
            return "Message #war-room";
          case "TRIAL":
            return localPlayer?.id === gameState.nominatedPlayer
              ? "Present your defense..."
              : "Question the nominated player...";
          case "NIGHT":
            return "Channel locked during Night Phase";
          default:
            return `Channel locked during ${gameState.phase.type}`;
        }
      };

      const handleSendButtonClick = () => {
        if (chatInput.trim() && isChatEnabled()) {
          handleSendMessage();
        }
      };

      return (
        <div className="border-t border-border bg-background-primary p-3">
          {replyingTo && (
            <div className="flex items-center justify-between bg-background-secondary border border-border rounded-md px-3 py-2 mb-3 text-sm">
              <div className="flex items-center gap-2 text-text-secondary">
                <span className="text-text-muted">↩️ Replying to</span>
                <span className="font-semibold text-text-primary">
                  {replyingTo.playerName}
                </span>
                <span className="text-text-muted truncate max-w-xs">
                  "{replyingTo.message}"
                </span>
              </div>
              <button
                onClick={cancelReply}
                className="text-text-muted hover:text-text-primary transition-colors"
                title="Cancel reply"
              >
                ✕
              </button>
            </div>
          )}

          {rateLimitError && (
            <div className="flex items-center justify-between bg-red-500/10 border border-red-500/30 rounded-md px-3 py-2 mb-3 text-sm">
              <div className="flex items-center gap-2 text-red-600">
                <span>⚠️</span>
                <span>{rateLimitError}</span>
              </div>
            </div>
          )}

          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <input
                className="flex-1 bg-background-secondary border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-primary focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)] disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-150"
                type="text"
                placeholder={getPlaceholder()}
                value={chatInput}
                maxLength={280}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyDown={(e) => handleKeyDown(e as any)} // Cast event type for compatibility
                disabled={!isChatEnabled()}
              />
              <Button
                variant="primary"
                size="sm"
                onClick={handleSendButtonClick}
                disabled={
                  !isChatEnabled() ||
                  !chatInput.trim() ||
                  chatInput.length > 280
                }
                rightIcon={
                  <svg
                    className="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                    />
                  </svg>
                }
                title="Send message"
              >
                Send
              </Button>
            </div>
            {isChatEnabled() && (
              <div className="flex justify-end">
                <span
                  className={`text-xs ${chatInput.length > 280 ? "text-red-500" : "text-text-muted"}`}
                >
                  {chatInput.length}/280
                </span>
              </div>
            )}
          </div>
        </div>
      );
  }
};
