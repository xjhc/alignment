import React from 'react';
import styles from './ThreatMeter.module.css';

interface ThreatMeterProps {
  tokens: number;
  aiEquity: number;
}

export const ThreatMeter: React.FC<ThreatMeterProps> = ({ tokens, aiEquity }) => {
  const getBarWidth = () => {
    if (tokens === 0) return 100; // If no tokens, AI would take over immediately
    return Math.min((aiEquity / tokens) * 100, 100);
  };

  const getStatusIndicator = () => {
    const percentage = getBarWidth();
    if (percentage >= 100) return { text: 'ALIGNED', className: 'aligned' };
    if (percentage >= 80) return { text: 'CRITICAL', className: 'critical' };
    if (percentage >= 60) return { text: 'DANGER', className: 'danger' };
    return { text: 'SAFE', className: 'safe' };
  };

  const status = getStatusIndicator();

  return (
    <div className={styles.hudSection}>
      <div className={styles.sectionHeader}>
        <span className={styles.sectionTitle}>🌀 ALIGNMENT</span>
      </div>
      <div className={styles.threatMeterCard}>
        <div className={styles.threatMeterHeader}>
          <span className={styles.threatLabel}>AI EXPOSURE</span>
          <span 
            className={`${styles.visibilityIcon} ${styles.private}`}
            title="If AI Exposure exceeds your Tokens, you will be Aligned."
          >
            🔒
          </span>
        </div>
        <div className={styles.threatMeterBar}>
          <div className={styles.threatBarBg}>
            <div 
              className={styles.threatBarFill} 
              style={{ width: `${getBarWidth()}%` }}
            />
          </div>
        </div>
        <div className={styles.threatMeterValue}>
          <span>Equity: {aiEquity}</span>
          <span>Tokens: {tokens}</span>
          <span className={`${styles.statusIndicator} ${styles[status.className]}`}>
            {status.text}
          </span>
        </div>
      </div>
    </div>
  );
};