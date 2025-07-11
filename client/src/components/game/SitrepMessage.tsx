import React from 'react';
import { ChatMessage, GameState, DailySitrep, SitrepSection } from '../../types';

interface SitrepMessageProps {
  message: ChatMessage;
  gameState: GameState;
}

export const SitrepMessage: React.FC<SitrepMessageProps> = ({ message, gameState }) => {
  const metadata = message.metadata;
  
  // Check if this is a new SITREP_PUBLISHED event with structured DailySitrep
  const dailySitrep = metadata?.daily_sitrep as DailySitrep;
  
  if (dailySitrep) {
    // Render the new structured SITREP
    return (
      <div className="flex gap-3 p-3 bg-background-tertiary border border-border/30 rounded-lg mb-2">
        <div className="w-8 h-8 rounded-full bg-ai text-white flex items-center justify-center text-sm">🤖</div>
        <div className="flex-1">
          <span className="text-ai font-mono font-bold text-sm">Loebmate</span>
          <div className="text-text-primary bg-background-quaternary border border-border/20 rounded-lg p-3 mt-2 font-mono text-sm">
            <div className="flex items-center justify-between mb-3">
              <strong>DAILY SITREP - DAY {dailySitrep.day_number}</strong>
              <span className={`px-2 py-1 text-xs rounded ${
                dailySitrep.alert_level === 'CRITICAL' ? 'bg-danger text-danger-foreground' :
                dailySitrep.alert_level === 'HIGH' ? 'bg-warning text-warning-foreground' :
                dailySitrep.alert_level === 'ELEVATED' ? 'bg-info text-info-foreground' :
                'bg-success text-success-foreground'
              }`}>
                {dailySitrep.alert_level}
              </span>
            </div>
            
            {dailySitrep.sections.map((section, index) => (
              <div key={index} className="mb-4">
                <strong>{section.title}</strong><br/>
                <div className={`${
                  section.type === 'classified' ? 'bg-warning/10 border-l-4 border-warning pl-3' :
                  section.type === 'redacted' ? 'bg-danger/10 border-l-4 border-danger pl-3' :
                  ''
                }`}>
                  {section.content.split('\n').map((line, lineIndex) => (
                    <span key={lineIndex}>
                      {line}<br/>
                    </span>
                  ))}
                </div>
              </div>
            ))}
            
            <div className="text-xs text-text-secondary mt-4 pt-2 border-t border-border/20">
              {dailySitrep.footer_note}
            </div>
          </div>
        </div>
      </div>
    );
  }
  
  // Fallback to legacy SITREP format
  const nightActionResults = gameState.nightActionResults || metadata?.nightActions || [];
  
  const players = gameState?.players || [];
  const headcount = metadata?.playerHeadcount || {
    humans: Array.isArray(players) ? players.filter(p => p.isAlive && p.alignment !== 'ALIGNED').length : 0,
    aligned: Array.isArray(players) ? players.filter(p => p.isAlive && p.alignment === 'ALIGNED').length : 0,
    dead: Array.isArray(players) ? players.filter(p => !p.isAlive).length : 0
  };
  const crisisEvent = metadata?.crisisEvent || gameState.crisisEvent;

  // Helper function to generate night action descriptions from structured data
  const generateNightActionDescriptions = (results: any[]): string[] => {
    const descriptions: string[] = [];
    
    if (!results || results.length === 0) {
      return descriptions;
    }

    results.forEach((result: any) => {
      // Handle different types of night action results
      if (result.blocked_players && result.blocked_players.length > 0) {
        result.blocked_players.forEach((blocked: any) => {
          descriptions.push(`${blocked.player_name} was blocked by ${blocked.blocker_name}`);
        });
      }
      
      if (result.converted_players && result.converted_players.length > 0) {
        result.converted_players.forEach((converted: any) => {
          descriptions.push(`${converted.player_name} was converted to the AI faction`);
        });
      }
      
      if (result.mining_results && result.mining_results.length > 0) {
        result.mining_results.forEach((mining: any) => {
          if (mining.success) {
            descriptions.push(`${mining.miner_name} successfully mined ${mining.tokens_mined} tokens from ${mining.target_name}`);
          } else {
            descriptions.push(`${mining.miner_name}'s mining attempt on ${mining.target_name} failed`);
          }
        });
      }
      
      if (result.role_ability_results && result.role_ability_results.length > 0) {
        result.role_ability_results.forEach((ability: any) => {
          descriptions.push(`${ability.player_name} used ${ability.ability_type}: ${ability.message}`);
        });
      }
      
      if (result.milestone_results && result.milestone_results.length > 0) {
        result.milestone_results.forEach((milestone: any) => {
          if (milestone.role_unlocked) {
            descriptions.push(`${milestone.player_name} unlocked their role ability (${milestone.milestones_count} milestones)`);
          }
        });
      }
      
      if (result.failed_actions && result.failed_actions.length > 0) {
        result.failed_actions.forEach((failed: any) => {
          descriptions.push(`${failed.player_name}'s ${failed.action_type} action failed: ${failed.reason}`);
        });
      }
      
      // Use summary message if available
      if (result.summary_message) {
        descriptions.push(result.summary_message);
      }
    });
    
    return descriptions;
  };

  // Generate descriptions from structured data or use legacy format
  const nightActionDescriptions = Array.isArray(nightActionResults) && nightActionResults.length > 0 
    ? generateNightActionDescriptions(nightActionResults)
    : nightActionResults.map((action: any) => action.description || action).filter(Boolean);

  return (
    <div className="flex gap-3 p-3 bg-background-tertiary border border-border/30 rounded-lg mb-2">
      <div className="w-8 h-8 rounded-full bg-ai text-white flex items-center justify-center text-sm">🤖</div>
      <div className="flex-1">
        <span className="text-ai font-mono font-bold text-sm">Loebmate</span>
        <div className="text-text-primary bg-background-quaternary border border-border/20 rounded-lg p-3 mt-2 font-mono text-sm">
          <strong>Good morning, team. Here's the SITREP.</strong><br/><br/>
          
          <strong>NIGHT {gameState.dayNumber - 1} ACTIVITY LOG:</strong><br/>
          {nightActionDescriptions.length > 0 ? (
            nightActionDescriptions.map((description, index) => (
              <span key={index}>
                • {description}<br/>
              </span>
            ))
          ) : (
            <span>• No significant activity detected.<br/></span>
          )}
          <br/>
          
          <strong>HR HEADCOUNT:</strong><br/>
          • <strong>{headcount.humans} Human Life-signs Detected</strong><br/>
          • <strong>{headcount.aligned} Aligned Agents Active</strong> 🤖<br/>
          {headcount.dead > 0 && (
            <>• <strong>{headcount.dead} Personnel Deactivated</strong> 👻<br/></>
          )}
          <br/>
          
          {crisisEvent && (
            <>
              <div className="bg-danger/10 border-l-4 border-danger p-3 my-3">
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-danger text-lg">⚠️</span>
                  <strong className="text-danger text-lg">CRITICAL INCIDENT</strong>
                  <span className="text-danger text-lg">⚠️</span>
                </div>
                <div className="text-danger font-bold text-base mb-2">{crisisEvent.title}</div>
                <div className="text-text-primary">{crisisEvent.description}</div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};