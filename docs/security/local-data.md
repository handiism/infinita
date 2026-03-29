# Local Data

## Storage Mode
- Infinita MVP uses **local-only** storage.
- Remote storage modes are not supported.

## Default Data Path
- Override path for tests or custom setups with `INFINITA_DATA_DIR`.
- Default runtime path uses the current user's config directory:
  - macOS: `~/Library/Application Support/infinita`
  - Linux: `$XDG_CONFIG_HOME/infinita` or `~/.config/infinita`
  - Windows: `%AppData%\infinita`

## Stored Files
- SQLite database file: `infinita.db`
- Database schema is created and migrated automatically on startup.

## Permission Model
- Data directory is created with user-only permissions where supported: `0700`.
- SQLite database file is set to user read/write only where supported: `0600`.

## Backup Guidance
- Exit the CLI before copying files.
- Back up the full data directory, especially `infinita.db`.
- Restore by replacing the data directory with the backup copy.

## Manual Deletion
- Exit the CLI.
- Remove the configured data directory or delete `infinita.db`.
- This permanently removes local financial data for the MVP installation.

## Privacy Notes
- Telemetry is off by default.
- Analytics consent is stored locally and can be reviewed or revoked with `infinita settings show` and `infinita settings analytics --opt-in <true|false>`.
