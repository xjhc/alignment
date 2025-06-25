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
    selectedVote,
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

    return (
      <div className="p-3 px-4 bg-gray-900 border-t border-gray-700 animate-[fadeIn_0.3s_ease]">
        <div className="mb-1.5">
          <h3 className="text-sm font-bold text-gray-100 normal-case tracking-normal text-left p-0 bg-transparent m-0 mb-3">Deactivate {nominatedPlayer.name}?</h3>
        </div>
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-3 px-4 py-2 rounded-md bg-gray-900 border border-gray-600 transition-all duration-150 hover:bg-gray-700">
            <span className="font-bold text-sm min-w-[50px] text-green-500">✔️ YES</span>
            <span className="font-mono text-sm font-bold min-w-[35px] text-right text-green-500">🪙 {yesVotes}</span>
            <div className="flex-grow flex gap-1 flex-wrap items-center">
              {gameState.voteState?.votes && Object.entries(gameState.voteState.votes)
                .filter(([, vote]) => vote === 'GUILTY')
                .map(([playerId]) => {
                  const voter = gameState.players.find(p => p.id === playerId);
                  if (!voter) return null;
                  return (
                    <div key={playerId} className={`flex flex-col items-center gap-0.5 px-1.5 py-1 bg-gray-700 border border-gray-600 rounded text-xs min-w-[40px] ${
                      voter.id === localPlayer.id ? 'bg-amber-500 border-amber-500 text-black' : ''
                    }`}>
                      <div className="flex items-center gap-0.5">
                        <span className="text-xs">
                          {voter.jobTitle === 'CISO' ? '👤' :
                           voter.jobTitle === 'Systems' ? '🧑‍💻' :
                           voter.jobTitle === 'Ethics' ? '🕵️' :
                           voter.jobTitle === 'CTO' ? '🤖' :
                           voter.jobTitle === 'COO' ? '🧑‍🚀' :
                           voter.jobTitle === 'CFO' ? '👩‍🔬' : '👤'}
                        </span>
                        <span className="font-mono font-bold text-xs">{voter.tokens}</span>
                      </div>
                      <div className="font-mono text-xs text-gray-500 text-center max-w-[35px] overflow-hidden text-ellipsis whitespace-nowrap">{voter.name}</div>
                    </div>
                  );
                })}
            </div>
            <Button
              variant={selectedVote === 'GUILTY' ? 'primary' : 'secondary'}
              size="sm"
              onClick={() => {
                setSelectedVote('GUILTY');
                handleVote();
              }}
              className={`text-xs font-bold ${
                selectedVote === 'GUILTY' ? 'bg-amber-500 border-amber-500 text-black' : 'hover:enabled:bg-green-500 hover:enabled:border-green-500 hover:enabled:text-white'
              }`}
              onMouseDown={(e) => {
                e.currentTarget.classList.add('animation-pulse');
                setTimeout(() => e.currentTarget.classList.remove('animation-pulse'), 600);
              }}
            >
              VOTE
            </Button>
          </div>
          <div className="flex items-center gap-3 px-4 py-2 rounded-md bg-gray-900 border border-gray-600 transition-all duration-150 hover:bg-gray-700">
            <span className="font-bold text-sm min-w-[50px] text-red-500">❌ NO</span>
            <span className="font-mono text-sm font-bold min-w-[35px] text-right text-red-500">🪙 {noVotes}</span>
            <div className="flex-grow flex gap-1 flex-wrap items-center">
              {gameState.voteState?.votes && Object.entries(gameState.voteState.votes)
                .filter(([, vote]) => vote === 'INNOCENT')
                .map(([playerId]) => {
                  const voter = gameState.players.find(p => p.id === playerId);
                  if (!voter) return null;
                  return (
                    <div key={playerId} className={`flex flex-col items-center gap-0.5 px-1.5 py-1 bg-gray-700 border border-gray-600 rounded text-xs min-w-[40px] ${
                      voter.id === localPlayer.id ? 'bg-amber-500 border-amber-500 text-black' : ''
                    }`}>
                      <div className="flex items-center gap-0.5">
                        <span className="text-xs">
                          {voter.jobTitle === 'CISO' ? '👤' :
                           voter.jobTitle === 'Systems' ? '🧑‍💻' :
                           voter.jobTitle === 'Ethics' ? '🕵️' :
                           voter.jobTitle === 'CTO' ? '🤖' :
                           voter.jobTitle === 'COO' ? '🧑‍🚀' :
                           voter.jobTitle === 'CFO' ? '👩‍🔬' : '👤'}
                        </span>
                        <span className="font-mono font-bold text-xs">{voter.tokens}</span>
                      </div>
                      <div className="font-mono text-xs text-gray-500 text-center max-w-[35px] overflow-hidden text-ellipsis whitespace-nowrap">{voter.name}</div>
                    </div>
                  );
                })}
            </div>
            <Button
              variant={selectedVote === 'INNOCENT' ? 'primary' : 'secondary'}
              size="sm"
              onClick={() => {
                setSelectedVote('INNOCENT');
                handleVote();
              }}
              className={`text-xs font-bold ${
                selectedVote === 'INNOCENT' ? 'bg-amber-500 border-amber-500 text-black' : 'hover:enabled:bg-red-500 hover:enabled:border-red-500 hover:enabled:text-white'
              }`}
              onMouseDown={(e) => {
                e.currentTarget.classList.add('animation-pulse');
                setTimeout(() => e.currentTarget.classList.remove('animation-pulse'), 600);
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