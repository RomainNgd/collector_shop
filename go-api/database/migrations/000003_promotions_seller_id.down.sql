DROP INDEX IF EXISTS idx_promotions_seller_id;

ALTER TABLE promotions
    DROP CONSTRAINT IF EXISTS fk_promotions_seller,
    DROP COLUMN IF EXISTS seller_id;
