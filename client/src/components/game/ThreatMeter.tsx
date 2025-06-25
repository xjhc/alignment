import React, { useEffect, useState } from 'react';

interface ThreatMeterProps {
  tokens: number;
  aiEquity: number;
}

export const ThreatMeter: React.FC<ThreatMeterProps> = ({ tokens, aiEquity }) => {
  const [animatedWidth, setAnimatedWidth] = useState(0);
  const [previousWidth, setPreviousWidth] = useState(0);

  const getBarWidth = () => {
    if (tokens === 0) return 100; // If no tokens, AI would take over immediately
    return Math.min((aiEquity / tokens) * 100, 100);
  };

  const currentWidth = getBarWidth();

  useEffect(() => {
    // Animate bar width changes
    if (currentWidth !== previousWidth) {
      const startTime = Date.now();
      const startWidth = animatedWidth;
      const targetWidth = currentWidth;
      const duration = 500; // 500ms animation

      const animate = () => {
        const elapsed = Date.now() - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        // Easing function for smooth animation
        const easeOut = 1 - Math.pow(1 - progress, 3);
        const newWidth = startWidth + (targetWidth - startWidth) * easeOut;
        
        setAnimatedWidth(newWidth);
        
        if (progress < 1) {
          requestAnimationFrame(animate);
        } else {
          setPreviousWidth(currentWidth);
        }
      };
      
      requestAnimationFrame(animate);
    }
  }, [currentWidth, previousWidth, animatedWidth]);

  const getStatusIndicator = () => {
    const percentage = getBarWidth();
    if (percentage >= 100) return { text: 'ALIGNED', className: 'aligned' };
    if (percentage >= 80) return { text: 'CRITICAL', className: 'critical' };
    if (percentage >= 60) return { text: 'DANGER', className: 'danger' };
    return { text: 'SAFE', className: 'safe' };
  };

  const status = getStatusIndicator();

  return (
    <div className="animate-fade-in">
      <div className="flex justify-between items-center mb-2">
        <span className="text-[11px] font-bold text-text-muted uppercase">🌀 ALIGNMENT</span>
      </div>
      <div className="bg-background-tertiary border border-border rounded-lg p-2.5 px-3">
        <div className="flex justify-between items-center mb-1.5">
          <span className="text-[10px] font-bold text-text-muted uppercase">AI EXPOSURE</span>
          <span 
            className="text-[8px] opacity-60 cursor-help text-pink-500"
            title="If AI Exposure exceeds your Tokens, you will be Aligned."
          >
            🔒
          </span>
        </div>
        <div className="mb-1.5">
          <div className="w-full h-1.5 bg-background-primary rounded-sm overflow-hidden">
            <div 
              className="h-full bg-pink-500 rounded-sm transition-all duration-300 ease-out"
              style={{ 
                width: `${animatedWidth}%`,
              }}
            />
          </div>
        </div>
        <div className="flex justify-between items-center font-mono text-[10px] text-text-muted">
          <span>Equity: {aiEquity}</span>
          <span>Tokens: {tokens}</span>
          <span className={`font-bold px-1 py-0.5 rounded text-[9px] ${
            status.className === 'safe' 
              ? 'text-success bg-success/10' 
              : status.className === 'danger'
              ? 'text-danger bg-danger/10'
              : status.className === 'critical'
              ? 'text-danger bg-danger/10 animate-pulse'
              : status.className === 'aligned'
              ? 'text-ai bg-ai/10'
              : ''
          }`}>
            {status.text}
          </span>
        </div>
      </div>
    </div>
  );
};