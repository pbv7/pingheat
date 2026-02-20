package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pbv7/pingheat/internal/app"
	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/pkg/version"
)

var (
	errMissingTarget    = errors.New("target host required")
	errIntervalTooShort = errors.New("interval must be at least 100ms")
	errIntervalTooLong  = errors.New("interval must be at most 1 hour")
	errInvalidTarget    = errors.New("invalid target format")
	errInvalidPort      = errors.New("port must be between 1 and 65535")
)

// hostnameRe validates RFC 1123 compliant hostnames.
// Allows: letters, digits, hyphens, dots
// Each label: starts/ends with alphanumeric, max 63 chars
var hostnameRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)

// parseResult carries the parsed config and usage handler for errors.
type parseResult struct {
	cfg         config.Config
	showVersion bool
	usage       func()
}

func main() {
	result, err := parseArgs(os.Args[1:], os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if errors.Is(err, errMissingTarget) {
			fmt.Fprintln(os.Stderr)
			result.usage()
		}
		os.Exit(1)
	}

	if result.showVersion {
		fmt.Println("pingheat", version.Info())
		os.Exit(0)
	}

	// Run application
	application := app.New(result.cfg)
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs parses CLI arguments into a config without side effects.
func parseArgs(args []string, program string) (parseResult, error) {
	cfg := config.DefaultConfig()
	fs := flag.NewFlagSet(program, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	intervalShort := fs.Duration("i", cfg.Interval, "Ping interval (shorthand for -interval)")
	intervalLong := fs.Duration("interval", cfg.Interval, "Ping interval")
	historySize := fs.Int("history", cfg.HistorySize, "History buffer size (samples)")
	exporterAddr := fs.String("exporter", "", "Enable Prometheus exporter on address (e.g., :9090)")
	pprofAddr := fs.String("pprof", "", "Enable pprof server on address (e.g., :6060 binds to localhost)")
	showVersion := fs.Bool("version", false, "Show version")
	showHelp := fs.Bool("help", false, "Show help on startup")

	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <target>\n\n", program)
		fmt.Fprintf(os.Stderr, "pingheat - Network latency heatmap visualizer\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s google.com                    # Ping google.com with default settings\n", program)
		fmt.Fprintf(os.Stderr, "  %s -i 500ms 8.8.8.8              # Ping every 500ms (short form)\n", program)
		fmt.Fprintf(os.Stderr, "  %s -interval 500ms 8.8.8.8       # Ping every 500ms (long form)\n", program)
		fmt.Fprintf(os.Stderr, "  %s -exporter :9090 1.1.1.1       # Enable Prometheus metrics on :9090\n", program)
		fmt.Fprintf(os.Stderr, "  %s -pprof :6060 google.com       # Enable pprof server on localhost:6060\n", program)
	}
	fs.Usage = usage

	if err := fs.Parse(args); err != nil {
		return parseResult{usage: usage}, err
	}

	if *showVersion {
		return parseResult{cfg: cfg, showVersion: true, usage: usage}, nil
	}

	if len(fs.Args()) < 1 {
		return parseResult{usage: usage}, errMissingTarget
	}

	// Resolve interval: prefer -interval if set, otherwise use -i
	// Use flag.Visit to reliably detect which flags were actually provided
	interval := cfg.Interval
	intervalShortSet := false
	intervalLongSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "i" {
			intervalShortSet = true
		}
		if f.Name == "interval" {
			intervalLongSet = true
		}
	})

	if intervalShortSet {
		interval = *intervalShort
	}
	if intervalLongSet {
		interval = *intervalLong
	}

	if interval < 100*time.Millisecond {
		return parseResult{usage: usage}, errIntervalTooShort
	}
	if interval > time.Hour {
		return parseResult{usage: usage}, errIntervalTooLong
	}

	cfg.Target = fs.Args()[0]
	if err := validateTargetFormat(cfg.Target); err != nil {
		return parseResult{usage: usage}, err
	}
	cfg.Interval = interval
	cfg.HistorySize = *historySize
	cfg.ShowHelp = *showHelp

	if *exporterAddr != "" {
		if err := validateAddress(*exporterAddr, "exporter"); err != nil {
			return parseResult{usage: usage}, err
		}
		cfg.ExporterEnabled = true
		cfg.ExporterAddr = *exporterAddr
	}

	if *pprofAddr != "" {
		addr := *pprofAddr
		if err := validateAddress(addr, "pprof"); err != nil {
			return parseResult{usage: usage}, err
		}

		cfg.PprofEnabled = true
		// Security: Auto-bind to localhost when only port specified (:6060).
		// This prevents exposing pprof debugging endpoints to network.
		// To bind to all interfaces, explicitly use 0.0.0.0:6060 or ::.
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		}
		cfg.PprofAddr = addr
	}

	return parseResult{cfg: cfg, showVersion: *showVersion, usage: usage}, nil
}

