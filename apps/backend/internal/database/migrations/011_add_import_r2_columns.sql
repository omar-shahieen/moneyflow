-- Add R2 storage columns to imports table
ALTER TABLE imports ADD COLUMN storage_key VARCHAR(512);
ALTER TABLE imports ADD COLUMN checksum_sha256 VARCHAR(64);

-- Update status CHECK to include 'queued'
ALTER TABLE imports DROP CONSTRAINT IF EXISTS imports_status_check;
ALTER TABLE imports ADD CONSTRAINT imports_status_check
    CHECK (status IN ('pending', 'queued', 'processing', 'completed', 'failed'));

-- Idempotency: prevent the same file from being processed twice per user
CREATE UNIQUE INDEX idx_imports_user_checksum
    ON imports (user_id, checksum_sha256)
    WHERE checksum_sha256 IS NOT NULL;

---- create above / drop below ----

DROP INDEX IF EXISTS idx_imports_user_checksum;
ALTER TABLE imports DROP CONSTRAINT IF EXISTS imports_status_check;
ALTER TABLE imports ADD CONSTRAINT imports_status_check
    CHECK (status IN ('pending', 'processing', 'completed', 'failed'));
ALTER TABLE imports DROP COLUMN IF EXISTS checksum_sha256;
ALTER TABLE imports DROP COLUMN IF EXISTS storage_key;
