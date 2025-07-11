import { useWebSocketContext } from '../contexts/WebSocketContext';
import { useSessionContext } from '../contexts/SessionContext';
import { useState, useEffect } from 'react';
import { Button } from './ui';
import { websocketClient } from '../services/websocket';

interface ReconnectionOverlayProps {
  /** Whether to show the overlay */
  show: boolean;
  /** Optional callback when overlay is dismissed */
  onDismiss?: () => void;
}

export function ReconnectionOverlay({ show }: ReconnectionOverlayProps) {
  const { isReconnecting, lastError } = useWebSocketContext();
  const { onLeaveLobby } = useSessionContext();
  const [reconnectionStatus, setReconnectionStatus] = useState({
    isReconnecting: false,
    attempts: 0,
    maxAttempts: 10
  });
  const [isDismissed, setIsDismissed] = useState(false);
  const [isTimedOut, setIsTimedOut] = useState(false);
  const [timeoutCountdown, setTimeoutCountdown] = useState(10);

  useEffect(() => {
    // This effect ensures the overlay becomes visible if reconnection starts
    // while it's hidden. The "show" prop controls initial visibility.
    const initialStatus = websocketClient.getReconnectionStatus();
    setReconnectionStatus(initialStatus);
  }, []);

  // Timeout effect - starts countdown when show becomes true
  useEffect(() => {
    if (!show) {
      setIsTimedOut(false);
      setTimeoutCountdown(10);
      return;
    }

    let timeoutTimer: NodeJS.Timeout;
    let countdownTimer: NodeJS.Timeout;

    // Start the timeout countdown
    const startCountdown = () => {
      let countdown = 10;
      setTimeoutCountdown(countdown);
      
      const updateCountdown = () => {
        countdown -= 1;
        setTimeoutCountdown(countdown);
        
        if (countdown <= 0) {
          setIsTimedOut(true);
          clearInterval(countdownTimer);
        } else {
          countdownTimer = setTimeout(updateCountdown, 1000);
        }
      };
      
      countdownTimer = setTimeout(updateCountdown, 1000);
    };

    // Set main timeout
    timeoutTimer = setTimeout(() => {
      setIsTimedOut(true);
    }, 10000); // 10 seconds

    // Start countdown
    startCountdown();

    return () => {
      clearTimeout(timeoutTimer);
      clearTimeout(countdownTimer);
    };
  }, [show]);

  const handleDismiss = () => {
    setIsDismissed(true);
  };

  const handleForceReconnect = () => {
    // Reset timeout state when retrying
    setIsTimedOut(false);
    setTimeoutCountdown(10);
    websocketClient.forceReconnect();
  };

  const handleRefresh = () => {
    window.location.reload();
  };

  const handleReturnToLobbyList = () => {
    // Clear the broken session and return to lobby list
    onLeaveLobby();
  };

  // Show if explicitly told to, or if reconnection is actively happening (and not dismissed).
  const shouldShow = show || (isReconnecting && !isDismissed);

  if (!shouldShow) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 animation-fade-in">
      <div className="bg-background-secondary border border-border rounded-lg p-6 max-w-md w-full mx-4 shadow-lg animation-scale-in">
        {!isTimedOut ? (
          // Normal syncing state
          <>
            <div className="flex items-center gap-3 mb-4">
              <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
              <h2 className="text-lg font-bold text-text-primary">Syncing Session...</h2>
            </div>
            
            <div className="space-y-4">
              <div className="text-text-secondary">
                <p className="mb-2">
                  Restoring your session. Please wait...
                </p>
                {timeoutCountdown > 0 && (
                  <p className="text-xs text-text-muted">
                    Timeout in {timeoutCountdown} seconds
                  </p>
                )}
              </div>

              {isReconnecting && (
                <div className="bg-warning/10 border border-warning/20 rounded-lg p-3 text-sm">
                  <p className="text-warning">
                    Connection lost. Attempting to reconnect... 
                    (Attempt {reconnectionStatus.attempts}/{reconnectionStatus.maxAttempts})
                  </p>
                </div>
              )}

              {lastError && (
                <div className="bg-danger/10 border border-danger/20 rounded-lg p-3 text-sm">
                  <p className="text-danger">
                    Last error: {lastError}
                  </p>
                </div>
              )}

              <div className="bg-background-primary border border-border rounded-lg p-3">
                <h4 className="text-sm font-medium text-text-primary mb-2">While reconnecting:</h4>
                <ul className="text-xs text-text-secondary space-y-1">
                  <li>• Your spot in the game is reserved.</li>
                  <li>• Other players will see you as "reconnecting"</li>
                </ul>
              </div>

              {/* Action Buttons */}
              <div className="flex gap-2">
                <Button
                  onClick={handleForceReconnect}
                  variant="secondary"
                  size="sm"
                  className="flex-1"
                >
                  🔄 Retry Now
                </Button>
                <Button
                  onClick={handleRefresh}
                  variant="primary"
                  size="sm"
                  className="flex-1"
                >
                  🔄 Refresh Page
                </Button>
              </div>
              
              {isReconnecting && (
                <Button
                  onClick={handleDismiss}
                  variant="ghost"
                  size="sm"
                  className="w-full text-text-muted"
                >
                  Continue in Background
                </Button>
              )}
            </div>
          </>
        ) : (
          // Timeout error state
          <>
            <div className="flex items-center gap-3 mb-4">
              <div className="w-6 h-6 text-danger">⚠️</div>
              <h2 className="text-lg font-bold text-danger">Session Sync Failed</h2>
            </div>
            
            <div className="space-y-4">
              <div className="bg-danger/10 border border-danger/20 rounded-lg p-3">
                <p className="text-danger text-sm">
                  Failed to sync session. The server may be busy or the lobby may have closed.
                </p>
              </div>

              <div className="text-text-secondary text-sm">
                <p className="mb-2">What would you like to do?</p>
              </div>

              {/* Timeout Action Buttons */}
              <div className="space-y-2">
                <Button
                  onClick={handleForceReconnect}
                  variant="primary"
                  size="sm"
                  className="w-full"
                >
                  🔄 Retry Connection
                </Button>
                <Button
                  onClick={handleRefresh}
                  variant="secondary"
                  size="sm"
                  className="w-full"
                >
                  🔄 Refresh Page
                </Button>
                <Button
                  onClick={handleReturnToLobbyList}
                  variant="ghost"
                  size="sm"
                  className="w-full text-text-muted"
                >
                  ← Return to Lobby List
                </Button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}