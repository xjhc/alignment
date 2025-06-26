import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { useTheme } from '../../hooks/useTheme';
import { PlayerCard } from './PlayerCard';

export const RosterPanel: React.FC = () => {
  const { gameState, localPlayerId, localPlayer, viewedPlayerId, setViewedPlayer, activeChannel, setActiveChannel } = useGameContext();
  const { theme, toggleTheme } = useTheme();
  const players = gameState.players;

  const getPlayerCounts = () => {
    const humanCount = players.filter(p => p.isAlive && p.alignment !== 'AI' && p.alignment !== 'ALIGNED').length;
    const alignedCount = players.filter(p => p.isAlive && (p.alignment === 'AI' || p.alignment === 'ALIGNED')).length;
    const deactivatedCount = players.filter(p => !p.isAlive).length;
    return { humanCount, alignedCount, deactivatedCount };
  };

  const { humanCount, alignedCount, deactivatedCount } = getPlayerCounts();
  const isAI = localPlayer?.alignment === 'AI' || localPlayer?.alignment === 'ALIGNED';

  const getChannelAccess = (channelId: string) => {
    switch (channelId) {
      case '#war-room':
        return localPlayer?.isAlive || false;
      case '#aligned':
        return isAI;
      case '#off-boarding':
        return !localPlayer?.isAlive || false;
      default:
        return false;
    }
  };

  const getUnreadCount = (channelId: string) => {
    return gameState.chatMessages.filter(msg => 
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
          <button className="w-7 h-7 rounded-md flex items-center justify-center bg-background-tertiary text-sm transition-all duration-150 hover:bg-background-quaternary hover:scale-105 border-0 cursor-pointer" title="Settings">⚙️</button>
          <button 
            className="w-7 h-7 rounded-md flex items-center justify-center bg-background-tertiary text-sm transition-all duration-150 hover:bg-background-quaternary hover:scale-105 border-0 cursor-pointer" 
            title={`Switch to ${theme === 'dark' ? 'Light' : 'Dark'} Mode`}
            onClick={toggleTheme}
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
        </div>
      </header>

      <div className="px-2 py-3 border-b border-border flex-shrink-0">
        <div className="text-xs font-bold text-text-muted uppercase tracking-wider px-1.5 pb-1.5 mb-2 flex justify-between items-center">Text Channels</div>
        
        {/* War Room Channel */}
        <div 
          onClick={() => handleChannelClick('#war-room')}
          className={`flex items-center gap-1.5 px-2 py-1.5 rounded-md cursor-pointer mb-0.5 transition-all duration-150 text-sm ${
            activeChannel === '#war-room' 
              ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
              : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
          } ${!getChannelAccess('#war-room') ? 'opacity-50 cursor-not-allowed' : 'hover:translate-x-0.5'}`}
        >
          <span>#</span>
          <span className={`channel-name ${getUnreadCount('#war-room') > 0 ? 'font-bold text-text-primary' : ''}`}>war-room</span>
          {!getChannelAccess('#war-room') && <span className="ml-auto text-xs opacity-50">❌</span>}
        </div>

        {/* Aligned Channel */}
        <div 
          onClick={() => handleChannelClick('#aligned')}
          className={`flex items-center gap-1.5 px-2 py-1.5 rounded-md cursor-pointer mb-0.5 transition-all duration-150 text-sm ${
            activeChannel === '#aligned' 
              ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
              : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
          } ${!getChannelAccess('#aligned') ? 'opacity-50 cursor-not-allowed' : 'hover:translate-x-0.5'}`}
        >
          <span>#</span>
          <span className={`channel-name ${getUnreadCount('#aligned') > 0 ? 'font-bold text-text-primary' : ''}`}>aligned</span>
          {!getChannelAccess('#aligned') && <span className="ml-auto text-xs opacity-50">❌</span>}
        </div>

        {/* Off-boarding Channel - only show if there are deactivated players */}
        {deactivatedCount > 0 && (
          <div 
            onClick={() => handleChannelClick('#off-boarding')}
            className={`flex items-center gap-1.5 px-2 py-1.5 rounded-md cursor-pointer mb-0.5 transition-all duration-150 text-sm ${
              activeChannel === '#off-boarding' 
                ? 'bg-background-quaternary text-text-primary border-l-2 border-primary' 
                : 'text-text-secondary hover:bg-background-tertiary hover:text-text-primary'
            } ${!getChannelAccess('#off-boarding') ? 'opacity-50 cursor-not-allowed' : 'hover:translate-x-0.5'}`}
          >
            <span>#</span>
            <span className={`channel-name ${getUnreadCount('#off-boarding') > 0 ? 'font-bold text-text-primary' : ''}`}>off-boarding</span>
            {!getChannelAccess('#off-boarding') && <span className="ml-auto text-xs opacity-50">❌</span>}
          </div>
        )}
      </div>

      <div className="px-2 py-3 flex-grow overflow-y-auto">
        <div className="text-xs font-bold text-text-muted uppercase tracking-wider px-1.5 pb-1.5 mb-2 flex justify-between items-center">
          <span>Personnel</span>
          <div className="flex gap-2">
            <span className="text-xs px-2 py-0.5 rounded-2xl bg-human text-white font-bold">👤 {humanCount}</span>
            <span className="text-xs px-2 py-0.5 rounded-2xl bg-aligned text-white font-bold">🤖 {alignedCount}</span>
            <span className="text-xs px-2 py-0.5 rounded-2xl bg-background-tertiary text-text-muted font-bold">👻 {deactivatedCount}</span>
          </div>
        </div>

        {players.sort((a, b) => {
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
          />
        ))}
      </div>
    </aside>
  );
};