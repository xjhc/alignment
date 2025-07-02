-- Add FTUE (First Time User Experience) settings to players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS disable_loebmate_hints BOOLEAN DEFAULT FALSE;
ALTER TABLE players ADD COLUMN IF NOT EXISTS seen_hints JSONB DEFAULT '{}'::jsonb;