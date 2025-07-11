import React, { useState } from "react";
import { useGameContext } from "../../contexts/GameContext";

interface ExtensionVoteUIProps {
  remainingSeconds: number;
}

export const ExtensionVoteUI: React.FC<ExtensionVoteUIProps> = ({
  remainingSeconds,
}) => {
  const { gameState, localPlayer, handleExtensionVote } = useGameContext();
  const [selectedChoice, setSelectedChoice] = useState<
    "EXTEND" | "NOMINATE" | null
  >(null);
  const [hasVoted, setHasVoted] = useState(false);

  if (!localPlayer || !localPlayer.isAlive) return null;

  const handleVoteChoice = async (choice: "EXTEND" | "NOMINATE") => {
    if (hasVoted) return;

    setSelectedChoice(choice);
    setHasVoted(true);
    await handleExtensionVote(choice);
  };

  const extendVotes = gameState?.voteState?.results?.["EXTEND"] || 0;
  const nominateVotes = gameState?.voteState?.results?.["NOMINATE"] || 0;

  return (
    <div className="fixed bottom-4 left-1/2 transform -translate-x-1/2 bg-background-primary border border-yellow-500 rounded-lg p-4 shadow-lg z-50 min-w-96 animation-slide-in-up">
      <div className="text-center mb-3">
        <h3 className="font-bold text-yellow-500 mb-1">
          Discussion Time Running Out!
        </h3>
        <p className="text-text-secondary text-sm">
          {remainingSeconds} seconds remaining - Vote to extend or move to
          nomination
        </p>
      </div>
      <div className="flex gap-3 justify-center">
        <button
          onClick={() => handleVoteChoice("EXTEND")}
          disabled={hasVoted}
          className={`flex-1 px-4 py-3 rounded-md font-semibold transition-all ${selectedChoice === "EXTEND" ? "bg-green-500 text-white" : hasVoted ? "bg-background-secondary text-text-muted cursor-not-allowed" : "bg-green-500/20 text-green-400 border border-green-500 hover:bg-green-500/30"}`}
        >
          <div className="flex flex-col items-center gap-1">
            <span>🕐 Extend Discussion</span>
            <span className="text-xs">+1 minute</span>
            {extendVotes > 0 && (
              <span className="text-xs font-mono">({extendVotes} votes)</span>
            )}
          </div>
        </button>
        <button
          onClick={() => handleVoteChoice("NOMINATE")}
          disabled={hasVoted}
          className={`flex-1 px-4 py-3 rounded-md font-semibold transition-all ${selectedChoice === "NOMINATE" ? "bg-amber-500 text-white" : hasVoted ? "bg-background-secondary text-text-muted cursor-not-allowed" : "bg-amber-500/20 text-amber-400 border border-amber-500 hover:bg-amber-500/30"}`}
        >
          <div className="flex flex-col items-center gap-1">
            <span>⚡ Move to Nomination</span>
            <span className="text-xs">End discussion now</span>
            {nominateVotes > 0 && (
              <span className="text-xs font-mono">({nominateVotes} votes)</span>
            )}
          </div>
        </button>
      </div>
      {hasVoted && (
        <div className="mt-3 text-center text-text-secondary text-sm">
          ✓ Vote submitted! Waiting for other players...
        </div>
      )}
    </div>
  );
};
