import React from 'react';
import { useGameContext } from '../../contexts/GameContext';

interface MandateBannerProps {}

export const MandateBanner: React.FC<MandateBannerProps> = () => {
  const { gameState } = useGameContext();

  // Don't show banner if no mandate is active
  if (!gameState?.corporateMandate || !gameState.corporateMandate.isActive) {
    return null;
  }

  const mandate = gameState.corporateMandate;

  // Get mandate type for styling
  const getMandateStyle = (mandateType: string) => {
    switch (mandateType) {
      case 'AGGRESSIVE_GROWTH':
        return {
          bgColor: 'bg-green-900/20',
          borderColor: 'border-green-500',
          iconColor: 'text-green-400',
          icon: '📈'
        };
      case 'TOTAL_TRANSPARENCY':
        return {
          bgColor: 'bg-blue-900/20',
          borderColor: 'border-blue-500',
          iconColor: 'text-blue-400',
          icon: '👁️'
        };
      case 'SECURITY_LOCKDOWN':
        return {
          bgColor: 'bg-red-900/20',
          borderColor: 'border-red-500',
          iconColor: 'text-red-400',
          icon: '🔒'
        };
      default:
        return {
          bgColor: 'bg-gray-900/20',
          borderColor: 'border-gray-500',
          iconColor: 'text-gray-400',
          icon: '📋'
        };
    }
  };

  const style = getMandateStyle(mandate.type);

  // Format effects for display
  const formatEffects = (effects: Record<string, any>) => {
    const summaries = [];
    
    if (effects.starting_tokens_modifier) {
      const modifier = effects.starting_tokens_modifier;
      summaries.push(modifier > 0 ? `+${modifier} starting tokens` : `${modifier} starting tokens`);
    }
    
    if (effects.mining_success_modifier && effects.mining_success_modifier < 1) {
      const reduction = Math.round((1 - effects.mining_success_modifier) * 100);
      summaries.push(`-${reduction}% mining success`);
    }
    
    if (effects.reduced_mining_slots) {
      summaries.push('Reduced mining capacity');
    }
    
    if (effects.public_voting_only) {
      summaries.push('Public voting only');
    }
    
    if (effects.no_direct_messages) {
      summaries.push('No private messages');
    }
    
    if (effects.milestones_for_abilities && effects.milestones_for_abilities > 3) {
      summaries.push(`${effects.milestones_for_abilities} milestones required for abilities`);
    }
    
    if (effects.block_ai_odd_nights) {
      summaries.push('AI restricted on odd nights');
    }
    
    return summaries;
  };

  const effects = formatEffects(mandate.effects || {});

  return (
    <div className={`${style.bgColor} ${style.borderColor} border-2 rounded-lg p-3 mb-4 animate-[fadeIn_0.5s_ease]`}>
      <div className="flex items-center gap-3">
        <div className={`${style.iconColor} text-2xl flex-shrink-0`}>
          {style.icon}
        </div>
        <div className="flex-grow min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="text-sm font-bold text-white uppercase tracking-wide">
              Corporate Mandate
            </h3>
            <span className="text-xs bg-white/10 px-2 py-0.5 rounded-full text-white/80">
              ACTIVE
            </span>
          </div>
          <h4 className="text-base font-semibold text-white mb-1">
            {mandate.name}
          </h4>
          <p className="text-xs text-gray-300 mb-2 leading-relaxed">
            {mandate.description}
          </p>
          {effects.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {effects.map((effect, index) => (
                <span
                  key={index}
                  className="text-xs bg-white/10 px-2 py-0.5 rounded-full text-white/90 font-medium"
                >
                  {effect}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};