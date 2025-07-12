import React, { useState, useEffect, useCallback } from 'react';
import { Button } from './ui/Button';
import { useWebSocketContext } from '../contexts/WebSocketContext';

interface Friend {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'in_game' | 'in_lobby';
  lastSeen?: string;
}

interface FriendRequest {
  id: string;
  requesterName: string;
  requesterId: string;
  createdAt: string;
  status: 'pending' | 'accepted' | 'declined';
}

interface FriendsPanelProps {
  isVisible: boolean;
  onClose: () => void;
}

export const FriendsPanel: React.FC<FriendsPanelProps> = ({ isVisible, onClose }) => {
  const [activeTab, setActiveTab] = useState<'friends' | 'requests' | 'add'>('friends');
  const [friends, setFriends] = useState<Friend[]>([]);
  const [friendRequests, setFriendRequests] = useState<FriendRequest[]>([]);
  const [searchUsername, setSearchUsername] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [suggestions, setSuggestions] = useState<Friend[]>([]);
  
  const { subscribe } = useWebSocketContext();

  // WebSocket event handlers
  const handleFriendRequestReceived = useCallback((event: any) => {
    const { request } = event.payload;
    setFriendRequests(prev => [...prev, {
      id: request.id,
      requesterName: request.requester_name,
      requesterId: request.requester_id,
      createdAt: request.created_at,
      status: 'pending'
    }]);
    
    // Show notification
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification('New Friend Request', {
        body: `${request.requester_name} sent you a friend request`,
        icon: '/favicon.ico'
      });
    }
  }, []);

  const handleFriendRequestAccepted = useCallback((event: any) => {
    const { friend } = event.payload;
    setFriends(prev => [...prev, {
      id: friend.id,
      name: friend.name,
      status: friend.status || 'offline'
    }]);
    
    // Remove from requests if it was our outgoing request
    setFriendRequests(prev => prev.filter(req => req.requesterId !== friend.id));
  }, []);

  const handleFriendStatusUpdate = useCallback((event: any) => {
    const { friend_id, status } = event.payload;
    setFriends(prev => prev.map(friend => 
      friend.id === friend_id ? { ...friend, status } : friend
    ));
  }, []);

  // Subscribe to WebSocket events
  useEffect(() => {
    if (isVisible) {
      const unsubscribers = [
        subscribe('FRIEND_REQUEST_RECEIVED', handleFriendRequestReceived),
        subscribe('FRIEND_REQUEST_ACCEPTED', handleFriendRequestAccepted),
        subscribe('FRIEND_STATUS_UPDATE', handleFriendStatusUpdate),
      ];

      return () => {
        unsubscribers.forEach(unsub => unsub());
      };
    }
  }, [isVisible, subscribe, handleFriendRequestReceived, handleFriendRequestAccepted, handleFriendStatusUpdate]);

  // Load friends list on component mount
  useEffect(() => {
    if (isVisible) {
      loadFriends();
      loadFriendRequests();
      loadSuggestions();
      requestNotificationPermission();
    }
  }, [isVisible]);

  const requestNotificationPermission = () => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission();
    }
  };

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
        // Load friends with presence status
        const friendsWithStatus = data.friends || [];
        setFriends(friendsWithStatus);
        
        // Poll for status updates every 30 seconds
        if (friendsWithStatus.length > 0) {
          loadFriendStatuses(friendsWithStatus.map((f: any) => f.id));
        }
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

  const loadFriendStatuses = async (friendIds: string[]) => {
    try {
      const response = await fetch('/api/friends/status', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ friend_ids: friendIds }),
      });

      if (response.ok) {
        const data = await response.json();
        const statusMap = data.statuses || {};
        
        setFriends(prev => prev.map(friend => ({
          ...friend,
          status: statusMap[friend.id] || 'offline'
        })));
      }
    } catch (err) {
      console.error('Error loading friend statuses:', err);
    }
  };

  const loadSuggestions = async () => {
    try {
      const response = await fetch('/api/friends/suggestions', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSuggestions(data.suggestions || []);
      }
    } catch (err) {
      console.error('Error loading friend suggestions:', err);
    }
  };

  const loadFriendRequests = async () => {
    try {
      const response = await fetch('/api/friends/requests', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setFriendRequests(data.requests || []);
      }
    } catch (err) {
      console.error('Error loading friend requests:', err);
    }
  };

  const sendFriendRequest = async () => {
    if (!searchUsername.trim()) {
      setError('Please enter a username');
      return;
    }

    try {
      setIsLoading(true);
      const response = await fetch('/api/friends/requests', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          recipient_username: searchUsername.trim(),
        }),
      });

      if (response.ok) {
        setSearchUsername('');
        setError(null);
        // Show success message
        alert('Friend request sent successfully!');
      } else {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to send friend request');
      }
    } catch (err) {
      setError('Error sending friend request');
      console.error('Error sending friend request:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const acceptFriendRequest = async (requestId: string) => {
    try {
      const response = await fetch(`/api/friends/requests/${requestId}/accept`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        // Remove the request from the list and reload friends
        setFriendRequests(prev => prev.filter(req => req.id !== requestId));
        loadFriends();
      } else {
        setError('Failed to accept friend request');
      }
    } catch (err) {
      setError('Error accepting friend request');
      console.error('Error accepting friend request:', err);
    }
  };

  const declineFriendRequest = async (requestId: string) => {
    try {
      const response = await fetch(`/api/friends/requests/${requestId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        // Remove the request from the list
        setFriendRequests(prev => prev.filter(req => req.id !== requestId));
      } else {
        setError('Failed to decline friend request');
      }
    } catch (err) {
      setError('Error declining friend request');
      console.error('Error declining friend request:', err);
    }
  };

  const removeFriend = async (friendId: string) => {
    if (!confirm('Are you sure you want to remove this friend?')) {
      return;
    }

    try {
      const response = await fetch('/api/friends/remove', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          friend_id: friendId,
        }),
      });

      if (response.ok) {
        // Remove the friend from the list
        setFriends(prev => prev.filter(friend => friend.id !== friendId));
      } else {
        setError('Failed to remove friend');
      }
    } catch (err) {
      setError('Error removing friend');
      console.error('Error removing friend:', err);
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
          <h2 className="text-lg font-semibold text-text-primary">Friends</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-text-secondary hover:text-text-primary"
          >
            ✕
          </Button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-border">
          <button
            className={`flex-1 px-4 py-2 text-sm font-medium ${
              activeTab === 'friends'
                ? 'text-primary border-b-2 border-primary'
                : 'text-text-secondary hover:text-text-primary'
            }`}
            onClick={() => setActiveTab('friends')}
          >
            Friends ({friends.length})
          </button>
          <button
            className={`flex-1 px-4 py-2 text-sm font-medium ${
              activeTab === 'requests'
                ? 'text-primary border-b-2 border-primary'
                : 'text-text-secondary hover:text-text-primary'
            }`}
            onClick={() => setActiveTab('requests')}
          >
            Requests ({friendRequests.length})
          </button>
          <button
            className={`flex-1 px-4 py-2 text-sm font-medium ${
              activeTab === 'add'
                ? 'text-primary border-b-2 border-primary'
                : 'text-text-secondary hover:text-text-primary'
            }`}
            onClick={() => setActiveTab('add')}
          >
            Add Friend
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4">
          {error && (
            <div className="mb-4 p-3 bg-danger text-white rounded">
              {error}
            </div>
          )}

          {activeTab === 'friends' && (
            <div className="space-y-3">
              {isLoading ? (
                <div className="text-center text-text-secondary">Loading...</div>
              ) : friends.filter(friend => friend.status !== 'offline').length === 0 ? (
                <div className="text-center text-text-secondary">
                  <p>No friends are currently online.</p>
                  <p className="text-sm">Friends will appear here when they are in-game or in a lobby.</p>
                </div>
              ) : (
                friends.filter(friend => friend.status !== 'offline').map((friend) => (
                  <div
                    key={friend.id}
                    className="flex items-center justify-between p-3 bg-background-secondary rounded border"
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
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeFriend(friend.id)}
                      className="text-danger hover:bg-danger hover:text-white"
                    >
                      Remove
                    </Button>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'requests' && (
            <div className="space-y-3">
              {friendRequests.length === 0 ? (
                <div className="text-center text-text-secondary">
                  No pending friend requests
                </div>
              ) : (
                friendRequests.map((request) => (
                  <div
                    key={request.id}
                    className="p-3 bg-background-secondary rounded border"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-3">
                        <div className="w-8 h-8 bg-primary rounded-full flex items-center justify-center text-white font-semibold">
                          {request.requesterName[0].toUpperCase()}
                        </div>
                        <div>
                          <p className="font-medium text-text-primary">
                            {request.requesterName}
                          </p>
                          <p className="text-xs text-text-secondary">
                            {new Date(request.createdAt).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="flex space-x-2">
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={() => acceptFriendRequest(request.id)}
                        className="flex-1"
                      >
                        Accept
                      </Button>
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => declineFriendRequest(request.id)}
                        className="flex-1"
                      >
                        Decline
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'add' && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-text-primary mb-2">
                  Username
                </label>
                <input
                  type="text"
                  value={searchUsername}
                  onChange={(e) => setSearchUsername(e.target.value)}
                  placeholder="Enter username to add as friend"
                  className="w-full px-3 py-2 border border-border rounded bg-background-secondary text-text-primary placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-primary"
                  onKeyPress={(e) => e.key === 'Enter' && sendFriendRequest()}
                />
              </div>
              <Button
                variant="primary"
                className="w-full"
                onClick={sendFriendRequest}
                disabled={isLoading || !searchUsername.trim()}
              >
                {isLoading ? 'Sending...' : 'Send Friend Request'}
              </Button>

              {suggestions.length > 0 && (
                <div className="mt-6">
                  <h3 className="text-sm font-medium text-text-primary mb-3">
                    People You May Know
                  </h3>
                  <div className="space-y-2">
                    {suggestions.map((suggestion) => (
                      <div
                        key={suggestion.id}
                        className="flex items-center justify-between p-2 bg-background-secondary rounded border"
                      >
                        <div className="flex items-center space-x-2">
                          <div className="w-6 h-6 bg-primary rounded-full flex items-center justify-center text-white text-xs font-semibold">
                            {suggestion.name[0].toUpperCase()}
                          </div>
                          <span className="text-sm text-text-primary">{suggestion.name}</span>
                        </div>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => {
                            setSearchUsername(suggestion.name);
                            sendFriendRequest();
                          }}
                          className="text-xs"
                        >
                          Add
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};