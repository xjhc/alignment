import React, { useState } from "react";
import { useGameContext } from "../../contexts/GameContext";
import { Player } from "../../types";
import { Button } from "../ui";

interface NightActionSelectionProps {}

type ActionType = "mine" | "project" | "ability" | "bootcamp" | "shadow" | null;

export const NightActionSelection: React.FC<NightActionSelectionProps> = () => {
  const {
    gameState,
    localPlayer,
    setMiningTarget,
    handleMineTokens,
    handleUseAbility,
    handleProjectMilestones,
    canPlayerAffordAbility,
    isValidNightActionTarget,
  } = useGameContext();

  if (!localPlayer) return null;

  const [selectedAction, setSelectedAction] = useState<ActionType>(null);
  const [selectedTarget, setSelectedTarget] = useState<string>("");
  const [shadowTarget, setShadowTarget] = useState<string>("");
  const [shadowStep, setShadowStep] = useState<
    "select-source" | "select-target"
  >("select-source");
  const [isMinimized, setIsMinimized] = useState(false);

  const players = gameState?.players || [];
  const alivePlayers = Array.isArray(players)
    ? players.filter((p) => p.isAlive)
    : [];
  const hasUnlockedAbility =
    localPlayer.role?.type && localPlayer.projectMilestones >= 3;

  const isIntern = localPlayer.role?.type === "INTERN";
  const bootcampPoints = localPlayer.bootcampPoints || 0;
  const canUseShadow = isIntern && bootcampPoints >= 1;
  const eligibleShadowTargets = alivePlayers.filter(
    (p) =>
      p.role?.type &&
      p.projectMilestones >= 3 &&
      ["CISO", "CTO", "CEO", "COO", "CFO", "ETHICS", "PLATFORMS"].includes(
        p.role.type
      )
  );

  const handleActionSelect = (action: ActionType) => {
    setSelectedAction(action);
    setSelectedTarget("");
    setShadowTarget("");
    setShadowStep("select-source");

    if (action === "project") {
      handleProjectMilestones();
      setIsMinimized(true);
    } else if (action === "bootcamp") {
      // This would call a new `handleInternBootcamp` from context if it existed
      console.log("BOOTCAMP action submitted");
      setIsMinimized(true);
    }
  };

  const handleTargetSelect = (targetId: string) => {
    if (selectedAction === "shadow") {
      if (shadowStep === "select-source") {
        setSelectedTarget(targetId);
        setShadowStep("select-target");
      } else if (shadowStep === "select-target") {
        setShadowTarget(targetId);
        // This would call a new `handleInternShadow` from context
        console.log("SHADOW action submitted");
        setIsMinimized(true);
      }
    } else {
      setSelectedTarget(targetId);

      if (selectedAction === "mine") {
        setMiningTarget(targetId);
        handleMineTokens();
        setIsMinimized(true);
      } else if (selectedAction === "ability") {
        handleUseAbility(targetId);
        setIsMinimized(true);
      }
    }
  };

  const getActionSuccessRate = (action: ActionType) => {
    if (action === "mine") {
      if (alivePlayers.length === 0) return "N/A";
      const totalMiningAttempts = alivePlayers.length;
      const successRate = Math.max(
        33,
        Math.min(75, Math.floor((3 / totalMiningAttempts) * 100))
      );
      return `${successRate}% Success`;
    }
    return null;
  };

  const getPlayerIcon = (player: Player) => {
    switch (player.jobTitle) {
      case "CISO":
        return "👤";
      case "Systems":
        return "🧑‍💻";
      case "Ethics":
        return "🕵️";
      case "CTO":
        return "🤖";
      case "COO":
        return "🧑‍🚀";
      case "CFO":
        return "👩‍🔬";
      default:
        return "👤";
    }
  };

  if (isMinimized) {
    return (
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
        <div className="p-2">
          <div className="flex flex-col gap-1">
            <div
              className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md cursor-pointer transition-all duration-150 hover:bg-gray-600 border-amber-500 ${selectedAction === "mine" ? "bg-amber-500/10 border-amber-500" : ""}`}
              onClick={() => setIsMinimized(false)}
            >
              <span className="text-xs w-4 text-center">⛏️</span>
              <span className="font-medium text-xs text-gray-100 flex-grow">
                Mine for Player
              </span>
              <span className="ml-auto text-xs font-bold text-blue-500 font-mono">
                {selectedAction === "mine" && selectedTarget ? "SELECTED" : ""}
              </span>
            </div>
            <div
              className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md cursor-pointer transition-all duration-150 hover:bg-gray-600 ${selectedAction === "project" ? "bg-amber-500/10 border-amber-500" : ""}`}
              onClick={() => handleActionSelect("project")}
            >
              <span className="text-xs w-4 text-center">📈</span>
              <span className="font-medium text-xs text-gray-100 flex-grow">
                Project Milestones
              </span>
              <span className="text-xs text-gray-500 bg-gray-900 px-1 py-0.5 rounded-lg font-medium uppercase">
                Role
              </span>
            </div>
            {isIntern ? (
              <>
                <div
                  className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md cursor-pointer transition-all duration-150 hover:bg-gray-600 ${selectedAction === "bootcamp" ? "bg-amber-500/10 border-amber-500" : ""}`}
                  onClick={() => handleActionSelect("bootcamp")}
                >
                  <span className="text-xs w-4 text-center">📚</span>
                  <span className="font-medium text-xs text-gray-100 flex-grow">
                    Bootcamp
                  </span>
                  <span className="text-xs text-blue-500 bg-blue-500/10 px-1 py-0.5 rounded-lg font-medium">
                    +1 Point
                  </span>
                </div>
                <div
                  className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md transition-all duration-150 ${canUseShadow ? `cursor-pointer hover:bg-gray-600 ${selectedAction === "shadow" ? "bg-amber-500/10 border-amber-500" : ""}` : "opacity-60 cursor-not-allowed grayscale-30"}`}
                  onClick={() => canUseShadow && handleActionSelect("shadow")}
                >
                  <span className="text-xs w-4 text-center">👤</span>
                  <span className="font-medium text-xs text-gray-100 flex-grow">
                    Shadow
                  </span>
                  <span className="text-xs text-green-500 font-bold">
                    {canUseShadow ? "Ready" : `Need ${1 - bootcampPoints} pts`}
                  </span>
                </div>
              </>
            ) : (
              <div
                className={`flex items-center gap-1.5 px-2 py-1.5 bg-gray-700 border border-gray-600 rounded-md transition-all duration-150 ${hasUnlockedAbility && canPlayerAffordAbility(localPlayer.id) ? "cursor-pointer hover:bg-gray-600" : "opacity-60 cursor-not-allowed grayscale-30"}`}
              >
                <span className="text-xs w-4 text-center">🔒</span>
                <span className="font-medium text-xs text-gray-100 flex-grow">
                  {localPlayer.role?.type || "Role Ability"}
                </span>
                <span className="text-xs text-red-500 font-bold">
                  {hasUnlockedAbility
                    ? canPlayerAffordAbility(localPlayer.id)
                      ? "Ready"
                      : "No Tokens"
                    : "Locked"}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
      <div className="flex flex-col gap-4">
        <div className="m-0">
          <h3 className="text-sm font-bold text-gray-100 m-0 mb-3">
            Choose Night Action
          </h3>
        </div>

        <div className="flex flex-col gap-2">
          <div
            className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 cursor-pointer transition-all duration-150 hover:bg-gray-700 hover:-translate-y-0.5 ${selectedAction === "mine" ? "border-amber-500 bg-amber-500/10" : ""}`}
            onClick={() => handleActionSelect("mine")}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-base w-5 text-center">⛏️</span>
              <span className="font-semibold text-gray-100 text-sm flex-grow">
                Mine for Player
              </span>
              <span className="text-xs text-green-500 bg-green-500/10 px-1.5 py-0.5 rounded-2xl font-bold">
                {getActionSuccessRate("mine")}
              </span>
            </div>
            <div className="text-xs text-gray-400 leading-snug mb-1.5">
              Generate 1 Token for any player. Success depends on liquidity
              pool (currently 3 slots for {alivePlayers.length} players
              attempting).
            </div>
            <div className="text-xs text-gray-500">
              <strong className="text-gray-100 font-semibold">Command:</strong>{" "}
              <code className="bg-gray-700 px-1 py-0.5 rounded font-mono text-xs">
                Mine for [Player Name]
              </code>
            </div>
          </div>
          <div
            className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 cursor-pointer transition-all duration-150 hover:bg-gray-700 hover:-translate-y-0.5 ${selectedAction === "project" ? "border-amber-500 bg-amber-500/10" : ""}`}
            onClick={() => handleActionSelect("project")}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-base w-5 text-center">📈</span>
              <span className="font-semibold text-gray-100 text-sm flex-grow">
                Project Milestones
              </span>
              <span className="text-xs text-gray-500 bg-gray-700 px-1.5 py-0.5 rounded-2xl font-medium uppercase">
                Role Ability
              </span>
            </div>
            <div className="text-xs text-gray-400 leading-snug mb-1.5">
              Advance your role's project by 1 point. At 3 points, unlock your
              powerful role-specific ability for future nights.
            </div>
            <div className="text-xs text-gray-500">
              <strong className="text-gray-100 font-semibold">Command:</strong>{" "}
              <code className="bg-gray-700 px-1 py-0.5 rounded font-mono text-xs">
                Project Milestones
              </code>
            </div>
          </div>
          {isIntern ? (
            <>
              <div
                className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 cursor-pointer transition-all duration-150 hover:bg-gray-700 hover:-translate-y-0.5 ${selectedAction === "bootcamp" ? "border-amber-500 bg-amber-500/10" : ""}`}
                onClick={() => handleActionSelect("bootcamp")}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-base w-5 text-center">📚</span>
                  <span className="font-semibold text-gray-100 text-sm flex-grow">
                    Bootcamp Training
                  </span>
                  <span className="text-xs text-blue-500 bg-blue-500/10 px-1.5 py-0.5 rounded-2xl font-bold">
                    +1 Point
                  </span>
                </div>
                <div className="text-xs text-gray-400 leading-snug mb-1.5">
                  Spend the night in intensive training. Gain 1 Bootcamp Point
                  that can be spent to copy another player's role ability.
                  Current Points: {bootcampPoints}
                </div>
                <div className="text-xs text-gray-500">
                  <strong className="text-gray-100 font-semibold">
                    Command:
                  </strong>{" "}
                  <code className="bg-gray-700 px-1 py-0.5 rounded font-mono text-xs">
                    Bootcamp Training
                  </code>
                </div>
              </div>
              <div
                className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 transition-all duration-150 ${canUseShadow ? `cursor-pointer hover:bg-gray-700 hover:-translate-y-0.5 ${selectedAction === "shadow" ? "border-amber-500 bg-amber-500/10" : ""}` : "opacity-60 cursor-not-allowed grayscale-30"}`}
                onClick={() => canUseShadow && handleActionSelect("shadow")}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-base w-5 text-center">👤</span>
                  <span className="font-semibold text-gray-100 text-sm flex-grow">
                    Shadow Ability
                  </span>
                  <span className="text-xs text-green-500 bg-green-500/10 px-1.5 py-0.5 rounded-2xl font-bold">
                    {canUseShadow ? "Ready" : `Need ${1 - bootcampPoints} pts`}
                  </span>
                </div>
                <div className="text-xs text-gray-400 leading-snug mb-1.5">
                  {canUseShadow
                    ? `Copy another player's unlocked role ability and use it immediately. Costs 1 Bootcamp Point. Available targets: ${eligibleShadowTargets.length}`
                    : `Requires 1 Bootcamp Point to shadow another player's ability. Current Points: ${bootcampPoints}`}
                </div>
                <div className="text-xs text-gray-500">
                  <strong className="text-gray-100 font-semibold">
                    Command:
                  </strong>{" "}
                  <code className="bg-gray-700 px-1 py-0.5 rounded font-mono text-xs">
                    Shadow [Player] targeting [Target]
                  </code>
                </div>
              </div>
            </>
          ) : (
            <div
              className={`px-3 py-2 rounded-lg border border-gray-600 bg-gray-800 transition-all duration-150 ${hasUnlockedAbility && canPlayerAffordAbility(localPlayer.id) ? `cursor-pointer hover:bg-gray-700 hover:-translate-y-0.5 ${selectedAction === "ability" ? "border-amber-500 bg-amber-500/10" : ""}` : "opacity-60 cursor-not-allowed grayscale-30"}`}
              onClick={() =>
                hasUnlockedAbility &&
                canPlayerAffordAbility(localPlayer.id) &&
                handleActionSelect("ability")
              }
            >
              <div className="flex items-center gap-2 mb-2">
                <span className="text-base w-5 text-center">🔒</span>
                <span className="font-semibold text-gray-100 text-sm flex-grow">
                  {localPlayer.role?.type || "Role Ability"}
                </span>
                <span className="text-xs text-red-500 bg-red-500/10 px-1.5 py-0.5 rounded-2xl font-bold">
                  {hasUnlockedAbility
                    ? canPlayerAffordAbility(localPlayer.id)
                      ? "Ready"
                      : "No Tokens"
                    : "Locked"}
                </span>
              </div>
              <div className="text-xs text-gray-400 leading-snug mb-1.5">
                {hasUnlockedAbility
                  ? `Use your role-specific ability. ${canPlayerAffordAbility(localPlayer.id) ? "Available to use." : "Requires more tokens."}`
                  : "Role ability requires system access currently unavailable."}
              </div>
              <div className="text-xs text-gray-500">
                <strong className="text-gray-100 font-semibold">Status:</strong>{" "}
                {hasUnlockedAbility
                  ? "System access granted"
                  : "Network security protocols offline"}
              </div>
            </div>
          )}
        </div>
        {selectedAction === "mine" && (
          <div className="pt-3 border-t border-gray-600">
            <div className="mb-2">
              <h4 className="text-xs font-semibold text-gray-100 m-0">
                Select Mining Target
              </h4>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {alivePlayers.map((player) => (
                <Button
                  key={player.id}
                  variant={selectedTarget === player.id ? "primary" : "ghost"}
                  size="sm"
                  onClick={() => handleTargetSelect(player.id)}
                  className={`flex items-center gap-1.5 px-2.5 py-1.5 ${selectedTarget === player.id ? "border-amber-500 bg-amber-500/10" : ""} ${player.alignment === "ALIGNED" ? "bg-cyan-500/5 border-cyan-600" : ""}`}
                >
                  <span className="text-sm leading-none flex-shrink-0">
                    {getPlayerIcon(player)}
                  </span>
                  <span
                    className={`font-medium text-xs text-gray-100 flex-shrink-0 ${player.alignment === "ALIGNED" ? "text-cyan-600 animate-[glitch_1.5s_infinite]" : ""}`}
                  >
                    {player.name}
                  </span>
                  <span className="font-mono font-bold text-gray-400 text-xs ml-auto">
                    🪙 {player.tokens}
                  </span>
                </Button>
              ))}
            </div>
          </div>
        )}
        {selectedAction === "ability" && hasUnlockedAbility && (
          <div className="pt-3 border-t border-gray-600">
            <div className="mb-2">
              <h4 className="text-xs font-semibold text-gray-100 m-0">
                Select Ability Target
              </h4>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {alivePlayers.map((player) => (
                <Button
                  key={player.id}
                  variant={selectedTarget === player.id ? "primary" : "ghost"}
                  size="sm"
                  onClick={() =>
                    isValidNightActionTarget(
                      localPlayer.id,
                      player.id,
                      "ability"
                    ) && handleTargetSelect(player.id)
                  }
                  disabled={
                    !isValidNightActionTarget(
                      localPlayer.id,
                      player.id,
                      "ability"
                    )
                  }
                  className={`flex items-center gap-1.5 px-2.5 py-1.5 ${selectedTarget === player.id ? "border-amber-500 bg-amber-500/10" : ""} ${!isValidNightActionTarget(localPlayer.id, player.id, "ability") ? "grayscale-50" : ""}`}
                >
                  <span className="text-sm leading-none flex-shrink-0">
                    {getPlayerIcon(player)}
                  </span>
                  <span className="font-medium text-xs text-gray-100 flex-shrink-0">
                    {player.name}
                  </span>
                  <span className="font-mono font-bold text-gray-400 text-xs ml-auto">
                    🪙 {player.tokens}
                  </span>
                </Button>
              ))}
            </div>
          </div>
        )}
        {selectedAction === "shadow" && canUseShadow && (
          <div className="pt-3 border-t border-gray-600">
            {shadowStep === "select-source" && (
              <>
                <div className="mb-2">
                  <h4 className="text-xs font-semibold text-gray-100 m-0">
                    Step 1: Select Player to Shadow
                  </h4>
                  <p className="text-xs text-gray-400 mt-1">
                    Choose a player whose ability you want to copy. They must
                    have an unlocked role ability.
                  </p>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {eligibleShadowTargets.map((player) => (
                    <Button
                      key={player.id}
                      variant={
                        selectedTarget === player.id ? "primary" : "ghost"
                      }
                      size="sm"
                      onClick={() => handleTargetSelect(player.id)}
                      className={`flex items-center gap-1.5 px-2.5 py-1.5 ${selectedTarget === player.id ? "border-amber-500 bg-amber-500/10" : ""}`}
                    >
                      <span className="text-sm leading-none flex-shrink-0">
                        {getPlayerIcon(player)}
                      </span>
                      <span className="font-medium text-xs text-gray-100 flex-shrink-0">
                        {player.name}
                      </span>
                      <span className="text-xs text-purple-400 font-semibold">
                        {player.role?.type}
                      </span>
                      <span className="font-mono font-bold text-gray-400 text-xs ml-auto">
                        📈 {player.projectMilestones}
                      </span>
                    </Button>
                  ))}
                </div>
              </>
            )}
            {shadowStep === "select-target" && selectedTarget && (
              <>
                <div className="mb-2">
                  <h4 className="text-xs font-semibold text-gray-100 m-0">
                    Step 2: Select Target for{" "}
                    {
                      gameState.players.find((p) => p.id === selectedTarget)
                        ?.role?.type
                    }{" "}
                    Ability
                  </h4>
                  <p className="text-xs text-gray-400 mt-1">
                    You will copy{" "}
                    {
                      gameState.players.find((p) => p.id === selectedTarget)
                        ?.name
                    }
                    's{" "}
                    {
                      gameState.players.find((p) => p.id === selectedTarget)
                        ?.role?.type
                    }{" "}
                    ability and use it on this target.
                  </p>
                </div>
                <div className="flex flex-wrap gap-1.5 mb-3">
                  {alivePlayers.map((player) => (
                    <Button
                      key={player.id}
                      variant={shadowTarget === player.id ? "primary" : "ghost"}
                      size="sm"
                      onClick={() => handleTargetSelect(player.id)}
                      className={`flex items-center gap-1.5 px-2.5 py-1.5 ${shadowTarget === player.id ? "border-amber-500 bg-amber-500/10" : ""}`}
                    >
                      <span className="text-sm leading-none flex-shrink-0">
                        {getPlayerIcon(player)}
                      </span>
                      <span className="font-medium text-xs text-gray-100 flex-shrink-0">
                        {player.name}
                      </span>
                      <span className="font-mono font-bold text-gray-400 text-xs ml-auto">
                        🪙 {player.tokens}
                      </span>
                    </Button>
                  ))}
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setShadowStep("select-source");
                    setShadowTarget("");
                  }}
                  className="text-xs text-gray-400 hover:text-gray-100"
                >
                  ← Back to Step 1
                </Button>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
