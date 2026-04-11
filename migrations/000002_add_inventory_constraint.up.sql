ALTER TABLE inventory ADD CONSTRAINT positive_quantity CHECK (quantity >= 0);

ALTER TABLE inventory ADD CONSTRAINT positive_price CHECK (price >= 0);
