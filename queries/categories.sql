-- name: GetCategories :many
SELECT name FROM categories;

-- name: BulkLinkCategoriesByName :exec
INSERT INTO inventory_categories (item_id, category_id)
SELECT $1, id
FROM categories
WHERE name = ANY(@categories::text[]);

-- name: GetCategoriesByItemId :many
SELECT name FROM categories
LEFT JOIN inventory_categories ON categories.id = inventory_categories.category_id
LEFT JOIN inventory ON inventory_categories.item_id = inventory.id
WHERE inventory.id = @item_id;

-- name: DeleteItemCategories :exec
DELETE FROM inventory_categories
WHERE item_id = $1;
