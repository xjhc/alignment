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
    switch (type) {
      case 'Team Objective':
        return 'objective-card faction';
      case 'Personal KPI':
        return 'objective-card kpi';
      case 'Mandate':
        return 'objective-card mandate';
      default:
        return 'objective-card';
    }
  };

  return (
    <div className={getCardClassName()}>
      <div className="objective-type">{type}</div>
      <div className="objective-name">
        {name}
        {isPrivate && (
          <span 
            className="visibility-icon private"
            title="Your secret objective"
          >
            🔒
          </span>
        )}
      </div>
      <div className="objective-desc">{description}</div>
      {progressText && (
        <div className="objective-progress">{progressText}</div>
      )}
    </div>
  );
};