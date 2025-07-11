import React from 'react';
import { useSessionContext } from '../../contexts/SessionContext';
import { Button } from '../ui';

export function SpectatorHUD() {
  const { onBackToLogin } = useSessionContext();

  return (
    <div className="flex flex-col h-full bg-background-secondary">
      {/* Header */}
      <div className="p-4 border-b border-border">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-text-primary">👁️ Spectator Mode</h2>
            <p className="text-sm text-text-muted">Observing game in progress</p>
          </div>
          <Button
            variant="secondary"
            size="sm"
            onClick={onBackToLogin}
            className="text-xs"
          >
            Leave Game
          </Button>
        </div>
      </div>

      {/* Spectator Info */}
      <div className="flex-1 p-4 space-y-4">
        <div className="p-3 bg-background-tertiary rounded border border-border">
          <h3 className="text-sm font-medium text-text-primary mb-2">📖 Spectator Rules</h3>
          <ul className="text-xs text-text-secondary space-y-1">
            <li>• You can observe all public game activity</li>
            <li>• Chat with other spectators in #spectators</li>
            <li>• No game actions or voting allowed</li>
            <li>• Private information is hidden</li>
          </ul>
        </div>

        <div className="p-3 bg-background-tertiary rounded border border-border">
          <h3 className="text-sm font-medium text-text-primary mb-2">🎮 Game Viewing</h3>
          <ul className="text-xs text-text-secondary space-y-1">
            <li>• Watch #war-room discussions</li>
            <li>• See public voting and phase changes</li>
            <li>• Observe player eliminations</li>
            <li>• View token counts and status updates</li>
          </ul>
        </div>

        <div className="p-3 bg-amber/10 border border-amber/20 rounded">
          <h3 className="text-sm font-medium text-amber mb-1">⚠️ Note</h3>
          <p className="text-xs text-amber/80">
            You're viewing a live game. Player roles, alignments, and private communications are hidden to maintain game integrity.
          </p>
        </div>
      </div>
    </div>
  );
}