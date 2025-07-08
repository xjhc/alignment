import React from "react";
import { motion } from "framer-motion";
import { useGameContext } from "../../contexts/GameContext";
import { useSound } from "../../hooks/useSound";
import { Tooltip } from "../ui/Tooltip";
import { Button } from "../ui";

interface VoteUIProps {}

export const VoteUI: React.FC<VoteUIProps> = () => {
  const {
    gameState,
    localPlayer,
    selectedNominee,
    setSelectedNominee,
    setSelectedVote,
    handleNominate,
    handleVote,
  } = useGameContext();
  const { playSound } = useSound();

  if (!localPlayer) return null;
  const players = gameState?.players || [];
  const alivePlayers = Array.isArray(players)
    ? players.filter((p) => p.isAlive && p.id !== localPlayer.id)
    : [];

  if (gameState.phase.type === "NOMINATION") {
    return (
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
        <div className="mb-1.5">
          <h3 className="text-sm font-bold text-gray-100 normal-case tracking-normal text-left p-0 bg-transparent m-0 mb-3">
            Who should we deactivate?
          </h3>
        </div>
        <motion.div className="flex flex-wrap gap-1 justify-start" layout>
          {alivePlayers.map((player) => {
            const isSelected = selectedNominee === player.id;
            const votes = gameState.voteState?.votes || {};
            const playerVotes = Object.values(votes).filter(
              (vote) => vote === player.id
            ).length;
            return (
              <Tooltip
                key={player.id}
                content={`${player.name} (${playerVotes} vote${playerVotes !== 1 ? "s" : ""})`}
              >
                <motion.button
                  layout
                  className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-gray-600 bg-gray-800 cursor-pointer min-w-0 font-inherit ${isSelected ? "bg-amber-500/10 border-amber-500" : ""} ${player.alignment === "ALIGNED" ? "bg-cyan-500/5 border-cyan-600" : ""}`}
                  onClick={() => {
                    playSound("vote");
                    setSelectedNominee(player.id);
                    handleNominate();
                  }}
                  whileHover={{
                    backgroundColor: "rgba(55, 65, 81, 1)",
                    y: -2,
                    scale: 1.02,
                    transition: { duration: 0.15 },
                  }}
                  whileTap={{ scale: 0.95 }}
                  animate={
                    isSelected
                      ? {
                          borderColor: "rgba(245, 158, 11, 1)",
                          backgroundColor: "rgba(245, 158, 11, 0.1)",
                          boxShadow: "0 0 10px rgba(245, 158, 11, 0.3)",
                        }
                      : {}
                  }
                >
                  <span className="text-sm leading-none flex-shrink-0">
                    {/*...icon logic...*/}
                  </span>
                  <span
                    className={`font-medium text-xs text-gray-100 flex-shrink-0 ${player.alignment === "ALIGNED" ? "text-cyan-600 animate-[glitch_1.5s_infinite]" : ""}`}
                  >
                    {player.name}
                  </span>
                  <span
                    className={`font-mono font-bold text-gray-400 text-xs ml-auto ${isSelected ? "text-amber-500" : ""}`}
                  >
                    🪙 {playerVotes}
                  </span>
                </motion.button>
              </Tooltip>
            );
          })}
        </motion.div>
      </div>
    );
  }

  if (gameState.phase.type === "VERDICT") {
    const nominatedPlayer = gameState.players.find(
      (p) => p.id === gameState.nominatedPlayer
    );
    if (!nominatedPlayer) return null;

    const yesVotes = gameState.voteState?.results?.["GUILTY"] || 0;
    const noVotes = gameState.voteState?.results?.["INNOCENT"] || 0;
    const getPlayerAvatar = (jobTitle: string) => {
      switch (jobTitle) {
        case 'CEO': return '👑';
        case 'CTO': return '💻';
        case 'CFO': return '💰';
        case 'COO': return '⚙️';
        case 'CISO': return '🔒';
        case 'Ethics Officer': return '⚖️';
        case 'Platform Lead': return '🏗️';
        case 'Intern': return '🎓';
        default: return '👤';
      }
    };
    const renderVoteBlocks = (voteOption: string) => {
      if (!gameState.voteState?.votes || !gameState.voteState?.tokenWeights) return null;
      
      const votes = gameState.voteState.votes;
      const tokenWeights = gameState.voteState.tokenWeights;
      
      return Object.entries(votes)
        .filter(([, vote]) => vote === voteOption)
        .map(([playerId]) => {
          const tokenWeight = tokenWeights[playerId] || 0;
          const isMyVote = playerId === localPlayer.id;
          const player = gameState.players.find(p => p.id === playerId);
          
          return (
            <div
              key={playerId}
              className={`vote-block ${isMyVote ? 'my-vote' : ''}`}
            >
              <div className="block-header">
                <span className="block-icon">{getPlayerAvatar(player?.jobTitle || '')}</span>
                <span className="block-amount">{tokenWeight}</span>
              </div>
              <div className="block-hash">
                {isMyVote ? (
                  <div className="flex items-center gap-1">
                    <span className="text-yellow-500">⭐</span>
                    <span className="font-semibold">YOU</span>
                  </div>
                ) : (
                  <span className="text-text-muted text-xs">{player?.name || 'Unknown'}</span>
                )}
              </div>
            </div>
          );
        });
    };
    const hasVoted =
      gameState.voteState?.votes && localPlayer.id in gameState.voteState.votes;
    const myVote = hasVoted ? gameState.voteState?.votes[localPlayer.id] : null;

    return (
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animation-fade-in">
        <div className="mb-1.5">
          <h3 className="text-sm font-bold text-gray-100 normal-case tracking-normal text-left p-0 bg-transparent m-0 mb-3">
            Deactivate {nominatedPlayer.name}?
          </h3>
        </div>
        <div className="verdict-poll">
          <div className="vote-option yes">
            <span className="option-label">✔️ YES</span>
            <span className="vote-tally">🪙 {yesVotes}</span>
            <div className="blockchain-chain">{renderVoteBlocks("GUILTY")}</div>
            <Button
              variant={myVote === "GUILTY" ? "primary" : "secondary"}
              size="sm"
              disabled={hasVoted}
              hapticFeedback="vote"
              onClick={() => {
                playSound("vote");
                setSelectedVote("GUILTY");
                handleVote();
              }}
              className={`option-vote-btn ${myVote === "GUILTY" ? "voted bg-amber-500 border-amber-500 text-black" : hasVoted ? "opacity-50 cursor-not-allowed" : "hover:enabled:bg-green-500 hover:enabled:border-green-500 hover:enabled:text-white"}`}
            >
              VOTE
            </Button>
          </div>
          <div className="vote-option no">
            <span className="option-label">❌ NO</span>
            <span className="vote-tally">🪙 {noVotes}</span>
            <div className="blockchain-chain">
              {renderVoteBlocks("INNOCENT")}
            </div>
            <Button
              variant={myVote === "INNOCENT" ? "primary" : "secondary"}
              size="sm"
              disabled={hasVoted}
              hapticFeedback="vote"
              onClick={() => {
                playSound("vote");
                setSelectedVote("INNOCENT");
                handleVote();
              }}
              className={`option-vote-btn ${myVote === "INNOCENT" ? "voted bg-amber-500 border-amber-500 text-black" : hasVoted ? "opacity-50 cursor-not-allowed" : "hover:enabled:bg-red-500 hover:enabled:border-red-500 hover:enabled:text-white"}`}
            >
              VOTE
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return null;
};
