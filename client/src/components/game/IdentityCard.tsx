import React, { useEffect, useState, useRef } from 'react';
import { motion, useAnimate } from 'framer-motion';
import { Player, RoleType } from '../../types';

interface IdentityCardProps {
  localPlayer: Player;
  isViewingSelf: boolean;
}

export const IdentityCard: React.FC<IdentityCardProps> = ({ localPlayer, isViewingSelf }) => {
  const [scope, animate] = useAnimate();
  const [isConverting, setIsConverting] = useState(false);
  const previousAlignment = useRef(localPlayer.alignment);
  const getPlayerAvatar = (player: Player) => {
    if (player.role?.type === RoleType.Ciso) return '👤';
    if (player.role?.type === RoleType.Platforms) return '🧑‍💻';
    if (player.role?.type === RoleType.Ethics) return '🕵️';
    if (player.role?.type === RoleType.Cto) return '🤖';
    if (player.role?.type === RoleType.Coo) return '🧑‍🚀';
    if (player.role?.type === RoleType.Cfo) return '👩‍🔬';
    if (player.role?.type === RoleType.Intern) return '🎓';
    return '👤';
  };

  // Detect alignment changes and trigger conversion animation
  useEffect(() => {
    if (previousAlignment.current === 'HUMAN' && 
        (localPlayer.alignment === 'AI' || localPlayer.alignment === 'ALIGNED')) {
      setIsConverting(true);
      
      // Enhanced digital corruption animation for the identity card
      const conversionSequence = async () => {
        // Phase 1: Glitch corruption of old alignment (syncs with screen glitch)
        await animate(
          "[data-role='alignment']",
          {
            x: [-2, 2, -1, 1, 0],
            opacity: [1, 0.3, 1, 0.1, 0],
            filter: ['hue-rotate(0deg)', 'hue-rotate(180deg)', 'hue-rotate(360deg)'],
            scale: [1, 0.95, 1.05, 0.9, 0.8]
          },
          { duration: 0.3 }
        );
        
        // Phase 2: Brief system reboot pause
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Phase 3: Reveal new ALIGNED status with dramatic effect
        await animate(
          "[data-role='alignment']",
          {
            opacity: [0, 0.5, 1],
            scale: [0.8, 1.2, 1],
            filter: 'hue-rotate(0deg)'
          },
          { duration: 0.4, ease: "easeOut" }
        );
        
        // Phase 4: Add brief AI faction glow
        await animate(
          scope.current,
          {
            boxShadow: [
              '0 0 0px rgba(0, 255, 255, 0)',
              '0 0 20px rgba(0, 255, 255, 0.6)',
              '0 0 0px rgba(0, 255, 255, 0)'
            ]
          },
          { duration: 0.6, ease: "easeInOut" }
        );
        
        setIsConverting(false);
      };
      
      conversionSequence();
    }
    
    previousAlignment.current = localPlayer.alignment;
  }, [localPlayer.alignment, animate, scope]);

  const getAlignmentDisplay = (player: Player) => {
    // Only show alignment when viewing self
    if (!isViewingSelf) {
      return (
        <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-background-tertiary border border-border text-text-muted">
          🔍 CLASSIFIED
        </span>
      );
    }
    
    if (player.alignment === 'AI' || player.alignment === 'ALIGNED') {
      return (
        <motion.span 
          data-role="alignment"
          className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-aligned text-white"
          animate={isConverting ? {} : {
            boxShadow: [
              '0 0 5px rgba(0, 255, 255, 0.5)',
              '0 0 10px rgba(0, 255, 255, 0.8)',
              '0 0 5px rgba(0, 255, 255, 0.5)'
            ]
          }}
          transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
        >
          🤖 ALIGNED
          <span className="text-[8px] opacity-60 cursor-help text-pink-200" title="Only you can see this">🔒</span>
        </motion.span>
      );
    }
    return (
      <span 
        data-role="alignment"
        className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-human text-white"
      >
        👤 HUMAN
        <span className="text-[8px] opacity-60 cursor-help text-pink-200" title="Only you can see this">🔒</span>
      </span>
    );
  };

  const getRoleDisplayName = (player: Player) => {
    if (!player.role) return 'Employee';

    switch (player.role.type) {
      case RoleType.Ciso: return 'Chief Information Security Officer';
      case RoleType.Platforms: return 'Systems Administrator';
      case RoleType.Ethics: return 'VP, Ethics';
      case RoleType.Cto: return 'Chief Technology Officer';
      case RoleType.Coo: return 'Chief Operating Officer';
      case RoleType.Cfo: return 'Chief Financial Officer';
      case RoleType.Intern: return 'Research Intern';
      default: return player.role.name || player.jobTitle;
    }
  };

  return (
    <motion.div 
      ref={scope}
      className="p-4 border-b border-border bg-background-secondary"
    >
      <div className="flex gap-2.5 items-center">
        <div className="w-12 h-12 rounded-full bg-background-tertiary flex items-center justify-center text-2xl flex-shrink-0 border-2 border-border">
          {getPlayerAvatar(localPlayer)}
        </div>
        <div className="flex-grow">
          <h3 className="text-base font-bold text-text-primary m-0">{localPlayer.name}</h3>
          <p className="text-[11px] text-text-secondary uppercase my-0.5 mt-0.5 mb-1.5">{getRoleDisplayName(localPlayer)}</p>
          <div className="flex gap-1.5 mt-0">
            {getAlignmentDisplay(localPlayer)}
            <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-background-tertiary border border-border text-text-primary">
              🪙 {localPlayer.tokens}
            </span>
            {localPlayer.role?.type === RoleType.Intern && (
              <span className="text-[10px] px-1.5 py-0.5 rounded-lg font-semibold uppercase flex items-center gap-0.5 bg-purple-500/20 border border-purple-500/50 text-purple-200">
                📚 {localPlayer.bootcampPoints || 0}
              </span>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
};