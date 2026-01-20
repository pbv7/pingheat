# pingheat

A cross-platform terminal application that visualizes network latency as a real-time scrolling heatmap, with optional Prometheus metrics export.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/License-MIT-green)

## Features

- **Real-time Heatmap** - Visual representation of ping latency with color-coded blocks
- **Cross-platform** - Works on Linux, macOS, and Windows
- **Prometheus Metrics** - Optional export of 22+ metrics for monitoring dashboards
- **Comprehensive Statistics** - Min/Avg/Max RTT, jitter, percentiles (p50/p90/p95/p99), loss tracking
- **Instability Detection** - Tracks outages, brownouts, and packet loss bursts
- **Large History** - Stores up to 30,000 samples for scrollable review
- **Keyboard Navigation** - Vim-style controls for browsing history

## Installation

### From Source

```bash
git clone https://github.com/pbv7/pingheat.git
cd pingheat
make build
./bin/pingheat google.com
```

### Go Install

```bash
go install github.com/pbv7/pingheat/cmd/pingheat@latest
```

## Usage

```bash
# Basic ping
pingheat google.com

# Custom interval (500ms)
pingheat -i 500ms 8.8.8.8

# IPv6 literal (brackets optional)
pingheat 2001:db8::1
pingheat [2001:db8::1]

# IPv6 link-local (interface required)
pingheat fe80::1%en0

# Enable Prometheus metrics on port 9090
pingheat -exporter :9090 1.1.1.1

# Enable pprof profiling (binds to localhost)
pingheat -pprof :6060 google.com

# All options
pingheat -i 200ms -history 50000 -exporter :9090 -pprof :6060 cloudflare.com
```

### Command Line Options

| Flag        | Default | Description                                |
| ----------- | ------- | ------------------------------------------ |
| `-i`        | `1s`    | Ping interval (min: 100ms)                 |
| `-history`  | `30000` | Number of samples to keep in history       |
| `-exporter` | -       | Enable Prometheus exporter (e.g., `:9090`) |
| `-pprof`    | -       | Enable pprof server (e.g., `:6060` binds to localhost) |
| `-version`  | -       | Show version information                   |
| `-help`     | -       | Show help on startup                       |

## Keyboard Controls

| Key             | Action                |
| --------------- | --------------------- |
| `↑` / `k`       | Scroll up (older)     |
| `↓` / `j`       | Scroll down (newer)   |
| `PgUp` / `PgDn` | Page up / down        |
| `Home` / `g`    | Jump to oldest        |
| `End` / `G`     | Jump to newest        |
| `?` / `h`       | Toggle help           |
| `c`             | Clear history         |
| `q` / `Ctrl+C`  | Quit                  |

## Color Legend

| RTT       | Color (hex) | Classification |
| --------- | ----------- | -------------- |
| 0-30ms    | `#00FF00`   | Excellent      |
| 30-80ms   | `#7FFF00`   | Good           |
| 80-150ms  | `#FFFF00`   | Fair           |
| 150-300ms | `#FF8C00`   | Poor           |
| >300ms    | `#FF0000`   | Bad            |
| Timeout   | `#8B008B`   | No response    |

## Prometheus Metrics

When enabled with `-exporter :9090`, metrics are available at `http://localhost:9090/metrics`.
To restrict metrics to localhost, use `-exporter 127.0.0.1:9090`.

### Counters

- `pingheat_ping_sent_total` - Total packets sent
- `pingheat_ping_success_total` - Successful responses
- `pingheat_ping_timeout_total` - Timeouts

### Latency Gauges

- `pingheat_ping_latency_ms{stat="min|avg|max"}` - RTT statistics
- `pingheat_ping_stddev_ms` - Standard deviation
- `pingheat_ping_jitter_ms` - Jitter (mean absolute deviation)
- `pingheat_ping_last_rtt_ms` - Most recent RTT
- `pingheat_ping_latency_p50_ms` - Median latency
- `pingheat_ping_latency_p90_ms` - 90th percentile
- `pingheat_ping_latency_p95_ms` - 95th percentile
- `pingheat_ping_latency_p99_ms` - 99th percentile

### Availability

- `pingheat_ping_loss_percent` - Packet loss (0-100)
- `pingheat_ping_availability_percent` - Availability (0-100)
- `pingheat_ping_up` - Target reachability (1=up, 0=down)

### Streaks & Instability

- `pingheat_ping_current_streak` - Current streak (+success, -timeout)
- `pingheat_ping_longest_success_streak` - Record consecutive successes
- `pingheat_ping_longest_timeout_streak` - Record consecutive timeouts
- `pingheat_ping_loss_bursts_total` - Number of loss burst events
- `pingheat_ping_brownout_samples_total` - High-latency samples (>200ms)
- `pingheat_ping_brownout_bursts_total` - Number of brownout events
- `pingheat_ping_in_brownout` - Currently in brownout (1=yes)

### System

- `pingheat_uptime_seconds` - Monitoring duration

## Building

```bash
# Build for current platform
make build

# Run tests
make test

# Test coverage
make test-cover

# Lint
make lint

# Cross-compile all platforms
make build-all

# Build with GoReleaser (snapshot)
make release-snapshot
```

## Platform Support

| Platform                     | Tested | Notes                                |
| ---------------------------- | ------ | ------------------------------------ |
| Linux (amd64, arm64, armv7)  | Yes    | Full support                         |
| macOS (Intel, Apple Silicon) | Yes    | Full support                         |
| Windows (amd64, arm64)       | Yes    | Uses `-t` flag (no interval control) |

All platforms automatically force English locale for consistent output parsing.

## Architecture

```text
┌──────────────┐
│ Ping Runner  │ executes system ping command
└──────┬───────┘
       │ samples
┌──────▼───────┐
│ Distributor  │ broadcasts to consumers
└──────┬───────┘
       │
   ┌───┴───┬─────────────┐
   │       │             │
┌──▼──┐ ┌──▼───┐ ┌───────▼───────┐
│ UI  │ │Metrics│ │Prometheus     │
│     │ │Engine │ │Exporter (opt) │
└─────┘ └──────┘ └───────────────┘
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please open an issue or submit a pull request.
