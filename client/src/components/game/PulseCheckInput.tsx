import React from 'react';

interface PulseCheckInputProps {
  handlePulseCheck: (response: string) => Promise<void>;
  localPlayerName: string;
}

export const PulseCheckInput: React.FC<PulseCheckInputProps> = ({ handlePulseCheck, localPlayerName }) => {
  const responses = ["Nominal", "Elevated", "Critical"];

  return (
    <div className="p-3 px-4 bg-gray-800 border-t border-gray-600 animate-[fadeIn_0.3s_ease] text-center">
      <p className="text-sm text-gray-400 mb-3">
        As {localPlayerName}, your pulse check response is:
      </p>
      <div className="flex justify-center gap-2">
        {responses.map((response) => (
          <button
            key={response}
            className="px-4 py-2 text-sm font-semibold border border-gray-600 rounded-md bg-gray-900 text-gray-100 cursor-pointer transition-all duration-200 hover:bg-gray-700 hover:border-amber-500 hover:-translate-y-0.5"
            onClick={() => handlePulseCheck(response)}
          >
            {response}
          </button>
        ))}
      </div>
    </div>
  );
};