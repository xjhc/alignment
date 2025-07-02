import { useState } from 'react';
import { useSessionContext } from '../../contexts/SessionContext';

interface Player {
  id: string;
  name: string;
  avatar?: string;
}

interface ReportModalProps {
  isOpen: boolean;
  onClose: () => void;
  players: Player[];
  onSubmitReport: (targetPlayerId: string, reason: string, description: string) => void;
}

const REPORT_REASONS = [
  { value: 'harassment', label: 'Harassment or Bullying', description: 'Targeted harassment, insults, or bullying behavior' },
  { value: 'inappropriate_language', label: 'Inappropriate Language', description: 'Offensive, vulgar, or inappropriate language' },
  { value: 'cheating', label: 'Cheating or Exploiting', description: 'Using exploits, cheats, or unfair advantages' },
  { value: 'griefing', label: 'Griefing or Trolling', description: 'Deliberately disrupting the game or ruining others\' experience' },
  { value: 'spam', label: 'Spam or Flooding', description: 'Excessive messaging or chat flooding' },
  { value: 'discrimination', label: 'Discrimination', description: 'Discriminatory language or behavior' },
  { value: 'other', label: 'Other', description: 'Other inappropriate behavior not listed above' }
];

export function ReportModal({ isOpen, onClose, players, onSubmitReport }: ReportModalProps) {
  const [selectedPlayer, setSelectedPlayer] = useState<string>('');
  const [selectedReason, setSelectedReason] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { appState } = useSessionContext();

  if (!isOpen) return null;

  // Filter out the current player from the list
  const eligiblePlayers = players.filter(p => p.id !== appState.playerId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPlayer || !selectedReason || !description.trim()) return;

    setIsSubmitting(true);
    try {
      await onSubmitReport(selectedPlayer, selectedReason, description.trim());
      
      // Reset form
      setSelectedPlayer('');
      setSelectedReason('');
      setDescription('');
      onClose();
    } catch (error) {
      console.error('Failed to submit report:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const selectedReasonData = REPORT_REASONS.find(r => r.value === selectedReason);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background-primary border border-border-primary rounded-lg p-6 w-full max-w-lg mx-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-text-primary">Report Player</h2>
          <button
            onClick={onClose}
            className="text-text-muted hover:text-text-primary text-xl"
          >
            ×
          </button>
        </div>

        <div className="mb-4 p-3 bg-warning/10 border border-warning/20 rounded">
          <div className="text-sm text-warning">
            ⚠️ <strong>Important:</strong> Reports are reviewed by moderators. Please only report genuine violations of community guidelines.
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Player Selection */}
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              Select Player to Report
            </label>
            <select
              value={selectedPlayer}
              onChange={(e) => setSelectedPlayer(e.target.value)}
              className="w-full p-2 bg-background-secondary border border-border-secondary rounded text-text-primary"
              required
            >
              <option value="">Choose a player...</option>
              {eligiblePlayers.map(player => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>

          {/* Reason Selection */}
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              Reason for Report
            </label>
            <div className="space-y-2">
              {REPORT_REASONS.map(reason => (
                <label
                  key={reason.value}
                  className={`flex items-start p-3 rounded border cursor-pointer transition-colors ${
                    selectedReason === reason.value
                      ? 'bg-danger/10 border-danger/20'
                      : 'bg-background-secondary border-border-secondary hover:border-border-primary'
                  }`}
                >
                  <input
                    type="radio"
                    name="reason"
                    value={reason.value}
                    checked={selectedReason === reason.value}
                    onChange={(e) => setSelectedReason(e.target.value)}
                    className="mt-1 mr-3"
                  />
                  <div>
                    <div className="font-medium text-text-primary">{reason.label}</div>
                    <div className="text-sm text-text-secondary">{reason.description}</div>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Description */}
          {selectedReason && (
            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                Detailed Description <span className="text-danger">*</span>
              </label>
              {selectedReasonData && (
                <div className="text-xs text-text-secondary mb-2">
                  Please provide specific details about the {selectedReasonData.label.toLowerCase()} incident.
                </div>
              )}
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Please describe what happened, when it occurred, and any other relevant details..."
                className="w-full p-3 bg-background-secondary border border-border-secondary rounded text-text-primary resize-none"
                rows={4}
                maxLength={1000}
                required
              />
              <div className="text-xs text-text-muted mt-1">
                {description.length}/1000 characters
              </div>
            </div>
          )}

          {/* Submit Button */}
          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-background-secondary text-text-secondary rounded hover:bg-background-tertiary transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!selectedPlayer || !selectedReason || !description.trim() || isSubmitting}
              className="flex-1 px-4 py-2 bg-danger text-white rounded hover:bg-danger/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isSubmitting ? 'Submitting...' : 'Submit Report'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}