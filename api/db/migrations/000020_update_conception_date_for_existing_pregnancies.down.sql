-- Rollback: Set conception_date back to NULL for pregnancies that didn't have it originally
-- This is a safe rollback since we're only clearing conception dates that were calculated

UPDATE pregnancies 
SET conception_date = NULL
WHERE conception_date = DATE(due_date, '-266 days');