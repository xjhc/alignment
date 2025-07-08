import React, { useState, useEffect } from 'react';
import { Button } from './ui';

interface ConnectionQualityIndicatorProps {
  isConnected: boolean;
  lastPingTime?: number;
  reconnectAttempts?: number;
  onReconnect?: () => void;
}

interface ConnectionStats {
  latency: number;
  quality: 'excellent' | 'good' | 'poor' | 'offline';
  packetLoss: number;
  reconnectCount: number;
}

export const ConnectionQualityIndicator: React.FC<ConnectionQualityIndicatorProps> = ({
  isConnected,
  lastPingTime,
  reconnectAttempts = 0,
  onReconnect
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [stats, setStats] = useState<ConnectionStats>({
    latency: 0,
    quality: 'offline',
    packetLoss: 0,
    reconnectCount: reconnectAttempts
  });
  const [pingHistory, setPingHistory] = useState<number[]>([]);

  // Update connection quality based on ping and connection status
  useEffect(() => {
    if (!isConnected) {
      setStats(prev => ({
        ...prev,
        quality: 'offline',
        reconnectCount: reconnectAttempts
      }));
      return;
    }

    if (lastPingTime) {
      const newPing = lastPingTime;
      setPingHistory(prev => [...prev.slice(-9), newPing]); // Keep last 10 pings
      
      let quality: ConnectionStats['quality'] = 'excellent';
      if (newPing > 200) quality = 'poor';
      else if (newPing > 100) quality = 'good';
      
      setStats(prev => ({
        ...prev,
        latency: newPing,
        quality,
        packetLoss: Math.max(0, (reconnectAttempts * 5)) // Estimate packet loss
      }));
    }
  }, [isConnected, lastPingTime, reconnectAttempts]);

  const getQualityColor = (quality: ConnectionStats['quality']) => {
    switch (quality) {
      case 'excellent': return 'text-success';
      case 'good': return 'text-amber';
      case 'poor': return 'text-danger';
      case 'offline': return 'text-text-muted';
    }
  };

  const getQualityIcon = (quality: ConnectionStats['quality']) => {
    switch (quality) {
      case 'excellent': return '📶';
      case 'good': return '📶';
      case 'poor': return '📶';
      case 'offline': return '📵';
    }
  };

  const getQualityText = (quality: ConnectionStats['quality']) => {
    switch (quality) {
      case 'excellent': return 'Excellent';
      case 'good': return 'Good';
      case 'poor': return 'Poor';
      case 'offline': return 'Offline';
    }
  };

  const getConnectionBars = (quality: ConnectionStats['quality']) => {
    const bars = [];
    const barCount = quality === 'excellent' ? 4 : quality === 'good' ? 3 : quality === 'poor' ? 2 : 1;
    
    for (let i = 0; i < 4; i++) {
      bars.push(
        <div
          key={i}
          className={`w-1 h-3 rounded-sm ${
            i < barCount && isConnected ? 'bg-current' : 'bg-text-muted/30'
          }`}
          style={{ height: `${(i + 1) * 25}%` }}
        />
      );
    }
    
    return bars;
  };

  return (
    <div className="relative">
      {/* Connection Indicator */}
      <div 
        className={`flex items-center gap-2 px-3 py-2 rounded-lg border cursor-pointer transition-all ${
          isConnected 
            ? 'bg-background-secondary border-border hover:bg-background-tertiary' 
            : 'bg-danger/10 border-danger/20 hover:bg-danger/20'
        }`}
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className={`flex items-end gap-px ${getQualityColor(stats.quality)}`}>
          {getConnectionBars(stats.quality)}
        </div>
        <div className="text-sm">
          <span className={`font-medium ${getQualityColor(stats.quality)}`}>
            {getQualityText(stats.quality)}
          </span>
          {stats.latency > 0 && (
            <span className="text-text-muted ml-1">
              {stats.latency}ms
            </span>
          )}
        </div>
        <div className="text-xs text-text-muted">
          {isExpanded ? '▼' : '▶'}
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-background-secondary border border-border rounded-lg p-4 shadow-lg z-50 animation-fade-in">
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-text-primary mb-2">Connection Details</h4>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-text-muted">Status:</span>
                  <span className={`ml-2 font-medium ${getQualityColor(stats.quality)}`}>
                    {isConnected ? 'Connected' : 'Disconnected'}
                  </span>
                </div>
                <div>
                  <span className="text-text-muted">Quality:</span>
                  <span className={`ml-2 font-medium ${getQualityColor(stats.quality)}`}>
                    {getQualityText(stats.quality)}
                  </span>
                </div>
                <div>
                  <span className="text-text-muted">Latency:</span>
                  <span className="ml-2 font-medium text-text-primary">
                    {stats.latency > 0 ? `${stats.latency}ms` : 'N/A'}
                  </span>
                </div>
                <div>
                  <span className="text-text-muted">Reconnects:</span>
                  <span className="ml-2 font-medium text-text-primary">
                    {stats.reconnectCount}
                  </span>
                </div>
              </div>
            </div>

            {/* Ping History Graph */}
            {pingHistory.length > 0 && (
              <div>
                <h4 className="text-sm font-medium text-text-primary mb-2">Latency History</h4>
                <div className="flex items-end gap-1 h-12 p-2 bg-background-primary rounded border border-border">
                  {pingHistory.map((ping, index) => {
                    const height = Math.min((ping / 300) * 100, 100); // Max 300ms
                    return (
                      <div
                        key={index}
                        className="flex-1 bg-primary/60 rounded-sm min-h-px"
                        style={{ height: `${height}%` }}
                        title={`${ping}ms`}
                      />
                    );
                  })}
                </div>
                <div className="text-xs text-text-muted mt-1 text-center">
                  Last 10 pings
                </div>
              </div>
            )}

            {/* Connection Tips */}
            <div>
              <h4 className="text-sm font-medium text-text-primary mb-2">Connection Tips</h4>
              <div className="text-xs text-text-secondary space-y-1">
                {stats.quality === 'poor' && (
                  <>
                    <div>• Try moving closer to your router</div>
                    <div>• Close other applications using internet</div>
                    <div>• Check for network congestion</div>
                  </>
                )}
                {stats.quality === 'offline' && (
                  <>
                    <div>• Check your internet connection</div>
                    <div>• Try refreshing the page</div>
                    <div>• Contact support if issues persist</div>
                  </>
                )}
                {stats.quality === 'good' && (
                  <div>• Connection is stable for normal gameplay</div>
                )}
                {stats.quality === 'excellent' && (
                  <div>• Optimal connection for best experience</div>
                )}
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex gap-2 pt-2 border-t border-border">
              {!isConnected && onReconnect && (
                <Button
                  onClick={onReconnect}
                  variant="primary"
                  size="sm"
                  className="text-xs"
                >
                  🔄 Reconnect
                </Button>
              )}
              <Button
                onClick={() => window.location.reload()}
                variant="secondary"
                size="sm"
                className="text-xs"
              >
                🔄 Refresh Page
              </Button>
              <Button
                onClick={() => setIsExpanded(false)}
                variant="ghost"
                size="sm"
                className="text-xs"
              >
                Close
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};