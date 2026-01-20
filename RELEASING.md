# Release Process

## Prerequisites

Install GoReleaser:

```bash
# macOS
brew install goreleaser

# Linux (via snap)
snap install --classic goreleaser

# Or download from: https://github.com/goreleaser/goreleaser/releases
```

## Local Testing

Test the release process locally without publishing:

```bash
# Validate configuration
make release-check

# Build snapshot release (no tags required)
make release-snapshot
```

Built artifacts will be in `dist/` directory.

## Creating a Release

### 1. Ensure Everything is Committed

```bash
git status
```

### 2. Create and Push a Tag

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tag to GitHub
git push origin v1.0.0
```

### 3. Automated Release

GitHub Actions will automatically:
- Run tests
- Build binaries for all platforms
- Create checksums
- Generate changelog
- Create a **draft** GitHub release

### 4. Review and Publish

1. Go to https://github.com/pbv7/pingheat/releases
2. Find the draft release
3. Review the generated changelog and artifacts
4. Edit release notes if needed
5. Click "Publish release"

## Supported Platforms

GoReleaser builds for:

- **Linux**: amd64, arm64, armv7
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64, arm64

## Release Artifacts

Each release includes:

- Compressed binaries (`.tar.gz` for Unix, `.zip` for Windows)
- `checksums.txt` - SHA256 checksums for verification
- Auto-generated changelog
- Source code archives

## Manual Release

To create a release manually:

```bash
# Ensure you have a clean git state and a tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Run GoReleaser locally
make release
```

**Note**: Manual releases require `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN="your_github_token"
make release
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- `v1.0.0` - Major release (breaking changes)
- `v1.1.0` - Minor release (new features, backwards compatible)
- `v1.1.1` - Patch release (bug fixes)

## Troubleshooting

### "repository not initialized" error

Ensure you're in a git repository with at least one commit:

```bash
git init
git add .
git commit -m "Initial commit"
```

### "tag not found" error

GoReleaser requires a git tag. For testing, use snapshot mode:

```bash
make release-snapshot
```

### Missing GITHUB_TOKEN

For manual releases, create a personal access token at:
https://github.com/settings/tokens

Required scopes: `repo` (full control of private repositories)
