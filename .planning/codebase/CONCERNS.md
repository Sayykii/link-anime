# Codebase Concerns

**Analysis Date:** 2026-02-22

## Tech Debt

**In-Memory Session Storage:**
- Issue: Sessions are stored in a Go map (`internal/auth/auth.go` lines 24-26), not persisted. Server restart logs out all users.
- Files: `internal/auth/auth.go`
- Impact: Users must re-authenticate after every container restart or deploy
- Fix approach: Store sessions in SQLite alongside other data, or use signed JWTs

**Hardcoded Default Password:**
- Issue: Default password "changeme" is set in config if LA_PASSWORD env is not provided
- Files: `internal/config/config.go` line 40
- Impact: Risk of deployment with weak default credentials
- Fix approach: Require password to be explicitly set, refuse to start without it, or prompt on first run

**Notification Errors Silently Ignored:**
- Issue: `client.Do(req)` calls in notify package discard errors with `//nolint:errcheck`
- Files: `internal/notify/notify.go` lines 59, 120, 141
- Impact: Failed notifications are silent - user never knows if Discord/ntfy webhook failed
- Fix approach: Log notification errors, optionally track failure counts

**Linker Uses Reused Field Name:**
- Issue: `result.Linked` is reused to mean "removed count" in Unlink operations (line 339, 437)
- Files: `internal/linker/linker.go` lines 339, 437
- Impact: Confusing semantics, easy to introduce bugs
- Fix approach: Create separate `Removed` field or dedicated UnlinkResult struct

**Global Database Variable:**
- Issue: `database.DB` is a package-level global, making testing and dependency injection harder
- Files: `internal/database/database.go` line 13
- Impact: Difficult to unit test database-dependent code in isolation
- Fix approach: Pass DB connection through dependency injection, or use a DB interface

## Known Bugs

**RSS Poller Doesn't Detect Re-uploads:**
- Symptoms: If the same title is uploaded again with different quality/group, it won't be matched again
- Files: `internal/rss/poller.go` lines 127-133
- Trigger: Title hash is used for deduplication, not info hash
- Workaround: Manually download re-uploads via Nyaa search tab

**WebSocket Origin Check Allows All:**
- Symptoms: Any origin can establish WebSocket connections
- Files: `internal/ws/hub.go` lines 12-15
- Trigger: `CheckOrigin: func(r *http.Request) bool { return true }`
- Workaround: Rely on session auth middleware for protection

## Security Considerations

**Cookie Missing Secure Flag in Production:**
- Risk: Session cookie may be sent over HTTP if accessed without HTTPS
- Files: `internal/auth/auth.go` lines 117-126
- Current mitigation: HttpOnly and SameSite=Lax are set
- Recommendations: Add Secure flag when running behind HTTPS reverse proxy, or detect scheme

**qBittorrent Credentials Stored in Plain Text:**
- Risk: Database stores qBit password unencrypted
- Files: `internal/api/settings_handler.go` line 58, `internal/database/database.go`
- Current mitigation: Database file permissions, UI masks password display
- Recommendations: Encrypt sensitive settings values at rest

**No Rate Limiting on Login:**
- Risk: Brute-force password attacks possible
- Files: `internal/api/auth_handler.go`
- Current mitigation: None
- Recommendations: Add rate limiting middleware, account lockout after failed attempts

**Session Token Length:**
- Risk: 32 bytes (256 bits) is secure, but token isn't rotated on activity
- Files: `internal/auth/auth.go` line 66
- Current mitigation: 24-hour expiry
- Recommendations: Consider session rotation, sliding expiry

## Performance Bottlenecks

**Library Scan Walks Entire Tree:**
- Problem: `ScanLibrary()` and `LibrarySize()` walk all files synchronously
- Files: `internal/scanner/scanner.go` lines 35-98, 179-193
- Cause: No caching, full filesystem walk on every request
- Improvement path: Cache library metadata in DB, use inotify/fsnotify for incremental updates

