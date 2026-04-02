-- name: GetInitialBalance :one
SELECT id, initial_balance_minor, currency_code, initialized_at
FROM initial_balance
WHERE id = 1;

-- name: UpsertInitialBalance :one
INSERT INTO initial_balance (id, initial_balance_minor, currency_code, initialized_at)
VALUES (1, ?, ?, datetime('now'))
ON CONFLICT(id) DO UPDATE SET
    initial_balance_minor = excluded.initial_balance_minor,
    currency_code = excluded.currency_code,
    initialized_at = datetime('now')
RETURNING id, initial_balance_minor, currency_code, initialized_at;
