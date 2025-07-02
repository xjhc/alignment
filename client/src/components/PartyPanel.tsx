import React, { useState, useEffect, useCallback } from 'react';
import { Button } from './ui/Button';
import { useWebSocketContext } from '../contexts/WebSocketContext';

interface PartyMember {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'in_game' | 'in_lobby';
  isLeader: boolean;
}

interface Party {
  id: string;
  leaderId: string;
  members: PartyMember[];
  createdAt: string;
}

interface PartyPanelProps {
  isVisible: boolean;
  onClose: () => void;
  currentPlayerId?: string;
}

export const PartyPanel: React.FC<PartyPanelProps> = ({ isVisible, onClose, currentPlayerId }) => {
  const [party, setParty] = useState<Party | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [inviteCode, setInviteCode] = useState<string>('');
  
  const { subscribe, sendAction } = useWebSocketContext();

  // WebSocket event handlers
  const handlePartyUpdate = useCallback((event: any) => {
    const { party: updatedParty } = event.payload;
    setParty(updatedParty);
  }, []);

  const handlePartyDisbanded = useCallback(() => {
    setParty(null);
  }, []);

  const handlePartyInviteReceived = useCallback((event: any) => {
    const { invite } = event.payload;
    
    // Show notification
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification('Party Invite', {
        body: `${invite.inviter_name} invited you to join their party`,
        icon: '/favicon.ico'
      });
    }

    // Auto-show accept/decline UI (could be enhanced with a modal)
    const accepted = confirm(`${invite.inviter_name} invited you to join their party. Accept?`);
    if (accepted) {
      acceptPartyInvite(invite.id);
    }
  }, []);

  // Subscribe to WebSocket events
  useEffect(() => {
    if (isVisible) {
      const unsubscribers = [
        subscribe('PARTY_UPDATE', handlePartyUpdate),
        subscribe('PARTY_DISBANDED', handlePartyDisbanded),
        subscribe('PARTY_INVITE_RECEIVED', handlePartyInviteReceived),
      ];

      return () => {
        unsubscribers.forEach(unsub => unsub());
      };
    }
  }, [isVisible, subscribe, handlePartyUpdate, handlePartyDisbanded, handlePartyInviteReceived]);

  // Load party data on component mount
  useEffect(() => {
    if (isVisible) {
      loadParty();
    }
  }, [isVisible]);

  const loadParty = async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/api/party', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setParty(data.party || null);
      } else if (response.status !== 404) {
        setError('Failed to load party information');
      }
    } catch (err) {
      setError('Error loading party');
      console.error('Error loading party:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const createParty = () => {
    try {
      sendAction({
        type: 'CREATE_PARTY' as any,
        payload: {}
      });
    } catch (err) {
      setError('Failed to create party');
      console.error('Error creating party:', err);
    }
  };

  const leaveParty = () => {
    if (!confirm('Are you sure you want to leave the party?')) {
      return;
    }

    try {
      sendAction({
        type: 'LEAVE_PARTY' as any,
        payload: {}
      });
    } catch (err) {
      setError('Failed to leave party');
      console.error('Error leaving party:', err);
    }
  };

  const inviteToParty = (friendId: string) => {
    try {
      sendAction({
        type: 'INVITE_TO_PARTY' as any,
        payload: { friend_id: friendId }
      });
    } catch (err) {
      setError('Failed to send party invite');
      console.error('Error inviting to party:', err);
    }
  };

  const acceptPartyInvite = (inviteId: string) => {
    try {
      sendAction({
        type: 'ACCEPT_PARTY_INVITE' as any,
        payload: { invite_id: inviteId }
      });
    } catch (err) {
      setError('Failed to accept party invite');
      console.error('Error accepting party invite:', err);
    }
  };

  const promoteToLeader = (memberId: string) => {
    try {
      sendAction({
        type: 'PROMOTE_PARTY_LEADER' as any,
        payload: { member_id: memberId }
      });
    } catch (err) {
      setError('Failed to promote party member');
      console.error('Error promoting party member:', err);
    }
  };

  const kickFromParty = (memberId: string) => {
    if (!confirm('Are you sure you want to kick this member from the party?')) {
      return;
    }

    try {
      sendAction({
        type: 'KICK_FROM_PARTY' as any,
        payload: { member_id: memberId }
      });
    } catch (err) {
      setError('Failed to kick party member');
      console.error('Error kicking party member:', err);
    }
  };

  const joinLobbyWithParty = () => {
    try {
      sendAction({
        type: 'JOIN_LOBBY_WITH_PARTY' as any,
        payload: {}
      });
    } catch (err) {
      setError('Failed to join lobby with party');
      console.error('Error joining lobby with party:', err);
    }
  };

  const generateInviteLink = async () => {
    if (!party) return;

    try {
      const response = await fetch('/api/party/invite-link', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        const inviteUrl = `${window.location.origin}/party/join/${data.invite_code}`;
        
        if (navigator.clipboard) {
          await navigator.clipboard.writeText(inviteUrl);
          alert('Invite link copied to clipboard!');
        } else {
          setInviteCode(inviteUrl);
        }
      } else {
        setError('Failed to generate invite link');
      }
    } catch (err) {
      setError('Error generating invite link');
      console.error('Error generating invite link:', err);
    }
  };

  const getStatusBadge = (status: PartyMember['status']) => {
    const statusClasses = {
      online: 'bg-green text-white',
      offline: 'bg-gray text-white',
      in_game: 'bg-purple text-white',
      in_lobby: 'bg-blue text-white',
    };

    const statusText = {
      online: 'Online',
      offline: 'Offline',
      in_game: 'In Game',
      in_lobby: 'In Lobby',
    };

    return (
      <span className={`px-2 py-1 rounded text-xs ${statusClasses[status]}`}>
        {statusText[status]}
      </span>
    );
  };

  const isLeader = party && currentPlayerId && party.leaderId === currentPlayerId;

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-background-primary border border-border rounded-lg w-full max-w-md h-96 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-text-primary">Party</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-text-secondary hover:text-text-primary"
          >
            ✕
          </Button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4">
          {error && (
            <div className="mb-4 p-3 bg-danger text-white rounded">
              {error}
            </div>
          )}

          {isLoading ? (
            <div className="text-center text-text-secondary">Loading...</div>
          ) : !party ? (
            <div className="text-center space-y-4">
              <div className="text-text-secondary">
                <p>You're not in a party.</p>
                <p className="text-sm">Create or join a party to play with friends!</p>
              </div>
              <Button
                variant="primary"
                onClick={createParty}
                className="w-full"
              >
                Create Party
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Party Members */}
              <div>
                <h3 className="text-sm font-medium text-text-primary mb-3">
                  Members ({party.members.length}/6)
                </h3>
                <div className="space-y-2">
                  {party.members.map((member) => (
                    <div
                      key={member.id}
                      className="flex items-center justify-between p-3 bg-background-secondary rounded border"
                    >
                      <div className="flex items-center space-x-3">
                        <div className="w-8 h-8 bg-primary rounded-full flex items-center justify-center text-white font-semibold">
                          {member.name[0].toUpperCase()}
                        </div>
                        <div>
                          <div className="flex items-center space-x-2">
                            <p className="font-medium text-text-primary">{member.name}</p>
                            {member.isLeader && (
                              <span className="text-xs bg-yellow-500 text-black px-2 py-0.5 rounded">
                                Leader
                              </span>
                            )}
                          </div>
                          {getStatusBadge(member.status)}
                        </div>
                      </div>
                      
                      {isLeader && member.id !== currentPlayerId && (
                        <div className="flex space-x-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => promoteToLeader(member.id)}
                            className="text-xs text-yellow-500 hover:bg-yellow-500 hover:text-black"
                            title="Promote to leader"
                          >
                            👑
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => kickFromParty(member.id)}
                            className="text-xs text-danger hover:bg-danger hover:text-white"
                            title="Kick from party"
                          >
                            ✕
                          </Button>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              {/* Party Actions */}
              <div className="space-y-2">
                {isLeader && (
                  <>
                    <Button
                      variant="primary"
                      className="w-full"
                      onClick={joinLobbyWithParty}
                    >
                      Join Lobby as Party
                    </Button>
                    <Button
                      variant="secondary"
                      className="w-full"
                      onClick={generateInviteLink}
                    >
                      Generate Invite Link
                    </Button>
                  </>
                )}
                
                <Button
                  variant="danger"
                  className="w-full"
                  onClick={leaveParty}
                >
                  Leave Party
                </Button>
              </div>

              {inviteCode && (
                <div className="p-3 bg-background-secondary rounded border">
                  <p className="text-sm text-text-primary mb-2">Invite Link:</p>
                  <input
                    type="text"
                    value={inviteCode}
                    readOnly
                    className="w-full px-2 py-1 text-xs bg-background-primary border border-border rounded"
                    onClick={(e) => (e.target as HTMLInputElement).select()}
                  />
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};