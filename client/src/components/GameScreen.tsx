import { useState, useEffect } from 'react';
import { useGameEngine } from '../hooks/useGameEngine';
import { useWebSocket } from '../hooks/useWebSocket';
import { GameState, Player, Phase, ClientAction } from '../types';
import { PrivateNotifications } from './PrivateNotifications';

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

  const handleKeyDown = (e: React.KeyboardEvent) => {
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
              <p>Choose your night action. All actions are executed simultaneously at the end of the phase.</p>

              <div className="night-actions-grid">
                <div className="action-category">
                  <h3>Universal Actions</h3>
                  <div className="action-card" onClick={handleMineTokens}>
                    <div className="action-icon">⛏️</div>
                    <div className="action-details">
                      <div className="action-name">Mine Tokens</div>
                      <div className="action-description">Earn digital currency for abilities</div>
                      <div className="action-cost">Free</div>
                    </div>
                  </div>
                </div>

                {currentPlayer.role?.ability && (
                  <div className="action-category">
                    <h3>Role Ability</h3>
                    <div
                      className={`action-card ability ${canPlayerAffordAbility(playerId) ? '' : 'disabled'}`}
                      onClick={canPlayerAffordAbility(playerId) ? handleUseAbility : undefined}
                    >
                      <div className="action-icon">✨</div>
                      <div className="action-details">
                        <div className="action-name">{currentPlayer.role.ability.name}</div>
                        <div className="action-description">{currentPlayer.role.ability.description}</div>
                        <div className="action-cost">
                          {canPlayerAffordAbility(playerId) ? 'Available' : 'Insufficient tokens'}
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {currentPlayer.alignment === 'AI' && (
                  <div className="action-category ai-actions">
                    <h3>AI Operations</h3>
                    <div className="conversion-section">
                      <h4>Target for Conversion:</h4>
                      <div className="conversion-targets">
                        {gameState.players
                          .filter(p => p.isAlive && p.id !== playerId && p.alignment !== 'AI')
                          .map(player => (
                            <div
                              key={player.id}
                              className={`target-card ${conversionTarget === player.id ? 'selected' : ''}`}
                              onClick={() => setConversionTarget(player.id)}
                            >
                              <div className="player-avatar">👤</div>
                              <div className="target-details">
                                <div className="target-name">{player.name}</div>
                                <div className="target-job">{player.jobTitle}</div>
                                <div className="target-status">
                                  {player.alignment === 'HUMAN' ? '🧑‍💼 Human' : '❓ Unknown'}
                                </div>
                              </div>
                            </div>
                          ))}
                      </div>
                      {conversionTarget && (
                        <button
                          className="action-btn conversion"
                          onClick={handleConversionAttempt}
                          disabled={!isConnected}
                        >
                          🤖 Attempt Conversion
                        </button>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div className="night-phase-info">
                <div className="info-card">
                  <h4>⏰ Phase Timer</h4>
                  <p>Time remaining: {formatTimeRemaining(gameState.phase)}</p>
                </div>
                <div className="info-card">
                  <h4>📊 Your Status</h4>
                  <p>Tokens: 🪙 {currentPlayer.tokens}</p>
                  <p>Role: {currentPlayer.role?.name || 'Unknown'}</p>
                  <p>Alignment: {currentPlayer.alignment === 'AI' ? '🤖 AI' : '🧑‍💼 Human'}</p>
                </div>
              </div>
            </div>
          )}

          {gameState.phase.type === 'SITREP' && (
            <div className="sitrep-phase">
              <h2>{getPhaseDisplayName(gameState.phase.type)}</h2>
              <p>The day is starting. Review the situation and prepare for discussion.</p>

              {/* Night Action Results Summary */}
              {gameState.nightActionResults && gameState.nightActionResults.length > 0 && (
                <div className="night-results-summary">
                  <h3>🌙 Overnight Activity Report</h3>
                  <div className="night-results-grid">
                    {gameState.nightActionResults
                      .filter(result => result.isPublic)
                      .map(result => (
                        <div key={result.id} className={`night-result-card ${result.result}`}>
                          <div className="result-icon">
                            {result.type === 'MINE_TOKENS' ? '⛏️' :
                              result.type === 'BLOCK' ? '🚫' :
                                result.type === 'CONVERT' ? '🤖' :
                                  result.type === 'INVESTIGATE' ? '🔍' :
                                    result.type === 'PROTECT' ? '🛡️' : '❓'}
                          </div>
                          <div className="result-details">
                            <div className="result-action">{result.type.replace('_', ' ')}</div>
                            <div className="result-description">{result.description}</div>
                            <div className={`result-status ${result.result}`}>
                              {result.result === 'success' ? '✅ Success' :
                                result.result === 'failed' ? '❌ Failed' :
                                  result.result === 'blocked' ? '🚫 Blocked' : result.result}
                            </div>
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              )}

              {/* Crisis Event Display */}
              {gameState.crisisEvent && (
                <div className="crisis-alert">
                  <h3>🚨 Crisis Event: {gameState.crisisEvent.title}</h3>
                  <p className="crisis-description">{gameState.crisisEvent.description}</p>
                  {gameState.crisisEvent.effects && Object.keys(gameState.crisisEvent.effects).length > 0 && (
                    <div className="crisis-effects">
                      <h4>Active Effects:</h4>
                      <ul>
                        {Object.entries(gameState.crisisEvent.effects).map(([key, value]) => (
                          <li key={key}>
                            <strong>{key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:</strong>
                            {typeof value === 'boolean' ? (value ? ' Active' : ' Inactive') :
                              typeof value === 'number' ? ` ${value}` :
                                ` ${value}`}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}

              {/* Corporate Mandate Display */}
              {gameState.corporateMandate && gameState.corporateMandate.isActive && (
                <div className="mandate-alert">
                  <h3>📋 Corporate Mandate: {gameState.corporateMandate.name}</h3>
                  <p className="mandate-description">{gameState.corporateMandate.description}</p>
                  {gameState.corporateMandate.effects && Object.keys(gameState.corporateMandate.effects).length > 0 && (
                    <div className="mandate-effects">
                      <h4>Policy Changes:</h4>
                      <ul>
                        {Object.entries(gameState.corporateMandate.effects).map(([key, value]) => (
                          <li key={key}>
                            <strong>{key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:</strong>
                            {typeof value === 'boolean' ? (value ? ' Enabled' : ' Disabled') :
                              typeof value === 'number' ? ` ${value}` :
                                ` ${value}`}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
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
              <p>Select a team member to nominate for elimination.</p>

              {gameState.voteState && (
                <div className="voting-progress">
                  <h3>Current Nominations:</h3>
                  <div className="nomination-results">
                    {Object.entries(gameState.voteState.results || {}).map(([playerId, count]) => {
                      const nominee = gameState.players.find(p => p.id === playerId);
                      return nominee ? (
                        <div key={playerId} className="nomination-result">
                          <span className="nominee-name">{nominee.name}</span>
                          <span className="vote-count">{count} nomination{count !== 1 ? 's' : ''}</span>
                        </div>
                      ) : null;
                    })}
                  </div>
                </div>
              )}

              <div className="nomination-form">
                <div className="player-grid">
                  {gameState.players
                    .filter(p => p.isAlive && p.id !== playerId)
                    .map(player => (
                      <div
                        key={player.id}
                        className={`nominee-card ${selectedNominee === player.id ? 'selected' : ''}`}
                        onClick={() => setSelectedNominee(player.id)}
                      >
                        <div className="player-avatar">👤</div>
                        <div className="player-details">
                          <div className="player-name">{player.name}</div>
                          <div className="player-job">{player.jobTitle}</div>
                          <div className="player-tokens">🪙 {player.tokens}</div>
                        </div>
                      </div>
                    ))}
                </div>
                <button
                  className="action-btn primary"
                  onClick={handleNominate}
                  disabled={!selectedNominee || !isConnected}
                >
                  Nominate {selectedNominee ? gameState.players.find(p => p.id === selectedNominee)?.name : 'Player'}
                </button>
              </div>
            </div>
          )}

          {(gameState.phase.type === 'TRIAL' || gameState.phase.type === 'VERDICT') && (
            <div className="voting-phase">
              <h2>{getPhaseDisplayName(gameState.phase.type)}</h2>

              {gameState.nominatedPlayer && (
                <div className="nominated-player">
                  <h3>On Trial:</h3>
                  {(() => {
                    const nominee = gameState.players.find(p => p.id === gameState.nominatedPlayer);
                    return nominee ? (
                      <div className="nominee-info">
                        <div className="player-avatar large">👤</div>
                        <div className="nominee-details">
                          <div className="nominee-name">{nominee.name}</div>
                          <div className="nominee-job">{nominee.jobTitle}</div>
                          <div className="nominee-tokens">🪙 {nominee.tokens}</div>
                        </div>
                      </div>
                    ) : (
                      <div className="nominee-info">Unknown player</div>
                    );
                  })()}
                </div>
              )}

              <p>Cast your vote to determine their fate.</p>

              {gameState.voteState && (
                <div className="voting-progress">
                  <h3>Current Votes:</h3>
                  <div className="vote-tally">
                    <div className="vote-option">
                      <span className="vote-label guilty">Guilty</span>
                      <span className="vote-count">{gameState.voteState.results?.GUILTY || 0}</span>
                    </div>
                    <div className="vote-option">
                      <span className="vote-label innocent">Innocent</span>
                      <span className="vote-count">{gameState.voteState.results?.INNOCENT || 0}</span>
                    </div>
                  </div>
                </div>
              )}

              <div className="voting-form">
                <div className="vote-buttons">
                  <button
                    className={`vote-btn guilty ${selectedVote === 'GUILTY' ? 'selected' : ''}`}
                    onClick={() => setSelectedVote('GUILTY')}
                  >
                    <span className="vote-icon">⚡</span>
                    <span className="vote-text">GUILTY</span>
                    <span className="vote-sub">Eliminate</span>
                  </button>
                  <button
                    className={`vote-btn innocent ${selectedVote === 'INNOCENT' ? 'selected' : ''}`}
                    onClick={() => setSelectedVote('INNOCENT')}
                  >
                    <span className="vote-icon">🛡️</span>
                    <span className="vote-text">INNOCENT</span>
                    <span className="vote-sub">Keep alive</span>
                  </button>
                </div>
                <button
                  className="action-btn primary"
                  onClick={handleVote}
                  disabled={!selectedVote || !isConnected}
                >
                  Cast Vote: {selectedVote || 'Make Selection'}
                </button>
              </div>
            </div>
          )}
        </section>

        {/* Right Sidebar - Chat */}
        <aside className="chat-panel">
          <h3>Communications</h3>
          <div className="chat-messages">
            {isChatHistoryLoading && gameState.chatMessages.length === 0 && (
              <div className="chat-loading">
                <div className="loading-spinner">⏳</div>
                <div className="loading-text">Loading chat history...</div>
              </div>
            )}
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
              onKeyDown={handleKeyDown}
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