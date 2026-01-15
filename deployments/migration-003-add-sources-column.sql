-- Migration 003: Add sources column to servers table
-- For existing production databases

-- Add sources column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='servers' AND column_name='sources'
    ) THEN
        ALTER TABLE servers ADD COLUMN sources TEXT DEFAULT '';
    END IF;
END $$;

-- Create index for sources column for better performance
CREATE INDEX IF NOT EXISTS idx_servers_sources ON servers (sources);

-- Update existing records to have empty sources if NULL
UPDATE servers SET sources = '' WHERE sources IS NULL;
