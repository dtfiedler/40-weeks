-- Add username column back
ALTER TABLE users ADD COLUMN username TEXT;

-- Copy name to username
UPDATE users SET username = name WHERE username IS NULL;

-- Create table with username as NOT NULL
CREATE TABLE users_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data back
INSERT INTO users_old (id, username, password, email, is_admin, created)
SELECT id, name, password, email, is_admin, created FROM users;

-- Drop new table and rename old one
DROP TABLE users;
ALTER TABLE users_old RENAME TO users;