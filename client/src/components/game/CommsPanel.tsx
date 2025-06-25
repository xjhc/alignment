import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { usePhaseTimer } from '../../hooks/usePhaseTimer';
import { useGameActions } from '../../hooks/useGameActions';
import { ContextualInputArea } from './ContextualInputArea';
import { SitrepMessage } from './SitrepMessage';
import { VoteResultMessage } from './VoteResultMessage';
import { PulseCheckMessage } from './PulseCheckMessage';

export const CommsPanel: React.FC = () => {
  const { gameState, localPlayer } = useGameContext();
  const timeRemaining = usePhaseTimer(gameState.phase);
  const { startReply } = useGameActions();

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
  }, [gameState.chatMessages]);

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

  return (
    <section className="bg-background-primary flex flex-col">
      <header className="px-4 py-3 border-b border-border flex justify-between items-center flex-shrink-0">
        <div className="flex flex-col gap-0.5">
          <span className="font-mono font-bold text-text-primary text-sm">#war-room</span>
          <span className="text-xs text-text-secondary">Emergency ops • All comms logged</span>
        </div>
        <div className="flex flex-col items-end gap-1">
          <div className={`px-2 py-1 rounded text-xs font-bold uppercase tracking-wider ${getPhaseClass(gameState.phase.type)}`}>{phaseName}</div>
          <div className="flex items-center gap-1">
            <div className="text-xs text-text-muted uppercase">ENDS IN</div>
            <div className="font-mono font-bold text-text-primary text-sm animate-pulse">{timeRemaining}</div>
          </div>
        </div>
      </header>

      <div className="flex-1 p-4 overflow-y-auto flex flex-col gap-2 min-h-0" ref={chatLogRef}>
        {(!gameState.chatMessages || gameState.chatMessages.length === 0) ? (
          <div className="empty-chat-message">
            <span className="text-text-muted italic">
              No messages yet. Waiting for system initialization...
            </span>
          </div>
        ) : null}
        {gameState.chatMessages.map((msg, index) => {
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
            
            // Default chat message rendering
            const getMessageAvatar = (message: any) => {
              if (message.isSystem) return '🤖';
              const player = gameState.players.find((p: any) => p.name === message.playerName);
              if (!player?.isAlive) return '👻';
              return player?.avatar || '👤';
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
                    <span className={`font-semibold text-text-primary text-sm ${
                      msg.isSystem ? 'text-blue-500 font-bold' : ''
                    }`}>
                      {msg.playerName}
                    </span>
                    {!msg.isSystem && gameState.phase.type === 'DISCUSSION' && (
                      <button
                        onClick={() => startReply(msg.id || `${index}`, msg.playerName, msg.message)}
                        className="opacity-0 group-hover:opacity-100 transition-opacity text-xs text-text-muted hover:text-text-primary px-2 py-1 rounded hover:bg-background-tertiary"
                        title="Reply to this message"
                      >
                        ↩️ Reply
                      </button>
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
                                {parsed.quotedMessage}
                              </div>
                            </div>
                          )}
                          {parsed.replyContent && (
                            <div>{parsed.replyContent}</div>
                          )}
                        </>
                      );
                    })()}
                  </div>
                </div>
              </div>
            );
          })
        }
      </div>

      <ContextualInputArea />
    </section>
  );
};