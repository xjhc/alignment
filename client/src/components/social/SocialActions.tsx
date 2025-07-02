import { useState } from 'react';
import { KudosModal } from './KudosModal';
import { ReportModal } from './ReportModal';

interface Player {
  id: string;
  name: string;
  avatar?: string;
}

interface SocialActionsProps {
  players: Player[];
  onGiveKudos: (targetPlayerId: string, reason: string) => Promise<void>;
  onSubmitReport: (targetPlayerId: string, reason: string, description: string) => Promise<void>;
  className?: string;
}

export function SocialActions({ players, onGiveKudos, onSubmitReport, className = '' }: SocialActionsProps) {
  const [showKudosModal, setShowKudosModal] = useState(false);
  const [showReportModal, setShowReportModal] = useState(false);

  return (
    <div className={`flex gap-2 ${className}`}>
      <button
        onClick={() => setShowKudosModal(true)}
        className="px-3 py-2 bg-success/10 text-success border border-success/20 rounded hover:bg-success/20 transition-colors text-sm font-medium"
        title="Give kudos to a player"
      >
        👏 Kudos
      </button>
      
      <button
        onClick={() => setShowReportModal(true)}
        className="px-3 py-2 bg-danger/10 text-danger border border-danger/20 rounded hover:bg-danger/20 transition-colors text-sm font-medium"
        title="Report inappropriate behavior"
      >
        🚨 Report
      </button>

      <KudosModal
        isOpen={showKudosModal}
        onClose={() => setShowKudosModal(false)}
        players={players}
        onGiveKudos={onGiveKudos}
      />

      <ReportModal
        isOpen={showReportModal}
        onClose={() => setShowReportModal(false)}
        players={players}
        onSubmitReport={onSubmitReport}
      />
    </div>
  );
}