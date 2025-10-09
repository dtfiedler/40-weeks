-- This migration cannot be safely reversed as it corrects the week calculation logic
-- The previous migration (000021) used incorrect week calculations without LMP offset
-- If rollback is needed, you would need to restore from a database backup

-- Placeholder comment to indicate this migration is not reversible
SELECT 'Migration 000022 is not reversible - week numbers were corrected to include LMP offset' AS notice;