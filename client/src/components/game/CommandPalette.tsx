import React, { useState, useEffect, useRef, useMemo } from "react";
import { useGameContext } from "../../contexts/GameContext";
import { FADE_IN, SLIDE_IN_UP } from "../../utils/animations";

interface Command {
  id: string;
  label: string;
  shortcut?: string;
  description: string;
  action: () => void;
  category: "game" | "chat" | "navigation" | "debug";
  enabled: boolean;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({
  isOpen,
  onClose,
}) => {
  const {
    gameState,
    localPlayer,
    handleSkipPhase,
    setChatInput,
    getPhaseDisplayName,
  } = useGameContext();

  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const modalRef = useRef<HTMLDivElement>(null);

  const commands = useMemo((): Command[] => {
    const cmds: Command[] = [
      {
        id: "skip-phase",
        label: "Skip Phase",
        shortcut: "Ctrl+S",
        description: `Vote to skip the current ${getPhaseDisplayName(gameState.phase.type)} phase`,
        action: () => {
          handleSkipPhase();
          onClose();
        },
        category: "game",
        enabled:
          gameState.phase.type !== "TRIAL" &&
          gameState.phase.type !== "GAME_OVER" &&
          localPlayer?.isAlive === true,
      },
      {
        id: "send-status",
        label: "Set Status Message",
        shortcut: "/status",
        description: "Set your Slack status message",
        action: () => {
          setChatInput("/status ");
          onClose();
        },
        category: "chat",
        enabled: true,
      },
      {
        id: "clear-chat",
        label: "Clear Chat Input",
        shortcut: "Esc",
        description: "Clear the current chat message",
        action: () => {
          setChatInput("");
          onClose();
        },
        category: "chat",
        enabled: true,
      },
      {
        id: "focus-chat",
        label: "Focus Chat Input",
        shortcut: "Ctrl+/",
        description: "Focus the chat input field",
        action: () => {
          const chatInput = document.querySelector(
            'input[placeholder*="chat" i], textarea[placeholder*="chat" i]'
          ) as HTMLInputElement;
          if (chatInput) {
            chatInput.focus();
          }
          onClose();
        },
        category: "navigation",
        enabled: true,
      },
      {
        id: "open-settings",
        label: "Open Settings",
        description: "Open game settings and preferences",
        action: () => {
          const settingsEvent = new CustomEvent("open-settings-modal");
          window.dispatchEvent(settingsEvent);
          onClose();
        },
        category: "navigation",
        enabled: true,
      },
      {
        id: "debug-state",
        label: "Log Game State",
        description: "Log current game state to console",
        action: () => {
          console.log("Game State:", gameState);
          console.log("Local Player:", localPlayer);
          onClose();
        },
        category: "debug",
        enabled: process.env.NODE_ENV === "development",
      },
    ];

    if (gameState.phase.type === "NOMINATION" && localPlayer?.isAlive) {
      cmds.push({
        id: "nominate-help",
        label: "Nomination Help",
        description: "Show information about the nomination phase",
        action: () => {
          alert("Click on a player card to nominate them for elimination.");
          onClose();
        },
        category: "game",
        enabled: true,
      });
    }

    if (gameState.phase.type === "VERDICT" && localPlayer?.isAlive) {
      cmds.push({
        id: "vote-guilty",
        label: "Vote Guilty",
        shortcut: "G",
        description: "Vote to eliminate the nominated player",
        action: () => {
          const guiltButton = document.querySelector(
            '[data-vote="GUILTY"]'
          ) as HTMLButtonElement;
          if (guiltButton) guiltButton.click();
          onClose();
        },
        category: "game",
        enabled: true,
      });

      cmds.push({
        id: "vote-innocent",
        label: "Vote Innocent",
        shortcut: "I",
        description: "Vote to keep the nominated player",
        action: () => {
          const innocentButton = document.querySelector(
            '[data-vote="INNOCENT"]'
          ) as HTMLButtonElement;
          if (innocentButton) innocentButton.click();
          onClose();
        },
        category: "game",
        enabled: true,
      });
    }

    return cmds.filter((cmd) => cmd.enabled);
  }, [
    gameState,
    localPlayer,
    handleSkipPhase,
    setChatInput,
    getPhaseDisplayName,
    onClose,
  ]);

  const filteredCommands = useMemo(() => {
    if (!searchQuery.trim()) return commands;
    const query = searchQuery.toLowerCase();
    return commands.filter(
      (cmd) =>
        cmd.label.toLowerCase().includes(query) ||
        cmd.description.toLowerCase().includes(query) ||
        cmd.shortcut?.toLowerCase().includes(query)
    );
  }, [commands, searchQuery]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [filteredCommands]);

  useEffect(() => {
    if (isOpen && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          setSelectedIndex((prev) =>
            prev < filteredCommands.length - 1 ? prev + 1 : 0
          );
          break;
        case "ArrowUp":
          e.preventDefault();
          setSelectedIndex((prev) =>
            prev > 0 ? prev - 1 : filteredCommands.length - 1
          );
          break;
        case "Enter":
          e.preventDefault();
          if (filteredCommands[selectedIndex]) {
            filteredCommands[selectedIndex].action();
          }
          break;
        case "Escape":
          e.preventDefault();
          onClose();
          break;
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, filteredCommands, selectedIndex, onClose]);

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case "game":
        return "🎮";
      case "chat":
        return "💬";
      case "navigation":
        return "🧭";
      case "debug":
        return "🐛";
      default:
        return "⚡";
    }
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case "game":
        return "text-blue-500";
      case "chat":
        return "text-green-500";
      case "navigation":
        return "text-purple-500";
      case "debug":
        return "text-orange-500";
      default:
        return "text-gray-500";
    }
  };

  return (
    <div
      className={`fixed inset-0 bg-black/50 z-50 flex items-start justify-center pt-20 ${FADE_IN}`}
    >
      <div
        ref={modalRef}
        className={`bg-background-primary border border-border rounded-lg shadow-2xl w-full max-w-2xl mx-4 ${SLIDE_IN_UP}`}
      >
        <div className="p-4 border-b border-border">
          <div className="relative">
            <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-text-muted">
              ⌘
            </div>
            <input
              ref={searchInputRef}
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Type a command or search..."
              className="w-full pl-10 pr-4 py-3 bg-background-secondary border border-border rounded-md text-text-primary placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
            />
          </div>
        </div>
        <div className="max-h-96 overflow-y-auto">
          {filteredCommands.length === 0 ? (
            <div className="p-4 text-center text-text-muted">
              No commands found for "{searchQuery}"
            </div>
          ) : (
            <div className="p-2">
              {filteredCommands.map((command, index) => (
                <div
                  key={command.id}
                  className={`flex items-center gap-3 p-3 rounded-md transition-colors cursor-pointer ${index === selectedIndex ? "bg-primary text-white" : "hover:bg-background-secondary"}`}
                  onClick={() => command.action()}
                >
                  <div className="text-lg">
                    {getCategoryIcon(command.category)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span
                        className={`font-medium ${index === selectedIndex ? "text-white" : "text-text-primary"}`}
                      >
                        {command.label}
                      </span>
                      <span
                        className={`text-xs px-2 py-1 rounded ${getCategoryColor(command.category)} ${index === selectedIndex ? "bg-white/10" : "bg-background-tertiary"}`}
                      >
                        {command.category}
                      </span>
                    </div>
                    <div
                      className={`text-sm mt-1 ${index === selectedIndex ? "text-white/80" : "text-text-secondary"}`}
                    >
                      {command.description}
                    </div>
                  </div>
                  {command.shortcut && (
                    <div
                      className={`text-xs font-mono px-2 py-1 rounded border ${index === selectedIndex ? "bg-white/10 border-white/20 text-white" : "bg-background-tertiary border-border text-text-muted"}`}
                    >
                      {command.shortcut}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
        <div className="p-3 border-t border-border text-xs text-text-muted flex items-center justify-between">
          <div className="flex items-center gap-4">
            <span>↑↓ Navigate</span>
            <span>↵ Execute</span>
            <span>Esc Close</span>
          </div>
          <div>{filteredCommands.length} commands</div>
        </div>
      </div>
    </div>
  );
};
