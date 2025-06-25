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
  const [response, setResponse] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (response.trim() && !isSubmitting) {
      setIsSubmitting(true);
      try {
        await handlePulseCheck(response.trim());
      } catch (error) {
        console.error('Failed to submit pulse check:', error);
        setIsSubmitting(false);
      }
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <div className="p-4 bg-blue-900/20 border-t border-blue-500/30 animate-[fadeIn_0.3s_ease]">
      <div className="text-blue-400 font-mono font-bold text-sm mb-2">💭 PULSE CHECK</div>
      <div className="text-text-primary font-medium mb-3 italic">"{question}"</div>
      <p className="text-sm text-gray-400 mb-3">
        As {localPlayerName}, your response:
      </p>
      <div className="flex gap-2">
        <input
          type="text"
          value={response}
          onChange={(e) => setResponse(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder="Enter your response (max 200 characters)..."
          maxLength={200}
          className="flex-1 px-3 py-2 text-sm bg-gray-900 border border-gray-600 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          disabled={isSubmitting}
          autoFocus
        />
        <button
          onClick={handleSubmit}
          disabled={!response.trim() || isSubmitting}
          className="px-4 py-2 text-sm font-semibold border border-blue-600 rounded-md bg-blue-900 text-blue-100 cursor-pointer transition-all duration-200 hover:bg-blue-700 hover:border-blue-400 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-blue-900"
        >
          {isSubmitting ? 'Submitting...' : 'Submit'}
        </button>
      </div>
      <div className="text-xs text-gray-500 mt-2">
        {response.length}/200 characters • Press Enter to submit
      </div>
    </div>
  );
};