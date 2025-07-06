-- Create parties table
CREATE TABLE IF NOT EXISTS parties (
    id VARCHAR(36) PRIMARY KEY,
    leader_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create party_members table
CREATE TABLE IF NOT EXISTS party_members (
    party_id VARCHAR(36) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    player_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (party_id, player_id)
);

-- Create party_invites table
CREATE TABLE IF NOT EXISTS party_invites (
    id VARCHAR(36) PRIMARY KEY,
    party_id VARCHAR(36) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    inviter_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    invitee_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_party_members_party_id ON party_members(party_id);
CREATE INDEX IF NOT EXISTS idx_party_members_player_id ON party_members(player_id);
CREATE INDEX IF NOT EXISTS idx_party_invites_invitee_id ON party_invites(invitee_id);