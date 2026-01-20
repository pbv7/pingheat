# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Pingheat** is a cross-platform terminal application that visualizes network latency as a real-time scrolling heatmap with optional Prometheus metrics export. Written in Go 1.25+, supporting Linux, macOS, and Windows.

## Build & Development Commands

```bash
# Building
make build                    # Build for current platform → bin/pingheat
make build-all               # Cross-compile for Linux, macOS, Windows
make install                 # Install to GOPATH/bin

# Running
make run ARGS="google.com"   # Build and run with arguments
./bin/pingheat google.com    # Direct execution

# Testing
make test                    # Run all tests with race detector
make test-cover              # Generate HTML coverage report (coverage.html)
make cover-summary           # Print coverage percentage

# Linting & Dependencies
make lint                    # Run golangci-lint
make deps                    # Download and tidy modules

# Release (requires goreleaser installed)
make release-check           # Validate .goreleaser.yaml
make release-snapshot        # Test release build locally
make release                 # Create actual release (requires git tag)

# Cleanup
make clean                   # Remove bin/, coverage files
```

## Architecture

### Component-Based Design with Channel Communication

```
Ping Runner → [samples channel] → Distributor → [fan-out] → UI + Metrics + Exporter
```

**Key Components:**
- **Runner** (`internal/ping/runner.go`): Spawns system ping, reads stdout/stderr
- **Parser** (`internal/parser/`): Platform-specific output parsing (Linux/macOS/Windows)
- **Metrics Engine** (`internal/metrics/engine.go`): Aggregates 23+ statistics
- **UI** (`internal/ui/`): Bubble Tea TUI with heatmap visualization
- **Exporter** (`internal/exporter/prometheus.go`): Prometheus metrics HTTP server
- **App** (`internal/app/app.go`): Orchestrates lifecycle, goroutines, channels

### Data Flow Pattern

1. **Ping Runner** executes `ping` command via `exec.CommandContext`
2. **Parser** converts platform-specific output to unified `types.Sample` struct
3. **Distributor** (in app.go) fans out samples to multiple consumers using **non-blocking sends**:
   ```go
   select {
   case channel <- sample:
   default:
       // Skip if buffer full - prevents blocking
   }
   ```
4. **Metrics Engine** aggregates statistics (lock-protected with RWMutex)
5. **UI** renders heatmap grid + stats (Bubble Tea update/view loop)
6. **Exporter** exposes 23 Prometheus metrics on `/metrics` endpoint

### Platform-Specific Handling

**Critical: Locale Enforcement**

All platforms force English output for consistent regex parsing:
- **Linux/macOS**: Set `LC_ALL=C` and `LANG=C` environment variables
- **Windows**: Use `cmd.exe /C "chcp 437 >nul & ping -t <target>"`

**Parser Differences:**
- **Linux/macOS**: Parse `icmp_seq=N time=X.XXX ms` from output
- **Windows**: Parse `Reply from ... time=Xms`, manually track sequence numbers

**Interval Support:**
- **Linux/macOS**: `-i <interval>` flag (fractional seconds supported)
- **Windows**: No interval control (uses `-t` for continuous mode)

## Critical Implementation Details

### 1. Non-Blocking Distribution
The distributor uses `select` with `default` to avoid blocking. This means:
- Ping runner never waits for slow consumers
- Samples may be dropped if UI/metrics buffers are full (intentional design)
- Fast ping rates won't deadlock the system

### 2. Ring Buffer for History
`internal/buffer/ringbuffer.go` provides bounded memory:
- Default: 30,000 samples for UI display
- Thread-safe with RWMutex
- Automatically overwrites oldest data

### 3. Concurrency Safety
- **Metrics Engine**: RWMutex protects `Stats` reads during `Add()` writes
- **Ring Buffer**: RWMutex on all operations
- **Channels**: Buffered (100 slots for samples, 10 for metrics)

