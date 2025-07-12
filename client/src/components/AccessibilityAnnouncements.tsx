import React, { useEffect, useState } from 'react';
import { useGameContext } from '../contexts/GameContext';
import { useSessionContext } from '../contexts/SessionContext';

/**
 * Accessibility Announcements Component
 * 
 * This component provides ARIA live regions for screen readers to announce
 * critical game state changes without interrupting the user's current activity.
 * 
 * All announcements are made through aria-live regions that are visually hidden
 * but accessible to screen readers.
 */
export const AccessibilityAnnouncements: React.FC = () => {
  const { gameState, localPlayer } = useGameContext();
  const { appState } = useSessionContext();
  const [announcements, setAnnouncements] = useState<{
    polite: string;
    assertive: string;
  }>({
    polite: '',
    assertive: '',
  });

  // Track previous phase to detect changes
  const [previousPhase, setPreviousPhase] = useState<string | null>(null);
  const [previousPlayerCount, setPreviousPlayerCount] = useState<number>(0);

  useEffect(() => {
    if (!gameState?.phase) return;

    const currentPhase = gameState.phase.type;
    const currentPlayerCount = gameState.players?.filter(p => p.isAlive).length || 0;

    // Announce phase changes
    if (previousPhase && previousPhase !== currentPhase) {
      let phaseAnnouncement = '';
      
      switch (currentPhase) {
        case 'DISCUSSION':
          phaseAnnouncement = 'Discussion phase has begun. All players may now communicate.';
          break;
        case 'NOMINATION':
          phaseAnnouncement = 'Nomination phase has begun. Players may now nominate others for elimination.';
          break;
        case 'TRIAL':
          phaseAnnouncement = `Trial phase has begun. ${gameState.nominatedPlayer ? 'A player has been nominated and may defend themselves.' : ''}`;
          break;
        case 'VERDICT':
          phaseAnnouncement = 'Verdict phase has begun. All players must vote guilty or innocent.';
          break;
        case 'NIGHT':
          phaseAnnouncement = 'Night phase has begun. AI players may now take secret actions.';
          break;
        case 'SITREP':
          phaseAnnouncement = 'Situation report phase. Review the results of the previous phase.';
          break;
        default:
          phaseAnnouncement = `${currentPhase} phase has begun.`;
      }

      setAnnouncements(prev => ({
        ...prev,
        polite: phaseAnnouncement,
      }));
    }

    // Announce player eliminations
    if (previousPlayerCount > 0 && currentPlayerCount < previousPlayerCount) {
      const playersEliminated = previousPlayerCount - currentPlayerCount;
      setAnnouncements(prev => ({
        ...prev,
        assertive: `${playersEliminated} player${playersEliminated > 1 ? 's have' : ' has'} been eliminated.`,
      }));
    }

    setPreviousPhase(currentPhase);
    setPreviousPlayerCount(currentPlayerCount);
  }, [gameState?.phase?.type, gameState?.players, previousPhase, previousPlayerCount]);

  // Announce game end states
  useEffect(() => {
    if (appState.screen === 'GAME_OVER') {
      setAnnouncements(prev => ({
        ...prev,
        assertive: 'Game has ended. Moving to results screen.',
      }));
    }
  }, [appState.screen]);

  // Announce role assignments (only to the player themselves)
  useEffect(() => {
    if (localPlayer?.role && !previousPhase) { // Only on initial assignment
      const roleAnnouncement = `You have been assigned the role of ${localPlayer.role.name}. ${localPlayer.role.description}`;
      setAnnouncements(prev => ({
        ...prev,
        polite: roleAnnouncement,
      }));
    }
  }, [localPlayer?.role, previousPhase]);

  return (
    <div className="sr-only">
      {/* Polite announcements for general updates */}
      <div 
        aria-live="polite" 
        aria-relevant="additions text"
        className="sr-only"
      >
        {announcements.polite}
      </div>
      
      {/* Assertive announcements for critical updates */}
      <div 
        aria-live="assertive" 
        aria-relevant="additions text"
        className="sr-only"
      >
        {announcements.assertive}
      </div>
    </div>
  );
};