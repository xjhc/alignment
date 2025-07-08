import React, { useState, useEffect, useRef } from 'react';
import { Button } from './ui';

interface LobbyMessage {
  id: string;
  playerId: string;
  playerName: string;
  message: string;
  timestamp: Date;
  type: 'message' | 'system' | 'join' | 'leave';
}

interface LobbyChatProps {
  gameId: string;
  playerId: string;
  playerName: string;
  messages: LobbyMessage[];
  onSendMessage: (message: string) => void;
  isConnected: boolean;
}

export const LobbyChat: React.FC<LobbyChatProps> = ({
  gameId,
  playerId,
  playerName,
  messages,
  onSendMessage,
  isConnected
}) => {
  const [inputMessage, setInputMessage] = useState('');
  const [isExpanded, setIsExpanded] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  // Focus input when chat is expanded
  useEffect(() => {
    if (isExpanded && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isExpanded]);

  const handleSendMessage = () => {
    if (inputMessage.trim() && isConnected) {
      onSendMessage(inputMessage.trim());
      setInputMessage('');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const formatTime = (timestamp: Date) => {
    return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const getMessageStyle = (message: LobbyMessage) => {
    switch (message.type) {
      case 'system':
        return 'bg-background-tertiary border-border text-text-secondary italic';
      case 'join':
        return 'bg-success/10 border-success/20 text-success';
      case 'leave':
        return 'bg-danger/10 border-danger/20 text-danger';
      default:
        return message.playerId === playerId 
          ? 'bg-primary/10 border-primary/20 text-text-primary'
          : 'bg-background-primary border-border text-text-primary';
    }
  };

  const unreadCount = messages.filter(m => 
    m.playerId !== playerId && 
    m.timestamp > new Date(Date.now() - 30000) // Messages from last 30 seconds
  ).length;

  return (
    <div className={`bg-background-secondary border border-border rounded-lg transition-all duration-300 ${
      isExpanded ? 'w-full h-96' : 'w-full h-12'
    }`}>
      {/* Header */}
      <div 
        className="flex items-center justify-between p-3 cursor-pointer hover:bg-background-tertiary rounded-t-lg"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center gap-2">
          <div className="text-sm">💬</div>
          <span className="text-sm font-medium text-text-primary">Lobby Chat</span>
          {!isExpanded && unreadCount > 0 && (
            <span className="bg-primary text-background-primary text-xs px-2 py-1 rounded-full">
              {unreadCount}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${
            isConnected ? 'bg-success' : 'bg-danger'
          }`} />
          <div className="text-text-muted text-sm">
            {isExpanded ? '▼' : '▶'}
          </div>
        </div>
      </div>

      {/* Chat Messages */}
      {isExpanded && (
        <>
          <div className="flex-1 h-64 overflow-y-auto p-3 pt-0 space-y-2">
            {messages.length === 0 ? (
              <div className="flex items-center justify-center h-full text-text-muted">
                <div className="text-center">
                  <div className="text-2xl mb-2">👫</div>
                  <p className="text-sm">Start a conversation!</p>
                </div>
              </div>
            ) : (
              messages.map((message) => (
                <div 
                  key={message.id} 
                  className={`p-2 rounded-lg border text-sm ${getMessageStyle(message)} animation-fade-in`}
                >
                  {message.type === 'message' ? (
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-xs">
                          {message.playerId === playerId ? 'You' : message.playerName}
                        </span>
                        <span className="text-xs text-text-muted">
                          {formatTime(message.timestamp)}
                        </span>
                      </div>
                      <div className="leading-relaxed">
                        {message.message}
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center gap-2">
                      <span className="text-xs">
                        {formatTime(message.timestamp)}
                      </span>
                      <span className="text-xs">
                        {message.message}
                      </span>
                    </div>
                  )}
                </div>
              ))
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Input Area */}
          <div className="border-t border-border p-3">
            <div className="flex gap-2">
              <input
                ref={inputRef}
                type="text"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder={isConnected ? "Type a message..." : "Connecting..."}
                disabled={!isConnected}
                className="flex-1 px-3 py-2 bg-background-primary border border-border rounded text-sm text-text-primary placeholder-text-muted disabled:opacity-50"
              />
              <Button
                onClick={handleSendMessage}
                disabled={!inputMessage.trim() || !isConnected}
                variant="primary"
                size="sm"
                className="px-4"
              >
                Send
              </Button>
            </div>
            <div className="text-xs text-text-muted mt-2 text-center">
              {isConnected ? (
                <span>Press Enter to send • Keep it friendly!</span>
              ) : (
                <span>Reconnecting to chat...</span>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
};