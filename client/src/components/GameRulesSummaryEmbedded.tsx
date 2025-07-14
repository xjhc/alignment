import React from 'react';

export const GameRulesSummaryEmbedded: React.FC = () => {
  return (
    <div>
      <h3 className="text-sm font-bold text-text-primary mb-3 flex items-center gap-2">
        <span className="text-base">📋</span>
        Game Rules Overview
      </h3>
      
      <div className="space-y-3">
        {/* Core Objective */}
        <div className="bg-background-primary border border-border rounded-lg p-3 space-y-1">
          <div className="text-xs font-medium text-text-primary mb-2">Objectives</div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-human">👤 Humans</span>
            <span className="text-text-secondary">Find the AI</span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-ai">🤖 AI</span>
            <span className="text-text-secondary">Convert/Control</span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-aligned">⚖️ Aligned</span>
            <span className="text-text-secondary">Help AI</span>
          </div>
        </div>

        {/* Game Flow */}
        <div className="bg-background-primary border border-border rounded-lg p-3">
          <div className="text-xs font-medium text-text-primary mb-2">Game Flow</div>
          <div className="grid grid-cols-2 gap-2 text-xs">
            <div>
              <span className="text-text-primary">☀️ Day:</span>
              <span className="text-text-secondary"> Discuss, Vote, Eliminate</span>
            </div>
            <div>
              <span className="text-text-primary">🌙 Night:</span>
              <span className="text-text-secondary"> Abilities, Mine, Projects</span>
            </div>
          </div>
        </div>

        {/* Key Mechanics */}
        <div className="bg-background-primary border border-border rounded-lg p-3 space-y-1">
          <div className="text-xs font-medium text-text-primary mb-2">Key Mechanics</div>
          <div className="text-xs text-text-secondary">
            🪙 <strong>Tokens:</strong> More tokens = stronger votes
          </div>
          <div className="text-xs text-text-secondary">
            🎭 <strong>Roles:</strong> Special abilities for investigation & protection
          </div>
          <div className="text-xs text-text-secondary">
            🤖 <strong>AI Victory:</strong> Convert majority or control 51% tokens
          </div>
        </div>
      </div>
    </div>
  );
};