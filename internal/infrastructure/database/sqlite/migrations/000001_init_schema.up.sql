PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    normalized_key TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
    amount_minor INTEGER NOT NULL CHECK(amount_minor > 0),
    currency_code TEXT NOT NULL DEFAULT 'IDR',
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    category_name_snapshot TEXT NOT NULL,
    date TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_date_created ON transactions(date DESC, created_at DESC);

CREATE TABLE IF NOT EXISTS budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    month TEXT NOT NULL,
    monthly_limit_minor INTEGER NOT NULL CHECK(monthly_limit_minor > 0),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(category_id, month)
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS initial_balance (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    initial_balance_minor INTEGER NOT NULL CHECK(initial_balance_minor >= 0),
    currency_code TEXT NOT NULL DEFAULT 'IDR',
    initialized_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO categories (name, normalized_key, description)
VALUES
    ('Food', 'food', 'Everyday meals and dining expenses'),
    ('Transport', 'transport', 'Public and private transport costs'),
    ('Utilities', 'utilities', 'Utilities and recurring services'),
    ('Housing', 'housing', 'Rent, mortgage, and housing-related costs'),
    ('Savings', 'savings', 'Allocated savings and investments');

INSERT OR IGNORE INTO settings (key, value)
VALUES
    ('storage_mode', 'local'),
    ('report_timezone', 'Asia/Jakarta');

INSERT OR IGNORE INTO initial_balance (id, initial_balance_minor)
VALUES (1, 0);
