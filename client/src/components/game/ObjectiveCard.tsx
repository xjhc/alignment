import React from 'react';
import styles from './ObjectiveCard.module.css';

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
    const baseClass = styles.objectiveCard;
    switch (type) {
      case 'Team Objective':
        return `${baseClass} ${styles.faction}`;
      case 'Personal KPI':
        return `${baseClass} ${styles.kpi}`;
      case 'Mandate':
        return `${baseClass} ${styles.mandate}`;
      default:
        return baseClass;
    }
  };

  return (
    <div className={getCardClassName()}>
      <div className={styles.objectiveType}>{type}</div>
      <div className={styles.objectiveName}>
        {name}
        {isPrivate && (
          <span 
            className={`${styles.visibilityIcon} ${styles.private}`}
            title="Your secret objective"
          >
            🔒
          </span>
        )}
      </div>
      <div className={styles.objectiveDesc}>{description}</div>
      {progressText && (
        <div className={styles.objectiveProgress}>{progressText}</div>
      )}
    </div>
  );
};