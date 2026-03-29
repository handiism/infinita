-- name: CreateTransaction :exec
INSERT INTO transactions (id, type, amount_minor, currency_code, category_id, category_name_snapshot, date, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListTransactions :many
SELECT t.id, t.type, t.amount_minor, t.currency_code, t.category_id, t.category_name_snapshot, t.date, t.description, t.created_at
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE (?1 IS NULL OR c.normalized_key = ?1)
ORDER BY t.date DESC, t.created_at DESC
LIMIT ?2 OFFSET ?3;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE (?1 IS NULL OR c.normalized_key = ?1);

-- name: SumTransactionTotalsForDay :one
SELECT
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_total_minor,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_total_minor
FROM transactions
WHERE date = ?;

-- name: SumTransactionTotalsForMonth :one
SELECT
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_total_minor,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_total_minor
FROM transactions
WHERE substr(date, 1, 7) = ?;

-- name: SumCumulativeTotalsUpToDate :one
SELECT
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_total_minor,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_total_minor
FROM transactions
WHERE date <= ?;

-- name: TopCategoriesForMonth :many
SELECT
    c.name AS category_name,
    c.normalized_key AS category_key,
    COALESCE(SUM(t.amount_minor), 0) AS amount_minor
FROM categories c
LEFT JOIN transactions t ON t.category_id = c.id AND t.type = 'expense' AND substr(t.date, 1, 7) = ?
GROUP BY c.id
HAVING SUM(t.amount_minor) > 0
ORDER BY amount_minor DESC, category_key ASC;
