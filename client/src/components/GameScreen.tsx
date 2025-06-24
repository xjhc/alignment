import { useState, useEffect } from 'react';
import { useGameEngine } from '../hooks/useGameEngine';
import { useWebSocket } from '../hooks/useWebSocket';
import { GameState, Player, Phase, ClientAction } from '../types';
import { GameProvider } from '../contexts/GameContext';
import { PrivateNotifications } from './PrivateNotifications';
import { RosterPanel } from './game/RosterPanel';
import { CommsPanel } from './game/CommsPanel';
import { PlayerHUD } from './game/PlayerHUD';
import styles from './GameScreen.module.css';

interface GameScreenProps {
  gameState: GameState;
  playerId: string;
  isChatHistoryLoading?: boolean;
}

export function GameScreen({ gameState, playerId, isChatHistoryLoading = false }: GameScreenProps) {
  const [chatInput, setChatInput] = useState('');
  const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null);
  const [selectedNominee, setSelectedNominee] = useState<string>('');
  const [selectedVote, setSelectedVote] = useState<'GUILTY' | 'INNOCENT' | ''>('');
  const [conversionTarget, setConversionTarget] = useState<string>('');
  const [miningTarget, setMiningTarget] = useState<string>('');

  const {
    canPlayerAffordAbility,
    isValidNightActionTarget
  } = useGameEngine();

  const { sendAction, isConnected } = useWebSocket();

  useEffect(() => {
    const player = gameState.players.find(p => p.id === playerId);
    setCurrentPlayer(player || null);
  }, [gameState.players, playerId]);

  // WebSocket events are now handled automatically by gameEngine via websocket.ts
  // The useEffect in App.tsx syncs gameEngine state to React state
  // Individual event handlers here are no longer needed

  const getPhaseDisplayName = (phaseType: string) => {
    switch (phaseType) {
      case 'SITREP': return 'SITREP';
      case 'PULSE_CHECK': return 'PULSE CHECK';
      case 'DISCUSSION': return 'DISCUSSION';
      case 'NOMINATION': return 'NOMINATION';
      case 'TRIAL': return 'TRIAL';
      case 'VERDICT': return 'VERDICT';
      case 'NIGHT': return 'NIGHT PHASE';
      case 'GAME_OVER': return 'GAME OVER';
      default: return phaseType;
    }
  };

  const formatTimeRemaining = (phase: Phase) => {
    const now = new Date().getTime();
    const phaseStart = new Date(phase.startTime).getTime();
    const phaseEnd = phaseStart + (phase.duration * 1000);
    const remaining = Math.max(0, phaseEnd - now);

    const minutes = Math.floor(remaining / 60000);
    const seconds = Math.floor((remaining % 60000) / 1000);

    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  const handleSendMessage = async () => {
    if (!chatInput.trim() || !currentPlayer || !isConnected) {
      return;
    }

    try {
      const action: ClientAction = {
        type: 'POST_CHAT_MESSAGE',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          message: chatInput.trim(),
          player_name: currentPlayer.name,
        },
      };

      sendAction(action);
      setChatInput('');
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  };

  const handleMineTokens = async () => {
    if (!currentPlayer || !miningTarget || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          action_type: 'MINE_TOKENS',
          target_player_id: miningTarget,
        },
      };

      sendAction(action);
      setMiningTarget('');
    } catch (error) {
      console.error('Failed to mine tokens:', error);
    }
  };

  const handleUseAbility = async () => {
    if (!currentPlayer || !canPlayerAffordAbility(playerId) || !isConnected) {
      return;
    }

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          action_type: 'USE_ABILITY',
          ability_type: currentPlayer.role?.type || 'UNKNOWN',
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to use ability:', error);
    }
  };

  const handleConversionAttempt = async () => {
    if (!currentPlayer || !conversionTarget || !isConnected) return;

    // Validate that this is a valid target for conversion
    if (!isValidNightActionTarget(playerId, conversionTarget, 'ATTEMPT_CONVERSION')) {
      console.warn('Invalid conversion target');
      return;
    }

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          action_type: 'ATTEMPT_CONVERSION',
          target_player_id: conversionTarget,
        },
      };

      sendAction(action);
      setConversionTarget('');
    } catch (error) {
      console.error('Failed to attempt conversion:', error);
    }
  };

  const handleNominate = async () => {
    if (!selectedNominee || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_VOTE',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          target_id: selectedNominee,
          vote_type: 'NOMINATION',
        },
      };

      sendAction(action);
      setSelectedNominee('');
    } catch (error) {
      console.error('Failed to nominate player:', error);
    }
  };

  const handleVote = async () => {
    if (!selectedVote || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_VOTE',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          target_id: selectedVote, // For GUILTY/INNOCENT votes, the target is the vote itself
          vote_type: 'VERDICT',
        },
      };

      sendAction(action);
      setSelectedVote('');
    } catch (error) {
      console.error('Failed to cast vote:', error);
    }
  };

  const handlePulseCheck = async (response: string) => {
    if (!isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_PULSE_CHECK',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          response,
        },
      };

      sendAction(action);
    } catch (error) {
      console.error('Failed to submit pulse check:', error);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSendMessage();
    }
  };


  const localPlayer = gameState.players.find(p => p.id === playerId);

  if (!localPlayer) {
    return (
      <div className={styles.gameScreen}>
        <div className={styles.loading}>Loading game state...</div>
      </div>
    );
  }

  return (
    <GameProvider 
      gameState={gameState} 
      localPlayerId={playerId} 
      isConnected={isConnected}
    >
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
        
        <CommsPanel 
          chatInput={chatInput}
          setChatInput={setChatInput}
          handleSendMessage={handleSendMessage}
          handleKeyDown={handleKeyDown}
          selectedNominee={selectedNominee}
          setSelectedNominee={setSelectedNominee}
          selectedVote={selectedVote}
          setSelectedVote={setSelectedVote}
          conversionTarget={conversionTarget}
          setConversionTarget={setConversionTarget}
          miningTarget={miningTarget}
          setMiningTarget={setMiningTarget}
          handleNominate={handleNominate}
          handleVote={handleVote}
          handlePulseCheck={handlePulseCheck}
          handleMineTokens={handleMineTokens}
          handleUseAbility={handleUseAbility}
          handleConversionAttempt={handleConversionAttempt}
          getPhaseDisplayName={getPhaseDisplayName}
          formatTimeRemaining={formatTimeRemaining}
          canPlayerAffordAbility={canPlayerAffordAbility}
          isValidNightActionTarget={isValidNightActionTarget}
          isChatHistoryLoading={isChatHistoryLoading}
        />
        
        <PlayerHUD />
      </div>
    </GameProvider>
  );
}