# Changelog

All notable changes to MuxueTools will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.3.1] - 2026-01-21

### Added

- **Desktop Fixed Port**: Desktop version now uses configured port (default: 8080) instead of random port
  - Third-party apps (Cursor, etc.) only need one-time configuration
  - Falls back to random port if configured port is unavailable
  - Displays `127.0.0.1` instead of `localhost` for better compatibility

- **Server Port Settings UI**: Added port configuration in Settings → Security tab
  - Port range: 1024-65535
  - Restart required after changing port
  - Full i18n support (Chinese, English, Japanese)

- **Documentation**: Added Desktop fixed port configuration guide to README

### Changed

- Dashboard API endpoint now displays `http://127.0.0.1:PORT/v1` format
- Save confirmation now shows restart reminder when port is modified

### Technical

- Added `ServerConfigUpdate` type in backend for port configuration updates
- Added `stored_port` field to config API response
- Updated `cmd/desktop/main.go` with port fallback logic

---

## [0.3.0] - 2026-01-20

### Added

- **Stream Output Toggle**: Added setting to enable/disable streaming responses
  - Located in Settings → Model tab
  - Default: enabled (streaming)
  - Full i18n support

### Technical

- Added `stream_output` field to ModelSettingsConfig
- Frontend conditionally calls `streamChatCompletion` or `chatCompletion` based on setting

---

## [0.2.0] - 2026-01-19

### Added

- Multi-language support (Chinese, English, Japanese)
- Chat session persistence with SQLite storage
- Statistics dashboard with charts
- API key management with import/export

### Changed

- Renamed project from MxlnAPI to MuxueTools
- Updated UI to Claude-style design

---

## [0.1.0] - 2026-01-15

### Added

- Initial release
- OpenAI-compatible Gemini API proxy
- Basic chat interface
- Key pool management with multiple strategies
- Auto-update detection
