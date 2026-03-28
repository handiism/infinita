-- name: UpsertBudget :exec
INSERT INTO budgets (category_id, month, monthly_limit_minor, updated_at)
VALUES (?, ?, ?, datetime('now'))
ON CONFLICT(category_id, month) DO UPDATE SET
    monthly_limit_minor = excluded.monthly_limit_minor,
    updated_at = datetime('now');

-- name: GetBudgetsForMonth :many
SELECT
    c.name AS category_name,
    c.normalized_key AS category_key,
    b.month AS month,
    b.monthly_limit_minor AS monthly_limit_minor,
    COALESCE((
        SELECT SUM(amount_minor)
        FROM transactions t
        WHERE t.category_id = b.category_id
          AND t.type = 'expense'
          AND substr(t.date, 1, 7) = b.month
    ), 0) AS spent_month_to_date_minor
FROM budgets b
JOIN categories c ON c.id = b.category_id
WHERE b.month = ?
ORDER BY c.name COLLATE NOCASE ASC;
