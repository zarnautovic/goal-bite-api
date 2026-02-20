DROP INDEX IF EXISTS idx_foods_barcode_unique;

ALTER TABLE foods
    DROP CONSTRAINT IF EXISTS foods_barcode_format_check;

ALTER TABLE foods
    DROP COLUMN IF EXISTS barcode;
