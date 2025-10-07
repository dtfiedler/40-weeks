CREATE TABLE pregnancy_updates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pregnancy_id INTEGER NOT NULL,
    week_number INTEGER,
    title TEXT NOT NULL,
    content TEXT,
    update_type TEXT DEFAULT 'general', -- 'general', 'appointment', 'milestone', 'photo'
    appointment_type TEXT, -- '20_week_scan', '36_week_checkup', 'first_appointment', etc.
    is_shared BOOLEAN DEFAULT FALSE,
    shared_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pregnancy_id) REFERENCES pregnancies (id) ON DELETE CASCADE
);