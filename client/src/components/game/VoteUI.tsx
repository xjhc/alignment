import React from 'react';
import { useGameContext } from '../../contexts/GameContext';
import { useGameActions } from '../../hooks/useGameActions';
import { useSound } from '../../hooks/useSound';
import { Tooltip } from '../ui/Tooltip';
import { Button } from '../ui';

interface VoteUIProps {
  // No props needed - everything comes from context
}

export const VoteUI: React.FC<VoteUIProps> = () => {
  const { gameState, localPlayer } = useGameContext();
  const {
    selectedNominee,
    setSelectedNominee,
    setSelectedVote,
    handleNominate,
    handleVote,
  } = useGameActions();
  const { playSound } = useSound();
  
  if (!localPlayer) return null;
  const alivePlayers = gameState.players.filter(p => p.isAlive && p.id !== localPlayer.id);
  
  if (gameState.phase.type === 'NOMINATION') {
    return (
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
        <div className="mb-1.5">
          <h3 className="text-sm font-bold text-gray-100 normal-case tracking-normal text-left p-0 bg-transparent m-0 mb-3">Who should we deactivate?</h3>
        </div>
        <div className="flex flex-wrap gap-1 justify-start">
          {alivePlayers.map((player) => {
            const isSelected = selectedNominee === player.id;
            const playerVotes = gameState.voteState?.votes ? 
              Object.values(gameState.voteState.votes).filter(vote => vote === player.id).length : 0;
            
            return (
              <Tooltip
                key={player.id}
                content={`${player.name} (${playerVotes} vote${playerVotes !== 1 ? 's' : ''})`}
              >
                <button
                  className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-gray-600 bg-gray-800 transition-all duration-150 cursor-pointer min-w-0 font-inherit hover:bg-gray-700 hover:-translate-y-0.5 ${
                    isSelected ? 'bg-amber-500/10 border-amber-500' : ''
                  } ${
                    player.alignment === 'ALIGNED' ? 'bg-cyan-500/5 border-cyan-600' : ''
                  }`}
                  onClick={() => {
                    playSound('vote');
                    setSelectedNominee(player.id);
                    handleNominate();
                  }}
                  onMouseDown={(e) => {
                    e.currentTarget.classList.add('animation-scale-in');
                    setTimeout(() => e.currentTarget.classList.remove('animation-scale-in'), 150);
                  }}
              >
                <span className="text-sm leading-none flex-shrink-0">
                  {player.jobTitle === 'CISO' ? '👤' :
                   player.jobTitle === 'Systems' ? '🧑‍💻' :
                   player.jobTitle === 'Ethics' ? '🕵️' :
                   player.jobTitle === 'CTO' ? '🤖' :
                   player.jobTitle === 'COO' ? '🧑‍🚀' :
                   player.jobTitle === 'CFO' ? '👩‍🔬' : '👤'}
                </span>
                <span className={`font-medium text-xs text-gray-100 flex-shrink-0 ${
                  player.alignment === 'ALIGNED' ? 'text-cyan-600 animate-[glitch_1.5s_infinite]' : ''
                }`}>
                  {player.name}
                </span>
                <span className={`font-mono font-bold text-gray-400 text-xs ml-auto ${
                  isSelected ? 'text-amber-500' : ''
                }`}>🪙 {playerVotes}</span>
              </button>
              </Tooltip>
            );
          })}
        </div>
      </div>
    );
  }

  if (gameState.phase.type === 'VERDICT') {
    const nominatedPlayer = gameState.players.find(p => p.id === gameState.nominatedPlayer);
    if (!nominatedPlayer) return null;

    const yesVotes = gameState.voteState?.results?.['GUILTY'] || 0;
    const noVotes = gameState.voteState?.results?.['INNOCENT'] || 0;

    // Helper function to get player avatar emoji
    const getPlayerAvatar = (jobTitle: string) => {
      switch (jobTitle) {
        case 'CISO': return '👤';
        case 'Systems': return '🧑‍💻';
        case 'Ethics': return '🕵️';
        case 'CTO': return '🤖';
        case 'COO': return '🧑‍🚀';
        case 'CFO': return '👩‍🔬';
        default: return '👤';
      }
    };

    // Helper function to render blockchain vote blocks
    const renderVoteBlocks = (voteOption: string) => {
      if (!gameState.voteState?.votes) return null;
      
      return Object.entries(gameState.voteState.votes)
        .filter(([, vote]) => vote === voteOption)
        .map(([playerId]) => {
          const voter = gameState.players.find(p => p.id === playerId);
          if (!voter) return null;
          
          const isMyVote = voter.id === localPlayer.id;
          
          return (
            <div
              key={playerId}
              className={`vote-block ${isMyVote ? 'my-vote' : ''} animation-fade-in`}
            >
              <div className="block-header">
                <span className="block-icon">{getPlayerAvatar(voter.jobTitle)}</span>
                <span className="block-amount">{voter.tokens}</span>
              </div>
              <div className="block-hash">{voter.name}</div>
            </div>
          );
        });
    };

    // Check if local player has voted
    const hasVoted = gameState.voteState?.votes && localPlayer.id in gameState.voteState.votes;
    const myVote = hasVoted ? gameState.voteState?.votes[localPlayer.id] : null;

    return (
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animation-fade-in">
        <div className="mb-1.5">
          <h3 className="text-sm font-bold text-gray-100 normal-case tracking-normal text-left p-0 bg-transparent m-0 mb-3">
            Deactivate {nominatedPlayer.name}?
          </h3>
        </div>
        <div className="verdict-poll">
          {/* YES Vote Row */}
          <div className="vote-option yes">
            <span className="option-label">✔️ YES</span>
            <span className="vote-tally">🪙 {yesVotes}</span>
            <div className="blockchain-chain">
              {renderVoteBlocks('GUILTY')}
            </div>
            <Button
              variant={myVote === 'GUILTY' ? 'primary' : 'secondary'}
              size="sm"
              disabled={hasVoted}
              onClick={() => {
                playSound('vote');
                setSelectedVote('GUILTY');
                handleVote();
              }}
              className={`option-vote-btn ${
                myVote === 'GUILTY' ? 'voted bg-amber-500 border-amber-500 text-black' : 
                hasVoted ? 'opacity-50 cursor-not-allowed' :
                'hover:enabled:bg-green-500 hover:enabled:border-green-500 hover:enabled:text-white'
              }`}
              onMouseDown={(e) => {
                if (!hasVoted) {
                  e.currentTarget.classList.add('animation-pulse');
                  setTimeout(() => e.currentTarget.classList.remove('animation-pulse'), 600);
                }
              }}
            >
              VOTE
            </Button>
          </div>

          {/* NO Vote Row */}
          <div className="vote-option no">
            <span className="option-label">❌ NO</span>
            <span className="vote-tally">🪙 {noVotes}</span>
            <div className="blockchain-chain">
              {renderVoteBlocks('INNOCENT')}
            </div>
            <Button
              variant={myVote === 'INNOCENT' ? 'primary' : 'secondary'}
              size="sm"
              disabled={hasVoted}
              onClick={() => {
                playSound('vote');
                setSelectedVote('INNOCENT');
                handleVote();
              }}
              className={`option-vote-btn ${
                myVote === 'INNOCENT' ? 'voted bg-amber-500 border-amber-500 text-black' : 
                hasVoted ? 'opacity-50 cursor-not-allowed' :
                'hover:enabled:bg-red-500 hover:enabled:border-red-500 hover:enabled:text-white'
              }`}
              onMouseDown={(e) => {
                if (!hasVoted) {
                  e.currentTarget.classList.add('animation-pulse');
                  setTimeout(() => e.currentTarget.classList.remove('animation-pulse'), 600);
                }
              }}
            >
              VOTE
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return null;
};