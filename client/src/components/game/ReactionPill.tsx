import React from "react";

interface ReactionPillProps {
  emoji: string;
  count: number;
  players: string[];
  isReactedBySelf: boolean;
  onClick: () => void;
}

export const ReactionPill: React.FC<ReactionPillProps> = ({
  emoji,
  count,
  players,
  isReactedBySelf,
  onClick,
}) => {
  const getTooltipText = () => {
    if (players.length === 0) return `React with ${emoji}`;
    if (players.length === 1) {
      return `${players[0]} reacted with ${emoji}`;
    }
    if (players.length === 2) {
      return `${players[0]} and ${players[1]} reacted with ${emoji}`;
    }
    const otherCount = players.length - 2;
    return `${players[0]}, ${players[1]}, and ${otherCount} other${otherCount > 1 ? "s" : ""} reacted with ${emoji}`;
  };

  return (
    <button
      onClick={onClick}
      title={getTooltipText()}
      className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs border transition-all duration-150 hover:scale-105 ${
        isReactedBySelf
          ? "bg-blue-500/20 border-blue-500 text-blue-400 shadow-sm"
          : "bg-background-secondary hover:bg-background-tertiary border-border text-text-secondary"
      }`}
      aria-pressed={isReactedBySelf}
    >
      <span className="text-sm">{emoji}</span>
      <span className="font-medium text-xs">{count}</span>
    </button>
  );
};
