import React, { useState } from 'react';
import { Tooltip } from './Tooltip';

interface GlossaryTooltipProps {
  term: string;
  children: React.ReactNode;
  className?: string;
}

// Game terminology definitions
const GLOSSARY: Record<string, string> = {
  'System Shock': 'A temporary status effect that impairs a player\'s abilities. Can cause message corruption or action locks.',
  'LIAISON Protocol': 'An emergency AI detection system that reveals alignment information when specific triggers are met.',
  'Tokens': 'The in-game currency that determines your voting power. Earned through mining and role abilities.',
  'Pulse Check': 'A daily sentiment survey where players share their thoughts on the current crisis situation.',
  'KPI': 'Key Performance Indicator - your personal objective that provides additional victory conditions.',
  'War Room': 'The main chat channel where all players communicate during most phases.',
  'Night Actions': 'Special abilities that players can use during the Night phase, including mining and role powers.',
  'Mining': 'The process of earning tokens during the Night phase to increase your voting influence.',
  'Project Milestones': 'Collaborative objectives that players can advance during Night phases.',
  'AI Faction': 'Players who have been converted by the rogue AI and work to take control of the company.',
  'Human Faction': 'The original employees working to identify and eliminate the AI threat.',
  'Alignment': 'Your current faction - either Human or AI. Can change during the game through conversion.',
  'Role Abilities': 'Special powers unique to each corporate role (CEO, CTO, CISO, etc.).',
  'Crisis Events': 'Random challenges that affect gameplay and provide context for Pulse Check questions.',
  'Whistleblower Protocol': 'A voting system for eliminated players to influence which crisis occurs next.',
  'Extension Vote': 'A emergency vote during Discussion phase to extend the time limit.',
  'Nomination': 'The process of voting to select a player for potential elimination.',
  'Verdict': 'The final vote to determine whether a nominated player is eliminated.',
  'Elimination': 'The removal of a player from the game, revealing their role and alignment.',
  'Corporate Mandate': 'A game modifier that changes rules or communication restrictions for the entire match.'
};

export function GlossaryTooltip({ term, children, className = '' }: GlossaryTooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  
  const definition = GLOSSARY[term];
  
  if (!definition) {
    // If term not found, render children without tooltip
    return <>{children}</>;
  }

  return (
    <Tooltip
      content={
        <div className="max-w-xs">
          <div className="font-semibold text-sm mb-1 text-primary">{term}</div>
          <div className="text-xs text-text-secondary leading-relaxed">{definition}</div>
        </div>
      }
      position="top"
      className={className}
    >
      <span 
        className={`
          cursor-help border-b border-dotted border-primary/50 
          hover:border-primary transition-colors duration-200
          ${className}
        `}
        onMouseEnter={() => setIsVisible(true)}
        onMouseLeave={() => setIsVisible(false)}
      >
        {children}
      </span>
    </Tooltip>
  );
}

// Convenience component for wrapping text with glossary terms
export function GlossaryText({ 
  children, 
  terms = [] 
}: { 
  children: string; 
  terms?: string[] 
}) {
  let text = children;
  
  // Auto-detect common terms if none specified
  const autoTerms = terms.length > 0 ? terms : Object.keys(GLOSSARY);
  
  // Replace each term with a tooltip-wrapped version
  autoTerms.forEach(term => {
    const regex = new RegExp(`\\b${term}\\b`, 'gi');
    text = text.replace(regex, (match) => `<tooltip>${match}</tooltip>`);
  });
  
  // Split text and wrap tooltip markers
  const parts = text.split(/(<tooltip>.*?<\/tooltip>)/g);
  
  return (
    <>
      {parts.map((part, index) => {
        if (part.startsWith('<tooltip>') && part.endsWith('</tooltip>')) {
          const termText = part.slice(9, -10); // Remove <tooltip></tooltip>
          return (
            <GlossaryTooltip key={index} term={termText}>
              {termText}
            </GlossaryTooltip>
          );
        }
        return part;
      })}
    </>
  );
}