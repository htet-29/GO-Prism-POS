CREATE INDEX IF NOT EXISTS idx_items_name ON inventory USING gin (item_name gin_trgm_ops);
