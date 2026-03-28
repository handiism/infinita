-- name: CreateCategory :exec
INSERT INTO categories (name, normalized_key, description)
VALUES (?, ?, ?);

-- name: GetCategoryByNormalizedKey :one
SELECT id, name, normalized_key, description, created_at
FROM categories
WHERE normalized_key = ?;

-- name: ListCategories :many
SELECT id, name, normalized_key, description, created_at
FROM categories
ORDER BY name COLLATE NOCASE ASC;
