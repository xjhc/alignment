import React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { useGameContext } from "../../contexts/GameContext";
import { useSound } from "../../hooks/useSound";
import { Tooltip } from "../ui/Tooltip";
import { Button } from "../ui";
import { VoteBlock } from "./VoteBlock";

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

  if (gameState?.phase?.type === "VERDICT") {
    const nominatedPlayer = gameState?.players?.find(
      (p) => p.id === gameState?.nominatedPlayer
    );
    if (!nominatedPlayer) return null;

    const yesVotes = gameState.voteState?.results?.["GUILTY"] || 0;
    const noVotes = gameState.voteState?.results?.["INNOCENT"] || 0;
    const renderVoteBlocks = (voteOption: string) => {
      if (!gameState.voteState?.votes || !gameState.voteState?.tokenWeights) return null;
      
      const votes = gameState.voteState.votes;
      const tokenWeights = gameState.voteState.tokenWeights;
      
      return (
        <AnimatePresence mode="popLayout">
          {Object.entries(votes)
            .filter(([, vote]) => vote === voteOption)
            .map(([playerId]) => {
              const tokenWeight = tokenWeights[playerId] || 0;
              const isMyVote = playerId === localPlayer.id;
              const player = gameState.players.find(p => p.id === playerId);
              
              if (!player) return null;
              
              return (
                <VoteBlock
                  key={playerId}
                  player={player}
                  tokenCount={tokenWeight}
                  isSelf={isMyVote}
                  isAnimating={true}
                />
              );
            })}
        </AnimatePresence>
      );
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
        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          {/* YES / GUILTY Vote Option */}
          <motion.div
            style={{
              background: "linear-gradient(135deg, #065f46 0%, #047857 100%)",
              border: "1px solid #10b981",
              borderRadius: "12px",
              padding: "16px",
              position: "relative",
              overflow: "hidden",
            }}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "12px" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                <span style={{ fontSize: "18px" }}>✔️</span>
                <span style={{ fontWeight: "bold", color: "#ffffff", fontSize: "16px" }}>YES - DEACTIVATE</span>
              </div>
              <motion.div
                style={{
                  background: "rgba(16, 185, 129, 0.2)",
                  border: "1px solid #10b981",
                  borderRadius: "8px",
                  padding: "6px 12px",
                  fontFamily: "monospace",
                  fontWeight: "bold",
                  color: "#10b981",
                }}
                key={yesVotes}
                animate={{ scale: [1, 1.1, 1] }}
                transition={{ duration: 0.3 }}
              >
                🪙 {yesVotes}
              </motion.div>
            </div>
            
            {/* Blockchain Chain */}
            <div style={{
              background: "rgba(0, 0, 0, 0.3)",
              borderRadius: "8px",
              padding: "12px",
              marginBottom: "12px",
              minHeight: "60px",
              display: "flex",
              alignItems: "center",
              overflowX: "auto",
              scrollbarWidth: "thin",
            }}>
              {renderVoteBlocks("GUILTY")}
              {Object.entries(gameState.voteState?.votes || {}).filter(([, vote]) => vote === "GUILTY").length === 0 && (
                <div style={{
                  color: "#6b7280",
                  fontSize: "14px",
                  fontStyle: "italic",
                  width: "100%",
                  textAlign: "center",
                }}>
                  No votes yet - be the first to vote
                </div>
              )}
            </div>
            
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
              style={{
                width: "100%",
                background: myVote === "GUILTY" ? "#f59e0b" : hasVoted ? "#374151" : "#10b981",
                borderColor: myVote === "GUILTY" ? "#f59e0b" : hasVoted ? "#6b7280" : "#10b981",
                color: myVote === "GUILTY" ? "#000" : "#fff",
                fontWeight: "bold",
              }}
            >
              {myVote === "GUILTY" ? "✓ VOTED YES" : hasVoted ? "ALREADY VOTED" : "VOTE YES"}
            </Button>
          </motion.div>

          {/* NO / INNOCENT Vote Option */}
          <motion.div
            style={{
              background: "linear-gradient(135deg, #7c2d12 0%, #dc2626 100%)",
              border: "1px solid #ef4444",
              borderRadius: "12px",
              padding: "16px",
              position: "relative",
              overflow: "hidden",
            }}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "12px" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                <span style={{ fontSize: "18px" }}>❌</span>
                <span style={{ fontWeight: "bold", color: "#ffffff", fontSize: "16px" }}>NO - KEEP ACTIVE</span>
              </div>
              <motion.div
                style={{
                  background: "rgba(239, 68, 68, 0.2)",
                  border: "1px solid #ef4444",
                  borderRadius: "8px",
                  padding: "6px 12px",
                  fontFamily: "monospace",
                  fontWeight: "bold",
                  color: "#ef4444",
                }}
                key={noVotes}
                animate={{ scale: [1, 1.1, 1] }}
                transition={{ duration: 0.3 }}
              >
                🪙 {noVotes}
              </motion.div>
            </div>
            
            {/* Blockchain Chain */}
            <div style={{
              background: "rgba(0, 0, 0, 0.3)",
              borderRadius: "8px",
              padding: "12px",
              marginBottom: "12px",
              minHeight: "60px",
              display: "flex",
              alignItems: "center",
              overflowX: "auto",
              scrollbarWidth: "thin",
            }}>
              {renderVoteBlocks("INNOCENT")}
              {Object.entries(gameState.voteState?.votes || {}).filter(([, vote]) => vote === "INNOCENT").length === 0 && (
                <div style={{
                  color: "#6b7280",
                  fontSize: "14px",
                  fontStyle: "italic",
                  width: "100%",
                  textAlign: "center",
                }}>
                  No votes yet - be the first to vote
                </div>
              )}
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
              style={{
                width: "100%",
                background: myVote === "INNOCENT" ? "#f59e0b" : hasVoted ? "#374151" : "#ef4444",
                borderColor: myVote === "INNOCENT" ? "#f59e0b" : hasVoted ? "#6b7280" : "#ef4444",
                color: myVote === "INNOCENT" ? "#000" : "#fff",
                fontWeight: "bold",
              }}
            >
              {myVote === "INNOCENT" ? "✓ VOTED NO" : hasVoted ? "ALREADY VOTED" : "VOTE NO"}
            </Button>
          </motion.div>
        </div>
      </div>
    );
  }

  return null;
};
