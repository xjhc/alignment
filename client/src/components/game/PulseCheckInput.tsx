import React from 'react';
import styles from './PulseCheckInput.module.css';

interface PulseCheckInputProps {
  handlePulseCheck: (response: string) => Promise<void>;
  localPlayerName: string;
}

export const PulseCheckInput: React.FC<PulseCheckInputProps> = ({ handlePulseCheck, localPlayerName }) => {
  const responses = ["Nominal", "Elevated", "Critical"];

  return (
    <div className={styles.container}>
      <p className={styles.prompt}>
        As {localPlayerName}, your pulse check response is:
      </p>
      <div className={styles.buttonGroup}>
        {responses.map((response) => (
          <button
            key={response}
            className={styles.responseButton}
            onClick={() => handlePulseCheck(response)}
          >
            {response}
          </button>
        ))}
      </div>
    </div>
  );
};