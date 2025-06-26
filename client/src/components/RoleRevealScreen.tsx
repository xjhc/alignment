import { useState, useEffect } from 'react';
import { useSessionContext } from '../contexts/SessionContext';

interface RoleRevealScreenProps {
  onEnterGame: () => void;
}

export function RoleRevealScreen({ onEnterGame }: RoleRevealScreenProps) {
  const { roleAssignment: assignment } = useSessionContext();
  const [showDetails, setShowDetails] = useState(false);

  useEffect(() => {
    if (assignment) {
      // Auto-show details after a brief delay for dramatic effect
      setTimeout(() => setShowDetails(true), 750);
    }
  }, [assignment]);

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
    <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary relative">
      {/* Glowing background effect */}
      <div 
        className="absolute inset-0 opacity-10 animation-pulse"
        style={{
          background: `radial-gradient(circle at center, ${getAlignmentColor(assignment.alignment)} 0%, transparent 70%)`
        }}
      />
      
      <h1 className="font-mono text-3xl font-semibold tracking-[2px] animation-fade-in">
        LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
      </h1>
      
      <div className="flex flex-col gap-4 items-center w-96 relative z-10">
        <h2 className="text-amber drop-shadow-[0_0_8px_rgba(255,191,0,1)] mb-6 animation-fade-in">IDENTITY ASSIGNED</h2>
        
        {/* Digital Card Container */}
        <div 
          className="w-full bg-background-secondary rounded-xl border border-border p-8 text-center relative overflow-hidden animation-scale-in"
          style={{ animationDelay: '250ms' }}
        >
          {/* Card glow effect */}
          <div 
            className="absolute inset-0 opacity-20 animation-pulse"
            style={{
              background: `linear-gradient(135deg, ${getAlignmentColor(assignment.alignment)}22, transparent)`
            }}
          />
          
          <div className="relative z-10">
            {/* Avatar and Role Info */}
            <div className="mb-6">
              <div className="w-20 h-20 text-5xl mx-auto mb-4 bg-background-tertiary rounded-full flex items-center justify-center border-2 border-border animation-scale-in" style={{ animationDelay: '250ms' }}>
                {getAlignmentIcon(assignment.alignment)}
              </div>
              <h3 className="text-2xl font-bold mb-2 animation-fade-in" style={{ animationDelay: '250ms' }}>{assignment.role.name}</h3>
              <p className="text-text-secondary animation-fade-in" style={{ animationDelay: '250ms' }}>{assignment.role.description}</p>
            </div>
            
            {/* Role Details */}
            {showDetails && (
              <div className="w-full mb-6 text-left space-y-3 animation-fade-in" style={{ animationDelay: '750ms' }}>
                <div className="flex justify-between items-start px-4 py-3 bg-background-tertiary rounded-lg">
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ALIGNMENT:</span>
                  <span 
                    className="font-bold text-lg"
                    style={{ color: getAlignmentColor(assignment.alignment) }}
                  >
                    {assignment.alignment}
                  </span>
                </div>
                
                <div className="flex justify-between items-start px-4 py-3 bg-background-tertiary rounded-lg">
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ROLE TYPE:</span>
                  <span className="font-semibold text-text-primary">{assignment.role.type}</span>
                </div>
                
                {assignment.role.ability && (
                  <div className="flex justify-between items-start px-4 py-3 bg-background-tertiary rounded-lg">
                    <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ABILITY:</span>
                    <span className="font-semibold text-text-primary">{assignment.role.ability.name}</span>
                  </div>
                )}
                
                {assignment.personalKPI && (
                  <div className="px-4 py-3 bg-background-tertiary rounded-lg">
                    <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px] block mb-2">PERSONAL KPI:</span>
                    <div className="font-semibold text-text-primary mb-1">{assignment.personalKPI.type}</div>
                    <div className="text-text-secondary text-sm mb-2">{assignment.personalKPI.description}</div>
                    {assignment.personalKPI.reward && (
                      <div className="text-success text-sm">
                        <strong>Reward:</strong> {assignment.personalKPI.reward}
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
        
        {/* Objective Text */}
        {showDetails && (
          <div className="animation-fade-in" style={{ animationDelay: '750ms' }}>
            {assignment.alignment === 'HUMAN' ? (
              <p className="text-text-secondary text-sm text-center mb-6">
                Your objective is to identify and deactivate the rogue AI before it gains control.
              </p>
            ) : (
              <p className="text-magenta font-medium text-sm text-center mb-6">
                Your objective is to convert enough humans to achieve AI dominance.
                Act human. Trust no one.
              </p>
            )}
          </div>
        )}
        
        {/* Enter Button */}
        {showDetails && (
          <button 
            className="w-full px-6 py-3 text-base font-semibold text-black bg-amber rounded-lg transition-colors duration-200 cursor-pointer border-none hover:bg-amber-light disabled:opacity-50 disabled:cursor-not-allowed animation-scale-in" 
            onClick={onEnterGame}
            style={{ animationDelay: '1200ms' }}
          >
            [ &gt; ENTER WAR ROOM ]
          </button>
        )}
      </div>
    </div>
  );
}