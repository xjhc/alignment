import React, { useEffect, useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Player } from '../../types';
import { useMouseGlow } from '../../hooks/useMouseGlow';
import { useSound } from '../../hooks/useSound';
import { ELIMINATION_FADE, GLITCH } from '../../utils/animations';

interface PlayerCardProps {
  player: Player;
  isSelf: boolean;
  isSelected: boolean;
  onSelect: (playerId: string) => void;
}

export const PlayerCard: React.FC<PlayerCardProps> = ({ player, isSelf, isSelected, onSelect }) => {
  const [wasAlive, setWasAlive] = useState(player.isAlive);
  const [isPlayingEliminationAnimation, setIsPlayingEliminationAnimation] = useState(false);
  const cardRef = useRef<HTMLDivElement>(null);
  const { elementRef, getGlowStyle } = useMouseGlow();
  const { playSound } = useSound();

  useEffect(() => {
    // Trigger elimination animation when player becomes not alive
    if (wasAlive && !player.isAlive) {
      setIsPlayingEliminationAnimation(true);
      
      // Play elimination sound effect
      playSound('error'); // Using error sound as elimination feedback
      
      if (cardRef.current) {
        // Apply the CSS animation class
        cardRef.current.classList.add(ELIMINATION_FADE);
        
        // Listen for animation end to update state
        const handleAnimationEnd = () => {
          setIsPlayingEliminationAnimation(false);
          if (cardRef.current) {
            cardRef.current.removeEventListener('animationend', handleAnimationEnd);
          }
        };
        
        cardRef.current.addEventListener('animationend', handleAnimationEnd);
      }
    }
    setWasAlive(player.isAlive);
  }, [player.isAlive, wasAlive, playSound]);

  const getPlayerAvatar = (p: Player) => {
    // Check if player abandoned the game (different from elimination)
    if (!p.isAlive && p.statusMessage === 'ABANDONED') return '🚪';
    
    // If the player is not alive (eliminated), show the ghost
    if (!p.isAlive) return '👻';

    // Use the avatar from player object if available, otherwise fall back to generic icon
    return p.avatar || '👤';
  };

  const getPlayerClasses = () => {
    const classes = [
      'flex items-start gap-2 p-1.5 px-2 rounded-md cursor-pointer mb-0.5 min-h-10 will-change-transform'
    ];
    
    if (isSelf) {
      classes.push('bg-background-tertiary border-l-2 border-human');
    }
    
    if (isSelected) {
      classes.push('bg-background-secondary border-l-2 border-blue-500');
    }
    
    if (!player.isAlive && !isPlayingEliminationAnimation) {
      if (player.statusMessage === 'ABANDONED') {
        classes.push('opacity-70 bg-red-500/5 border-l-2 border-red-500');
      } else {
        classes.push('opacity-60 grayscale-[60%]');
      }
    }
    
    if (player.alignment === 'AI' || player.alignment === 'ALIGNED') {
      classes.push('bg-aligned/5 border-l-2 border-aligned');
      // Add glitch animation to AI-aligned players
      classes.push(GLITCH);
    }
    
    return classes.join(' ');
  };

  const getDataAttributes = () => {
    const attrs: { [key: string]: string } = { 'data-name': player.name };
    if (player.systemShocks?.some(shock => shock.isActive)) {
      attrs['data-status'] = 'shocked';
    }
    return attrs;
  };

  const renderProjectMilestones = () => {
    const maxMilestones = 3; // Assuming max 3 milestones
    const icons = [];
    for (let i = 0; i < maxMilestones; i++) {
      const filled = i < player.projectMilestones;
      const isAligned = player.alignment === 'AI' || player.alignment === 'ALIGNED';
      const progressClass = filled ? 
        `text-xs transition-all duration-150 ${isAligned ? 'text-aligned' : 'text-human'}` : 
        'text-xs text-text-muted transition-all duration-150';
      icons.push(
        <span key={i} className={progressClass} title={`Project Progress: ${player.projectMilestones} / ${maxMilestones}`}>
          {filled ? '⬢' : '⬡'}
        </span>
      );
    }
    return icons;
  };

  const displayName = isSelf ? `${player.name} (Me)` : player.name;
  const displayTokens = player.isAlive 
    ? `🪙 ${player.tokens}` 
    : (player.statusMessage === 'ABANDONED' ? '🚪' : '❌');

  const getAriaLabel = () => {
    const parts = [
      isSelf ? `${player.name}, yourself` : player.name,
      player.jobTitle || 'Employee',
      player.isAlive ? 'alive' : (player.statusMessage === 'ABANDONED' ? 'abandoned' : 'eliminated'),
      `${player.tokens} tokens`,
      `${player.projectMilestones} of 3 project milestones completed`
    ];
    
    if (player.isRolePubliclyRevealed && !isSelf) {
      parts.push('role publicly revealed');
    }
    
    if (player.statusMessage && player.statusMessage !== 'ABANDONED') {
      parts.push(`status: ${player.statusMessage}`);
    }
    
    return parts.join(', ');
  };

  return (
    <AnimatePresence>
      {(player.isAlive || isPlayingEliminationAnimation) && (
        <motion.button
          ref={(node) => {
            cardRef.current = node;
            if (elementRef) elementRef.current = node;
          }}
          className={`${getPlayerClasses()} relative overflow-hidden border-0 text-left w-full`}
          {...getDataAttributes()}
          onClick={() => {
            // Play selection sound for interactive feedback
            playSound('buttonClick');
            onSelect(player.id);
          }}
          aria-label={getAriaLabel()}
          aria-pressed={isSelected}
          layout
          initial={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          whileHover={!isPlayingEliminationAnimation ? { 
            x: 2,
            scale: 1.01,
            transition: { duration: 0.15 }
          } : {}}
          whileTap={!isPlayingEliminationAnimation ? { scale: 0.98 } : {}}
          style={getGlowStyle('rgba(59, 130, 246, 0.1)', 150)}
        >
          {/* Interactive glow overlay */}
          <div 
            className="absolute inset-0 pointer-events-none"
            style={getGlowStyle('rgba(59, 130, 246, 0.05)', 100)}
          />
      <motion.div 
        className="w-7 h-7 rounded-full bg-background-tertiary flex items-center justify-center text-sm flex-shrink-0 border border-border mt-0.5 relative z-10"
        whileHover={!isPlayingEliminationAnimation ? { scale: 1.1, transition: { duration: 0.15 } } : {}}
      >
        {getPlayerAvatar(player)}
      </motion.div>
      <div className="flex-grow flex flex-col justify-center min-h-9 relative z-10">
        <div className="flex items-center gap-1.5 text-xs leading-tight">
          <span className={`font-semibold flex-shrink-0 min-w-10 ${isSelf ? 'text-human' : 'text-text-primary'}`}>
            {displayName}
          </span>
          <span className={`font-medium uppercase tracking-wide flex-shrink-0 text-[10px] min-w-7.5 ${
            (player.isRolePubliclyRevealed && !isSelf) ? 'text-yellow-500 font-semibold' : 'text-text-secondary'
          }`}>
            {(player.isRolePubliclyRevealed && !isSelf) && '🔍 '}
            {/* Add alignment icon for non-color accessibility */}
            {(player.alignment === 'AI' || player.alignment === 'ALIGNED') && '🤖 '}
            {player.alignment === 'HUMAN' && '👤 '}
            {isSelf ? (player.role?.name || player.jobTitle || 'Employee') : 
             (player.isRolePubliclyRevealed ? player.role?.name || 'Unknown Role' : player.jobTitle || 'Employee')}
          </span>
          <div className="flex items-center gap-1.5 ml-auto">
            <motion.span 
              className={`font-semibold flex-shrink-0 text-[11px] min-w-6 ${
                player.alignment === 'AI' || player.alignment === 'ALIGNED' 
                  ? 'text-aligned' 
                  : 'text-human'
              }`}
              animate={player.alignment === 'AI' || player.alignment === 'ALIGNED' ? {
                opacity: [1, 0.6, 1],
                transition: { duration: 1.5, repeat: Infinity, ease: "easeInOut" }
              } : {}}
            >
              {displayTokens}
            </motion.span>
            <div className="flex gap-px items-center flex-shrink-0 min-w-6">
              {renderProjectMilestones()}
            </div>
          </div>
        </div>
        {player.statusMessage && (
          <div className={`text-[11px] italic mt-0.5 overflow-hidden text-ellipsis whitespace-nowrap leading-tight ${
            player.statusMessage === 'ABANDONED' ? 'text-red-500 font-semibold' :
            player.systemShocks?.some(shock => shock.isActive) ? 'text-pink-500' : 'text-text-muted'
          } ${!player.isAlive ? 'opacity-80' : ''}`}>
            {player.statusMessage === 'ABANDONED' ? '🚪 ABANDONED' : `"${player.statusMessage}"`}
          </div>
        )}
      </div>
        </motion.button>
      )}
    </AnimatePresence>
  );
};