import React from 'react';
import { ChatMessage, GameState } from '../../types';
import { Tooltip } from '../ui';

interface VoteResultMessageProps {
  message: ChatMessage;
  gameState: GameState;
  localPlayerId: string;
}

export const VoteResultMessage: React.FC<VoteResultMessageProps> = ({ message, gameState, localPlayerId }) => {
  const voteResult = message.metadata?.voteResult;
  
  if (!voteResult) {
    return null;
  }

  const { question, outcome, votes, tokenWeights, results, eliminatedPlayer } = voteResult;
  const yesVotes = results['GUILTY'] || results['YES'] || 0;
  const noVotes = results['INNOCENT'] || results['NO'] || 0;

  const getAlignmentDisplay = (alignment: string) => {
    switch (alignment) {
      case 'HUMAN':
        return <span className="px-2 py-1 text-xs font-bold text-white uppercase rounded-full bg-human">👤 Human</span>;
      case 'ALIGNED':
        return <span className="px-2 py-1 text-xs font-bold text-white uppercase rounded-full bg-aligned">🤖 Aligned</span>;
      case 'AI':
        return <span className="px-2 py-1 text-xs font-bold text-white uppercase rounded-full bg-ai">🤖 Original AI</span>;
      default:
        return <span className="px-2 py-1 text-xs font-bold text-white uppercase rounded-full bg-text-muted">❓ Unknown</span>;
    }
  };

  const renderVoteBlocks = (voteType: 'GUILTY' | 'YES' | 'INNOCENT' | 'NO') => {
    const relevantVoters = Object.entries(votes)
      .filter(([, vote]) => vote === voteType || (voteType === 'GUILTY' && vote === 'YES') || (voteType === 'INNOCENT' && vote === 'NO'))
      .map(([playerId]) => playerId);

    return relevantVoters.map((playerId) => {
      const tokenWeight = tokenWeights[playerId] || 0;
      const isLocalPlayer = playerId === localPlayerId;

      const block = (
        <div key={playerId} className={`bg-background-tertiary border border-border rounded-md p-2 text-center transition-transform duration-200 hover:-translate-y-0.5 ${isLocalPlayer ? 'bg-amber-400/20 border-amber-500' : ''}`}>
            <div className={`font-mono font-bold text-sm ${isLocalPlayer ? 'text-amber-500' : 'text-text-primary'}`}>{tokenWeight}</div>
            <div className="text-xs text-text-muted">Tokens</div>
        </div>
      );

      if (isLocalPlayer) {
        return (
          <Tooltip key={playerId} content="Your Vote (Only you can see this)" position="top">
            <div className="relative">
              {block}
              <span className="absolute -top-1 -right-1 text-xs">🔒</span>
            </div>
          </Tooltip>
        );
      }

      return block;
    });
  };

  return (
    <div className="bg-background-secondary border border-border rounded-lg p-4 my-2">
      <div className="font-bold text-info text-sm mb-2 uppercase tracking-wider">⚖️ Alignment Achieved</div>
      <div className="italic text-text-primary mb-4">{question}</div>
      
      <div className="space-y-3">
        {/* YES VOTES */}
        <div className="bg-background-primary p-3 rounded-md border border-border">
          <div className="flex items-center justify-between mb-2">
            <span className="font-bold text-success">✔️ YES</span>
            <span className="font-mono font-bold text-success">🪙 {yesVotes}</span>
          </div>
          <div className="flex items-center gap-2 flex-wrap">
            {renderVoteBlocks('GUILTY')}
          </div>
        </div>

        {/* NO VOTES */}
        <div className="bg-background-primary p-3 rounded-md border border-border">
          <div className="flex items-center justify-between mb-2">
            <span className="font-bold text-danger">❌ NO</span>
            <span className="font-mono font-bold text-danger">🪙 {noVotes}</span>
          </div>
          <div className="flex items-center gap-2 flex-wrap">
            {renderVoteBlocks('INNOCENT')}
          </div>
        </div>
      </div>

      <div className="mt-4 pt-4 border-t border-border text-center">
        <p className="font-bold text-text-primary">{outcome}</p>
        {eliminatedPlayer && (
          <div className="text-sm text-text-secondary mt-1">
            <span className="font-semibold">Revealed:</span> {eliminatedPlayer.role}{' '}
            {getAlignmentDisplay(eliminatedPlayer.alignment)}
          </div>
        )}
      </div>
    </div>
  );
};