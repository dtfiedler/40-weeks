CREATE TABLE milestones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pregnancy_id INTEGER NOT NULL,
    milestone_type TEXT NOT NULL, -- 'first_appointment', '12_week_scan', '20_week_scan', '36_week_appointment', 'due_date', 'induction_scheduled'
    title TEXT NOT NULL,
    scheduled_date DATE,
    completed_date DATE,
    is_completed BOOLEAN DEFAULT FALSE,
    notes TEXT,
    week_number INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pregnancy_id) REFERENCES pregnancies (id) ON DELETE CASCADE
);