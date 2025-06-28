import React, { useState } from 'react';
import { useGameContext } from '../contexts/GameContext';
import { PrivateNotifications } from './PrivateNotifications';
import { RosterPanel } from './game/RosterPanel';
import { CommsPanel } from './game/CommsPanel';
import { PlayerHUD } from './game/PlayerHUD';
import { ExitInterviewScreen } from './game/ExitInterviewScreen';

export function GameScreen() {
  const { gameState, localPlayer } = useGameContext();
  const [showExitInterview, setShowExitInterview] = useState(false);

  // Show exit interview for eliminated players who haven't submitted a parting shot
  React.useEffect(() => {
    if (localPlayer && !localPlayer.isAlive && !localPlayer.partingShot && !showExitInterview) {
      setShowExitInterview(true);
    }
  }, [localPlayer, showExitInterview]);

  if (!localPlayer) {
    return (
      <div className="w-screen h-screen flex flex-col bg-background-primary">
        <div className="flex items-center justify-center text-text-muted text-sm">
          <span>Loading game state...</span>
          <div className="loading-spinner large" style={{ marginLeft: '12px' }}></div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-screen h-screen grid grid-cols-[260px_1fr_320px] gap-px bg-border overflow-hidden">
      {/* Exit Interview Overlay */}
      {showExitInterview && (
        <ExitInterviewScreen
          onComplete={() => setShowExitInterview(false)}
        />
      )}

      {/* Private Notifications Overlay */}
      {gameState.privateNotifications && (
        <PrivateNotifications
          notifications={gameState.privateNotifications}
          onMarkAsRead={(notificationId) => {
            // TODO: Send action to mark notification as read
            console.log('Mark notification as read:', notificationId);
          }}
        />
      )}
      
      <RosterPanel />
      
      <CommsPanel />
      
      <PlayerHUD />
    </div>
  );
}