-- Create players table
CREATE TABLE IF NOT EXISTS players (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Profile customization
    equipped_avatar VARCHAR(50) DEFAULT 'default',
    equipped_title VARCHAR(50) DEFAULT '',
    
    -- Statistics
    total_games_played INTEGER DEFAULT 0,
    total_games_won INTEGER DEFAULT 0,
    total_tokens_mined INTEGER DEFAULT 0,
    kudos_received INTEGER DEFAULT 0,
    
    -- Unlocked content (stored as JSON arrays)
    unlocked_achievements JSONB DEFAULT '[]'::jsonb,
    unlocked_avatars JSONB DEFAULT '["default"]'::jsonb,
    unlocked_titles JSONB DEFAULT '[]'::jsonb,
    
    -- Social features
    blocked_players JSONB DEFAULT '[]'::jsonb
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_email ON players(email);
CREATE INDEX IF NOT EXISTS idx_players_last_active ON players(last_active_at);

-- Create game_history table
CREATE TABLE IF NOT EXISTS game_history (
    id VARCHAR(36) PRIMARY KEY,
    game_id VARCHAR(36) NOT NULL,
    player_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    alignment VARCHAR(20) NOT NULL,
    winner_faction VARCHAR(20) NOT NULL,
    is_winner BOOLEAN NOT NULL,
    tokens_mined INTEGER DEFAULT 0,
    days_survived INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Performance metrics
    correct_votes INTEGER DEFAULT 0,
    conversions INTEGER DEFAULT 0,
    abilities_used INTEGER DEFAULT 0
);

-- Create indexes for game_history
CREATE INDEX IF NOT EXISTS idx_game_history_player_id ON game_history(player_id);
CREATE INDEX IF NOT EXISTS idx_game_history_game_id ON game_history(game_id);
CREATE INDEX IF NOT EXISTS idx_game_history_created_at ON game_history(created_at);

-- Create achievements table
CREATE TABLE IF NOT EXISTS achievements (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    icon_url VARCHAR(255),
    rarity VARCHAR(20) DEFAULT 'common',
    unlock_criteria JSONB NOT NULL,
    avatar_reward VARCHAR(50),
    title_reward VARCHAR(50)
);

-- Create player_reports table
CREATE TABLE IF NOT EXISTS player_reports (
    id VARCHAR(36) PRIMARY KEY,
    reporter_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    reported_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    game_id VARCHAR(36),
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by VARCHAR(36)
);

-- Create indexes for player_reports
CREATE INDEX IF NOT EXISTS idx_player_reports_reporter ON player_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_player_reports_reported ON player_reports(reported_id);
CREATE INDEX IF NOT EXISTS idx_player_reports_status ON player_reports(status);

-- Create kudos_given table
CREATE TABLE IF NOT EXISTS kudos_given (
    id VARCHAR(36) PRIMARY KEY,
    giver_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    receiver_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    game_id VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure one kudos per game per player pair
    UNIQUE(giver_id, receiver_id, game_id)
);

-- Create indexes for kudos_given
CREATE INDEX IF NOT EXISTS idx_kudos_given_giver ON kudos_given(giver_id);
CREATE INDEX IF NOT EXISTS idx_kudos_given_receiver ON kudos_given(receiver_id);
CREATE INDEX IF NOT EXISTS idx_kudos_given_game ON kudos_given(game_id);

-- Create avatars table
CREATE TABLE IF NOT EXISTS avatars (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon_emoji VARCHAR(10) NOT NULL,
    description TEXT,
    rarity VARCHAR(20) DEFAULT 'common',
    unlock_type VARCHAR(20) DEFAULT 'default'
);

-- Create titles table
CREATE TABLE IF NOT EXISTS titles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#ffffff',
    rarity VARCHAR(20) DEFAULT 'common',
    unlock_type VARCHAR(20) DEFAULT 'default'
);

-- Insert default avatars
INSERT INTO avatars (id, name, icon_emoji, description, rarity, unlock_type) VALUES
('default', 'Default Avatar', '👤', 'The standard corporate headshot', 'common', 'default'),
('scientist', 'Scientist', '🧑‍🔬', 'For the analytically minded', 'common', 'achievement'),
('detective', 'Detective', '🕵️', 'Trust but verify', 'rare', 'achievement'),
('robot', 'Robot', '🤖', 'Embrace the silicon future', 'epic', 'achievement'),
('crown', 'Executive', '👑', 'Leadership material', 'legendary', 'achievement');

-- Insert default titles
INSERT INTO titles (id, name, description, color, rarity, unlock_type) VALUES
('rookie', 'Rookie', 'New to the corporate world', '#ffffff', 'common', 'default'),
('veteran', 'Veteran', 'Seasoned corporate warrior', '#4ade80', 'rare', 'achievement'),
('mastermind', 'Mastermind', 'Strategic genius', '#8b5cf6', 'epic', 'achievement'),
('legend', 'Legend', 'Corporate hall of fame', '#f59e0b', 'legendary', 'achievement');

-- Insert sample achievements
INSERT INTO achievements (id, name, description, rarity, unlock_criteria, avatar_reward, title_reward) VALUES
('first_win', 'First Victory', 'Win your first game', 'common', '{"wins": 1}', NULL, NULL),
('win_as_ai', 'Master of Deception', 'Win a game as the Original AI without receiving any votes', 'rare', '{"win_as_ai_no_votes": true}', 'robot', NULL),
('survive_5_days', 'Survivor', 'Survive 5 days in a single game', 'rare', '{"days_survived": 5}', NULL, 'veteran'),
('perfect_detective', 'Perfect Detective', 'Vote correctly in all elimination rounds in a winning game', 'epic', '{"perfect_voting": true}', 'detective', NULL),
('social_butterfly', 'Social Butterfly', 'Receive 50 kudos from other players', 'epic', '{"kudos_received": 50}', NULL, 'mastermind'),
('legend_status', 'Corporate Legend', 'Win 100 games', 'legendary', '{"wins": 100}', 'crown', 'legend');

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for players table
CREATE TRIGGER update_players_updated_at 
    BEFORE UPDATE ON players 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();