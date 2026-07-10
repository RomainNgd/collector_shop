ALTER TABLE promotions
    ADD COLUMN seller_id BIGINT,
    ADD CONSTRAINT fk_promotions_seller FOREIGN KEY (seller_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL;

CREATE INDEX idx_promotions_seller_id ON promotions USING btree (seller_id);
