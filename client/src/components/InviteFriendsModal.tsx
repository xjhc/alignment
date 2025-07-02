import React, { useState, useEffect } from 'react';
import { Button } from './ui/Button';

interface Friend {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'in_game' | 'in_lobby';
}

interface InviteFriendsModalProps {
  isVisible: boolean;
  onClose: () => void;
  lobbyId?: string;
}

export const InviteFriendsModal: React.FC<InviteFriendsModalProps> = ({ 
  isVisible, 
  onClose, 
  lobbyId 
}) => {
  const [friends, setFriends] = useState<Friend[]>([]);
  const [selectedFriends, setSelectedFriends] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isInviting, setIsInviting] = useState(false);

  useEffect(() => {
    if (isVisible) {
      loadFriends();
    }
  }, [isVisible]);

  const loadFriends = async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/api/friends', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        // Filter to only show online friends who can be invited
        const availableFriends = (data.friends || []).filter(
          (friend: Friend) => friend.status === 'online' || friend.status === 'in_lobby'
        );
        setFriends(availableFriends);
      } else {
        setError('Failed to load friends list');
      }
    } catch (err) {
      setError('Error loading friends');
      console.error('Error loading friends:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const toggleFriendSelection = (friendId: string) => {
    const newSelected = new Set(selectedFriends);
    if (newSelected.has(friendId)) {
      newSelected.delete(friendId);
    } else {
      newSelected.add(friendId);
    }
    setSelectedFriends(newSelected);
  };

  const sendInvites = async () => {
    if (selectedFriends.size === 0 || !lobbyId) return;

    try {
      setIsInviting(true);
      setError(null);

      const invitePromises = Array.from(selectedFriends).map(friendId =>
        fetch('/api/lobby/invite', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            lobby_id: lobbyId,
            friend_id: friendId,
          }),
        })
      );

      const responses = await Promise.all(invitePromises);
      const failedInvites = responses.filter(r => !r.ok);

      if (failedInvites.length === 0) {
        // All invites sent successfully
        alert(`Invites sent to ${selectedFriends.size} friend(s)!`);
        setSelectedFriends(new Set());
        onClose();
      } else {
        setError(`Failed to send ${failedInvites.length} invite(s). Please try again.`);
      }
    } catch (err) {
      setError('Error sending invites');
      console.error('Error sending invites:', err);
    } finally {
      setIsInviting(false);
    }
  };

  const getStatusBadge = (status: Friend['status']) => {
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

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-background-primary border border-border rounded-lg w-full max-w-md h-96 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-text-primary">Invite Friends to Lobby</h2>
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
            <div className="text-center text-text-secondary">Loading friends...</div>
          ) : friends.length === 0 ? (
            <div className="text-center text-text-secondary">
              <p>No available friends to invite.</p>
              <p className="text-sm mt-1">Friends must be online to receive lobby invites.</p>
            </div>
          ) : (
            <>
              <div className="mb-4 text-sm text-text-secondary">
                Select friends to invite to this lobby ({selectedFriends.size} selected):
              </div>
              <div className="space-y-2">
                {friends.map((friend) => (
                  <div
                    key={friend.id}
                    className={`flex items-center justify-between p-3 rounded border cursor-pointer transition-colors ${
                      selectedFriends.has(friend.id)
                        ? 'bg-primary/10 border-primary'
                        : 'bg-background-secondary border-border hover:bg-background-tertiary'
                    }`}
                    onClick={() => toggleFriendSelection(friend.id)}
                  >
                    <div className="flex items-center space-x-3">
                      <div className="w-8 h-8 bg-primary rounded-full flex items-center justify-center text-white font-semibold">
                        {friend.name[0].toUpperCase()}
                      </div>
                      <div>
                        <p className="font-medium text-text-primary">{friend.name}</p>
                        {getStatusBadge(friend.status)}
                      </div>
                    </div>
                    <div className="text-primary">
                      {selectedFriends.has(friend.id) ? '✓' : '○'}
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        {friends.length > 0 && (
          <div className="flex justify-between p-4 border-t border-border">
            <Button
              variant="secondary"
              onClick={onClose}
              disabled={isInviting}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={sendInvites}
              disabled={selectedFriends.size === 0 || isInviting}
            >
              {isInviting 
                ? 'Sending...' 
                : `Send Invite${selectedFriends.size !== 1 ? 's' : ''} (${selectedFriends.size})`
              }
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};