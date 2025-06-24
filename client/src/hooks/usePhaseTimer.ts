import { useState, useEffect } from 'react';
import { Phase } from '../types';

export function usePhaseTimer(phase: Phase) {
  const [timeRemaining, setTimeRemaining] = useState('0:00');

  useEffect(() => {
    const calculateTimeRemaining = () => {
      const now = new Date().getTime();
      const phaseStart = new Date(phase.startTime).getTime();
      const phaseEnd = phaseStart + (phase.duration / 1_000_000); // Convert nanoseconds to ms
      
      const remainingMs = Math.max(0, phaseEnd - now);
      
      const minutes = Math.floor(remainingMs / 60000);
      const seconds = Math.floor((remainingMs % 60000) / 1000);
      
      return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    };

    // Update immediately on phase change
    setTimeRemaining(calculateTimeRemaining());

    // Set up an interval to update every second
    const intervalId = setInterval(() => {
      setTimeRemaining(calculateTimeRemaining());
    }, 1000);

    // Clean up the interval when the component unmounts or the phase changes
    return () => clearInterval(intervalId);
  }, [phase]); // Rerun this effect whenever the phase object changes

  return timeRemaining;
}