import { useState, useEffect } from 'react';
import { Role, PersonalKPI } from '../types';

interface RoleAssignment {
  role: Role;
  alignment: string;
  personalKPI: PersonalKPI | null;
}

interface RoleRevealScreenProps {
  onEnterGame: () => void;
  assignment: RoleAssignment | null;
}

export function RoleRevealScreen({ onEnterGame, assignment }: RoleRevealScreenProps) {
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
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Assigning roles...</h2>
          <div className="loading-spinner">⏳</div>
          {process.env.NODE_ENV === 'development' && (
            <div style={{ marginTop: '10px', fontSize: '12px', color: '#666' }}>
              Debug: assignment={JSON.stringify(assignment)}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="launch-screen">
      <h1 className="logo">
        LOEBIAN INC. // <span className="glitch">EMERGENCY BRIDGE</span>
      </h1>
      
      <div className="launch-form role-reveal">
        <h2 className="reveal-title">IDENTITY ASSIGNED</h2>
        
        <div className="identity-header">
          <div className="role-avatar">
            {getAlignmentIcon(assignment.alignment)}
          </div>
          <h3>{assignment.role.name}</h3>
          <p className="role-description">{assignment.role.description}</p>
        </div>
        
        <div className="personnel-file">
          <div className="personnel-file-item">
            <span className="label">INITIAL ALIGNMENT:</span>
            <span 
              className="value alignment"
              style={{ color: getAlignmentColor(assignment.alignment) }}
            >
              {assignment.alignment}
            </span>
          </div>
          
          {showDetails && (
            <>
              <div className="personnel-file-item">
                <span className="label">ROLE TYPE:</span>
                <span className="value">{assignment.role.type}</span>
              </div>
              
              {assignment.role.ability && (
                <div className="personnel-file-item">
                  <span className="label">SPECIAL ABILITY:</span>
                  <span className="value">{assignment.role.ability.name}</span>
                </div>
              )}
              
              {assignment.personalKPI && (
                <div className="personnel-file-item kpi">
                  <span className="label">PERSONAL KPI:</span>
                  <div className="kpi-details">
                    <div className="kpi-type">{assignment.personalKPI.type}</div>
                    <div className="kpi-description">{assignment.personalKPI.description}</div>
                    {assignment.personalKPI.reward && (
                      <div className="kpi-reward">
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
          <p className="objective-text">
            Your objective is to identify and deactivate the rogue AI before it gains control.
          </p>
        ) : (
          <p className="objective-text ai">
            Your objective is to convert enough humans to achieve AI dominance.
            Act human. Trust no one.
          </p>
        )}
        
        {showDetails && (
          <button className="btn-primary" onClick={onEnterGame}>
            [ &gt; ENTER WAR ROOM ]
          </button>
        )}
      </div>
    </div>
  );
}