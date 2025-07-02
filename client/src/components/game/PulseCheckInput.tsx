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
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [response, setResponse] = useState('');

  const handleSubmit = async () => {
    if (!isSubmitting && response.trim()) {
      setIsSubmitting(true);
      try {
        await handlePulseCheck(response.trim());
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
        As {localPlayerName}, provide your response:
      </p>
      <div className="space-y-3">
        <textarea
          value={response}
          onChange={(e) => setResponse(e.target.value)}
          disabled={isSubmitting}
          maxLength={280}
          placeholder="Share your thoughts, feelings, and strategy..."
          className="w-full px-3 py-2 text-sm bg-gray-800 border border-gray-600 rounded-md text-gray-100 placeholder-gray-400 resize-none focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          rows={4}
        />
        <div className="flex items-center justify-between">
          <div className="text-xs text-gray-400">
            {response.length}/280 characters
          </div>
          <button
            onClick={handleSubmit}
            disabled={isSubmitting || !response.trim() || response.length > 280}
            className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-md transition-all duration-200 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-blue-600"
          >
            {isSubmitting ? 'Submitting...' : 'Submit Response'}
          </button>
        </div>
      </div>
    </div>
  );
};