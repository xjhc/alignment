-- Create friend_requests table
CREATE TABLE IF NOT EXISTS friend_requests (
    id VARCHAR(36) PRIMARY KEY,
    requester_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    recipient_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'accepted', 'declined'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Prevent duplicate requests
    UNIQUE(requester_id, recipient_id)
);

-- Create friends table (bidirectional friendship)
CREATE TABLE IF NOT EXISTS friends (
    id VARCHAR(36) PRIMARY KEY,
    player1_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    player2_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure friendship is unique and bidirectional
    UNIQUE(player1_id, player2_id),
    -- Prevent self-friendship
    CHECK (player1_id != player2_id),
    -- Ensure consistent ordering (player1_id < player2_id) to avoid duplicates
    CHECK (player1_id < player2_id)
);

-- Create player_presence table for online status tracking
CREATE TABLE IF NOT EXISTS player_presence (
    player_id VARCHAR(36) PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'offline',  -- 'offline', 'online', 'in_lobby', 'in_game'
    lobby_id VARCHAR(36),
    game_id VARCHAR(36),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_friend_requests_requester ON friend_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_recipient ON friend_requests(recipient_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_status ON friend_requests(status);

CREATE INDEX IF NOT EXISTS idx_friends_player1 ON friends(player1_id);
CREATE INDEX IF NOT EXISTS idx_friends_player2 ON friends(player2_id);

CREATE INDEX IF NOT EXISTS idx_player_presence_status ON player_presence(status);
CREATE INDEX IF NOT EXISTS idx_player_presence_last_seen ON player_presence(last_seen);

-- Create trigger for friend_requests updated_at
CREATE TRIGGER update_friend_requests_updated_at 
    BEFORE UPDATE ON friend_requests 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for player_presence updated_at
CREATE TRIGGER update_player_presence_updated_at 
    BEFORE UPDATE ON player_presence 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();