import React from "react";
import { motion } from "framer-motion";
import { Player } from "../../types/generated";

interface VoteBlockProps {
  player: Player;
  tokenCount: number;
  isSelf: boolean;
  isAnimating?: boolean;
}

export const VoteBlock: React.FC<VoteBlockProps> = ({
  player,
  tokenCount,
  isSelf,
  isAnimating = false,
}) => {
  const getPlayerAvatar = (jobTitle: string) => {
    switch (jobTitle) {
      case "CEO":
        return "👑";
      case "CTO":
        return "💻";
      case "CFO":
        return "💰";
      case "COO":
        return "⚙️";
      case "CISO":
        return "🔒";
      case "Ethics Officer":
        return "⚖️";
      case "Platform Lead":
        return "🏗️";
      case "Intern":
        return "🎓";
      default:
        return "👤";
    }
  };

  return (
    <motion.div
      className={`vote-block ${isSelf ? "my-vote" : ""}`}
      initial={isAnimating ? { opacity: 0, x: 20, scale: 0.8 } : false}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: -20, scale: 0.8 }}
      transition={{
        type: "spring",
        stiffness: 400,
        damping: 25,
        duration: 0.3,
      }}
      style={{
        display: "inline-block",
        margin: "0 4px",
        padding: "8px 10px",
        background: isSelf
          ? "linear-gradient(135deg, var(--accent-amber) 0%, var(--accent-amber-dark) 100%)"
          : "linear-gradient(135deg, var(--bg-tertiary) 0%, var(--bg-quaternary) 100%)",
        border: isSelf ? "2px solid var(--accent-amber)" : "1px solid var(--border)",
        borderRadius: "8px",
        minWidth: "64px",
        boxShadow: isSelf
          ? "0 0 16px var(--accent-amber-light), inset 0 1px 0 rgba(255, 255, 255, 0.1)"
          : "0 2px 8px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.05)",
        position: "relative",
        overflow: "hidden",
      }}
    >
      {/* Blockchain-style header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: "4px",
        }}
      >
        <span
          style={{
            fontSize: "16px",
            lineHeight: "1",
          }}
        >
          {getPlayerAvatar(player.jobTitle)}
        </span>
        <span
          style={{
            fontSize: "12px",
            fontWeight: "bold",
            fontFamily: "monospace",
            color: isSelf ? "var(--bg-primary)" : "var(--accent-amber)",
          }}
        >
          🪙{tokenCount}
        </span>
      </div>

      {/* Player identifier */}
      <div
        style={{
          fontSize: "10px",
          fontFamily: "monospace",
          color: isSelf ? "var(--bg-primary)" : "var(--text-muted)",
          textAlign: "center",
          borderTop: "1px solid " + (isSelf ? "var(--accent-amber-dark)" : "var(--bg-tertiary)"),
          paddingTop: "4px",
          lineHeight: "1.2",
        }}
      >
        {isSelf ? (
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: "2px",
            }}
          >
            <span style={{ color: "var(--bg-primary)", fontSize: "10px" }}>⭐</span>
            <span style={{ fontWeight: "bold", color: "var(--bg-primary)" }}>YOU</span>
          </div>
        ) : (
          <span style={{ fontSize: "9px" }}>
            {player.name.length > 8 ? player.name.slice(0, 8) + "..." : player.name}
          </span>
        )}
      </div>

      {/* Subtle blockchain hash pattern background */}
      {!isSelf && (
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background:
              "repeating-linear-gradient(45deg, rgba(255,255,255,0.02) 0px, rgba(255,255,255,0.02) 1px, transparent 1px, transparent 8px)",
            pointerEvents: "none",
          }}
        />
      )}

      {/* Animated glow for self */}
      {isSelf && (
        <motion.div
          style={{
            position: "absolute",
            top: "-2px",
            left: "-2px",
            right: "-2px",
            bottom: "-2px",
            background:
              "linear-gradient(45deg, var(--accent-amber), var(--accent-amber-dark), var(--accent-amber), var(--accent-amber-dark))",
            borderRadius: "10px",
            zIndex: -1,
            backgroundSize: "400% 400%",
          }}
          animate={{
            backgroundPosition: ["0% 50%", "100% 50%", "0% 50%"],
          }}
          transition={{
            duration: 3,
            repeat: Infinity,
            ease: "linear",
          }}
        />
      )}
    </motion.div>
  );
};