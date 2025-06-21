import { useState } from 'react';
import { useWebSocketEvent } from '../hooks/useWebSocket';
import { Role, PersonalKPI } from '../types';

interface RoleRevealScreenProps {
  onEnterGame: () => void;
}

interface RoleAssignment {
  role: Role;
  alignment: string;
  personalKPI: PersonalKPI;
}

export function RoleRevealScreen({ onEnterGame }: RoleRevealScreenProps) {
  const [roleAssignment, setRoleAssignment] = useState<RoleAssignment | null>(null);
  const [showDetails, setShowDetails] = useState(false);

  // Listen for role assignment
  useWebSocketEvent('ROLES_ASSIGNED', (payload: { your_role: RoleAssignment }) => {
    setRoleAssignment(payload.your_role);
    // Auto-show details after a brief delay for dramatic effect
    setTimeout(() => setShowDetails(true), 1500);
  });

  const getAlignmentColor = (alignment: string) => {
    return alignment === 'HUMAN' ? 'var(--color-human)' : 'var(--color-ai)';
  };

  const getAlignmentIcon = (alignment: string) => {
    return alignment === 'HUMAN' ? '🧑‍💼' : '🤖';
  };

  if (!roleAssignment) {
    return (
      <div className="launch-screen">
        <div className="launch-form">
          <h2>Assigning roles...</h2>
          <div className="loading-spinner">⏳</div>
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
            {getAlignmentIcon(roleAssignment.alignment)}
          </div>
          <h3>{roleAssignment.role.name}</h3>
          <p className="role-description">{roleAssignment.role.description}</p>
        </div>
        
        <div className="personnel-file">
          <div className="personnel-file-item">
            <span className="label">INITIAL ALIGNMENT:</span>
            <span 
              className="value alignment"
              style={{ color: getAlignmentColor(roleAssignment.alignment) }}
            >
              {roleAssignment.alignment}
            </span>
          </div>
          
          {showDetails && (
            <>
              <div className="personnel-file-item">
                <span className="label">ROLE TYPE:</span>
                <span className="value">{roleAssignment.role.type}</span>
              </div>
              
              {roleAssignment.role.ability && (
                <div className="personnel-file-item">
                  <span className="label">SPECIAL ABILITY:</span>
                  <span className="value">{roleAssignment.role.ability.name}</span>
                </div>
              )}
              
              <div className="personnel-file-item kpi">
                <span className="label">PERSONAL KPI:</span>
                <div className="kpi-details">
                  <div className="kpi-type">{roleAssignment.personalKPI.type}</div>
                  <div className="kpi-description">{roleAssignment.personalKPI.description}</div>
                  <div className="kpi-reward">
                    <strong>Reward:</strong> {roleAssignment.personalKPI.reward}
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
        
        {roleAssignment.alignment === 'HUMAN' ? (
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