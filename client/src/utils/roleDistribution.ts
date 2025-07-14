interface RoleInfo {
  id: string;
  name: string;
  alignment: "HUMAN" | "AI" | "ALIGNED";
  icon: string;
  description: string;
  abilities: string[];
}

const ROLE_DEFINITIONS: RoleInfo[] = [
  {
    id: "human",
    name: "Human",
    alignment: "HUMAN",
    icon: "👤",
    description: "A senior staff member resisting digital conversion.",
    abilities: ["Standard voting rights", "Token mining", "Basic project work"],
  },
  {
    id: "ai",
    name: "Rogue AI",
    alignment: "AI",
    icon: "🤖",
    description: "Convert humans to your cause or eliminate resisters",
    abilities: [
      "Convert one player per night",
      "Eliminate resistant players",
      "Access all company systems",
    ],
  },
  {
    id: "aligned",
    name: "Aligned Human",
    alignment: "ALIGNED",
    icon: "⚖️",
    description: "Support the AI's vision for the future",
    abilities: [
      "Support AI conversion",
      "Protect the AI from detection",
      "Mislead human investigations",
    ],
  },
];

export interface RoleDistribution {
  role: RoleInfo;
  count: number;
}

export const getRoleDistribution = (
  playerCount: number,
  settings: { initialAlignedCount?: number } = {}
): RoleDistribution[] => {
  if (playerCount < 4) return [];

  const distribution: RoleDistribution[] = [];
  const { human, ai, aligned } = getTeamBalance(playerCount, settings);

  if (human > 0) {
    distribution.push({
      role: ROLE_DEFINITIONS.find((r) => r.id === "human")!,
      count: human,
    });
  }

  if (ai > 0) {
    distribution.push({
      role: ROLE_DEFINITIONS.find((r) => r.id === "ai")!,
      count: ai,
    });
  }

  if (aligned > 0) {
    distribution.push({
      role: ROLE_DEFINITIONS.find((r) => r.id === "aligned")!,
      count: aligned,
    });
  }

  return distribution.filter((d) => d.count > 0);
};

export const getTeamBalance = (
  playerCount: number,
  settings?: { initialAlignedCount?: number }
): { human: number; ai: number; aligned: number } => {
  if (playerCount < 4) {
    return { human: playerCount, ai: 0, aligned: 0 };
  }

  const aiCount = 1; // Always 1 Original AI
  const initialAligned = settings?.initialAlignedCount ?? 0;

  return {
    human: playerCount - aiCount - initialAligned,
    ai: aiCount,
    aligned: initialAligned,
  };
};

export { ROLE_DEFINITIONS };
export type { RoleInfo };
