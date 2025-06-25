import React from 'react';

interface ObjectiveCardProps {
  type: 'Team Objective' | 'Personal KPI' | 'Mandate';
  name: string;
  description: string;
  progressText?: string;
  isPrivate?: boolean;
}

export const ObjectiveCard: React.FC<ObjectiveCardProps> = ({ 
  type, 
  name, 
  description, 
  progressText,
  isPrivate = false 
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
      {progressText && (
        <div className="text-[10px] text-success font-medium">{progressText}</div>
      )}
    </div>
  );
};