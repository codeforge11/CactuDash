## [0.1.0a]

### Changed
- Remove sudo from clearOldLogs command.

### Added
- Add logout function.
- After reboot or shutdown, the user will be redirected to the login page and the cookies will be deleted.
- Add log creation after WebSocket connection is closed.

### Performance
- Optimize variable.

### Fixed
- Remove unused WebSocket message handler.
- Remove non-executable error sending.