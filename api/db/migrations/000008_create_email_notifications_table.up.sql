CREATE TABLE email_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pregnancy_id INTEGER NOT NULL,
    village_member_id INTEGER NOT NULL,
    update_id INTEGER,
    milestone_id INTEGER,
    email_type TEXT NOT NULL, -- 'update', 'milestone', 'announcement', 'welcome'
    subject TEXT NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delivery_status TEXT DEFAULT 'sent', -- 'sent', 'delivered', 'bounced', 'failed'
    ses_message_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pregnancy_id) REFERENCES pregnancies (id) ON DELETE CASCADE,
    FOREIGN KEY (village_member_id) REFERENCES village_members (id) ON DELETE CASCADE,
    FOREIGN KEY (update_id) REFERENCES pregnancy_updates (id) ON DELETE SET NULL,
    FOREIGN KEY (milestone_id) REFERENCES milestones (id) ON DELETE SET NULL
);