### 4. Variance Calculation
Uses accumulated sums for efficiency:
```go
// Variance = E[X²] - (E[X])²
varianceUs := (sumRTTSquares / n) - (meanUs * meanUs)
```

### 5. Brownout Detection
High-latency state tracked separately from timeouts:
- **Brownout Threshold**: RTT > 200ms
- **Brownout Burst**: Transition into/out of brownout state
- Useful for detecting degraded (but not failed) connections

## Testing Patterns

### Running Specific Tests
```bash
# Test single package
go test -v ./internal/metrics

# Test specific function
go test -v -run TestEngine_Add ./internal/metrics

# Race detection (always use in CI)
go test -race ./...

# Coverage for specific package
go test -coverprofile=cover.out ./internal/parser
go tool cover -func=cover.out
```

### Parser Testing
Use `testdata/` fixtures:
- `testdata/linux.txt` - Linux ping output samples
- `testdata/darwin.txt` - macOS ping output samples
- `testdata/windows.txt` - Windows ping output samples

When adding new parsers or fixing bugs, update these fixtures.

## Common Gotchas

### 1. Locale Bugs
**Problem**: Ping output in non-English locales breaks regex parsing.

**Solution**: Always enforce C locale (already implemented in runner.go).

**Symptom**: Zero metrics despite ping running.

### 2. Windows Interval
**Problem**: Windows doesn't support `-i` flag.

**Workaround**: Use `-t` (continuous), rely on OS scheduling. Custom intervals not enforced.

### 3. Stats Reset Behavior
- `metrics.Engine.Reset()`: Clears ALL data including timestamps
- UI clear command (`c` key): Resets only ring buffer, not metrics

Choose the appropriate reset based on use case.

### 4. Sequence Numbers
- **Linux/macOS**: From ping output (1-based, from `icmp_seq`)
- **Windows**: Synthetic counter (manually incremented per sample)
- Don't rely on sequence for ordering—use timestamps instead

## Adding New Features

### Adding a New Metric
1. Update `internal/metrics/engine.go`:
   - Add field to `Stats` struct
   - Update calculation in `Add()` method
   - Update `Reset()` method
2. Update `internal/exporter/prometheus.go`:
   - Add metric descriptor in `newExporter()`
   - Update `Collect()` to expose new metric
3. Update `README.md` metrics documentation

### Adding UI Elements
1. Modify `internal/ui/view.go` for rendering
2. Update grid calculations if changing reserved space:
   ```go
   availableHeight = height - 7  // Header(1) + Stats(2) + Status(1) + Borders(2) + Help(1)
   ```
3. Add keyboard shortcuts in `internal/ui/update.go`
4. Update help overlay in `view.go`

### Supporting New Platform
1. Create new parser in `internal/parser/<platform>.go`
2. Implement `Parser` interface
3. Update `parser.New()` factory for `runtime.GOOS` detection
4. Add test fixtures in `testdata/<platform>.txt`
5. Update `.goreleaser.yaml` build matrix

## Version Information

Version details injected at build time via LDFLAGS:
```
-X github.com/pbv7/pingheat/pkg/version.Version
-X github.com/pbv7/pingheat/pkg/version.Commit
-X github.com/pbv7/pingheat/pkg/version.BuildTime
```

Access via `pkg/version/version.go`.

## Dependencies

**UI Framework**: `github.com/charmbracelet/bubbletea` + `lipgloss`

**Metrics**: `github.com/prometheus/client_golang`

**Standard Library**: `time`, `context`, `os/exec`, `regexp`, `sync`

**No external ping library** - uses system `ping` command for maximum portability.

## Release Process

See `RELEASING.md` for detailed instructions.

**Quick version:**
1. Commit all changes
2. Create annotated tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
3. Push tag: `git push origin v1.0.0`
4. GitHub Actions automatically builds and creates draft release
5. Review and publish on GitHub

**Local snapshot testing:**
```bash
make release-snapshot
# Check dist/ directory for builds
```
