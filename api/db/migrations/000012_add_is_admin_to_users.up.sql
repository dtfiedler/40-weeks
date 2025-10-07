ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;

-- Make the existing admin user an admin
UPDATE users SET is_admin = TRUE WHERE email = 'admin@example.com' OR username = 'admin';