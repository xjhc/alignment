import React, { useEffect, useState } from 'react';
import { motion, useAnimate, AnimatePresence } from 'framer-motion';
import { Player } from '../../types';
import { useMouseGlow } from '../../hooks/useMouseGlow';

interface PlayerCardProps {
  player: Player;
  isSelf: boolean;
  isSelected: boolean;
  onSelect: (playerId: string) => void;
}

export const PlayerCard: React.FC<PlayerCardProps> = ({ player, isSelf, isSelected, onSelect }) => {
  const [wasAlive, setWasAlive] = useState(player.isAlive);
  const [showEliminationAnimation, setShowEliminationAnimation] = useState(false);
  const [scope, animate] = useAnimate();
  const { elementRef, getGlowStyle } = useMouseGlow();

  useEffect(() => {
    // Trigger elimination animation when player becomes not alive
    if (wasAlive && !player.isAlive) {
      setShowEliminationAnimation(true);
      
      // Digital erasure sequence
      const digitalErasure = async () => {
        // Glitch effect - rapid text/color changes
        await animate(
          scope.current,
          { 
            x: [-2, 2, -1, 1, 0],
            filter: ['hue-rotate(0deg)', 'hue-rotate(180deg)', 'hue-rotate(360deg)'],
            opacity: [1, 0.7, 1, 0.5, 1]
          },
          { duration: 0.4, ease: "easeInOut" }
        );
        
        // Vertical shrink to zero height
        await animate(
          scope.current,
          { 
            height: 0,
            opacity: 0,
            marginBottom: 0,
            paddingTop: 0,
            paddingBottom: 0
          },
          { duration: 0.6, ease: "easeInOut" }
        );
        
        setShowEliminationAnimation(false);
      };
      
      digitalErasure();
    }
    setWasAlive(player.isAlive);
  }, [player.isAlive, wasAlive, animate, scope]);

  const getPlayerAvatar = (p: Player) => {
    // If the player is not alive, always show the ghost.
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
    
    if (!player.isAlive && !showEliminationAnimation) {
      classes.push('opacity-60 grayscale-[60%]');
    }
    
    if (player.alignment === 'AI' || player.alignment === 'ALIGNED') {
      classes.push('bg-aligned/5 border-l-2 border-aligned');
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
  const displayTokens = player.isAlive ? `🪙 ${player.tokens}` : '❌';

  return (
    <AnimatePresence>
      {(player.isAlive || showEliminationAnimation) && (
        <motion.div
          ref={(node) => {
            if (scope) scope.current = node;
            if (elementRef) elementRef.current = node;
          }}
          className={`${getPlayerClasses()} relative overflow-hidden`}
          {...getDataAttributes()}
          onClick={() => onSelect(player.id)}
          title={`View Dossier for ${player.name}${player.isRolePubliclyRevealed && !isSelf ? ' (Role Publicly Revealed)' : ''}`}
          layout
          initial={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          whileHover={!showEliminationAnimation ? { 
            x: 2,
            scale: 1.01,
            transition: { duration: 0.15 }
          } : {}}
          whileTap={!showEliminationAnimation ? { scale: 0.98 } : {}}
          style={getGlowStyle('rgba(59, 130, 246, 0.1)', 150)}
        >
          {/* Interactive glow overlay */}
          <div 
            className="absolute inset-0 pointer-events-none"
            style={getGlowStyle('rgba(59, 130, 246, 0.05)', 100)}
          />
      <motion.div 
        className="w-7 h-7 rounded-full bg-background-tertiary flex items-center justify-center text-sm flex-shrink-0 border border-border mt-0.5 relative z-10"
        whileHover={!showEliminationAnimation ? { scale: 1.1, transition: { duration: 0.15 } } : {}}
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
          <div className={`text-[11px] text-text-muted italic mt-0.5 overflow-hidden text-ellipsis whitespace-nowrap leading-tight ${
            player.systemShocks?.some(shock => shock.isActive) ? 'text-pink-500' : ''
          } ${!player.isAlive ? 'opacity-80' : ''}`}>
            "{player.statusMessage}"
          </div>
        )}
      </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};