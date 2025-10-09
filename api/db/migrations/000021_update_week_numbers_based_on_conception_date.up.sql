-- Update week numbers for all pregnancy updates based on conception dates
UPDATE pregnancy_updates 
SET week_number = (
    SELECT CASE 
        -- Use update_date if available, otherwise fall back to created_at
        WHEN pregnancy_updates.update_date IS NOT NULL THEN
            CAST((julianday(pregnancy_updates.update_date) - julianday(pregnancies.conception_date)) / 7 + 1 AS INTEGER)
        ELSE
            CAST((julianday(pregnancy_updates.created_at) - julianday(pregnancies.conception_date)) / 7 + 1 AS INTEGER)
    END
    FROM pregnancies 
    WHERE pregnancies.id = pregnancy_updates.pregnancy_id
    AND pregnancies.conception_date IS NOT NULL
)
WHERE EXISTS (
    SELECT 1 FROM pregnancies 
    WHERE pregnancies.id = pregnancy_updates.pregnancy_id 
    AND pregnancies.conception_date IS NOT NULL
);