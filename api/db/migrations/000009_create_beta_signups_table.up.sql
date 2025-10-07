CREATE TABLE beta_signups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,
    ip_address TEXT
);

CREATE INDEX idx_beta_signups_email ON beta_signups(email);
CREATE INDEX idx_beta_signups_created_at ON beta_signups(created_at);