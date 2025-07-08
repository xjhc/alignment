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

  // Helper function to get player avatar emoji
  const getPlayerAvatar = (playerId: string) => {
    const player = gameState.players.find(p => p.id === playerId);
    if (!player) return '👤';
    
    switch (player.jobTitle) {
      case 'CISO': return '👤';
      case 'Systems': return '🧑‍💻';
      case 'Ethics': return '🕵️';
      case 'CTO': return '🤖';
      case 'COO': return '🧑‍🚀';
      case 'CFO': return '👩‍🔬';
      default: return '👤';
    }
  };

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

  // Helper function to render blockchain vote blocks (matching VoteUI structure)
  const renderVoteBlocks = (voteOption: 'GUILTY' | 'YES' | 'INNOCENT' | 'NO') => {
    if (!votes) return null;
    
    return Object.entries(votes)
      .filter(([, vote]) => vote === voteOption)
      .map(([playerId]) => {
        const tokenWeight = tokenWeights[playerId] || 0;
        const isMyVote = playerId === localPlayerId;
        
        return (
          <div
            key={playerId}
            className={`vote-block ${isMyVote ? 'my-vote' : ''}`}
          >
            <div className="block-header">
              <span className="block-icon">{getPlayerAvatar(playerId)}</span>
              <span className="block-amount">{tokenWeight}</span>
            </div>
            <div className="block-hash">
              {isMyVote ? (
                <div className="flex items-center gap-1">
                  <span>(YOU)</span>
                  <Tooltip content="Only you can see this" position="top">
                    <span className="text-xs">🔒</span>
                  </Tooltip>
                </div>
              ) : (
                '######'
              )}
            </div>
          </div>
        );
      });
  };

  return (
    <div className="bg-background-secondary border border-border rounded-lg p-4 my-2">
      <div className="font-bold text-info text-sm mb-2 uppercase tracking-wider">⚖️ Vote Complete</div>
      <div className="italic text-text-primary mb-4">{question || 'Final vote results'}</div>
      
      <div className="verdict-poll">
        {/* YES Vote Row */}
        <div className="vote-option yes">
          <span className="option-label">✔️ YES</span>
          <span className="vote-tally">🪙 {yesVotes}</span>
          <div className="blockchain-chain">
            {renderVoteBlocks('GUILTY')}
          </div>
        </div>

        {/* NO Vote Row */}
        <div className="vote-option no">
          <span className="option-label">❌ NO</span>
          <span className="vote-tally">🪙 {noVotes}</span>
          <div className="blockchain-chain">
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