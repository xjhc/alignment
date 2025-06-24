import { useGameContext } from '../contexts/GameContext';
import { PrivateNotifications } from './PrivateNotifications';
import { RosterPanel } from './game/RosterPanel';
import { CommsPanel } from './game/CommsPanel';
import { PlayerHUD } from './game/PlayerHUD';
import styles from './GameScreen.module.css';

export function GameScreen() {
  const { gameState, localPlayer } = useGameContext();

  // All game actions are now handled by the useGameActions hook


  if (!localPlayer) {
    return (
      <div className={styles.gameScreen}>
        <div className={styles.loading}>
          <span>Loading game state...</span>
          <div className="loading-spinner large" style={{ marginLeft: '12px' }}></div>
        </div>
      </div>
    );
  }

  return (
    <div className={`${styles.gameScreen} ${styles.gameLayoutDesktop}`}>
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