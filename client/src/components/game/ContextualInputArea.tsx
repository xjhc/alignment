import React, { KeyboardEvent, useState, useRef, useEffect } from "react";
import { useGameContext } from "../../contexts/GameContext";
import { useSessionContext } from "../../contexts/SessionContext";
import { VoteUI } from "./VoteUI";
import { NightActionSelection } from "./NightActionSelection";
import { PulseCheckInput } from "./PulseCheckInput";
import { Button } from "../ui/Button";
import { CrisisEvent } from "../../types";
import { MarkdownRenderer } from "./MarkdownRenderer";

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
    activeChannel,
  } = useGameContext();
  
  const { appState } = useSessionContext();
  const isSpectating = appState.isSpectating;

  // Enhanced input features state
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [showMarkdownPreview, setShowMarkdownPreview] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  // Quick emoji options for insertion
  const quickEmojis = ['😀', '😔', '🤔', '👍', '👎', '❤️', '🔥', '💯', '🎯', '⚠️', '🤖', '👤'];

  // Markdown formatting helpers
  const insertMarkdown = (prefix: string, suffix: string = '') => {
    if (!inputRef.current) return;
    
    const input = inputRef.current;
    const start = input.selectionStart || 0;
    const end = input.selectionEnd || 0;
    const selectedText = chatInput.substring(start, end);
    const newText = chatInput.substring(0, start) + prefix + selectedText + suffix + chatInput.substring(end);
    
    setChatInput(newText);
    
    // Restore cursor position
    setTimeout(() => {
      const newCursorPos = start + prefix.length + selectedText.length + suffix.length;
      input.setSelectionRange(newCursorPos, newCursorPos);
      input.focus();
    }, 0);
  };

  const insertEmoji = (emoji: string) => {
    if (!inputRef.current) return;
    
    const input = inputRef.current;
    const start = input.selectionStart || 0;
    const newText = chatInput.substring(0, start) + emoji + chatInput.substring(start);
    
    setChatInput(newText);
    setShowEmojiPicker(false);
    
    // Restore cursor position
    setTimeout(() => {
      const newCursorPos = start + emoji.length;
      input.setSelectionRange(newCursorPos, newCursorPos);
      input.focus();
    }, 0);
  };

  // Close emoji picker when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (showEmojiPicker && inputRef.current && !inputRef.current.closest('.relative')?.contains(event.target as Node)) {
        setShowEmojiPicker(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [showEmojiPicker]);

  // For spectators, we don't need a localPlayer. For regular players, we do.
  if (!isSpectating && !localPlayer) return null;

  // Spectators can only chat, no voting or night actions
  if (isSpectating) {
    // Only show chat input for spectators in the #spectators channel
    if (activeChannel !== '#spectators') {
      return (
        <div className="p-4 text-center text-text-muted text-sm">
          <p>Switch to #spectators to chat with other spectators</p>
        </div>
      );
    }
    // Show spectator chat input
  } else {
    // Regular player logic - determine what special UI to show
    let specialUI = null;
    
    switch (gameState?.phase?.type) {
      case "NOMINATION":
        return <VoteUI />;

      case "VERDICT":
        // During VERDICT, show voting UI above the chat input
        specialUI = <VoteUI />;
        break;

      case "NIGHT":
        return <NightActionSelection />;

      case "PULSE_CHECK":
        if (!localPlayer?.hasSubmittedPulseCheck) {
          const pulseCheckQuestion = generateCrisisQuestion(
            gameState?.crisisEvent
          );
          return (
            <PulseCheckInput
              handlePulseCheck={handlePulseCheck}
              localPlayerName={localPlayer.name}
              question={pulseCheckQuestion}
            />
          );
        }
      // After submission, fall through to show chat input
      // Fallthrough is intentional here.

      case "SITREP":
      case "DISCUSSION":
      case "TRIAL":
      default:
        break; // Continue to chat input below
    }

    // If we have special UI (like VERDICT voting), render it above the chat input
    if (specialUI) {
      return (
        <div>
          {specialUI}
          {/* Chat input section below */}
          {renderChatInput()}
        </div>
      );
    }
  }

  const isChatEnabled = () => {
    if (!isConnected) return false;

    switch (gameState.phase.type) {
      case "SITREP":
      case "DISCUSSION":
      case "TRIAL":
      case "VERDICT":
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
      case "VERDICT":
        return "Discuss the verdict...";
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

  // Extract chat input rendering into a separate function for reuse
  const renderChatInput = () => (
        <div className="border-t border-border bg-background-primary p-3">
          {replyingTo && (
            <div className="bg-background-secondary border border-border rounded-md px-3 py-2 mb-3 text-sm">
              <div className="flex items-center justify-between mb-2">
                <span className="text-text-muted">↩️ Replying to <span className="font-semibold text-text-primary">{replyingTo.playerName}</span></span>
                <button
                  onClick={cancelReply}
                  className="text-text-muted hover:text-text-primary transition-colors"
                  title="Cancel reply"
                >
                  ✕
                </button>
              </div>
              <div className="bg-background-tertiary border-l-2 border-primary rounded px-2 py-1 text-text-secondary text-xs">
                <MarkdownRenderer content={replyingTo.message} />
              </div>
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

          {/* Corporate Mandate Communication Warning */}
          {gameState?.corporateMandate?.isActive && gameState.corporateMandate.effects?.no_direct_messages && (
            <div className="flex items-center justify-between bg-yellow-500/10 border border-yellow-500/30 rounded-md px-3 py-2 mb-3 text-sm">
              <div className="flex items-center gap-2 text-yellow-600">
                <span>👁️</span>
                <span>Total Transparency Initiative: Private communications suspended</span>
              </div>
            </div>
          )}

          <div className="space-y-2">
            {/* Markdown Toolbar */}
            {isChatEnabled() && (
              <div className="flex items-center gap-1 px-2 py-1 bg-background-secondary border border-border rounded-md">
                {/* Markdown formatting buttons */}
                <button
                  onClick={() => insertMarkdown('**', '**')}
                  className="px-2 py-1 text-xs font-bold rounded hover:bg-background-tertiary transition-colors"
                  title="Bold (**text**)"
                >
                  B
                </button>
                <button
                  onClick={() => insertMarkdown('*', '*')}
                  className="px-2 py-1 text-xs italic rounded hover:bg-background-tertiary transition-colors"
                  title="Italic (*text*)"
                >
                  I
                </button>
                <button
                  onClick={() => insertMarkdown('~', '~')}
                  className="px-2 py-1 text-xs line-through rounded hover:bg-background-tertiary transition-colors"
                  title="Strikethrough (~text~)"
                >
                  S
                </button>
                <button
                  onClick={() => insertMarkdown('`', '`')}
                  className="px-2 py-1 text-xs font-mono rounded hover:bg-background-tertiary transition-colors"
                  title="Code (`text`)"
                >
                  {'</>'}
                </button>
                
                <div className="w-px h-4 bg-border mx-1" />
                
                {/* Emoji picker button */}
                <div className="relative">
                  <button
                    onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                    className="px-2 py-1 text-xs rounded hover:bg-background-tertiary transition-colors"
                    title="Add emoji"
                  >
                    😀
                  </button>
                  
                  {/* Emoji picker dropdown */}
                  {showEmojiPicker && (
                    <div className="absolute bottom-full left-0 mb-2 bg-background-primary border border-border rounded-lg p-2 shadow-lg z-50">
                      <div className="grid grid-cols-6 gap-1">
                        {quickEmojis.map((emoji) => (
                          <button
                            key={emoji}
                            onClick={() => insertEmoji(emoji)}
                            className="w-8 h-8 text-sm hover:bg-background-secondary rounded transition-colors"
                            title={`Insert ${emoji}`}
                          >
                            {emoji}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
                
                <div className="w-px h-4 bg-border mx-1" />
                
                {/* Preview toggle */}
                <button
                  onClick={() => setShowMarkdownPreview(!showMarkdownPreview)}
                  className={`px-2 py-1 text-xs rounded transition-colors ${
                    showMarkdownPreview ? 'bg-primary text-white' : 'hover:bg-background-tertiary'
                  }`}
                  title="Toggle preview"
                >
                  👁️
                </button>
              </div>
            )}
            
            {/* Input area */}
            <div className="flex items-center gap-2">
              <div className="flex-1">
                <input
                  ref={inputRef}
                  className="w-full bg-background-secondary border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-primary focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)] disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-150"
                  type="text"
                  placeholder={getPlaceholder()}
                  value={chatInput}
                  maxLength={280}
                  onChange={(e) => setChatInput(e.target.value)}
                  onKeyDown={(e) => handleKeyDown(e as any)}
                  disabled={!isChatEnabled()}
                />
                
                {/* Character counter */}
                <div className="flex justify-between items-center mt-1 px-1">
                  <div className="text-xs text-text-muted">
                    Markdown: **bold** *italic* ~strike~ `code`
                  </div>
                  <div className={`text-xs ${chatInput.length > 260 ? 'text-warning' : chatInput.length > 280 ? 'text-danger' : 'text-text-muted'}`}>
                    {chatInput.length}/280
                  </div>
                </div>
              </div>
              
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
            
            {/* Markdown Preview */}
            {showMarkdownPreview && chatInput.trim() && (
              <div className="bg-background-secondary border border-border rounded-md p-3">
                <div className="text-xs text-text-muted mb-2 flex items-center gap-2">
                  <span>👁️ Preview:</span>
                </div>
                <div className="text-sm">
                  <MarkdownRenderer content={chatInput} />
                </div>
              </div>
            )}
          </div>
        </div>
      );

      return renderChatInput();
};
