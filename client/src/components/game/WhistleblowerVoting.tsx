import React, { useState } from "react";
import {
  GeneratedWhistleblowerVoting,
  GeneratedCrisisEventOption,
} from "../../types";
import { useGameContext } from "../../contexts/GameContext";

interface WhistleblowerVotingProps {
  whistleblowerVoting: GeneratedWhistleblowerVoting;
  localPlayerName: string;
  hasVoted: boolean;
}

export const WhistleblowerVoting: React.FC<WhistleblowerVotingProps> = ({
  whistleblowerVoting,
  localPlayerName,
  hasVoted,
}) => {
  const [selectedCrisis, setSelectedCrisis] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { submitWhistleblowerVote } = useGameContext();

  const handleSubmitVote = async () => {
    if (!selectedCrisis || isSubmitting) return;

    setIsSubmitting(true);
    try {
      await submitWhistleblowerVote(selectedCrisis);
    } catch (error) {
      console.error("Failed to submit whistleblower vote:", error);
      setIsSubmitting(false);
    }
  };

  const getVoteCount = (crisisType: string): number => {
    return whistleblowerVoting.voteResults[crisisType] || 0;
  };

  const getTotalVotes = (): number => {
    return Object.values(whistleblowerVoting.voteResults).reduce(
      (sum, count) => sum + count,
      0
    );
  };

  if (!whistleblowerVoting.isActive) {
    return null;
  }

  return (
    <div className="bg-purple-900/20 border border-purple-500/30 rounded-lg p-4 mb-3">
      <div className="text-purple-400 font-mono font-bold text-sm mb-2">
        🕵️ WHISTLEBLOWER PROTOCOL
      </div>

      {!hasVoted && !whistleblowerVoting.isComplete ? (
        <>
          <div className="text-text-primary font-medium mb-3">
            As a deactivated employee, you have access to three leaked memos
            about potential incidents. Vote on which crisis should affect
            tomorrow's operations:
          </div>

          <div className="space-y-3 mb-4">
            {whistleblowerVoting.crisisOptions.map(
              (option: GeneratedCrisisEventOption) => (
                <div key={option.type} className="relative">
                  <label className="block">
                    <input
                      type="radio"
                      name="crisis-choice"
                      value={option.type}
                      checked={selectedCrisis === option.type}
                      onChange={(e) => setSelectedCrisis(e.target.value)}
                      disabled={isSubmitting}
                      className="sr-only"
                    />
                    <div
                      className={`p-3 border rounded-lg cursor-pointer transition-all duration-200 ${selectedCrisis === option.type ? "border-purple-500 bg-purple-900/30" : "border-border/30 bg-background-secondary/50 hover:border-purple-500/50 hover:bg-purple-900/10"} ${isSubmitting ? "opacity-50 cursor-not-allowed" : ""}`}
                    >
                      <div className="font-semibold text-purple-400 mb-1">
                        {option.title}
                      </div>
                      <div className="text-xs text-text-secondary">
                        {option.description}
                      </div>
                      {getVoteCount(option.type) > 0 && (
                        <div className="text-xs text-purple-400 mt-1">
                          {getVoteCount(option.type)} vote
                          {getVoteCount(option.type) !== 1 ? "s" : ""}
                        </div>
                      )}
                    </div>
                  </label>
                </div>
              )
            )}
          </div>

          <button
            onClick={handleSubmitVote}
            disabled={!selectedCrisis || isSubmitting}
            className="w-full px-4 py-2 bg-purple-600 text-white font-medium rounded-md hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
          >
            {isSubmitting ? "Submitting Vote..." : "Submit Whistleblower Vote"}
          </button>

          <div className="text-xs text-gray-400 mt-2">
            Your vote as {localPlayerName} will influence tomorrow's crisis
            event.
          </div>
        </>
      ) : (
        <>
          <div className="text-text-primary font-medium mb-3">
            {hasVoted
              ? "You have voted in the Whistleblower Protocol."
              : "Whistleblower voting has ended."}
          </div>
          <div className="text-sm text-gray-400 mb-3">
            Vote results ({getTotalVotes()} total vote
            {getTotalVotes() !== 1 ? "s" : ""}):
          </div>
          <div className="space-y-2">
            {whistleblowerVoting.crisisOptions.map(
              (option: GeneratedCrisisEventOption) => {
                const votes = getVoteCount(option.type);
                const percentage =
                  getTotalVotes() > 0 ? (votes / getTotalVotes()) * 100 : 0;
                const isWinning =
                  whistleblowerVoting.selectedCrisis === option.type;
                return (
                  <div
                    key={option.type}
                    className={`p-2 border rounded border-border/30 bg-background-secondary/30 ${isWinning ? "border-purple-500 bg-purple-900/20" : ""}`}
                  >
                    <div className="flex justify-between items-center">
                      <div className="font-medium text-sm">
                        {option.title}
                        {isWinning && (
                          <span className="text-purple-400 ml-2">
                            👑 SELECTED
                          </span>
                        )}
                      </div>
                      <div className="text-sm text-gray-400">
                        {votes} vote{votes !== 1 ? "s" : ""} (
                        {percentage.toFixed(0)}%)
                      </div>
                    </div>
                    {votes > 0 && (
                      <div className="w-full bg-gray-700 rounded-full h-1 mt-1">
                        <div
                          className={`h-1 rounded-full transition-all duration-300 ${isWinning ? "bg-purple-500" : "bg-gray-500"}`}
                          style={{ width: `${percentage}%` }}
                        />
                      </div>
                    )}
                  </div>
                );
              }
            )}
          </div>
        </>
      )}
    </div>
  );
};
