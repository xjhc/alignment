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
        try {
          // First show the card with layout animation
          const cardElement = document.querySelector("[data-role='card']");
          if (cardElement) {
            await animate(
              "[data-role='card']",
              { opacity: 1, scale: 1 },
              { duration: 0.6, ease: "easeOut" }
            );
          }
          
          // Then animate alignment text with glitch effect if AI
          if (assignment.alignment === 'AI' || assignment.alignment === 'ALIGNED') {
            const alignmentElement = document.querySelector("[data-role='alignment']");
            if (alignmentElement) {
              await animate(
                "[data-role='alignment']",
                { x: [-2, 2, -1, 1, 0], opacity: [0.8, 1, 0.9, 1] },
                { duration: 0.3, ease: "easeInOut" }
              );
            }
          }
          
          // Show details with staggered animation
          setShowDetails(true);
          
          // Wait for the details to be rendered before animating
          await new Promise(resolve => setTimeout(resolve, 100));
          
          const detailElements = document.querySelectorAll("[data-role='detail']");
          if (detailElements.length > 0) {
            await animate(
              "[data-role='detail']",
              { opacity: 1, y: 0 },
              { duration: 0.4, delay: stagger(0.1), ease: "easeOut" }
            );
          }
        } catch (error) {
          console.error('Animation sequence error:', error);
          // Fallback: just show the details without animation
          setShowDetails(true);
        }
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

  const getRoleIcon = (roleType: string) => {
    switch (roleType) {
      case 'CEO': return '👑';
      case 'CTO': return '⚙️';
      case 'COO': return '🏢';
      case 'CFO': return '💰';
      case 'CISO': return '🔐';
      case 'ETHICS': return '⚖️';
      case 'PLATFORMS': return '🌐';
      case 'INTERN': return '📚';
      default: return '🧑‍💼';
    }
  };

  const getEnhancedRoleInfo = (roleType: string) => {
    switch (roleType) {
      case 'CEO':
        return {
          title: 'Chief Executive Officer',
          description: 'The visionary leader responsible for the company\'s strategic direction and overall success.',
          abilities: 'Can use executive privilege to influence critical decisions and has access to high-level intelligence.',
          keyPowers: ['Executive Override', 'Strategic Intel', 'Company Direction']
        };
      case 'CTO':
        return {
          title: 'Chief Technology Officer',
          description: 'The technical mastermind overseeing all technology infrastructure and innovation.',
          abilities: 'Can analyze system logs and technical data to identify anomalies and potential security threats.',
          keyPowers: ['Technical Analysis', 'System Monitoring', 'Infrastructure Control']
        };
      case 'COO':
        return {
          title: 'Chief Operating Officer',
          description: 'The operational expert ensuring smooth day-to-day business functions.',
          abilities: 'Can coordinate team activities and identify operational irregularities.',
          keyPowers: ['Operations Oversight', 'Team Coordination', 'Process Control']
        };
      case 'CFO':
        return {
          title: 'Chief Financial Officer',
          description: 'The financial guardian responsible for the company\'s fiscal health and resource allocation.',
          abilities: 'Can track financial anomalies and resource expenditures that may indicate AI activity.',
          keyPowers: ['Financial Analysis', 'Budget Control', 'Resource Allocation']
        };
      case 'CISO':
        return {
          title: 'Chief Information Security Officer',
          description: 'The security specialist defending against cyber threats and protecting sensitive data.',
          abilities: 'Can investigate security breaches and identify potential infiltration attempts.',
          keyPowers: ['Security Investigation', 'Threat Detection', 'Access Control']
        };
      case 'ETHICS':
        return {
          title: 'VP, Ethics & Alignment',
          description: 'The moral compass ensuring all AI systems remain aligned with human values.',
          abilities: 'Can detect behavioral anomalies and assess alignment status of team members.',
          keyPowers: ['Alignment Assessment', 'Behavioral Analysis', 'Ethics Oversight']
        };
      case 'PLATFORMS':
        return {
          title: 'VP, Platforms',
          description: 'The platform architect responsible for the infrastructure that powers the company.',
          abilities: 'Can analyze platform usage patterns and identify suspicious activities.',
          keyPowers: ['Platform Analysis', 'Usage Monitoring', 'System Architecture']
        };
      case 'INTERN':
        return {
          title: 'Intern',
          description: 'The eager newcomer learning the ropes and proving their worth to the organization.',
          abilities: 'May have limited abilities initially but can grow in power through successful contributions.',
          keyPowers: ['Learning Opportunities', 'Growth Potential', 'Fresh Perspective']
        };
      default:
        return {
          title: assignment?.role.name || 'Unknown Role',
          description: assignment?.role.description || 'Role information not available.',
          abilities: 'Role abilities are being determined...',
          keyPowers: ['To Be Determined']
        };
    }
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
              <div className="flex items-center justify-center gap-3 mb-4">
                <motion.div 
                  className="w-20 h-20 text-5xl bg-background-tertiary rounded-full flex items-center justify-center border-2 border-border"
                  initial={{ scale: 0, rotate: -180 }}
                  animate={{ scale: 1, rotate: 0 }}
                  transition={{ duration: 0.8, delay: 0.6, type: "spring", stiffness: 100 }}
                >
                  {getRoleIcon(assignment.role.type)}
                </motion.div>
                <motion.div 
                  className="w-16 h-16 text-3xl bg-background-tertiary rounded-full flex items-center justify-center border-2 border-border"
                  initial={{ scale: 0, rotate: 180 }}
                  animate={{ scale: 1, rotate: 0 }}
                  transition={{ duration: 0.8, delay: 0.8, type: "spring", stiffness: 100 }}
                  style={{ borderColor: getAlignmentColor(assignment.alignment) }}
                >
                  {getAlignmentIcon(assignment.alignment)}
                </motion.div>
              </div>
              <motion.h3 
                className="text-2xl font-bold mb-2"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.8 }}
              >
                {getEnhancedRoleInfo(assignment.role.type).title}
              </motion.h3>
              <motion.p 
                className="text-text-secondary"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.9 }}
              >
                {getEnhancedRoleInfo(assignment.role.type).description}
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

                <motion.div 
                  data-role="detail"
                  className="px-4 py-3 bg-background-tertiary rounded-lg"
                  initial={{ opacity: 0, y: 20 }}
                  style={{ opacity: 0, y: 20 }}
                >
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px] block mb-2">ROLE OVERVIEW:</span>
                  <div className="text-text-secondary text-sm mb-3">{getEnhancedRoleInfo(assignment.role.type).description}</div>
                  <div className="text-text-secondary text-sm mb-3">{getEnhancedRoleInfo(assignment.role.type).abilities}</div>
                  <div className="text-xs font-bold text-text-muted uppercase tracking-[0.5px] mb-2">KEY POWERS:</div>
                  <div className="flex flex-wrap gap-1">
                    {getEnhancedRoleInfo(assignment.role.type).keyPowers.map((power, index) => (
                      <span key={index} className="text-xs bg-background-primary px-2 py-1 rounded text-text-primary">
                        {power}
                      </span>
                    ))}
                  </div>
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