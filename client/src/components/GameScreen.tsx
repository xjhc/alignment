import React, { useState, useEffect, useRef } from 'react';
import { useGameContext } from '../contexts/GameContext';
import { useSound } from '../hooks/useSound';
import { PrivateNotifications } from './PrivateNotifications';
import { RosterPanel } from './game/RosterPanel';
import { CommsPanel } from './game/CommsPanel';
import { PlayerHUD } from './game/PlayerHUD';
import { ExitInterviewScreen } from './game/ExitInterviewScreen';
import { AlignmentConversionOverlay } from './game/AlignmentConversionOverlay';
import { ExtensionVoteUI } from './game/ExtensionVoteUI';
import { EventType } from '../types/generated';

export function GameScreen() {
  const { gameState, localPlayer } = useGameContext();
  const { playSound } = useSound();
  const [showExitInterview, setShowExitInterview] = useState(false);
  const [showConversionOverlay, setShowConversionOverlay] = useState(false);
  const [showExtensionVoting, setShowExtensionVoting] = useState(false);
  const [extensionRemainingSeconds, setExtensionRemainingSeconds] = useState(15);
  const previousAlignment = useRef(localPlayer?.alignment);
  const previousPhase = useRef(gameState?.phase?.type);
  const previousMessageCount = useRef(0);
  const ambientSoundInitialized = useRef(false);

  // Show exit interview for eliminated players who haven't submitted a parting shot
  React.useEffect(() => {
    if (localPlayer && !localPlayer.isAlive && !localPlayer.partingShot && !showExitInterview) {
      setShowExitInterview(true);
    }
  }, [localPlayer, showExitInterview]);

  // Detect alignment conversion and show system takeover overlay
  useEffect(() => {
    if (localPlayer && 
        previousAlignment.current === 'HUMAN' && 
        (localPlayer.alignment === 'AI' || localPlayer.alignment === 'ALIGNED')) {
      setShowConversionOverlay(true);
      // Play conversion sound effect
      playSound('error'); // Dramatic sound for conversion
    }
    
    if (localPlayer) {
      previousAlignment.current = localPlayer.alignment;
    }
  }, [localPlayer?.alignment, playSound]);

  // Detect phase changes and play transition sounds
  useEffect(() => {
    if (gameState?.phase?.type && 
        previousPhase.current && 
        previousPhase.current !== gameState.phase.type) {
      // Play phase change sound
      playSound('phaseChange');
    }
    
    if (gameState?.phase?.type) {
      previousPhase.current = gameState.phase.type;
    }
  }, [gameState?.phase?.type, playSound]);

  // Start ambient background sound when game starts
  useEffect(() => {
    if (gameState && localPlayer && !ambientSoundInitialized.current) {
      // Start background ambiance
      import('../services/soundManager').then(({ soundManager }) => {
        soundManager.playMusic('lobby'); // Use lobby track as ambient sound
      });
      ambientSoundInitialized.current = true;
    }
  }, [gameState, localPlayer]);

  // Detect new chat messages and play notification sound
  useEffect(() => {
    const messages = gameState?.chatMessages || (gameState as any)?.chat_messages || [];
    const currentMessageCount = Array.isArray(messages) ? messages.length : 0;
    
    if (previousMessageCount.current > 0 && currentMessageCount > previousMessageCount.current) {
      // New message arrived - play notification sound
      playSound('message');
    }
    
    previousMessageCount.current = currentMessageCount;
  }, [gameState?.chatMessages, (gameState as any)?.chat_messages, playSound]);

  // Detect extension voting trigger events
  useEffect(() => {
    // Check for extension voting triggered based on phase and vote state
    if (gameState.phase.type === 'DISCUSSION' && 
        gameState.voteState?.type === 'EXTENSION' && 
        !gameState.voteState?.isComplete) {
      setShowExtensionVoting(true);
      // The remaining seconds would typically come from the event payload
      // For now, we'll use a default value
      setExtensionRemainingSeconds(15);
    } else {
      setShowExtensionVoting(false);
    }
  }, [gameState.phase.type, gameState.voteState?.type, gameState.voteState?.isComplete]);

  if (!localPlayer) {
    return (
      <div className="w-screen h-screen flex flex-col bg-background-primary">
        <div className="flex items-center justify-center text-text-muted text-sm">
          <span>Loading game state...</span>
          <div className="loading-spinner large" style={{ marginLeft: '12px' }}></div>
        </div>
      </div>
    );
  }

  return (
    <main className="w-screen h-screen grid grid-cols-[260px_1fr_320px] gap-px bg-border overflow-hidden">
      {/* Exit Interview Overlay */}
      {showExitInterview && (
        <ExitInterviewScreen
          onComplete={() => setShowExitInterview(false)}
        />
      )}

      {/* Private Notifications Overlay */}
      {gameState.privateNotifications && (
        <PrivateNotifications
          notifications={gameState.privateNotifications}
          onMarkAsRead={(notificationId) => {
            // TODO: Send action to mark notification as read
            console.log('Mark notification as read:', notificationId);
          }}
        />
      )}

      {/* Alignment Conversion System Takeover Overlay */}
      <AlignmentConversionOverlay
        isVisible={showConversionOverlay}
        onComplete={() => setShowConversionOverlay(false)}
      />

      {/* Extension Voting UI */}
      {showExtensionVoting && (
        <ExtensionVoteUI remainingSeconds={extensionRemainingSeconds} />
      )}
      
      <RosterPanel />
      
      <CommsPanel />
      
      <PlayerHUD />
    </main>
  );
}