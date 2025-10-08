CREATE TABLE pregnancy_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pregnancy_id INTEGER NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_title VARCHAR(255) NOT NULL,
    event_description TEXT,
    event_data TEXT, -- JSON as TEXT in SQLite
    week_number INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    
    FOREIGN KEY (pregnancy_id) REFERENCES pregnancies(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_pregnancy_events_pregnancy_id ON pregnancy_events(pregnancy_id);
CREATE INDEX idx_pregnancy_events_type ON pregnancy_events(event_type);
CREATE INDEX idx_pregnancy_events_created_at ON pregnancy_events(created_at DESC);