-- name: CreateItem :one
INSERT INTO inventory (sku, item_name, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *;
