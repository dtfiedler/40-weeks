CREATE TABLE village_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pregnancy_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    relationship TEXT NOT NULL, -- e.g., 'mother', 'sister', 'friend', 'coworker'
    is_told BOOLEAN DEFAULT FALSE,
    told_date DATETIME,
    is_subscribed BOOLEAN DEFAULT TRUE,
    unsubscribe_token TEXT UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pregnancy_id) REFERENCES pregnancies (id) ON DELETE CASCADE
);