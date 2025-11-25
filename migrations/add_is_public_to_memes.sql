
ALTER TABLE memes ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT true;

-- Create index for better query performance on public memes
CREATE INDEX IF NOT EXISTS idx_memes_is_public ON memes(is_public);

-- Create composite index for public memes ordered by creation date
CREATE INDEX IF NOT EXISTS idx_memes_public_created ON memes(is_public, created_at DESC) WHERE is_public = true;

-- Update existing memes to be public by default
UPDATE memes SET is_public = true WHERE is_public IS NULL;
