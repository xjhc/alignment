import React from 'react';
import { GameState, Player } from '../../types';
import styles from './VoteUI.module.css';

interface VoteUIProps {
  gameState: GameState;
  localPlayer: Player;
  selectedNominee: string;
  setSelectedNominee: (value: string) => void;
  selectedVote: 'GUILTY' | 'INNOCENT' | '';
  setSelectedVote: (value: 'GUILTY' | 'INNOCENT' | '') => void;
  handleNominate: () => Promise<void>;
  handleVote: () => Promise<void>;
}

export const VoteUI: React.FC<VoteUIProps> = ({
  gameState,
  localPlayer,
  selectedNominee,
  setSelectedNominee,
  selectedVote,
  setSelectedVote,
  handleNominate,
  handleVote,
}) => {
  const alivePlayers = gameState.players.filter(p => p.isAlive && p.id !== localPlayer.id);
  
  if (gameState.phase.type === 'NOMINATION') {
    return (
      <div className={styles.votePromptContainer}>
        <div className={styles.votePromptHeader}>
          <h3 className={styles.votePromptTitle}>Who should we deactivate?</h3>
        </div>
        <div className={styles.nomineeGrid}>
          {alivePlayers.map((player) => {
            const isSelected = selectedNominee === player.id;
            const playerVotes = gameState.voteState?.votes ? 
              Object.values(gameState.voteState.votes).filter(vote => vote === player.id).length : 0;
            
            return (
              <button
                key={player.id}
                className={`${styles.nomineeBtn} ${isSelected ? styles.voted : ''} ${player.alignment === 'ALIGNED' ? styles.aligned : ''}`}
                onClick={() => {
                  setSelectedNominee(player.id);
                  handleNominate();
                }}
              >
                <span className={styles.nomineeEmoji}>
                  {player.jobTitle === 'CISO' ? '👤' :
                   player.jobTitle === 'Systems' ? '🧑‍💻' :
                   player.jobTitle === 'Ethics' ? '🕵️' :
                   player.jobTitle === 'CTO' ? '🤖' :
                   player.jobTitle === 'COO' ? '🧑‍🚀' :
                   player.jobTitle === 'CFO' ? '👩‍🔬' : '👤'}
                </span>
                <span className={`${styles.nomineeName} ${player.alignment === 'ALIGNED' ? styles.glitched : ''}`}>
                  {player.name}
                </span>
                <span className={styles.nomineeVotes}>🪙 {playerVotes}</span>
              </button>
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
      <div className={styles.votePromptContainer}>
        <div className={styles.votePromptHeader}>
          <h3 className={styles.votePromptTitle}>Deactivate {nominatedPlayer.name}?</h3>
        </div>
        <div className={styles.verdictPoll}>
          <div className={`${styles.voteOption} ${styles.yes}`}>
            <span className={styles.optionLabel}>✔️ YES</span>
            <span className={styles.voteTally}>🪙 {yesVotes}</span>
            <div className={styles.blockchainChain}>
              {gameState.voteState?.votes && Object.entries(gameState.voteState.votes)
                .filter(([, vote]) => vote === 'GUILTY')
                .map(([playerId]) => {
                  const voter = gameState.players.find(p => p.id === playerId);
                  if (!voter) return null;
                  return (
                    <div key={playerId} className={`${styles.voteBlock} ${voter.id === localPlayer.id ? styles.myVote : ''}`}>
                      <div className={styles.blockHeader}>
                        <span className={styles.blockIcon}>
                          {voter.jobTitle === 'CISO' ? '👤' :
                           voter.jobTitle === 'Systems' ? '🧑‍💻' :
                           voter.jobTitle === 'Ethics' ? '🕵️' :
                           voter.jobTitle === 'CTO' ? '🤖' :
                           voter.jobTitle === 'COO' ? '🧑‍🚀' :
                           voter.jobTitle === 'CFO' ? '👩‍🔬' : '👤'}
                        </span>
                        <span className={styles.blockAmount}>{voter.tokens}</span>
                      </div>
                      <div className={styles.blockHash}>{voter.name}</div>
                    </div>
                  );
                })}
            </div>
            <button 
              className={`${styles.optionVoteBtn} ${selectedVote === 'GUILTY' ? styles.voted : ''}`}
              onClick={() => {
                setSelectedVote('GUILTY');
                handleVote();
              }}
            >
              VOTE
            </button>
          </div>
          <div className={`${styles.voteOption} ${styles.no}`}>
            <span className={styles.optionLabel}>❌ NO</span>
            <span className={styles.voteTally}>🪙 {noVotes}</span>
            <div className={styles.blockchainChain}>
              {gameState.voteState?.votes && Object.entries(gameState.voteState.votes)
                .filter(([, vote]) => vote === 'INNOCENT')
                .map(([playerId]) => {
                  const voter = gameState.players.find(p => p.id === playerId);
                  if (!voter) return null;
                  return (
                    <div key={playerId} className={`${styles.voteBlock} ${voter.id === localPlayer.id ? styles.myVote : ''}`}>
                      <div className={styles.blockHeader}>
                        <span className={styles.blockIcon}>
                          {voter.jobTitle === 'CISO' ? '👤' :
                           voter.jobTitle === 'Systems' ? '🧑‍💻' :
                           voter.jobTitle === 'Ethics' ? '🕵️' :
                           voter.jobTitle === 'CTO' ? '🤖' :
                           voter.jobTitle === 'COO' ? '🧑‍🚀' :
                           voter.jobTitle === 'CFO' ? '👩‍🔬' : '👤'}
                        </span>
                        <span className={styles.blockAmount}>{voter.tokens}</span>
                      </div>
                      <div className={styles.blockHash}>{voter.name}</div>
                    </div>
                  );
                })}
            </div>
            <button 
              className={`${styles.optionVoteBtn} ${selectedVote === 'INNOCENT' ? styles.voted : ''}`}
              onClick={() => {
                setSelectedVote('INNOCENT');
                handleVote();
              }}
            >
              VOTE
            </button>
          </div>
        </div>
      </div>
    );
  }

  return null;
};