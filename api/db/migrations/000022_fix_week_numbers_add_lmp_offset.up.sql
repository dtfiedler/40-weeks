-- Fix week numbers by adding 14 days (2 weeks) to account for LMP calculation
-- Gestational age is calculated from Last Menstrual Period (LMP), which is ~14 days before conception
UPDATE pregnancy_updates 
SET week_number = (
    SELECT CASE 
        -- Use update_date if available, otherwise fall back to created_at
        -- Add 14 days (2 weeks) to account for LMP offset
        WHEN pregnancy_updates.update_date IS NOT NULL THEN
            CAST((julianday(pregnancy_updates.update_date) - julianday(pregnancies.conception_date) + 14) / 7 + 1 AS INTEGER)
        ELSE
            CAST((julianday(pregnancy_updates.created_at) - julianday(pregnancies.conception_date) + 14) / 7 + 1 AS INTEGER)
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