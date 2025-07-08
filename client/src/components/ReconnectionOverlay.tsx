import { useWebSocketContext } from '../contexts/WebSocketContext';

interface ReconnectionOverlayProps {
  /** Whether to show the overlay */
  show: boolean;
}

export function ReconnectionOverlay({ show }: ReconnectionOverlayProps) {
  const { isReconnecting, lastError } = useWebSocketContext();

  if (!show || !isReconnecting) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 animation-fade-in">
      <div className="bg-surface-primary border border-stroke rounded-lg p-6 max-w-md mx-4 text-center animation-scale-in">
        <div className="flex items-center justify-center mb-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-accent-primary"></div>
        </div>
        
        <h2 className="text-lg font-semibold text-text-primary mb-2">
          Reconnecting...
        </h2>
        
        <p className="text-text-secondary text-sm mb-4">
          Attempting to reconnect to the game. Please wait a moment.
        </p>
        
        {lastError && (
          <div className="bg-danger-bg border border-danger-border rounded p-3 text-sm">
            <p className="text-danger-text">
              Connection issue: {lastError}
            </p>
            <p className="text-danger-text mt-1">
              Retrying automatically...
            </p>
          </div>
        )}
      </div>
    </div>
  );
}