import { useState, useEffect } from 'react';
import styles from './RoleRevealScreen.module.css';
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
      <div className={styles.launchScreen}>
        <div className={styles.launchForm}>
          <h2>Assigning roles...</h2>
          <div className="loading-spinner large"></div>
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
    <div className={styles.launchScreen}>
      <h1 className={styles.logo}>
        LOEBIAN INC. // <span className={styles.glitch}>EMERGENCY BRIDGE</span>
      </h1>
      
      <div className={`${styles.launchForm} ${styles.roleReveal}`}>
        <h2 className={styles.revealTitle}>IDENTITY ASSIGNED</h2>
        
        <div className={`${styles.identityHeader} animate-card-flip-in`}>
          <div className={styles.roleAvatar}>
            {getAlignmentIcon(assignment.alignment)}
          </div>
          <h3>{assignment.role.name}</h3>
          <p className={styles.roleDescription}>{assignment.role.description}</p>
        </div>
        
        <div className={styles.personnelFile}>
          <div className={`${styles.personnelFileItem} animate-stagger-reveal`}>
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
              <div className={`${styles.personnelFileItem} animate-stagger-reveal`} style={{ animationDelay: '0.2s' }}>
                <span className="label">ROLE TYPE:</span>
                <span className="value">{assignment.role.type}</span>
              </div>
              
              {assignment.role.ability && (
                <div className={`${styles.personnelFileItem} animate-stagger-reveal`} style={{ animationDelay: '0.4s' }}>
                  <span className="label">SPECIAL ABILITY:</span>
                  <span className="value">{assignment.role.ability.name}</span>
                </div>
              )}
              
              {assignment.personalKPI && (
                <div className={`${styles.personnelFileItem} ${styles.kpi} animate-stagger-reveal`} style={{ animationDelay: '0.6s' }}>
                  <span className="label">PERSONAL KPI:</span>
                  <div className={styles.kpiDetails}>
                    <div className={styles.kpiType}>{assignment.personalKPI.type}</div>
                    <div className={styles.kpiDescription}>{assignment.personalKPI.description}</div>
                    {assignment.personalKPI.reward && (
                      <div className={styles.kpiReward}>
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
          <p className={styles.objectiveText}>
            Your objective is to identify and deactivate the rogue AI before it gains control.
          </p>
        ) : (
          <p className={`${styles.objectiveText} ${styles.ai}`}>
            Your objective is to convert enough humans to achieve AI dominance.
            Act human. Trust no one.
          </p>
        )}
        
        {showDetails && (
          <button 
            className={`${styles.btnPrimary} animate-bounce`} 
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