**Torrent Polling Every 5 Seconds:**
- Problem: Fixed 5-second poll interval regardless of download activity
- Files: `cmd/server/main.go` line 95, `internal/monitor/monitor.go`
- Cause: Simple timer-based polling
- Improvement path: Adaptive polling - poll faster when downloads active, slower when idle

**RSS Poller Runs All Rules Serially:**
- Problem: Each RSS rule is checked sequentially with network requests
- Files: `internal/rss/poller.go` lines 94-107
- Cause: Simple loop over rules
- Improvement path: Parallelize rule checking with worker pool

## Fragile Areas

**Release Name Parser:**
- Files: `internal/parser/parser.go`
- Why fragile: Complex regex-based parsing of highly variable torrent naming conventions
- Safe modification: Extend test cases in `internal/parser/parser_test.go` before changes
- Test coverage: Good (213 lines of tests), but anime naming is infinitely variable

**Linker Multi-Season Detection:**
- Files: `internal/linker/linker.go` lines 34-46, 151-188
- Why fragile: Season directory detection relies on naming patterns, edge cases abound
- Safe modification: Test with actual directory structures, add more FindSeasonDirs tests
- Test coverage: No unit tests for linker package

**WebSocket State Synchronization:**
- Files: `internal/ws/hub.go`, `frontend/src/composables/useWebSocket.ts`
- Why fragile: Client reconnection can miss state updates, no message queuing
- Safe modification: Test reconnection scenarios manually
- Test coverage: None

## Scaling Limits

**Single SQLite Connection:**
- Current capacity: SetMaxOpenConns(1) - one connection
- Limit: Concurrent write operations will serialize/queue
- Scaling path: Appropriate for single-user homelab; for multi-user, consider PostgreSQL

**In-Memory Session Map:**
- Current capacity: Thousands of sessions (Go map)
- Limit: Memory-bound, lost on restart
- Scaling path: Move to database-backed sessions

**WebSocket Client Buffer:**
- Current capacity: 256 messages per client (`internal/ws/hub.go` line 49)
- Limit: Slow clients will be disconnected when buffer fills
- Scaling path: Increase buffer or implement backpressure

## Dependencies at Risk

**Nyaa.si Scraping:**
- Risk: Direct HTML scraping of Nyaa, no official API
- Impact: Any Nyaa HTML change breaks search
- Migration plan: Switch to RSS-only mode, or use alternative sources

**modernc.org/sqlite:**
- Risk: Pure-Go SQLite, less battle-tested than mattn/go-sqlite3
- Impact: Potential edge cases in complex queries or under load
- Migration plan: Swap to mattn/go-sqlite3 if CGO is acceptable

## Missing Critical Features

**No Backup/Export:**
- Problem: No way to backup history, RSS rules, settings
- Blocks: Disaster recovery, migration to new server

**No Multi-User Support:**
- Problem: Single shared password for all users
- Blocks: Household sharing with per-user tracking

**No Automatic Linking:**
- Problem: RSS matches download but don't auto-link to library
- Blocks: Fully automated anime acquisition pipeline

## Test Coverage Gaps

**No Backend Unit Tests Except Parser:**
- What's not tested: linker, scanner, auth, database, API handlers, qbit client, shoko client
- Files: Only `internal/parser/parser_test.go` exists
- Risk: Regressions in core linking/unlinking logic undetected
- Priority: High - linker.go is 560 lines with complex filesystem operations

**No Frontend Tests:**
- What's not tested: Vue components, stores, composables
- Files: No test files in `frontend/`
- Risk: UI regressions, broken user flows
- Priority: Medium - UI is relatively simple

**No Integration Tests:**
- What's not tested: Full API flows, database interactions, WebSocket communication
- Files: None
- Risk: End-to-end breakage not caught before deploy
- Priority: Medium - manual testing currently covers this

**No CI/CD Pipeline:**
- What's not tested: Automated test runs, linting, build verification
- Files: No `.github/workflows/` or similar
- Risk: Broken builds can be pushed to main
- Priority: Low for homelab use, high for collaboration

---

*Concerns audit: 2026-02-22*
