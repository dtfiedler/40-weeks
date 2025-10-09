-- Update existing pregnancies to calculate and set conception dates
-- Conception date is calculated as 266 days (38 weeks) before due date

UPDATE pregnancies 
SET conception_date = DATE(due_date, '-266 days')
WHERE conception_date IS NULL;