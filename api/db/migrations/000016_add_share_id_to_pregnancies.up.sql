-- Add share_id column for unique, URL-safe pregnancy sharing
ALTER TABLE pregnancies ADD COLUMN share_id TEXT;

-- Generate unique share_id for existing pregnancies using SQLite's randomblob
-- This creates a 12-character URL-safe string (8 bytes of randomness)
UPDATE pregnancies 
SET share_id = lower(hex(randomblob(6)))
WHERE share_id IS NULL;

-- Make share_id NOT NULL after populating existing records
-- SQLite doesn't support ALTER COLUMN, so we need to recreate the table
CREATE TABLE pregnancies_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    partner_name TEXT,
    partner_email TEXT,
    due_date DATE NOT NULL,
    conception_date DATE,
    current_week INTEGER DEFAULT 1,
    baby_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    share_id TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Copy data from old table
INSERT INTO pregnancies_new 
SELECT id, user_id, partner_name, partner_email, due_date, conception_date, 
       current_week, baby_name, is_active, share_id, created_at, updated_at
FROM pregnancies;

-- Drop old table and rename new one
DROP TABLE pregnancies;
ALTER TABLE pregnancies_new RENAME TO pregnancies;

-- Recreate the index on the new table
CREATE INDEX idx_pregnancies_share_id ON pregnancies(share_id);
CREATE INDEX idx_pregnancies_user_id ON pregnancies(user_id);