// validateTargetFormat validates target is a valid IP address or hostname.
// Does NOT perform DNS lookups - only format validation.
// Supports IPv6 zone IDs (e.g., fe80::1%en0 or [fe80::1%en0]).
func validateTargetFormat(target string) error {
	if target == "" {
		return errInvalidTarget
	}

	// Check if it's a valid IP (IPv4 or IPv6 without zone)
	if net.ParseIP(target) != nil {
		return nil
	}

	// Handle IPv6 literals with brackets.
	// If it has brackets, it MUST be an IP (with optional zone) and not a hostname.
	if strings.HasPrefix(target, "[") && strings.HasSuffix(target, "]") {
		host := target[1 : len(target)-1]
		// Strip zone ID if present (e.g., fe80::1%en0 -> fe80::1)
		if zoneIndex := strings.Index(host, "%"); zoneIndex != -1 {
			// Reject empty zone IDs (e.g., [fe80::1%])
			if zoneIndex == len(host)-1 {
				return fmt.Errorf("%w: %q has empty zone identifier", errInvalidTarget, target)
			}
			host = host[:zoneIndex]
		}
		if net.ParseIP(host) != nil {
			return nil // Valid bracketed IPv6 (with optional zone)
		}
		// Invalid bracketed value (not an IP)
		return fmt.Errorf("%w: %q must be a valid IP address or hostname", errInvalidTarget, target)
	}

	// Check for IPv6 with zone ID (e.g., fe80::1%en0)
	if zoneIndex := strings.Index(target, "%"); zoneIndex != -1 {
		// Reject empty zone IDs (e.g., fe80::1%)
		if zoneIndex == len(target)-1 {
			return fmt.Errorf("%w: %q has empty zone identifier", errInvalidTarget, target)
		}
		host := target[:zoneIndex]
		if net.ParseIP(host) != nil {
			return nil // Valid IPv6 with zone ID
		}
		// If a '%' is present, it must be a valid zoned IPv6 address. Hostnames cannot contain '%'.
		return fmt.Errorf("%w: %q must be a valid zoned IPv6 address (hostnames cannot contain '%%')", errInvalidTarget, target)
	}

	// Allow absolute FQDNs with trailing dot (e.g., example.com. or localhost.)
	// Strip trailing dot before hostname validation
	hostname := strings.TrimSuffix(target, ".")

	// Validate hostname format (RFC 1123 compliant)
	if !hostnameRe.MatchString(hostname) {
		return fmt.Errorf("%w: %q must be a valid IP address or hostname", errInvalidTarget, target)
	}

	return nil
}

// validateAddress validates that an address string contains a valid port (1-65535).
// Supports formats: ":9090", "localhost:9090", "0.0.0.0:9090", "[::1]:9090"
func validateAddress(addr, name string) error {
	// Handle port-only format (":9090") by adding temporary host for parsing
	hostPort := addr
	if strings.HasPrefix(addr, ":") {
		hostPort = "localhost" + addr
	}

	// Parse host:port using standard library
	_, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("invalid %s address %q: %w", name, addr, err)
	}

	// Parse port number
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid %s port %q: %w", name, portStr, err)
	}

	// Validate port range (allow all valid ports including privileged 1-1023)
	if port < 1 || port > 65535 {
		return fmt.Errorf("%w for %s: %d", errInvalidPort, name, port)
	}

	return nil
}
