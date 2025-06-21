import { useState, useEffect } from 'react';
import { useGameEngine } from '../hooks/useGameEngine';
import { useWebSocket, useGameEvents } from '../hooks/useWebSocket';
import { GameState, Player, Phase, ClientAction } from '../types';

interface GameScreenProps {
  gameState: GameState;
  playerId: string;
}

export function GameScreen({ gameState, playerId }: GameScreenProps) {
  const [chatInput, setChatInput] = useState('');
  const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null);
  const [selectedNominee, setSelectedNominee] = useState<string>('');
  const [selectedVote, setSelectedVote] = useState<'GUILTY' | 'INNOCENT' | ''>('');
  const [conversionTarget, setConversionTarget] = useState<string>('');
  
  const { 
    canPlayerAffordAbility,
    isValidNightActionTarget
  } = useGameEngine();
  
  const { sendAction, isConnected } = useWebSocket();
  
  // Subscribe to all game events automatically
  useGameEvents();

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
    const phaseEnd = phaseStart + phase.duration;
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
    if (!currentPlayer || !isConnected) return;

    try {
      const action: ClientAction = {
        type: 'SUBMIT_NIGHT_ACTION',
        payload: {
          game_id: gameState.id,
          player_id: playerId,
          action_type: 'MINE_TOKENS',
          difficulty: 0.3,
        },
      };

      sendAction(action);
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
          vote_type: 'NOMINATION',
          nominee_id: selectedNominee,
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
          vote_type: 'TRIAL',
          vote: selectedVote,
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

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSendMessage();
    }
  };


  if (!currentPlayer) {
    return (
      <div className="game-screen">
        <div className="loading">Loading game state...</div>
      </div>
    );
  }

  return (
    <div className="game-screen">
      {/* Header */}
      <header className="game-header">
        <div className="game-title">
          <h1>LOEBIAN INC. // WAR ROOM</h1>
          <div className="game-id">Game: {gameState.id.substring(0, 6)}</div>
        </div>
        
        <div className="phase-info">
          <div className="phase-name">{getPhaseDisplayName(gameState.phase.type)}</div>
          <div className="day-counter">Day {gameState.dayNumber}</div>
          <div className="time-remaining">{formatTimeRemaining(gameState.phase)}</div>
        </div>
        
        <div className="player-status">
          <div className="player-name">{currentPlayer.name}</div>
          <div className="player-role">{currentPlayer.role?.name || 'Unknown'}</div>
          <div className="player-tokens">🪙 {currentPlayer.tokens}</div>
          <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
          </div>
        </div>
      </header>

      {/* Main Game Area */}
      <main className="game-main">
        {/* Left Sidebar - Players */}
        <aside className="player-list">
          <h3>Personnel ({gameState.players.filter(p => p.isAlive).length} alive)</h3>
          {gameState.players.map(player => (
            <div 
              key={player.id} 
              className={`player-card ${!player.isAlive ? 'eliminated' : ''} ${player.id === playerId ? 'self' : ''}`}
            >
              <div className="player-avatar">👤</div>
              <div className="player-info">
                <div className="player-name">{player.name}</div>
                <div className="player-job">{player.jobTitle}</div>
                <div className="player-tokens">🪙 {player.tokens}</div>
              </div>
              {player.statusMessage && (
                <div className="player-status-msg">{player.statusMessage}</div>
              )}
            </div>
          ))}
        </aside>

        {/* Center - Phase Content */}
        <section className="phase-content">
          {gameState.phase.type === 'DISCUSSION' && (
            <div className="discussion-phase">
              <h2>Discussion Phase</h2>
              <p>Share information and discuss who might be the AI.</p>
              
              {gameState.crisisEvent && (
                <div className="crisis-event">
                  <h3>🚨 {gameState.crisisEvent.title}</h3>
                  <p>{gameState.crisisEvent.description}</p>
                </div>
              )}
            </div>
          )}

          {gameState.phase.type === 'NIGHT' && (
            <div className="night-phase">
              <h2>Night Phase</h2>
              <p>Submit your night action.</p>
              
              <div className="night-actions">
                <button 
                  className="action-btn" 
                  onClick={handleMineTokens}
                  disabled={!isConnected}
                >
                  Mine Tokens
                </button>
                {currentPlayer.role?.ability && canPlayerAffordAbility(playerId) && (
                  <button 
                    className="action-btn"
                    onClick={handleUseAbility}
                    disabled={!isConnected}
                  >
                    Use {currentPlayer.role.ability.name}
                  </button>
                )}
                {currentPlayer.alignment === 'ALIGNED' && (
                  <div className="conversion-form">
                    <select 
                      value={conversionTarget} 
                      onChange={(e) => setConversionTarget(e.target.value)}
                      className="player-select"
                    >
                      <option value="">Select conversion target...</option>
                      {gameState.players
                        .filter(p => p.isAlive && p.id !== playerId)
                        .map(player => (
                          <option key={player.id} value={player.id}>
                            {player.name} ({player.jobTitle})
                          </option>
                        ))}
                    </select>
                    <button 
                      className="action-btn"
                      onClick={handleConversionAttempt}
                      disabled={!conversionTarget || !isConnected}
                    >
                      Attempt Conversion
                    </button>
                  </div>
                )}
              </div>
            </div>
          )}

          {gameState.phase.type === 'SITREP' && (
            <div className="sitrep-phase">
              <h2>{getPhaseDisplayName(gameState.phase.type)}</h2>
              <p>The day is starting. Review the situation and prepare for discussion.</p>
            </div>
          )}

          {gameState.phase.type === 'PULSE_CHECK' && (
            <div className="pulse-check-phase">
              <h2>{getPhaseDisplayName(gameState.phase.type)}</h2>
              <p>Submit your pulse check response.</p>
              <div className="pulse-check-actions">
                <button 
                  className="action-btn"
                  onClick={() => handlePulseCheck('POSITIVE')}
                  disabled={!isConnected}
                >
                  Positive
                </button>
                <button 
                  className="action-btn"
                  onClick={() => handlePulseCheck('NEGATIVE')}
                  disabled={!isConnected}
                >
                  Negative
                </button>
                <button 
                  className="action-btn"
                  onClick={() => handlePulseCheck('NEUTRAL')}
                  disabled={!isConnected}
                >
                  Neutral
                </button>
              </div>
            </div>
          )}

          {gameState.phase.type === 'NOMINATION' && (
            <div className="nomination-phase">
              <h2>Nomination Phase</h2>
              <p>Select a player to nominate for elimination.</p>
              <div className="nomination-form">
                <select 
                  value={selectedNominee} 
                  onChange={(e) => setSelectedNominee(e.target.value)}
                  className="player-select"
                >
                  <option value="">Select player to nominate...</option>
                  {gameState.players
                    .filter(p => p.isAlive && p.id !== playerId)
                    .map(player => (
                      <option key={player.id} value={player.id}>
                        {player.name} ({player.jobTitle})
                      </option>
                    ))}
                </select>
                <button 
                  className="action-btn"
                  onClick={handleNominate}
                  disabled={!selectedNominee || !isConnected}
                >
                  Nominate
                </button>
              </div>
            </div>
          )}

          {(gameState.phase.type === 'TRIAL' || gameState.phase.type === 'VERDICT') && (
            <div className="voting-phase">
              <h2>{getPhaseDisplayName(gameState.phase.type)}</h2>
              <p>Cast your vote on the nominated player.</p>
              <div className="voting-form">
                <div className="vote-options">
                  <label>
                    <input
                      type="radio"
                      name="vote"
                      value="GUILTY"
                      checked={selectedVote === 'GUILTY'}
                      onChange={(e) => setSelectedVote(e.target.value as 'GUILTY')}
                    />
                    Guilty (Eliminate)
                  </label>
                  <label>
                    <input
                      type="radio"
                      name="vote"
                      value="INNOCENT"
                      checked={selectedVote === 'INNOCENT'}
                      onChange={(e) => setSelectedVote(e.target.value as 'INNOCENT')}
                    />
                    Innocent (Keep alive)
                  </label>
                </div>
                <button 
                  className="action-btn"
                  onClick={handleVote}
                  disabled={!selectedVote || !isConnected}
                >
                  Cast Vote
                </button>
              </div>
            </div>
          )}
        </section>

        {/* Right Sidebar - Chat */}
        <aside className="chat-panel">
          <h3>Communications</h3>
          <div className="chat-messages">
            {gameState.chatMessages.map(message => (
              <div 
                key={message.id} 
                className={`chat-message ${message.isSystem ? 'system' : ''} ${message.playerID === playerId ? 'own' : ''}`}
              >
                <div className="message-header">
                  <span className="sender">{message.playerName}</span>
                  <span className="timestamp">
                    {new Date(message.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div className="message-content">{message.message}</div>
              </div>
            ))}
          </div>
          
          <div className="chat-input">
            <input
              type="text"
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Type message..."
              maxLength={200}
            />
            <button 
              onClick={handleSendMessage}
              disabled={!isConnected || !chatInput.trim()}
            >
              Send
            </button>
          </div>
        </aside>
      </main>
    </div>
  );
}