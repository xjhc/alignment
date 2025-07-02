import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from './ui/Button';
import { useWebSocketContext } from '../contexts/WebSocketContext';

interface PartyInfo {
  id: string;
  leaderId: string;
  leaderName: string;
  memberCount: number;
  maxMembers: number;
  createdAt: string;
}

export function PartyJoinScreen() {
  const { inviteCode } = useParams<{ inviteCode: string }>();
  const navigate = useNavigate();
  const [partyInfo, setPartyInfo] = useState<PartyInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isJoining, setIsJoining] = useState(false);
  
  const { sendAction } = useWebSocketContext();

  useEffect(() => {
    if (inviteCode) {
      loadPartyInfo();
    }
  }, [inviteCode]);

  const loadPartyInfo = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const response = await fetch(`/api/party/invite/${inviteCode}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setPartyInfo(data.party);
      } else if (response.status === 404) {
        setError('Party invite not found or has expired');
      } else if (response.status === 410) {
        setError('This party invite has expired');
      } else {
        setError('Failed to load party information');
      }
    } catch (err) {
      setError('Error loading party information');
      console.error('Error loading party info:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const joinParty = async () => {
    if (!inviteCode) return;

    try {
      setIsJoining(true);
      setError(null);

      sendAction({
        type: 'JOIN_PARTY_BY_INVITE' as any,
        payload: { invite_code: inviteCode }
      });

      // Navigate to lobby list after joining
      setTimeout(() => {
        navigate('/lobby-list');
      }, 1000);

    } catch (err) {
      setError('Error joining party');
      console.error('Error joining party:', err);
    } finally {
      setIsJoining(false);
    }
  };

  return (
    <div className="w-screen h-screen flex flex-col items-center justify-center gap-6 bg-background-primary text-text-primary">
      <h1 className="font-mono text-3xl font-semibold tracking-[2px]">
        LOEBIAN INC. // <span className="inline-block animate-pulse">PARTY INVITE</span>
      </h1>

      <div className="flex flex-col gap-6 max-w-md mx-auto text-center">
        {isLoading ? (
          <div className="flex flex-col gap-4 items-center">
            <div className="loading-spinner large"></div>
            <p className="text-text-secondary">Loading party information...</p>
          </div>
        ) : error ? (
          <div className="flex flex-col gap-4 items-center">
            <div className="text-4xl">❌</div>
            <h2 className="text-xl font-semibold text-danger">Invite Invalid</h2>
            <p className="text-text-secondary">{error}</p>
            <Button
              variant="secondary"
              onClick={() => navigate('/lobby-list')}
            >
              Go to Lobby List
            </Button>
          </div>
        ) : partyInfo ? (
          <div className="flex flex-col gap-4 items-center">
            <div className="text-4xl">🎉</div>
            <h2 className="text-xl font-semibold">Party Invitation</h2>
            
            <div className="bg-background-secondary border border-border rounded-lg p-4 text-left w-full">
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-text-secondary">Leader:</span>
                  <span className="font-medium">{partyInfo.leaderName}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-secondary">Members:</span>
                  <span className="font-medium">{partyInfo.memberCount}/{partyInfo.maxMembers}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-secondary">Created:</span>
                  <span className="font-medium">
                    {new Date(partyInfo.createdAt).toLocaleDateString()}
                  </span>
                </div>
              </div>
            </div>

            <p className="text-text-secondary text-sm">
              You've been invited to join <strong>{partyInfo.leaderName}'s</strong> party!
            </p>

            <div className="flex gap-3 w-full">
              <Button
                variant="secondary"
                onClick={() => navigate('/lobby-list')}
                disabled={isJoining}
                className="flex-1"
              >
                Decline
              </Button>
              <Button
                variant="primary"
                onClick={joinParty}
                disabled={isJoining || partyInfo.memberCount >= partyInfo.maxMembers}
                className="flex-1"
              >
                {isJoining ? 'Joining...' : 'Join Party'}
              </Button>
            </div>

            {partyInfo.memberCount >= partyInfo.maxMembers && (
              <p className="text-sm text-danger">
                This party is currently full.
              </p>
            )}
          </div>
        ) : (
          <div className="flex flex-col gap-4 items-center">
            <div className="text-4xl">❓</div>
            <h2 className="text-xl font-semibold">Party Not Found</h2>
            <p className="text-text-secondary">This party invite could not be found.</p>
            <Button
              variant="secondary"
              onClick={() => navigate('/lobby-list')}
            >
              Go to Lobby List
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}