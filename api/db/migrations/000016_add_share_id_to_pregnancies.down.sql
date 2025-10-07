-- Revert to original pregnancies table without share_id
CREATE TABLE pregnancies_original (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    partner_name TEXT,
    partner_email TEXT,
    due_date DATE NOT NULL,
    conception_date DATE,
    current_week INTEGER DEFAULT 1,
    baby_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Copy data from current table (excluding share_id)
INSERT INTO pregnancies_original 
SELECT id, user_id, partner_name, partner_email, due_date, conception_date, 
       current_week, baby_name, is_active, created_at, updated_at
FROM pregnancies;

-- Drop current table and rename original
DROP TABLE pregnancies;
ALTER TABLE pregnancies_original RENAME TO pregnancies;

-- Recreate original index
CREATE INDEX idx_pregnancies_user_id ON pregnancies(user_id);