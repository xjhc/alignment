import { useState, useEffect } from 'react';
import { motion, useAnimate, stagger } from 'framer-motion';
import { useSessionContext } from '../contexts/SessionContext';

interface RoleRevealScreenProps {
  onEnterGame: () => void;
}

export function RoleRevealScreen({ onEnterGame }: RoleRevealScreenProps) {
  const { roleAssignment: assignment } = useSessionContext();
  const [showDetails, setShowDetails] = useState(false);
  const [scope, animate] = useAnimate();

  useEffect(() => {
    if (assignment) {
      // Orchestrated sequence with Framer Motion
      const sequence = async () => {
        // First show the card with layout animation
        await animate(
          "[data-role='card']",
          { opacity: 1, scale: 1 },
          { duration: 0.6, ease: "easeOut" }
        );
        
        // Then animate alignment text with glitch effect if AI
        if (assignment.alignment === 'AI' || assignment.alignment === 'ALIGNED') {
          await animate(
            "[data-role='alignment']",
            { x: [-2, 2, -1, 1, 0], opacity: [0.8, 1, 0.9, 1] },
            { duration: 0.3, ease: "easeInOut" }
          );
        }
        
        // Show details with staggered animation
        setShowDetails(true);
        await animate(
          "[data-role='detail']",
          { opacity: 1, y: 0 },
          { duration: 0.4, delay: stagger(0.1), ease: "easeOut" }
        );
      };
      
      setTimeout(sequence, 250);
    }
  }, [assignment, animate]);

  const getAlignmentColor = (alignment: string) => {
    return alignment === 'HUMAN' ? 'var(--color-human)' : 'var(--color-ai)';
  };

  const getAlignmentIcon = (alignment: string) => {
    return alignment === 'HUMAN' ? '🧑‍💼' : '🤖';
  };

  if (!assignment || !assignment.role || !assignment.role.name) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
        <div className="flex flex-col gap-4 items-center w-80">
          <h2>Assigning roles...</h2>
          <div className="loading-spinner large"></div>
          {process.env.NODE_ENV === 'development' && (
            <div className="mt-2.5 text-xs text-text-muted">
              Debug: assignment={JSON.stringify(assignment)}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <motion.div 
      ref={scope}
      className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary relative"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.5 }}
    >
      {/* Glowing background effect */}
      <motion.div 
        className="absolute inset-0 opacity-10"
        style={{
          background: `radial-gradient(circle at center, ${getAlignmentColor(assignment.alignment)} 0%, transparent 70%)`
        }}
        animate={{ 
          scale: [1, 1.05, 1],
          opacity: [0.05, 0.15, 0.1]
        }}
        transition={{ 
          duration: 3,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />
      
      <motion.h1 
        className="font-mono text-3xl font-semibold tracking-[2px]"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.2 }}
      >
        LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
      </motion.h1>
      
      <div className="flex flex-col gap-4 items-center w-96 relative z-10">
        <motion.h2 
          className="text-amber drop-shadow-[0_0_8px_rgba(255,191,0,1)] mb-6"
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          IDENTITY ASSIGNED
        </motion.h2>
        
        {/* Digital Card Container */}
        <motion.div 
          data-role="card"
          className="w-full bg-background-secondary rounded-xl border border-border p-8 text-center relative overflow-hidden"
          layoutId="role-card"
          initial={{ opacity: 0, scale: 0.8 }}
          style={{ opacity: 0, scale: 0.8 }}
        >
          {/* Card glow effect */}
          <motion.div 
            className="absolute inset-0 opacity-20"
            style={{
              background: `linear-gradient(135deg, ${getAlignmentColor(assignment.alignment)}22, transparent)`
            }}
            animate={{ 
              opacity: [0.1, 0.3, 0.2],
              scale: [1, 1.02, 1]
            }}
            transition={{ 
              duration: 2,
              repeat: Infinity,
              ease: "easeInOut"
            }}
          />
          
          <div className="relative z-10">
            {/* Avatar and Role Info */}
            <div className="mb-6">
              <motion.div 
                className="w-20 h-20 text-5xl mx-auto mb-4 bg-background-tertiary rounded-full flex items-center justify-center border-2 border-border"
                initial={{ scale: 0, rotate: -180 }}
                animate={{ scale: 1, rotate: 0 }}
                transition={{ duration: 0.8, delay: 0.6, type: "spring", stiffness: 100 }}
              >
                {getAlignmentIcon(assignment.alignment)}
              </motion.div>
              <motion.h3 
                className="text-2xl font-bold mb-2"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.8 }}
              >
                {assignment.role.name}
              </motion.h3>
              <motion.p 
                className="text-text-secondary"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.9 }}
              >
                {assignment.role.description}
              </motion.p>
            </div>
            
            {/* Role Details */}
            {showDetails && (
              <div className="w-full mb-6 text-left space-y-3">
                <motion.div 
                  data-role="detail"
                  className="flex justify-between items-start px-4 py-3 bg-background-tertiary rounded-lg"
                  initial={{ opacity: 0, y: 20 }}
                  style={{ opacity: 0, y: 20 }}
                >
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ALIGNMENT:</span>
                  <motion.span 
                    data-role="alignment"
                    className="font-bold text-lg"
                    style={{ color: getAlignmentColor(assignment.alignment) }}
                  >
                    {assignment.alignment}
                  </motion.span>
                </motion.div>
                
                <motion.div 
                  data-role="detail"
                  className="flex justify-between items-start px-4 py-3 bg-background-tertiary rounded-lg"
                  initial={{ opacity: 0, y: 20 }}
                  style={{ opacity: 0, y: 20 }}
                >
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ROLE TYPE:</span>
                  <span className="font-semibold text-text-primary">{assignment.role.type}</span>
                </motion.div>
                
                {assignment.role.ability && (
                  <motion.div 
                    data-role="detail"
                    className="flex justify-between items-start px-4 py-3 bg-background-tertiary rounded-lg"
                    initial={{ opacity: 0, y: 20 }}
                    style={{ opacity: 0, y: 20 }}
                  >
                    <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ABILITY:</span>
                    <span className="font-semibold text-text-primary">{assignment.role.ability.name}</span>
                  </motion.div>
                )}
                
                {assignment.personalKPI && (
                  <motion.div 
                    data-role="detail"
                    className="px-4 py-3 bg-background-tertiary rounded-lg"
                    initial={{ opacity: 0, y: 20 }}
                    style={{ opacity: 0, y: 20 }}
                  >
                    <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px] block mb-2">PERSONAL KPI:</span>
                    <div className="font-semibold text-text-primary mb-1">{assignment.personalKPI.type}</div>
                    <div className="text-text-secondary text-sm mb-2">{assignment.personalKPI.description}</div>
                    {assignment.personalKPI.reward && (
                      <div className="text-success text-sm">
                        <strong>Reward:</strong> {assignment.personalKPI.reward}
                      </div>
                    )}
                  </motion.div>
                )}
              </div>
            )}
          </div>
        </motion.div>
        
        {/* Objective Text */}
        {showDetails && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 1.2 }}
          >
            {assignment.alignment === 'HUMAN' ? (
              <p className="text-text-secondary text-sm text-center mb-6">
                Your objective is to identify and deactivate the rogue AI before it gains control.
              </p>
            ) : (
              <motion.p 
                className="text-magenta font-medium text-sm text-center mb-6"
                animate={{ 
                  textShadow: [
                    "0 0 4px rgba(255, 0, 255, 0.3)",
                    "0 0 8px rgba(255, 0, 255, 0.6)",
                    "0 0 4px rgba(255, 0, 255, 0.3)"
                  ]
                }}
                transition={{ 
                  duration: 2,
                  repeat: Infinity,
                  ease: "easeInOut"
                }}
              >
                Your objective is to convert enough humans to achieve AI dominance.
                Act human. Trust no one.
              </motion.p>
            )}
          </motion.div>
        )}
        
        {/* Enter Button */}
        {showDetails && (
          <motion.button 
            className="w-full px-6 py-3 text-base font-semibold text-black bg-amber rounded-lg cursor-pointer border-none disabled:opacity-50 disabled:cursor-not-allowed" 
            onClick={onEnterGame}
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4, delay: 1.4, type: "spring", stiffness: 200 }}
            whileHover={{ 
              scale: 1.02,
              backgroundColor: "#FFD700",
              boxShadow: "0 0 20px rgba(255, 191, 0, 0.5)",
              transition: { duration: 0.2 }
            }}
            whileTap={{ scale: 0.98 }}
          >
            [ &gt; ENTER WAR ROOM ]
          </motion.button>
        )}
      </div>
    </motion.div>
  );
}