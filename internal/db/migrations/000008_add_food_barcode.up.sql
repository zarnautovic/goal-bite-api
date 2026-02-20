ALTER TABLE foods
    ADD COLUMN IF NOT EXISTS barcode TEXT;

ALTER TABLE foods
    ADD CONSTRAINT foods_barcode_format_check CHECK (
        barcode IS NULL OR (
            barcode ~ '^[0-9]+$'
            AND char_length(barcode) IN (8, 12, 13, 14)
        )
    );

CREATE UNIQUE INDEX IF NOT EXISTS idx_foods_barcode_unique
    ON foods(barcode)
    WHERE barcode IS NOT NULL;
