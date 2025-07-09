import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useSessionContext } from '../contexts/SessionContext';
import { getUserIdForApi } from '../services/guestIdentity';

interface JoinLobbyScreenProps {}

export function JoinLobbyScreen({}: JoinLobbyScreenProps) {
  const { lobbyId } = useParams<{ lobbyId: string }>();
  const navigate = useNavigate();
  const { onJoinLobby, appState, sessionState } = useSessionContext();
  const [error, setError] = useState<string | null>(null);
  const [isJoining, setIsJoining] = useState(false);

  useEffect(() => {
    // If user is not logged in, redirect to login with return URL
    if (!appState.playerName) {
      const returnUrl = encodeURIComponent(`/join/${lobbyId}`);
      navigate(`/login?return=${returnUrl}`, { replace: true });
      return;
    }

    // If user is already in a session, handle appropriately
    if (sessionState === 'IN_LOBBY' || sessionState === 'IN_GAME') {
      setError('You are already in a game session. Please leave your current session to join a new lobby.');
      return;
    }

    // Attempt to join the lobby
    if (lobbyId && !isJoining) {
      setIsJoining(true);
      attemptJoinLobby(lobbyId);
    }
  }, [lobbyId, appState.playerName, sessionState, isJoining]);

  const attemptJoinLobby = async (targetLobbyId: string) => {
    try {
      const userId = getUserIdForApi();
      
      // First, try to get lobby info to get the lobby name
      let lobbyName = "";
      try {
        const lobbyListResponse = await fetch("/api/games");
        if (lobbyListResponse.ok) {
          const lobbyListData = await lobbyListResponse.json();
          const lobby = lobbyListData.lobbies?.find((l: any) => l.id === targetLobbyId);
          lobbyName = lobby?.name || "";
        }
      } catch (lobbyInfoErr) {
        console.warn("Could not fetch lobby info for invite link join:", lobbyInfoErr);
      }
      
      const response = await fetch(`/api/games/${targetLobbyId}/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          player_name: appState.playerName,
          player_avatar: appState.playerAvatar || '',
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to join lobby');
      }

      const data = await response.json();
      // Call the session context with the proper parameters, including lobby name if available
      onJoinLobby(data.game_id, data.player_id, data.session_token, lobbyName);
    } catch (err) {
      console.error('Failed to join lobby via invite link:', err);
      setError(err instanceof Error ? err.message : 'Failed to join lobby');
      setIsJoining(false);
    }
  };

  const handleGoToLobbyList = () => {
    navigate('/lobby-list');
  };

  const handleRetry = () => {
    if (lobbyId) {
      setError(null);
      setIsJoining(true);
      attemptJoinLobby(lobbyId);
    }
  };

  if (!lobbyId) {
    return (
      <div className="min-h-screen bg-background-primary flex items-center justify-center">
        <div className="bg-background-secondary rounded-lg p-8 max-w-md w-full mx-4 text-center">
          <h1 className="text-xl font-semibold text-text-primary mb-4">
            Invalid Invite Link
          </h1>
          <p className="text-text-secondary mb-6">
            This invite link is not valid or has expired.
          </p>
          <button
            onClick={handleGoToLobbyList}
            className="bg-primary text-white px-6 py-2 rounded-md hover:bg-primary-hover transition-colors"
          >
            Browse Available Lobbies
          </button>
        </div>
      </div>
    );
  }

  if (isJoining && !error) {
    return (
      <div className="min-h-screen bg-background-primary flex items-center justify-center">
        <div className="bg-background-secondary rounded-lg p-8 max-w-md w-full mx-4 text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <h1 className="text-xl font-semibold text-text-primary mb-2">
            Joining Lobby...
          </h1>
          <p className="text-text-secondary">
            Please wait while we connect you to the game.
          </p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background-primary flex items-center justify-center">
        <div className="bg-background-secondary rounded-lg p-8 max-w-md w-full mx-4 text-center">
          <div className="w-12 h-12 mx-auto mb-4 text-danger">
            <svg className="w-full h-full" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-text-primary mb-4">
            Unable to Join Lobby
          </h1>
          <p className="text-text-secondary mb-6">
            {error}
          </p>
          <div className="space-y-3">
            <button
              onClick={handleRetry}
              className="w-full bg-primary text-white px-6 py-2 rounded-md hover:bg-primary-hover transition-colors"
            >
              Try Again
            </button>
            <button
              onClick={handleGoToLobbyList}
              className="w-full border border-border-primary text-text-primary px-6 py-2 rounded-md hover:bg-background-tertiary transition-colors"
            >
              Browse Available Lobbies
            </button>
          </div>
        </div>
      </div>
    );
  }

  // This shouldn't normally be reached, but just in case
  return (
    <div className="min-h-screen bg-background-primary flex items-center justify-center">
      <div className="bg-background-secondary rounded-lg p-8 max-w-md w-full mx-4 text-center">
        <h1 className="text-xl font-semibold text-text-primary mb-4">
          Processing Invite...
        </h1>
      </div>
    </div>
  );
}