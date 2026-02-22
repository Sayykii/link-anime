# Testing Patterns

**Analysis Date:** 2026-02-22

## Test Framework

### Go Backend

**Runner:**
- Go standard `testing` package
- No external testing framework

**Assertion Library:**
- Standard `t.Errorf()` assertions
- No assertion libraries (testify, etc.)

**Run Commands:**
```bash
go test ./...                    # Run all tests
go test ./internal/parser/...    # Run specific package tests
go test -v ./...                 # Verbose output
go test -cover ./...             # With coverage
```

### TypeScript Frontend

**Runner:**
- No test framework configured
- No `jest.config.*` or `vitest.config.*` files present
- `package.json` contains no test scripts

**Status:** Frontend testing is not implemented

## Test File Organization

**Go Backend Location:**
- Co-located with source files
- Same package (no `_test` package separation)

**Go Naming:**
- `*_test.go` suffix
- Example: `parser.go` → `parser_test.go`

**Current Test Coverage:**
```
internal/parser/parser_test.go   # Only test file in codebase
```

## Test Structure

**Go Suite Organization:**
```go
func TestParseReleaseName(t *testing.T) {
    tests := []struct {
        input          string
        expectedName   string
        expectedSeason *int
    }{
        // Test cases with descriptive comments
        {
            "[SubsPlease] Frieren S01 (1080p) [HEVC]",
            "Frieren",
            intPtr(1),
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := ParseReleaseName(tt.input)

            if result.Name != tt.expectedName {
                t.Errorf("Name: got %q, want %q", result.Name, tt.expectedName)
            }
            // More assertions...
        })
    }
}
```

**Patterns:**
- Table-driven tests with anonymous structs
- Subtests via `t.Run()` for each case
- Descriptive test case names (uses input as subtest name)
- Helper functions for pointer values: `func intPtr(n int) *int`

## Mocking

**Framework:** None used

**What's Mocked:**
- Nothing currently - only pure functions tested

**What's NOT Mocked:**
- HTTP clients (`qbit.Client`, `shoko.Client`)
- Database operations
- File system operations
- WebSocket connections

## Fixtures and Factories

**Test Data:**
```go
// Inline test cases with real-world examples
tests := []struct {
    input          string
    expectedName   string
    expectedSeason *int
}{
    // Basic group tag stripping
    {
        "[SubsPlease] Frieren S01 (1080p) [HEVC]",
        "Frieren",
        intPtr(1),
    },
    // === Real-world test cases ===
    {
        "Frieren.Beyond.Journeys.End.S02E05.Logistics.in.the.Northern.Plateau.1080p.NF.WEB-DL.JPN.AAC2.0.H.264.MSubs-ToonsHub.mkv",
        "Frieren Beyond Journeys End",
        intPtr(2),
    },
}
```

**Location:**
- No separate fixtures directory
- Test data inline in test files

**Helper Functions:**
```go
func intPtr(n int) *int {
    return &n
}
```

## Coverage

**Requirements:** None enforced

**Current State:**
- Only `internal/parser/` has tests
- Untested packages: `api`, `auth`, `database`, `linker`, `monitor`, `notify`, `nyaa`, `qbit`, `rss`, `scanner`, `shoko`, `upscale`, `ws`

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out    # HTML report
```

## Test Types

**Unit Tests:**
- `internal/parser/parser_test.go` - Tests `ParseReleaseName()` function
- Pure function testing with no side effects
- Comprehensive input/output verification

**Integration Tests:**
- Not implemented
- No tests for API handlers
- No tests for database operations
- No tests for external service integrations

**E2E Tests:**
- Not implemented
- No frontend testing framework
- No Playwright/Cypress configuration

## Common Patterns

**Async Testing:**
```go
// Not currently used - no async tests in codebase
```

**Error Testing:**
```go
// Pattern for nil pointer handling
if tt.expectedSeason == nil && result.Season != nil {
    t.Errorf("Season: got %d, want nil", *result.Season)
} else if tt.expectedSeason != nil && result.Season == nil {
    t.Errorf("Season: got nil, want %d", *tt.expectedSeason)
} else if tt.expectedSeason != nil && result.Season != nil && *result.Season != *tt.expectedSeason {
    t.Errorf("Season: got %d, want %d", *result.Season, *tt.expectedSeason)
}
```

**Table-Driven Pattern:**
```go
tests := []struct {
    name     string  // or use input as name
    input    Type
    expected Type
}{
    {"case description", inputVal, expectedVal},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := FunctionUnderTest(tt.input)
        if got != tt.expected {
            t.Errorf("got %v, want %v", got, tt.expected)
        }
    })
}
```

## Adding New Tests

**For Go packages:**
1. Create `*_test.go` file in same directory as source
2. Use same package name (not `package foo_test`)
3. Follow table-driven test pattern
4. Use `t.Run()` for subtests

**For API handlers (when adding):**
```go
// Recommended pattern for handlers
func TestHandleLogin(t *testing.T) {
    // Setup: create test server with mock dependencies
    // Execute: make HTTP request
    // Assert: check response status and body
}
```

**For Frontend (when adding):**
- Consider Vitest for Vue component testing
- Add to `package.json` scripts
- Co-locate tests: `ComponentName.test.ts` or `ComponentName.spec.ts`

## Test Gaps to Address

**High Priority:**
- `internal/api/*` - HTTP handlers have no tests
- `internal/auth/*` - Auth logic (session, password) untested
- `internal/linker/*` - Core linking logic untested

**Medium Priority:**
- `internal/database/*` - Database operations
- `internal/scanner/*` - File system scanning

**Lower Priority (external integrations):**
- `internal/qbit/*` - qBittorrent client
- `internal/shoko/*` - Shoko client
- `internal/nyaa/*` - Nyaa scraper

---

*Testing analysis: 2026-02-22*
