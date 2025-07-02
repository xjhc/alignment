import { useState } from 'react';
import { useSessionContext } from '../../contexts/SessionContext';

interface Player {
  id: string;
  name: string;
  avatar?: string;
}

interface KudosModalProps {
  isOpen: boolean;
  onClose: () => void;
  players: Player[];
  onGiveKudos: (targetPlayerId: string, reason: string) => void;
}

const KUDOS_REASONS = [
  { value: 'great_teamwork', label: 'Great Teamwork', emoji: '🤝' },
  { value: 'excellent_strategy', label: 'Excellent Strategy', emoji: '🧠' },
  { value: 'good_communication', label: 'Good Communication', emoji: '💬' },
  { value: 'fair_play', label: 'Fair Play', emoji: '⚖️' },
  { value: 'fun_to_play_with', label: 'Fun to Play With', emoji: '😄' },
  { value: 'creative_thinking', label: 'Creative Thinking', emoji: '💡' },
  { value: 'helpful_attitude', label: 'Helpful Attitude', emoji: '🙋' },
  { value: 'good_sportsmanship', label: 'Good Sportsmanship', emoji: '🏆' }
];

export function KudosModal({ isOpen, onClose, players, onGiveKudos }: KudosModalProps) {
  const [selectedPlayer, setSelectedPlayer] = useState<string>('');
  const [selectedReason, setSelectedReason] = useState<string>('');
  const [customMessage, setCustomMessage] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { appState } = useSessionContext();

  if (!isOpen) return null;

  // Filter out the current player from the list
  const eligiblePlayers = players.filter(p => p.id !== appState.playerId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPlayer || !selectedReason) return;

    setIsSubmitting(true);
    try {
      const reasonData = KUDOS_REASONS.find(r => r.value === selectedReason);
      const fullReason = customMessage 
        ? `${reasonData?.label}: ${customMessage}` 
        : reasonData?.label || selectedReason;
      
      await onGiveKudos(selectedPlayer, fullReason);
      
      // Reset form
      setSelectedPlayer('');
      setSelectedReason('');
      setCustomMessage('');
      onClose();
    } catch (error) {
      console.error('Failed to give kudos:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background-primary border border-border-primary rounded-lg p-6 w-full max-w-md mx-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-text-primary">Give Kudos</h2>
          <button
            onClick={onClose}
            className="text-text-muted hover:text-text-primary text-xl"
          >
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Player Selection */}
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              Select Player
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
              Reason for Kudos
            </label>
            <div className="grid grid-cols-2 gap-2">
              {KUDOS_REASONS.map(reason => (
                <button
                  key={reason.value}
                  type="button"
                  onClick={() => setSelectedReason(reason.value)}
                  className={`p-2 text-left rounded border transition-colors ${
                    selectedReason === reason.value
                      ? 'bg-primary/20 border-primary text-primary'
                      : 'bg-background-secondary border-border-secondary text-text-secondary hover:border-border-primary'
                  }`}
                >
                  <div className="text-sm">
                    {reason.emoji} {reason.label}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Optional Custom Message */}
          {selectedReason && (
            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                Additional Message (Optional)
              </label>
              <textarea
                value={customMessage}
                onChange={(e) => setCustomMessage(e.target.value)}
                placeholder="Add a personal note..."
                className="w-full p-2 bg-background-secondary border border-border-secondary rounded text-text-primary resize-none"
                rows={3}
                maxLength={200}
              />
              <div className="text-xs text-text-muted mt-1">
                {customMessage.length}/200 characters
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
              disabled={!selectedPlayer || !selectedReason || isSubmitting}
              className="flex-1 px-4 py-2 bg-success text-white rounded hover:bg-success/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isSubmitting ? 'Sending...' : 'Give Kudos 👏'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}