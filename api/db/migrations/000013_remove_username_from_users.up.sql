-- Add name column if it doesn't exist
ALTER TABLE users ADD COLUMN name TEXT;

-- Copy username to name for existing users
UPDATE users SET name = username WHERE name IS NULL;

-- Make name NOT NULL (SQLite doesn't support adding NOT NULL directly)
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data to new table
INSERT INTO users_new (id, name, password, email, is_admin, created)
SELECT id, name, password, email, is_admin, created FROM users;

-- Drop old table and rename new one
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;