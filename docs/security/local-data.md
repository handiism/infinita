# Local Data

## Storage Mode
- Infinita MVP supports **local** and **remote** storage modes.
- Use `--mode` flag to specify mode (local or remote).
- Use `--server-url` flag to specify server URL for remote mode.
- Use `--api-key` flag to set API key for remote mode authentication.
- Local mode stores all data on the user's device.
- Remote mode connects to a remote server at the specified URL.

## API Key Authentication
- API key is stored in `settings.yaml` as `api_key`.
- Required for remote mode; optional for local mode.
- Server validates `X-API-Key` header on all endpoints except `/health` when api_key is configured.
- API key is masked in API responses (shown as `****`).
- Never commit api_key to version control.

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
- All data stays local on the user's device with no remote transmission.
