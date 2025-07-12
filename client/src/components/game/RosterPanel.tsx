import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useGameContext } from '../../contexts/GameContext';
import { useSessionContext } from '../../contexts/SessionContext';
import { useTheme } from '../../hooks/useTheme';
import { PlayerCard } from './PlayerCard';
import { soundManager } from '../../services/soundManager';
import { CHANNEL_UNLOCK, applyStaggeredAnimation } from '../../utils/animations';

export const RosterPanel: React.FC = () => {
  const { gameState, localPlayerId, localPlayer, viewedPlayerId, setViewedPlayer, activeChannel, setActiveChannel } = useGameContext();
  const { appState } = useSessionContext();
  const { theme, toggleTheme } = useTheme();
  const isSpectating = appState.isSpectating;
  const players = gameState?.players || [];
  const [isMuted, setIsMuted] = useState(soundManager.isMutedState());
  const [alignedChannelJustUnlocked, setAlignedChannelJustUnlocked] = useState(false);

  // Update local state when sound manager mute state changes
  useEffect(() => {
    setIsMuted(soundManager.isMutedState());
  }, []);

  // Track when aligned channel becomes available for animation
  useEffect(() => {
    const wasAI = getChannelAccess('#aligned');
    
    // If aligned channel just became available, trigger unlock animation
    if (wasAI && !alignedChannelJustUnlocked) {
      setAlignedChannelJustUnlocked(true);
      
      // Remove the animation state after the animation completes
      setTimeout(() => {
        setAlignedChannelJustUnlocked(false);
      }, 1000);
    }
  }, [localPlayer?.alignment]);

  const handleToggleMute = () => {
    const newMutedState = soundManager.toggleMute();
    setIsMuted(newMutedState);
  };

  const getPlayerCounts = () => {
    if (!Array.isArray(players)) return { humanCount: 0, alignedCount: 0, deactivatedCount: 0 };
    
    return players.reduce((counts, player) => {
      if (!player.isAlive) {
        counts.deactivatedCount++;
      } else if (player.alignment === 'AI' || player.alignment === 'ALIGNED') {
        counts.alignedCount++;
      } else {
        counts.humanCount++;
      }
      return counts;
    }, { humanCount: 0, alignedCount: 0, deactivatedCount: 0 });
  };

  const { humanCount, alignedCount, deactivatedCount } = getPlayerCounts();
  const isAI = localPlayer?.alignment === 'AI' || localPlayer?.alignment === 'ALIGNED';

  const getChannelAccess = (channelId: string) => {
    switch (channelId) {
      case '#war-room':
        return localPlayer?.isAlive || isSpectating || false;
      case '#aligned':
        return isAI;
      case '#off-boarding':
        return !localPlayer?.isAlive || false;
      case '#spectators':
        return isSpectating || false;
      default:
        return false;
    }
  };

  const getUnreadCount = (channelId: string) => {
    // Handle both chatMessages (camelCase) and chat_messages (snake_case) 
    const messages = gameState?.chatMessages || (gameState as any)?.chat_messages || [];
    if (!Array.isArray(messages)) return 0;
    
    return messages.filter(msg => 
      msg.channelID === channelId && !msg.isSystem
    ).length; // Simplified for now - would need read state tracking for real unread count
  };

  const handleChannelClick = (channelId: string) => {
    if (getChannelAccess(channelId)) {
      setActiveChannel(channelId);
    }
  };

  return (
    <aside className="flex flex-col bg-background-secondary overflow-hidden select-none">
      <header className="px-4 py-3 border-b border-border flex justify-between items-center flex-shrink-0">
        <div className="header-left">
          <span className="font-mono font-bold text-base tracking-widest text-text-primary">LOEBIAN</span>
        </div>
        <div className="flex gap-1">
          <button 
            className="w-7 h-7 rounded-md flex items-center justify-center bg-background-tertiary text-sm transition-all duration-150 hover:bg-background-quaternary hover:scale-105 border-0 cursor-pointer" 
            aria-label={isMuted ? "Unmute Audio" : "Mute Audio"}
            onClick={handleToggleMute}
          >
            {isMuted ? '🔇' : '🔊'}
          </button>
          <button 
            className="w-7 h-7 rounded-md flex items-center justify-center bg-background-tertiary text-sm transition-all duration-150 hover:bg-background-quaternary hover:scale-105 border-0 cursor-pointer" 
            aria-label="Settings"
          >
            ⚙️
          </button>
          <button 
            className="w-7 h-7 rounded-md flex items-center justify-center bg-background-tertiary text-sm transition-all duration-150 hover:bg-background-quaternary hover:scale-105 border-0 cursor-pointer" 
            aria-label={`Switch to ${theme === 'dark' ? 'Light' : 'Dark'} Mode`}
            onClick={toggleTheme}
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
        </div>
      </header>

      <nav className="px-2 py-3 border-b border-border flex-shrink-0" aria-label="Chat channels">
        <div className="text-xs font-bold text-text-muted uppercase tracking-wider px-1.5 pb-1.5 mb-2 flex justify-between items-center">Text Channels</div>
        
        {/* War Room Channel */}
        <button
          onClick={() => handleChannelClick('#war-room')}
          disabled={!getChannelAccess('#war-room')}
          aria-current={activeChannel === '#war-room' ? 'page' : undefined}
          aria-label={`War room channel${getUnreadCount('#war-room') > 0 ? ` (${getUnreadCount('#war-room')} unread messages)` : ''}${!getChannelAccess('#war-room') ? ' (access denied)' : ''}`}
          className={`w-full flex items-center gap-1.5 px-2 py-1.5 rounded-md mb-0.5 transition-all duration-150 text-sm border-0 ${
            activeChannel === '#war-room' 
              ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
              : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
          } ${!getChannelAccess('#war-room') ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer hover:translate-x-0.5'}`}
        >
          <span>#</span>
          <span className={`channel-name ${getUnreadCount('#war-room') > 0 ? 'font-bold text-text-primary' : ''}`}>war-room</span>
          {!getChannelAccess('#war-room') && <span className="ml-auto text-xs opacity-50">❌</span>}
        </button>

        {/* Aligned Channel */}
        <button
          onClick={() => handleChannelClick('#aligned')}
          disabled={!getChannelAccess('#aligned')}
          aria-current={activeChannel === '#aligned' ? 'page' : undefined}
          aria-label={`Aligned channel${getUnreadCount('#aligned') > 0 ? ` (${getUnreadCount('#aligned')} unread messages)` : ''}${!getChannelAccess('#aligned') ? ' (access denied)' : ''}${alignedChannelJustUnlocked ? ' (recently unlocked)' : ''}`}
          className={`w-full flex items-center gap-1.5 px-2 py-1.5 rounded-md mb-0.5 transition-all duration-150 text-sm border-0 ${
            activeChannel === '#aligned' 
              ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
              : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
          } ${!getChannelAccess('#aligned') ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer hover:translate-x-0.5'} ${
            alignedChannelJustUnlocked ? CHANNEL_UNLOCK : ''
          }`}
        >
          <span>#</span>
          <span className={`channel-name ${getUnreadCount('#aligned') > 0 ? 'font-bold text-text-primary' : ''}`}>aligned</span>
          {!getChannelAccess('#aligned') && <span className="ml-auto text-xs opacity-50">❌</span>}
          {getChannelAccess('#aligned') && alignedChannelJustUnlocked && (
            <span className="ml-auto text-xs text-aligned">🔓</span>
          )}
        </button>

        {/* Off-boarding Channel - only show if there are deactivated players */}
        {deactivatedCount > 0 && (
          <button
            onClick={() => handleChannelClick('#off-boarding')}
            disabled={!getChannelAccess('#off-boarding')}
            aria-current={activeChannel === '#off-boarding' ? 'page' : undefined}
            aria-label={`Off-boarding channel${getUnreadCount('#off-boarding') > 0 ? ` (${getUnreadCount('#off-boarding')} unread messages)` : ''}${!getChannelAccess('#off-boarding') ? ' (access denied)' : ''}`}
            className={`w-full flex items-center gap-1.5 px-2 py-1.5 rounded-md mb-0.5 transition-all duration-150 text-sm border-0 ${
              activeChannel === '#off-boarding' 
                ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
                : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
            } ${!getChannelAccess('#off-boarding') ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer hover:translate-x-0.5'}`}
          >
            <span>#</span>
            <span className={`channel-name ${getUnreadCount('#off-boarding') > 0 ? 'font-bold text-text-primary' : ''}`}>off-boarding</span>
            {!getChannelAccess('#off-boarding') && <span className="ml-auto text-xs opacity-50">❌</span>}
          </button>
        )}

        {/* Spectators Channel - only show for spectators */}
        {isSpectating && (
          <button
            onClick={() => handleChannelClick('#spectators')}
            disabled={!getChannelAccess('#spectators')}
            aria-current={activeChannel === '#spectators' ? 'page' : undefined}
            aria-label={`Spectators channel${getUnreadCount('#spectators') > 0 ? ` (${getUnreadCount('#spectators')} unread messages)` : ''}`}
            className={`w-full flex items-center gap-1.5 px-2 py-1.5 rounded-md mb-0.5 transition-all duration-150 text-sm border-0 ${
              activeChannel === '#spectators' 
                ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
                : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
            } cursor-pointer hover:translate-x-0.5`}
          >
            <span>#</span>
            <span className={`channel-name ${getUnreadCount('#spectators') > 0 ? 'font-bold text-text-primary' : ''}`}>spectators</span>
            <span className="ml-auto text-xs">👁️</span>
          </button>
        )}
      </nav>

      <div className="px-2 py-3 flex-grow overflow-y-auto">
        <div className="text-xs font-bold text-text-muted uppercase tracking-wider px-1.5 pb-1.5 mb-2 flex justify-between items-center">
          <span>Personnel</span>
          <div className="flex gap-2">
            <span className="text-xs px-2 py-0.5 rounded-2xl bg-human text-white font-bold">👤 {humanCount}</span>
            <span className="text-xs px-2 py-0.5 rounded-2xl bg-aligned text-white font-bold">🤖 {alignedCount}</span>
            <span className="text-xs px-2 py-0.5 rounded-2xl bg-background-tertiary text-text-muted font-bold">👻 {deactivatedCount}</span>
          </div>
        </div>

        <motion.div 
          layout
          ref={(ref) => {
            // Apply staggered animation to player cards when they mount
            if (ref) {
              const playerCards = ref.querySelectorAll('[data-player-card]');
              if (playerCards.length > 0) {
                applyStaggeredAnimation(playerCards, 75);
              }
            }
          }}
        >
          <AnimatePresence mode="popLayout">
            {Array.isArray(players) && players.sort((a, b) => {
              // Sort self to top, then by alive status, then by name
              if (a.id === localPlayerId) return -1;
              if (b.id === localPlayerId) return 1;
              if (a.isAlive !== b.isAlive) return a.isAlive ? -1 : 1;
              return a.name.localeCompare(b.name);
            }).map((player) => (
              <PlayerCard
                key={player.id}
                player={player}
                isSelf={player.id === localPlayerId}
                isSelected={player.id === viewedPlayerId}
                onSelect={setViewedPlayer}
                isOnTrial={gameState?.phase?.type === 'TRIAL' && gameState?.nominatedPlayer === player.id}
              />
            ))}
          </AnimatePresence>
        </motion.div>
      </div>
    </aside>
  );
};