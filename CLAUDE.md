# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Pingheat** is a cross-platform terminal application that visualizes network latency as a real-time scrolling heatmap.
It optionally exports Prometheus metrics and targets Go 1.25+ on Linux, macOS, and Windows.

## Development Dependencies

**Required:** Go 1.25+, Git

**Optional:**

- **golangci-lint** - Go code linting (installable via brew, go install, or npx)
- **Node.js/npx** - Markdown linting via markdownlint-cli2
- **actionlint** - GitHub Actions workflow validation (installable via brew, go install, or npx)
- **yamllint** - YAML syntax validation (installable via brew or pip)
- **GoReleaser** - Release binary building (installable via brew or from goreleaser.com)
- **gh** - GitHub CLI for creating PRs from terminal (`brew install gh` or from cli.github.com)

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
make lint                    # Run golangci-lint on Go code
make lint-md                 # Lint markdown files
make lint-workflows          # Lint GitHub Actions workflows (requires actionlint)
make lint-all                # Run all linters (Go + markdown + workflows)
make deps                    # Download and tidy modules

# Release (requires goreleaser installed)
make release-check           # Validate .goreleaser.yaml
make release-snapshot        # Test release build locally
make release                 # Create actual release (requires git tag)

# Cleanup
make clean                   # Remove bin/ and coverage files
make clean-dist              # Remove dist/ (GoReleaser output)
make clean-all               # Remove everything (clean + clean-dist)
```

## Development Workflow

**IMPORTANT**: The `main` branch is protected. All changes must go through pull requests.

### Making Changes

1. **Create a feature branch**:

   ```bash
   git checkout main
   git pull
   git checkout -b feature/my-feature  # or fix/, chore/, docs/, etc.
   ```

2. **Make your changes and commit**:

   ```bash
   # Make changes to files
   git add .
   git commit -m "feat: add awesome feature"
   ```

3. **Validate changes BEFORE pushing**:

   ```bash
   make test         # Run tests with race detector
   make lint-all     # Run all linters (Go + markdown + workflows)
   ```

   **CRITICAL**: Always run `make lint-all` before creating a PR. This catches issues locally that CI will check anyway.

4. **Push the branch**:

   ```bash
   git push -u origin feature/my-feature
   ```

5. **Create a pull request**:

   **Option A: Via GitHub web UI**

   - GitHub will show a banner with "Compare & pull request" button
   - Or go to: <https://github.com/pbv7/pingheat/pulls> → "New pull request"

   **Option B: Via GitHub CLI** (recommended):

   ```bash
   gh pr create --title "feat: add awesome feature" --body "Description of changes"

   # Or interactive mode:
   gh pr create
   ```

6. **Wait for CI checks**:
   - All 6 status checks must pass:
     - Test (ubuntu-latest, macos-latest, windows-latest)
     - Lint
     - Security Scan
     - Build Verification
   - Codecov will comment with coverage report
   - Dependency Review will scan for vulnerabilities

7. **Approve the PR** (required for solo dev workflow):

   **Note**: GitHub doesn't allow self-approval via CLI. You must use the web UI:

   - Go to the PR page
   - Click **"Files changed"** tab
   - Click **"Review changes"** button (top right)
   - Select **"Approve"**
   - Click **"Submit review"**

   Or use admin bypass (see alternative workflow below).

8. **Merge the PR**:

   ```bash
   # Via GitHub CLI (after checks pass and approval):
   gh pr merge --squash --delete-branch

   # Or use the "Squash and merge" button on GitHub web UI
   ```

9. **Update local main**:

   ```bash
   git checkout main
   git pull
   ```

### Alternative: Admin Bypass Workflow (Solo Dev)

If you have repository admin permissions and enabled "Repository admin" in the bypass list, you can skip manual approval:

```bash
# After creating PR and CI passes:
gh pr merge --squash --delete-branch --admin

