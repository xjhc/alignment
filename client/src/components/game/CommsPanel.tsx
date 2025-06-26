import React, { useState } from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { usePhaseTimer } from '../../hooks/usePhaseTimer';
import { useGameActions } from '../../hooks/useGameActions';
import { ContextualInputArea } from './ContextualInputArea';
import { SitrepMessage } from './SitrepMessage';
import { VoteResultMessage } from './VoteResultMessage';
import { PulseCheckMessage } from './PulseCheckMessage';
import { PulseCheckSubmissionMessage } from './PulseCheckSubmissionMessage';
import { IncitingIncidentMessage } from './IncitingIncidentMessage';
import { LoebmateMessage } from './LoebmateMessage';
import { EmojiPicker } from './EmojiPicker';
import { MarkdownRenderer } from './MarkdownRenderer';
import { EmojiReaction } from '../../types';

export const CommsPanel: React.FC = () => {
  const { gameState, localPlayer, activeChannel } = useGameContext();
  const timeRemaining = usePhaseTimer(gameState.phase);
  const { startReply, handleSkipPhase, handleEmojiReaction } = useGameActions();
  
  const [emojiPickerState, setEmojiPickerState] = useState<{
    isOpen: boolean;
    messageId: string | null;
    anchorElement: HTMLElement | null;
  }>({
    isOpen: false,
    messageId: null,
    anchorElement: null,
  });

  if (!localPlayer) {
    return <div>Loading...</div>;
  }
  
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

  const phaseName = getPhaseDisplayName(gameState.phase.type);

  // Skip vote calculations
  const skipVotes = gameState.skipVotes || {};
  const skipVoteCount = Object.keys(skipVotes).length;
  const livingHumans = gameState.players.filter(p => p.isAlive && p.controlType === 'HUMAN').length;
  const hasLocalPlayerVoted = skipVotes[localPlayer.id] || false;
  
  // Determine if skip button should be shown
  const canShowSkipButton = gameState.phase.type !== 'TRIAL' && 
                           gameState.phase.type !== 'GAME_OVER' && 
                           gameState.phase.type !== 'LOBBY' &&
                           localPlayer.isAlive;

  // Filter messages by active channel
  const filteredMessages = gameState.chatMessages.filter(msg => {
    // If message has no channelID, assume it's for #war-room (legacy support)
    const messageChannel = msg.channelID || '#war-room';
    return messageChannel === activeChannel;
  });

  const getChannelDescription = (channelId: string) => {
    switch (channelId) {
      case '#war-room':
        return 'Emergency ops • All comms logged';
      case '#aligned':
        return 'Private AI coordination • Encrypted';
      case '#off-boarding':
        return 'Spectator discussion • Post-elimination';
      default:
        return 'Channel communication';
    }
  };

  const getPhaseClass = (phaseType: string) => {
    switch (phaseType) {
      case 'DISCUSSION':
        return 'bg-green-500 text-white';
      case 'NOMINATION':
        return 'bg-yellow-500 text-white';
      case 'TRIAL':
        return 'bg-red-500 text-white';
      case 'VERDICT':
        return 'bg-red-500 text-white';
      case 'NIGHT':
        return 'bg-cyan-500 text-white';
      case 'PULSE_CHECK':
        return 'bg-cyan-600 text-white';
      default:
        return 'bg-blue-500 text-white';
    }
  };

  const chatLogRef = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    if (chatLogRef.current) {
      chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
    }
  }, [filteredMessages]);

  const parseMessageContent = (message: string) => {
    const quoteRegex = /[\[]quote=([^\]]+)[\]]([^[]*)[\[][\/]quote[\]]\n?(.*)/s;
    const match = message.match(quoteRegex);
    
    if (match) {
      const [, quotedPlayerName, quotedMessage, replyContent] = match;
      return {
        hasQuote: true,
        quotedPlayerName,
        quotedMessage: quotedMessage.trim(),
        replyContent: replyContent.trim(),
      };
    }
    
    return {
      hasQuote: false,
      replyContent: message,
    };
  };

  const openEmojiPicker = (messageId: string, anchorElement: HTMLElement) => {
    setEmojiPickerState({
      isOpen: true,
      messageId,
      anchorElement,
    });
  };

  const closeEmojiPicker = () => {
    setEmojiPickerState({
      isOpen: false,
      messageId: null,
      anchorElement: null,
    });
  };

  const handleEmojiSelect = (emoji: string) => {
    if (emojiPickerState.messageId) {
      handleEmojiReaction(emojiPickerState.messageId, emoji, activeChannel);
    }
    closeEmojiPicker();
  };

  const aggregateReactions = (reactions: EmojiReaction[] = []) => {
    const aggregated: Record<string, { count: number; players: string[] }> = {};
    
    reactions.forEach(reaction => {
      if (!aggregated[reaction.emoji]) {
        aggregated[reaction.emoji] = { count: 0, players: [] };
      }
      aggregated[reaction.emoji].count++;
      aggregated[reaction.emoji].players.push(reaction.playerName);
    });
    
    return Object.entries(aggregated).map(([emoji, data]) => ({
      emoji,
      count: data.count,
      players: data.players,
    }));
  };

  return (
    <section className="bg-background-primary flex flex-col min-h-0 overflow-y-auto">
      <header className="px-4 py-3 border-b border-border flex justify-between items-center flex-shrink-0">
        <div className="flex flex-col gap-0.5">
          <span className="font-mono font-bold text-text-primary text-sm">{activeChannel}</span>
          <span className="text-xs text-text-secondary">{getChannelDescription(activeChannel)}</span>
        </div>
        <div className="flex flex-col items-end gap-1">
          <div className={`px-2 py-1 rounded text-xs font-bold uppercase tracking-wider ${getPhaseClass(gameState.phase.type)}`}>{phaseName}</div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1">
              <div className="text-xs text-text-muted uppercase">ENDS IN</div>
              <div className="font-mono font-bold text-text-primary text-sm animate-pulse">{timeRemaining}</div>
            </div>
            {canShowSkipButton && (
              <button
                onClick={handleSkipPhase}
                disabled={hasLocalPlayerVoted}
                className={`px-2 py-1 rounded text-xs font-bold transition-all ${
                  hasLocalPlayerVoted 
                    ? 'bg-amber-500 text-white cursor-not-allowed opacity-75' 
                    : 'bg-background-tertiary text-text-primary hover:bg-background-secondary border border-border hover:border-border-secondary'
                }`}
                title={hasLocalPlayerVoted 
                  ? `You voted to skip (${skipVoteCount}/${livingHumans} ready)` 
                  : `Vote to skip this phase (${skipVoteCount}/${livingHumans} ready)`
                }
              >
                »
              </button>
            )}
          </div>
        </div>
      </header>

      <div className="flex-1 p-4 overflow-y-auto flex flex-col gap-2 min-h-0" ref={chatLogRef}>
        {(!filteredMessages || filteredMessages.length === 0) ? (
          <div className="empty-chat-message">
            <span className="text-text-muted italic">
              No messages in {activeChannel} yet...
            </span>
          </div>
        ) : null}
        {filteredMessages.map((msg, index) => {
            // Render specialized system messages
            if (msg.isSystem && msg.type === 'SITREP') {
              return (
                <div key={msg.id || index} className="stagger-child" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <SitrepMessage 
                    message={msg} 
                    gameState={gameState} 
                  />
                </div>
              );
            }
            
            if (msg.isSystem && msg.type === 'VOTE_RESULT') {
              return (
                <div key={msg.id || index} className="stagger-child" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <VoteResultMessage 
                    message={msg} 
                    gameState={gameState} 
                    localPlayerId={localPlayer.id}
                  />
                </div>
              );
            }
            
            if (msg.isSystem && msg.type === 'PULSE_CHECK') {
              return (
                <div key={msg.id || index} className="stagger-child" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <PulseCheckMessage 
                    message={msg} 
                    gameState={gameState} 
                  />
                </div>
              );
            }
            
            if (msg.isSystem && msg.type === 'INCITING_INCIDENT') {
              return (
                <div key={msg.id || index} className="stagger-child" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <IncitingIncidentMessage 
                    message={msg} 
                  />
                </div>
              );
            }
            
            if (msg.isSystem && msg.type === 'LOEBMATE_MESSAGE') {
              return (
                <div key={msg.id || index} className="stagger-child" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <LoebmateMessage 
                    message={msg} 
                  />
                </div>
              );
            }
            
            // Handle pulse check submission messages (identified by being system messages ending with "_chat")
            if (msg.isSystem && msg.id?.endsWith('_chat') && msg.playerName && msg.message) {
              return (
                <div key={msg.id || index} className="stagger-child" style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}>
                  <PulseCheckSubmissionMessage 
                    message={msg} 
                  />
                </div>
              );
            }
            
            // Default chat message rendering
            const getMessageAvatar = (message: any) => {
              if (message.isSystem) return '🤖';
              // Look up player by playerID first, then fallback to playerName
              const player = gameState.players.find((p: any) => 
                p.id === message.playerID || p.name === message.playerName
              );
              if (!player) return '👤'; // Default avatar if player not found
              if (!player.isAlive) return '👻'; // Ghost avatar for dead players
              return player.avatar || '👤'; // Player's avatar or default
            };

            const formatTimestamp = (timestamp: string) => {
              const date = new Date(timestamp);
              return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            };

            return (
              <div 
                key={msg.id || index} 
                className={`group flex items-start gap-2.5 px-2 py-1.5 rounded-md transition-all duration-150 mb-0.5 hover:bg-background-secondary hover:translate-x-0.5 ${
                  msg.isSystem ? 'border-l-2 border-blue-500 bg-blue-500/5 pl-3' : ''
                } animation-slide-in-left`}
                style={{ animationDelay: `${Math.min(index * 50, 500)}ms` }}
              >
                <div className={`w-6 h-6 rounded-full bg-background-tertiary flex items-center justify-center text-sm flex-shrink-0 border border-border shadow-sm ${
                  msg.isSystem ? 'bg-blue-500 text-white border-blue-500 shadow-blue-500/30' : ''
                }`}>
                  {getMessageAvatar(msg)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-0.5">
                    <div className="flex items-center gap-2">
                      <span className={`font-semibold text-text-primary text-sm ${
                        msg.isSystem ? 'text-blue-500 font-bold' : ''
                      }`}>
                        {msg.playerName}
                      </span>
                      {msg.timestamp && (
                        <span className="text-text-muted text-xs">
                          {formatTimestamp(msg.timestamp)}
                        </span>
                      )}
                    </div>
                    {!msg.isSystem && gameState.phase.type === 'DISCUSSION' && (
                      <div className="opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
                        <button
                          onClick={() => startReply(msg.id || `${index}`, msg.playerName, msg.message)}
                          className="text-xs text-text-muted hover:text-text-primary px-2 py-1 rounded hover:bg-background-tertiary"
                          title="Reply to this message"
                        >
                          ↩️ Reply
                        </button>
                        <button
                          onClick={(e) => openEmojiPicker(msg.id || `${index}`, e.currentTarget)}
                          className="text-xs text-text-muted hover:text-text-primary px-2 py-1 rounded hover:bg-background-tertiary"
                          title="React with emoji"
                        >
                          😊 React
                        </button>
                      </div>
                    )}
                  </div>
                  <div className="text-text-secondary text-sm leading-relaxed break-words mt-0.5">
                    {(() => {
                      const parsed = parseMessageContent(msg.message);
                      return (
                        <>
                          {parsed.hasQuote && (
                            <div className="bg-background-secondary border-l-2 border-border pl-3 py-2 mb-2 rounded-r">
                              <div className="text-xs text-text-muted font-semibold mb-1">
                                {parsed.quotedPlayerName}:
                              </div>
                              <div className="text-text-secondary text-xs italic">
                                <MarkdownRenderer content={parsed.quotedMessage || ''} />
                              </div>
                            </div>
                          )}
                          {parsed.replyContent && (
                            <MarkdownRenderer content={parsed.replyContent} />
                          )}
                        </>
                      );
                    })()}
                  </div>
                  {/* Reactions display */}
                  {msg.reactions && msg.reactions.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {aggregateReactions(msg.reactions).map(({ emoji, count, players }) => (
                        <button
                          key={emoji}
                          className="inline-flex items-center gap-1 px-2 py-1 rounded-full bg-background-secondary hover:bg-background-tertiary text-xs border border-border transition-colors"
                          title={`${players.join(', ')} reacted with ${emoji}`}
                          onClick={(e) => openEmojiPicker(msg.id || `${index}`, e.currentTarget)}
                        >
                          <span>{emoji}</span>
                          <span className="font-medium">{count}</span>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        }
      </div>

      <ContextualInputArea />
      
      <EmojiPicker
        isOpen={emojiPickerState.isOpen}
        onClose={closeEmojiPicker}
        onEmojiSelect={handleEmojiSelect}
        anchorElement={emojiPickerState.anchorElement}
      />
    </section>
  );
};