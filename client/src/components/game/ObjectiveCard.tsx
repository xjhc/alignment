import React from 'react';
import { GeneratedPersonalKPI, KPIType } from '../../types/generated';

interface ObjectiveCardProps {
  type: 'Team Objective' | 'Personal KPI' | 'Mandate';
  name: string;
  description: string;
  progressText?: string;
  isPrivate?: boolean;
  personalKPI?: GeneratedPersonalKPI;
}

export const ObjectiveCard: React.FC<ObjectiveCardProps> = ({ 
  type, 
  name, 
  description, 
  progressText,
  isPrivate = false,
  personalKPI
}) => {
  const getCardClassName = () => {
    const baseClass = 'bg-background-tertiary border border-border rounded-lg p-2.5 mb-1.5';
    switch (type) {
      case 'Team Objective':
        return `${baseClass} border-l-2 border-l-blue-500`;
      case 'Personal KPI':
        return `${baseClass} border-l-2 border-l-success`;
      case 'Mandate':
        return `${baseClass} border-l-2 border-l-info`;
      default:
        return baseClass;
    }
  };

  const getKPIProgressText = () => {
    if (type !== 'Personal KPI' || !personalKPI) return progressText;
    
    if (personalKPI.isCompleted) {
      return '✓ Completed';
    }
    
    // Generate progress text based on KPI type
    switch (personalKPI.type) {
      case KPIType.Inquisitor:
        return `Correct votes: ${personalKPI.progress}/${personalKPI.target}`;
      case KPIType.Guardian:
        return `CISO survived: Day ${personalKPI.progress}/${personalKPI.target}`;
      case KPIType.Capitalist:
        return 'Progress tracked at game end';
      case KPIType.SuccessionPlanner:
        return 'Progress tracked at game end';
      case KPIType.Scapegoat:
        return 'Progress tracked during voting';
      default:
        return `Progress: ${personalKPI.progress}/${personalKPI.target}`;
    }
  };

  const getProgressPercentage = () => {
    if (type !== 'Personal KPI' || !personalKPI || personalKPI.target === 0) return 0;
    return Math.min(100, (personalKPI.progress / personalKPI.target) * 100);
  };

  return (
    <div className={getCardClassName()}>
      <div className="text-[9px] font-bold uppercase text-text-muted mb-0.5">{type}</div>
      <div className="font-bold mb-0.5 text-xs flex items-center gap-1">
        {name}
        {isPrivate && (
          <span 
            className="text-[8px] opacity-60 cursor-help text-pink-500"
            title="Your secret objective"
          >
            🔒
          </span>
        )}
      </div>
      <div className="text-text-secondary text-[11px] leading-tight mb-0.5">{description}</div>
      {getKPIProgressText() && (
        <div className="text-[10px] text-success font-medium mb-1">{getKPIProgressText()}</div>
      )}
      {type === 'Personal KPI' && personalKPI && personalKPI.target > 0 && !personalKPI.isCompleted && (
        <div className="w-full bg-background-secondary rounded-full h-1.5">
          <div 
            className="bg-success h-1.5 rounded-full transition-all duration-300 ease-out"
            style={{ width: `${getProgressPercentage()}%` }}
          />
        </div>
      )}
    </div>
  );
};