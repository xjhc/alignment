import React from 'react';

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
    <div className="hud-section">
      <div className="section-header">
        <span className="section-title">🌀 ALIGNMENT</span>
      </div>
      <div className="threat-meter-card">
        <div className="threat-meter-header">
          <span className="threat-label">AI EXPOSURE</span>
          <span 
            className="visibility-icon private"
            title="If AI Exposure exceeds your Tokens, you will be Aligned."
          >
            🔒
          </span>
        </div>
        <div className="threat-meter-bar">
          <div className="threat-bar-bg">
            <div 
              className="threat-bar-fill" 
              style={{ width: `${getBarWidth()}%` }}
            />
          </div>
        </div>
        <div className="threat-meter-value">
          <span>Equity: {aiEquity}</span>
          <span>Tokens: {tokens}</span>
          <span className={`status-indicator ${status.className}`}>
            {status.text}
          </span>
        </div>
      </div>
    </div>
  );
};