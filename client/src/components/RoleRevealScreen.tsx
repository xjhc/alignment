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
      setTimeout(() => setShowDetails(true), 1500);
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
    <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
      <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
        LOEBIAN INC. // <span className="inline-block animate-pulse">EMERGENCY BRIDGE</span>
      </h1>
      
      <div className="flex flex-col gap-4 items-center w-96">
        <h2 className="text-amber drop-shadow-[0_0_8px_rgba(255,191,0,1)] mb-6">IDENTITY ASSIGNED</h2>
        
        <div className="w-full my-4 text-center animate-card-flip-in">
          <div className="w-20 h-20 text-5xl mx-auto mb-4 bg-background-secondary rounded-full flex items-center justify-center border-2 border-border">
            {getAlignmentIcon(assignment.alignment)}
          </div>
          <h3>{assignment.role.name}</h3>
          <p className="text-text-secondary mb-6">{assignment.role.description}</p>
        </div>
        
        <div className="w-full mb-4 text-left">
          <div className="flex justify-between items-start px-3 py-2 bg-background-secondary rounded mb-1 stagger-child">
            <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">INITIAL ALIGNMENT:</span>
            <span 
              className="font-bold text-text-primary"
              style={{ color: getAlignmentColor(assignment.alignment) }}
            >
              {assignment.alignment}
            </span>
          </div>
          
          {showDetails && (
            <>
              <div className="flex justify-between items-start px-3 py-2 bg-background-secondary rounded mb-1 stagger-child" style={{ animationDelay: '0.2s' }}>
                <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">ROLE TYPE:</span>
                <span className="font-semibold text-text-primary">{assignment.role.type}</span>
              </div>
              
              {assignment.role.ability && (
                <div className="flex justify-between items-start px-3 py-2 bg-background-secondary rounded mb-1 stagger-child" style={{ animationDelay: '0.4s' }}>
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">SPECIAL ABILITY:</span>
                  <span className="font-semibold text-text-primary">{assignment.role.ability.name}</span>
                </div>
              )}
              
              {assignment.personalKPI && (
                <div className="flex flex-col items-start px-3 py-2 bg-background-secondary rounded mb-1 stagger-child" style={{ animationDelay: '0.6s' }}>
                  <span className="text-xs font-bold text-text-muted uppercase tracking-[0.5px]">PERSONAL KPI:</span>
                  <div className="mt-2 w-full">
                    <div className="font-semibold text-text-primary mb-1">{assignment.personalKPI.type}</div>
                    <div className="text-text-secondary text-xs mb-2">{assignment.personalKPI.description}</div>
                    {assignment.personalKPI.reward && (
                      <div className="text-success text-xs">
                        <strong>Reward:</strong> {assignment.personalKPI.reward}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </div>
        
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
        
        {showDetails && (
          <button 
            className="w-full px-6 py-3 text-base font-semibold text-black bg-amber rounded transition-colors duration-200 cursor-pointer border-none hover:bg-amber-light disabled:opacity-50 disabled:cursor-not-allowed animation-pulse" 
            onClick={onEnterGame}
            style={{ animationDelay: '0.8s' }}
          >
            [ &gt; ENTER WAR ROOM ]
          </button>
        )}
      </div>
    </div>
  );
}