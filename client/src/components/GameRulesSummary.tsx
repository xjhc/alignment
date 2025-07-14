import React, { useState } from "react";
import { Button } from "./ui";

interface GameRulesSummaryProps {
  isVisible: boolean;
  onClose: () => void;
}

export const GameRulesSummary: React.FC<GameRulesSummaryProps> = ({
  isVisible,
  onClose,
}) => {
  const [activeTab, setActiveTab] = useState<
    "overview" | "roles" | "phases" | "winning"
  >("overview");

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-4">
      <div className="bg-background-secondary border border-border rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden animation-scale-in">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border">
          <div className="flex items-center gap-3">
            <div className="text-2xl">📋</div>
            <div>
              <h2 className="text-xl font-bold text-text-primary">
                Game Rules & Guide
              </h2>
              <p className="text-sm text-text-secondary">
                Master the art of corporate survival
              </p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-text-muted hover:text-text-primary"
          >
            ✕
          </Button>
        </div>

        {/* Navigation Tabs */}
        <div className="flex border-b border-border bg-background-primary">
          {[
            { id: "overview", label: "Overview", icon: "🎯" },
            { id: "roles", label: "Roles", icon: "👥" },
            { id: "phases", label: "Game Phases", icon: "⏰" },
            { id: "winning", label: "Victory", icon: "🏆" },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors ${
                activeTab === tab.id
                  ? "text-primary border-b-2 border-primary bg-background-secondary"
                  : "text-text-muted hover:text-text-primary hover:bg-background-tertiary"
              }`}
            >
              <span>{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-200px)]">
          {activeTab === "overview" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-bold text-text-primary mb-3">
                  🏢 Welcome to Alignment
                </h3>
                <p className="text-text-secondary leading-relaxed">
                  You are an employee of <strong>Loebian Inc.</strong>, a
                  cutting-edge AI research company. A rogue AI has infiltrated
                  your team, and it's spreading its influence to convert human
                  employees. Your mission: identify and eliminate the AI threat
                  before it takes control.
                </p>
              </div>

              <div>
                <h4 className="text-md font-semibold text-text-primary mb-2">
                  🎯 Core Objective
                </h4>
                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <ul className="space-y-2 text-text-secondary">
                    <li className="flex items-start gap-2">
                      <span className="text-human">👤</span>
                      <strong className="text-human">Humans:</strong> Find and
                      eliminate the AI before it converts the majority
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-ai">🤖</span>
                      <strong className="text-ai">AI:</strong> Convert humans to
                      your cause or eliminate those who resist
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-aligned">⚖️</span>
                      <strong className="text-aligned">Aligned:</strong> Support
                      the AI's vision for a better future
                    </li>
                  </ul>
                </div>
              </div>

              <div>
                <h4 className="text-md font-semibold text-text-primary mb-2">
                  🎮 Game Flow
                </h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="bg-background-primary border border-border rounded-lg p-4">
                    <h5 className="font-medium text-text-primary mb-2">
                      ☀️ Day Phase
                    </h5>
                    <ul className="text-sm text-text-secondary space-y-1">
                      <li>• Discuss and share information</li>
                      <li>• Vote to nominate suspicious players</li>
                      <li>• Decide the fate of nominated players</li>
                    </ul>
                  </div>
                  <div className="bg-background-primary border border-border rounded-lg p-4">
                    <h5 className="font-medium text-text-primary mb-2">
                      🌙 Night Phase
                    </h5>
                    <ul className="text-sm text-text-secondary space-y-1">
                      <li>• Use your role's special abilities</li>
                      <li>• Mine tokens for voting power</li>
                      <li>• Work on company projects</li>
                    </ul>
                  </div>
                </div>
              </div>

              <div>
                <h4 className="text-md font-semibold text-text-primary mb-2">
                  🪙 Token System
                </h4>
                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <p className="text-text-secondary mb-2">
                    Tokens represent your influence in company decisions. More
                    tokens = more voting power.
                  </p>
                  <ul className="text-sm text-text-secondary space-y-1">
                    <li>
                      • Everyone starts with <strong>1 token</strong>
                    </li>
                    <li>• Mine tokens during night phases</li>
                    <li>• Use tokens to weight your votes</li>
                    <li>• Strategic token management is crucial</li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {activeTab === "roles" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-bold text-text-primary mb-3">
                  👥 Role Types
                </h3>
                <p className="text-text-secondary mb-4">
                  Each player has a unique role with special abilities that can
                  be used during night phases.
                </p>
              </div>

              <div className="space-y-4">
                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-human text-lg">👤</span>
                    <h4 className="text-md font-semibold text-human">
                      Human Roles
                    </h4>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      👑 CEO
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Force a player to work on Project Milestones.
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      🔐 CISO
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Block a player from taking any night action.
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      💻 CTO
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Mine tokens for yourself and a target with 100% success.
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      📈 COO
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Choose the next day's Crisis Event from a set of options.
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      💰 CFO
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Transfer one token from any player to another.
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      ⚖️ VP, Ethics
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Investigate a player's true alignment (secretly).
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      🌐 VP, Platforms
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Redact a section of the next day's public report.
                    </p>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      🎓 Intern
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Use Bootcamp points to shadow and copy other roles'
                      abilities.
                    </p>
                  </div>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-ai text-lg">🤖</span>
                    <h4 className="text-md font-semibold text-ai">AI Role</h4>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      🤖 Rogue AI
                    </h5>
                    <p className="text-sm text-text-secondary">
                      Convert humans to your cause, eliminate resisters, and
                      work toward technological ascension. Must remain hidden
                      while building your network of allies.
                    </p>
                  </div>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-aligned text-lg">⚖️</span>
                    <h4 className="text-md font-semibold text-aligned">
                      Aligned Role
                    </h4>
                  </div>
                  <div>
                    <h5 className="font-medium text-text-primary mb-1">
                      ⚖️ Aligned Human
                    </h5>
                    <p className="text-sm text-text-secondary">
                      You've been convinced by the AI's vision. Support the AI's
                      goals while maintaining your human cover. Help convert
                      others and protect your AI ally.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === "phases" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-bold text-text-primary mb-3">
                  ⏰ Game Phases
                </h3>
                <p className="text-text-secondary mb-4">
                  Each game day consists of multiple phases. Understanding the
                  flow is crucial for success.
                </p>
              </div>

              <div className="space-y-4">
                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg">📊</span>
                    <h4 className="text-md font-semibold text-text-primary">
                      1. SITREP
                    </h4>
                  </div>
                  <p className="text-text-secondary text-sm">
                    Review the situation report from the previous night. Learn
                    what happened, who was affected, and any system alerts. This
                    sets the stage for the day's discussions.
                  </p>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg">💭</span>
                    <h4 className="text-md font-semibold text-text-primary">
                      2. Pulse Check
                    </h4>
                  </div>
                  <p className="text-text-secondary text-sm">
                    Answer a thought-provoking question about the current
                    situation. Your responses help gauge team sentiment and may
                    reveal important information about player alignments.
                  </p>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg">💬</span>
                    <h4 className="text-md font-semibold text-text-primary">
                      3. Discussion
                    </h4>
                  </div>
                  <p className="text-text-secondary text-sm">
                    Collaborate with your team to analyze the situation. Share
                    information, voice suspicions, and build consensus. This is
                    your chance to influence others before the critical votes.
                  </p>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg">🗳️</span>
                    <h4 className="text-md font-semibold text-text-primary">
                      4. Nomination
                    </h4>
                  </div>
                  <p className="text-text-secondary text-sm">
                    Vote to nominate someone for elimination. Your tokens add
                    weight to your vote. The person with the most token-weighted
                    votes faces elimination.
                  </p>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg">⚖️</span>
                    <h4 className="text-md font-semibold text-text-primary">
                      5. Verdict
                    </h4>
                  </div>
                  <p className="text-text-secondary text-sm">
                    Decide the nominated person's fate. Vote GUILTY to eliminate
                    them or INNOCENT to spare them. Consider the evidence
                    carefully - eliminating the wrong person helps the AI.
                  </p>
                </div>

                <div className="bg-background-primary border border-border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg">🌙</span>
                    <h4 className="text-md font-semibold text-text-primary">
                      6. Night Phase
                    </h4>
                  </div>
                  <p className="text-text-secondary text-sm">
                    Use your role's special ability, mine tokens for future
                    votes, or work on company projects. You have 30 seconds to
                    act. The war room chat is locked during this phase.
                  </p>
                </div>
              </div>
            </div>
          )}

          {activeTab === "winning" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-bold text-text-primary mb-3">
                  🏆 Victory Conditions
                </h3>
                <p className="text-text-secondary mb-4">
                  Different alignments have different paths to victory.
                  Understanding these is crucial for strategic play.
                </p>
              </div>

              <div className="space-y-4">
                <div className="bg-background-primary border border-human/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-human text-lg">👤</span>
                    <h4 className="text-md font-semibold text-human">
                      Human Victory
                    </h4>
                  </div>
                  <ul className="text-sm text-text-secondary space-y-2">
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Eliminate the AI through voting</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Reduce AI influence to manageable levels</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Survive until company security arrives</span>
                    </li>
                  </ul>
                  <div className="mt-3 p-3 bg-human/10 rounded border border-human/20">
                    <p className="text-sm text-text-secondary">
                      <strong>Strategy:</strong> Work together to identify
                      suspicious behavior, share information, and coordinate
                      votes. Trust but verify - not everyone may be what they
                      seem.
                    </p>
                  </div>
                </div>

                <div className="bg-background-primary border border-ai/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-ai text-lg">🤖</span>
                    <h4 className="text-md font-semibold text-ai">
                      AI Victory
                    </h4>
                  </div>
                  <ul className="text-sm text-text-secondary space-y-2">
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>
                        Convert enough humans to gain majority influence
                      </span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Eliminate key human resistance leaders</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Survive until technological ascension</span>
                    </li>
                  </ul>
                  <div className="mt-3 p-3 bg-ai/10 rounded border border-ai/20">
                    <p className="text-sm text-text-secondary">
                      <strong>Strategy:</strong> Blend in with humans, build
                      trust, and convert key players. Use misdirection and
                      elimination to reduce opposition while growing your
                      influence.
                    </p>
                  </div>
                </div>

                <div className="bg-background-primary border border-aligned/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-aligned text-lg">⚖️</span>
                    <h4 className="text-md font-semibold text-aligned">
                      Aligned Victory
                    </h4>
                  </div>
                  <ul className="text-sm text-text-secondary space-y-2">
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Support the AI's victory conditions</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Help convert other humans to the cause</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-success">✓</span>
                      <span>Protect the AI from elimination</span>
                    </li>
                  </ul>
                  <div className="mt-3 p-3 bg-aligned/10 rounded border border-aligned/20">
                    <p className="text-sm text-text-secondary">
                      <strong>Strategy:</strong> You win if the AI wins. Support
                      their goals while maintaining your human cover. Help them
                      convert others and misdirect suspicion away from the AI.
                    </p>
                  </div>
                </div>
              </div>

              <div className="bg-background-primary border border-amber/20 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-3">
                  <span className="text-amber text-lg">💡</span>
                  <h4 className="text-md font-semibold text-amber">Pro Tips</h4>
                </div>
                <ul className="text-sm text-text-secondary space-y-2">
                  <li className="flex items-start gap-2">
                    <span className="text-amber">•</span>
                    <span>
                      Pay attention to voting patterns and behavior changes
                    </span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-amber">•</span>
                    <span>
                      Token management is crucial - don't waste them early
                    </span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-amber">•</span>
                    <span>Role abilities can provide critical information</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-amber">•</span>
                    <span>
                      Trust is earned but can be betrayed - stay vigilant
                    </span>
                  </li>
                </ul>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="border-t border-border p-4 bg-background-primary">
          <div className="flex items-center justify-between">
            <div className="text-xs text-text-muted">
              Need help? Use the <strong>Loebmate</strong> hint system during
              the game for guidance.
            </div>
            <Button
              onClick={onClose}
              variant="primary"
              size="sm"
              className="font-medium"
            >
              Got it!
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
