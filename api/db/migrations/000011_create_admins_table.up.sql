CREATE TABLE admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Make the existing admin user an admin
INSERT INTO admins (user_id) 
SELECT id FROM users WHERE username = 'admin' LIMIT 1;