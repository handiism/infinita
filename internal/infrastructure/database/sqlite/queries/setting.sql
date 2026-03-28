-- name: GetSetting :one
SELECT key, value
FROM settings
WHERE key = ?;

-- name: UpsertSetting :exec
INSERT INTO settings (key, value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;

-- name: GetAnalyticsConsent :one
SELECT id, analytics_opt_in
FROM analytics_consent
WHERE id = 1;

-- name: SetAnalyticsConsent :exec
INSERT INTO analytics_consent (id, analytics_opt_in)
VALUES (1, ?)
ON CONFLICT(id) DO UPDATE SET analytics_opt_in = excluded.analytics_opt_in;
