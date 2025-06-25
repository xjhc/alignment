import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { useTheme } from '../../hooks/useTheme';
import { PlayerCard } from './PlayerCard';

export const RosterPanel: React.FC = () => {
  const { gameState, localPlayerId, localPlayer, viewedPlayerId, setViewedPlayer } = useGameContext();
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
        <div className="flex items-center gap-1.5 px-2 py-1.5 rounded-md text-white font-normal cursor-pointer mb-0.5 transition-all duration-150 text-sm bg-amber-500 hover:bg-amber-600 hover:translate-x-0.5">
          <span>#</span>
          <span className="channel-name">war-room</span>
        </div>
        <div className={`flex items-center gap-1.5 px-2 py-1.5 rounded-md cursor-pointer mb-0.5 transition-all duration-150 text-sm ${
          isAI 
            ? 'text-cyan-600 bg-cyan-500/10 opacity-100' 
            : 'text-text-muted opacity-60'
        } hover:bg-background-tertiary hover:text-text-primary hover:translate-x-0.5`}>
          <span>#</span>
          <span className="channel-name">aligned</span>
          {!isAI && <span className="ml-auto text-xs opacity-50">❌</span>}
        </div>
        {deactivatedCount > 0 && (
          <div className="flex items-center gap-1.5 px-2 py-1.5 rounded-md text-text-secondary font-normal cursor-pointer mb-0.5 transition-all duration-150 text-sm hover:bg-background-tertiary hover:text-text-primary hover:translate-x-0.5">
            <span>#</span>
            <span className="channel-name">off-boarding</span>
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