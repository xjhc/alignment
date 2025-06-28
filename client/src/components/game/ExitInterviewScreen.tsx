import React, { useState } from 'react';
import { useGameActions } from '../../hooks/useGameActions';
import { Button, Input } from '../ui';
import { Card } from '../ui/Card';

interface ExitInterviewScreenProps {
  onComplete: () => void;
}

export const ExitInterviewScreen: React.FC<ExitInterviewScreenProps> = ({ onComplete }) => {
  const [partingShot, setPartingShot] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { handleSubmitPartingShot } = useGameActions();

  const handleSubmit = async () => {
    if (!partingShot.trim()) {
      return;
    }

    setIsSubmitting(true);
    try {
      await handleSubmitPartingShot(partingShot.trim());
      onComplete();
    } catch (error) {
      console.error('Error submitting parting shot:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/90 flex items-center justify-center z-50 p-4">
      <Card className="max-w-md w-full bg-background-primary border-danger-primary">
        <div className="p-6 space-y-6">
          {/* Header */}
          <div className="text-center space-y-2">
            <div className="text-4xl">💀</div>
            <h2 className="text-xl font-bold text-text-primary">Exit Interview</h2>
            <p className="text-sm text-text-muted">
              You have been eliminated from the game. Leave your mark with a final message.
            </p>
          </div>

          {/* Parting Shot Input */}
          <div className="space-y-3">
            <label className="block text-sm font-medium text-text-primary">
              Parting Shot
              <span className="text-text-muted text-xs ml-1">(max 50 characters)</span>
            </label>
            
            <Input
              value={partingShot}
              onChange={(e) => setPartingShot(e.target.value.slice(0, 50))}
              placeholder="Your final words..."
              className="w-full"
              maxLength={50}
              autoFocus
            />
            
            <div className="text-xs text-text-muted text-right">
              {partingShot.length}/50 characters
            </div>
          </div>

          {/* Submit Button */}
          <div className="flex justify-center pt-4">
            <Button
              onClick={handleSubmit}
              disabled={!partingShot.trim() || isSubmitting}
              className="px-8 py-2"
            >
              {isSubmitting ? 'Submitting...' : 'Submit Final Message'}
            </Button>
          </div>

          {/* Note */}
          <div className="text-xs text-text-muted text-center space-y-1">
            <p>This message will be visible to all players for the rest of the game.</p>
            <p>Choose your words carefully - they cannot be changed.</p>
          </div>
        </div>
      </Card>
    </div>
  );
};