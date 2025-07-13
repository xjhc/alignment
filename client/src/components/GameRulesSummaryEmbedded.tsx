import React from 'react';

export const GameRulesSummaryEmbedded: React.FC = () => {
  return (
    <div className="bg-background-secondary/90 backdrop-blur-sm border border-border rounded-2xl p-6 shadow-lg">
      <h3 className="text-lg font-bold text-text-primary mb-4 flex items-center gap-2">
        <span className="text-base">📋</span>
        Game Rules Overview
      </h3>
      
      <div className="space-y-4">
        {/* Core Objective */}
        <div>
          <h4 className="text-sm font-semibold text-text-primary mb-2">🎯 Objective</h4>
          <div className="bg-background-primary border border-border rounded-lg p-3 space-y-2">
            <div className="flex items-start gap-2 text-xs">
              <span className="text-human">👤</span>
              <span className="text-text-secondary"><strong className="text-human">Humans:</strong> Find and eliminate the AI</span>
            </div>
            <div className="flex items-start gap-2 text-xs">
              <span className="text-ai">🤖</span>
              <span className="text-text-secondary"><strong className="text-ai">AI:</strong> Convert humans or control 51% of tokens</span>
            </div>
            <div className="flex items-start gap-2 text-xs">
              <span className="text-aligned">⚖️</span>
              <span className="text-text-secondary"><strong className="text-aligned">Aligned:</strong> Support the AI's goals</span>
            </div>
          </div>
        </div>

        {/* Game Flow */}
        <div>
          <h4 className="text-sm font-semibold text-text-primary mb-2">🔄 Game Flow</h4>
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-background-primary border border-border rounded-lg p-3">
              <div className="text-xs font-medium text-text-primary mb-1">☀️ Day Phase</div>
              <ul className="text-xs text-text-secondary space-y-1">
                <li>• Discuss & share info</li>
                <li>• Vote to nominate</li>
                <li>• Decide fate</li>
              </ul>
            </div>
            <div className="bg-background-primary border border-border rounded-lg p-3">
              <div className="text-xs font-medium text-text-primary mb-1">🌙 Night Phase</div>
              <ul className="text-xs text-text-secondary space-y-1">
                <li>• Use role abilities</li>
                <li>• Mine tokens</li>
                <li>• Work on projects</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Token System */}
        <div>
          <h4 className="text-sm font-semibold text-text-primary mb-2">🪙 Tokens & Voting</h4>
          <div className="bg-background-primary border border-border rounded-lg p-3">
            <ul className="text-xs text-text-secondary space-y-1">
              <li>• Start with <strong>1 token</strong></li>
              <li>• More tokens = stronger votes</li>
              <li>• Mine during night phases</li>
              <li>• AI wins at 51% token control</li>
            </ul>
          </div>
        </div>

        {/* Key Roles */}
        <div>
          <h4 className="text-sm font-semibold text-text-primary mb-2">👥 Role Types</h4>
          <div className="space-y-2">
            <div className="bg-background-primary border border-border rounded-lg p-2">
              <div className="text-xs">
                <span className="text-human font-medium">👤 Human Roles:</span> 
                <span className="text-text-secondary"> Detective 🔍, Security 🛡️, Manager 💼, Engineer 🔧</span>
              </div>
            </div>
            <div className="bg-background-primary border border-border rounded-lg p-2">
              <div className="text-xs">
                <span className="text-ai font-medium">🤖 AI Role:</span> 
                <span className="text-text-secondary"> Convert & eliminate while staying hidden</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};