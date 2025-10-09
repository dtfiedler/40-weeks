-- This migration cannot be safely reversed as we don't have the original week numbers
-- The week numbers were recalculated based on conception dates, and we don't store the original values
-- If rollback is needed, you would need to restore from a database backup

-- Placeholder comment to indicate this migration is not reversible
SELECT 'Migration 000021 is not reversible - week numbers were recalculated based on conception dates' AS notice;