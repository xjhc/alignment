import React from 'react';
import { ChatMessage, GameState } from '../../types';

interface VoteResultMessageProps {
  message: ChatMessage;
  gameState: GameState;
}

export const VoteResultMessage: React.FC<VoteResultMessageProps> = ({ message, gameState }) => {
  const voteResult = message.metadata?.voteResult;
  
  if (!voteResult) {
    return null;
  }

  const { question, outcome, votes, results, eliminatedPlayer } = voteResult;
  const yesVotes = results['GUILTY'] || results['YES'] || 0;
  const noVotes = results['INNOCENT'] || results['NO'] || 0;

  const getPlayerIcon = (playerId: string) => {
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

  const getAlignmentIcon = (alignment: string) => {
    switch (alignment) {
      case 'HUMAN': return '👤 HUMAN';
      case 'ALIGNED': return '🤖 ALIGNED AGENT';
      case 'AI': return '🤖 ORIGINAL AI';
      default: return '❓ UNKNOWN';
    }
  };

  const getAlignmentClass = (alignment: string) => {
    switch (alignment) {
      case 'HUMAN': return 'human';
      case 'ALIGNED': return 'aligned';
      case 'AI': return 'ai';
      default: return '';
    }
  };

  return (
    <div className="vote-result-system">
      <div className="vote-result-header">⚖️ ALIGNMENT ACHIEVED</div>
      <div className="vote-result-question">{question}</div>
      <div className="vote-result-details">
        <div className="vote-result-row yes">
          <span className="vote-option-label">✔️ YES</span>
          <span className="vote-total">🪙 {yesVotes}</span>
          <div className="blockchain-chain">
            {Object.entries(votes)
              .filter(([, vote]) => vote === 'GUILTY' || vote === 'YES')
              .map(([playerId]) => {
                const voter = gameState.players.find(p => p.id === playerId);
                if (!voter) return null;
                return (
                  <div key={playerId} className="vote-block">
                    <div className="block-header">
                      <span className="block-icon">{getPlayerIcon(playerId)}</span>
                      <span className="block-amount">{voter.tokens}</span>
                    </div>
                    <div className="block-hash">{voter.name}</div>
                  </div>
                );
              })}
          </div>
        </div>
        <div className="vote-result-row no">
          <span className="vote-option-label">❌ NO</span>
          <span className="vote-total">🪙 {noVotes}</span>
          <div className="blockchain-chain">
            {Object.entries(votes)
              .filter(([, vote]) => vote === 'INNOCENT' || vote === 'NO')
              .map(([playerId]) => {
                const voter = gameState.players.find(p => p.id === playerId);
                if (!voter) return null;
                return (
                  <div key={playerId} className="vote-block">
                    <div className="block-header">
                      <span className="block-icon">{getPlayerIcon(playerId)}</span>
                      <span className="block-amount">{voter.tokens}</span>
                    </div>
                    <div className="block-hash">{voter.name}</div>
                  </div>
                );
              })}
          </div>
        </div>
      </div>
      <div className="vote-outcome">
        <strong>{outcome}</strong><br/>
        {eliminatedPlayer && (
          <>
            <strong>REVEALED:</strong> {eliminatedPlayer.role}{' '}
            <span className={`outcome-identity ${getAlignmentClass(eliminatedPlayer.alignment)}`}>
              {getAlignmentIcon(eliminatedPlayer.alignment)}
            </span>
          </>
        )}
      </div>
    </div>
  );
};