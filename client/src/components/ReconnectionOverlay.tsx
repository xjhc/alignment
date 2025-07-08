import { useWebSocketContext } from '../contexts/WebSocketContext';
import { useState, useEffect } from 'react';
import { Button } from './ui';
import { websocketClient } from '../services/websocket';

interface ReconnectionOverlayProps {
  /** Whether to show the overlay */
  show: boolean;
  /** Optional callback when overlay is dismissed */
  onDismiss?: () => void;
}

export function ReconnectionOverlay({ show, onDismiss }: ReconnectionOverlayProps) {
  const { isReconnecting, lastError } = useWebSocketContext();
  const [reconnectionStatus, setReconnectionStatus] = useState({
    isReconnecting: false,
    attempts: 0,
    maxAttempts: 10
  });
  const [isDismissed, setIsDismissed] = useState(false);

  useEffect(() => {
    const handleReconnectionChange = (isReconnecting: boolean, attempts: number) => {
      setReconnectionStatus(prev => ({
        ...prev,
        isReconnecting,
        attempts
      }));
      
      // Reset dismissed state when reconnection starts
      if (isReconnecting) {
        setIsDismissed(false);
      }
    };

    websocketClient.onReconnect(handleReconnectionChange);
    
    // Get initial status
    const initialStatus = websocketClient.getReconnectionStatus();
    setReconnectionStatus(initialStatus);

    return () => {
      websocketClient.offReconnect(handleReconnectionChange);
    };
  }, []);

  const handleDismiss = () => {
    setIsDismissed(true);
    onDismiss?.();
  };

  const handleForceReconnect = () => {
    websocketClient.forceReconnect();
  };

  const handleRefresh = () => {
    window.location.reload();
  };

  const shouldShow = show && isReconnecting && !isDismissed;

  if (!shouldShow) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animation-fade-in">
      <div className="bg-background-secondary border border-danger/20 rounded-lg p-6 max-w-md w-full mx-4 shadow-lg animation-scale-in">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-6 h-6 border-2 border-danger border-t-transparent rounded-full animate-spin"></div>
          <h2 className="text-lg font-bold text-danger">Connection Lost</h2>
        </div>
        
        <div className="space-y-4">
          <div className="text-text-secondary">
            <p className="mb-2">
              Lost connection to the lobby. Attempting to reconnect...
            </p>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-text-muted">Attempt:</span>
              <span className="font-medium text-text-primary">
                {reconnectionStatus.attempts} / {reconnectionStatus.maxAttempts}
              </span>
            </div>
          </div>

          {/* Progress Bar */}
          <div className="w-full bg-background-primary rounded-full h-2">
            <div 
              className="bg-danger h-2 rounded-full transition-all duration-300"
              style={{ 
                width: `${(reconnectionStatus.attempts / reconnectionStatus.maxAttempts) * 100}%` 
              }}
            />
          </div>

          {lastError && (
            <div className="bg-danger/10 border border-danger/20 rounded-lg p-3 text-sm">
              <p className="text-danger">
                Connection issue: {lastError}
              </p>
              <p className="text-danger mt-1">
                Retrying automatically...
              </p>
            </div>
          )}

          {/* Reconnection Tips */}
          <div className="bg-background-primary border border-border rounded-lg p-3">
            <h4 className="text-sm font-medium text-text-primary mb-2">While reconnecting:</h4>
            <ul className="text-xs text-text-secondary space-y-1">
              <li>• Your spot in the lobby is reserved</li>
              <li>• Other players will see you as "reconnecting"</li>
              <li>• The game will wait for you to reconnect</li>
              <li>• You won't miss any important updates</li>
            </ul>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-2">
            <Button
              onClick={handleForceReconnect}
              variant="primary"
              size="sm"
              className="flex-1"
            >
              🔄 Retry Now
            </Button>
            <Button
              onClick={handleRefresh}
              variant="secondary"
              size="sm"
              className="flex-1"
            >
              🔄 Refresh Page
            </Button>
          </div>
          
          <Button
            onClick={handleDismiss}
            variant="ghost"
            size="sm"
            className="w-full text-text-muted"
          >
            Continue in Background
          </Button>
        </div>
      </div>
    </div>
  );
}