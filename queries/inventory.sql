-- name: CreateItem :one
INSERT INTO inventory (sku, item_name, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetItemByID :one
SELECT * FROM inventory 
WHERE id = $1;

-- name: UpdateItem :one
UPDATE inventory
SET sku  = $1, item_name = $2, quantity = $3, price = $4, version = version + 1
WHERE id = $5 AND version = $6
RETURNING *;