# Update local main
git checkout main
git pull
```

**Note**: The `--admin` flag uses administrator privileges to bypass branch protection rules.

### Common Workflows

**Updating dependencies**:

```bash
git checkout -b chore/update-dependencies
go get -u ./...
go mod tidy
make test        # Verify everything works
make lint-all    # Run all linters
git add go.mod go.sum
git commit -m "chore: update dependencies to latest versions"
git push -u origin chore/update-dependencies
gh pr create --title "chore: update dependencies" --body "Updated all dependencies to latest versions"

# After CI passes, approve via web UI, then:
gh pr merge --squash --delete-branch

# Or use admin bypass:
gh pr merge --squash --delete-branch --admin

# Update local main
git checkout main
git pull
```

**Quick fix**:

```bash
git checkout -b fix/typo-in-readme
# Fix the typo
make lint-all    # Validate changes
git add README.md
git commit -m "docs: fix typo in installation instructions"
git push -u origin fix/typo-in-readme
gh pr create

# After CI passes, approve via web UI or use admin bypass:
gh pr merge --squash --delete-branch --admin
git checkout main && git pull
```

### Branch Protection Rules

The `main` branch has these protections:

- ❌ Direct pushes blocked
- ❌ Force pushes blocked
- ✅ Requires pull request
- ✅ Requires all CI checks to pass
- ✅ Requires branch to be up-to-date
- ✅ Linear history enforced

**If you try to push directly to main**:

```bash
git push origin main
# Error: GH006: Protected branch update failed
# Error: Changes must be made through a pull request
```

## Architecture

### Component-Based Design with Channel Communication

```text
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

#### Critical: Locale Enforcement

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

## Before Committing

Always validate code quality before committing:

```bash
make test          # Run tests with race detector
make cover-summary # Check coverage (aim for >60%)
make lint-all      # Lint Go code, markdown files, and workflows
```

**Critical:** Always run `make lint-md` before committing markdown changes. Formatting errors break documentation rendering on GitHub.

**Workflow validation (if modifying .github/workflows/):**

```bash
make lint-workflows  # Validates GitHub Actions workflows with actionlint
```

### Commit Message Conventions

**IMPORTANT:** Do NOT include "Co-Authored-By: Claude..." lines in commit messages for this project.

Commit messages should follow conventional commits format when applicable:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `ci:` - CI/CD changes
- `refactor:` - Code refactoring
- `test:` - Test changes

## CI/CD Pipelines

GitHub Actions workflows automate testing, linting, security scanning, and releases.

### Workflows

| Workflow | Trigger | Purpose |
| -------- | ------- | ------- |
| `test.yml` | Push to main, PRs | Multi-platform tests, linting, security scan, build verification |
| `dependency-review.yml` | PRs only | Scan dependencies for vulnerabilities before merge |
| `release.yml` | Tag push (v*) | Pre-release checks + GoReleaser for binary distribution |

### Test Workflow Jobs (`test.yml`)

- **test**: Runs `go test -race -coverprofile=coverage.txt` on Linux, macOS, and Windows
  - Uploads coverage reports to Codecov (ubuntu-latest only)
- **lint**: golangci-lint + markdownlint
- **security**: govulncheck for vulnerability scanning
- **build**: Cross-compilation verification for all platforms

### Release Workflow (`release.yml`)

The release workflow runs pre-release checks before GoReleaser:

1. Multi-platform tests (Linux, macOS, Windows)
2. Linting with golangci-lint
3. Security scan with govulncheck
4. GoReleaser builds and publishes binaries

All pre-release jobs must pass before GoReleaser runs.

### Required Secrets

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions (no configuration needed)
- `CODECOV_TOKEN`: Required for uploading coverage reports to Codecov
  - Obtain from [codecov.io](https://app.codecov.io/gh/pbv7/pingheat/settings)
  - Add as repository secret in GitHub Settings > Secrets and variables > Actions
- `HOMEBREW_TAP_TOKEN`: Fine-grained PAT for updating Homebrew tap (release workflow only)
  - Required permissions: Contents (read/write) on `pbv7/homebrew-tap`

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

```text
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
