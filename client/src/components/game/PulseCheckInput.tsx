import React, { useState } from 'react';

interface PulseCheckInputProps {
  handlePulseCheck: (response: string) => Promise<void>;
  localPlayerName: string;
  question?: string;
}

export const PulseCheckInput: React.FC<PulseCheckInputProps> = ({ 
  handlePulseCheck, 
  localPlayerName,
  question = "What is your immediate response to the current crisis?"
}) => {
  // Pulse check responses are now handled via button selection
  const [isSubmitting, setIsSubmitting] = useState(false);

  const responses = [
    { id: 'confident', label: 'Confident', value: 'I feel confident about our current situation and next steps.' },
    { id: 'concerned', label: 'Concerned', value: 'I have concerns about how things are progressing right now.' },
    { id: 'suspicious', label: 'Suspicious', value: 'Something feels off and I think we need to be more careful.' }
  ];

  const handleSubmit = async (responseValue: string) => {
    if (!isSubmitting) {
      setIsSubmitting(true);
      try {
        await handlePulseCheck(responseValue);
      } catch (error) {
        console.error('Failed to submit pulse check:', error);
        setIsSubmitting(false);
      }
    }
  };

  return (
    <div className="p-4 bg-blue-900/20 border-t border-blue-500/30 animate-[fadeIn_0.3s_ease]">
      <div className="text-blue-400 font-mono font-bold text-sm mb-2">💭 PULSE CHECK</div>
      <div className="text-text-primary font-medium mb-3 italic">"{question}"</div>
      <p className="text-sm text-gray-400 mb-3">
        As {localPlayerName}, choose your response:
      </p>
      <div className="flex flex-col gap-2">
        {responses.map((response) => (
          <button
            key={response.id}
            onClick={() => handleSubmit(response.value)}
            disabled={isSubmitting}
            className="px-4 py-3 text-sm font-medium border border-gray-600 rounded-md bg-gray-800 text-gray-100 transition-all duration-200 hover:bg-gray-700 hover:border-gray-500 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-gray-800 text-left"
          >
            <div className="font-semibold text-blue-400">{response.label}</div>
            <div className="text-xs text-gray-400 mt-1">{response.value}</div>
          </button>
        ))}
      </div>
      {isSubmitting && (
        <div className="text-xs text-blue-400 mt-2 font-mono">
          Submitting response...
        </div>
      )}
    </div>
  );
};