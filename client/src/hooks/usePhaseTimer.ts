import { useState, useEffect } from "react";
import { Phase } from "../types";

export function usePhaseTimer(phase: Phase) {
  const [timeRemaining, setTimeRemaining] = useState("0:00");

  useEffect(() => {
    // Guard against invalid or missing phase data.
    // If the phase is not valid, we can't calculate a time, so we reset to 0:00 and do nothing.
    if (!phase || !phase.startTime || phase.duration <= 0) {
      setTimeRemaining("0:00");
      return; // Exit the effect early.
    }

    const calculateTimeRemaining = () => {
      const now = new Date().getTime();
      const phaseStart = new Date(phase.startTime).getTime();
      // The duration from the server is in nanoseconds. Convert to milliseconds.
      const phaseDurationMs = phase.duration / 1_000_000;
      const phaseEnd = phaseStart + phaseDurationMs;

      const remainingMs = Math.max(0, phaseEnd - now);

      const minutes = Math.floor(remainingMs / 60000);
      const seconds = Math.floor((remainingMs % 60000) / 1000);

      return `${minutes}:${seconds.toString().padStart(2, "0")}`;
    };

    // Update immediately on phase change
    setTimeRemaining(calculateTimeRemaining());

    const intervalId = setInterval(() => {
      setTimeRemaining(calculateTimeRemaining());
    }, 1000);

    return () => clearInterval(intervalId);
  }, [phase]);

  return timeRemaining;
}
