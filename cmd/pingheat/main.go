package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pbv7/pingheat/internal/app"
	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/pkg/version"
)

var (
	errMissingTarget    = errors.New("target host required")
	errIntervalTooShort = errors.New("interval must be at least 100ms")
)

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

func parseArgs(args []string, program string) (parseResult, error) {
	cfg := config.DefaultConfig()
	fs := flag.NewFlagSet(program, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	interval := fs.Duration("i", cfg.Interval, "Ping interval")
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
		fmt.Fprintf(os.Stderr, "  %s -i 500ms 8.8.8.8              # Ping every 500ms\n", program)
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

	if *interval < 100*time.Millisecond {
		return parseResult{usage: usage}, errIntervalTooShort
	}

	cfg.Target = fs.Args()[0]
	cfg.Interval = *interval
	cfg.HistorySize = *historySize
	cfg.ShowHelp = *showHelp

	if *exporterAddr != "" {
		cfg.ExporterEnabled = true
		cfg.ExporterAddr = *exporterAddr
	}

	if *pprofAddr != "" {
		cfg.PprofEnabled = true
		addr := *pprofAddr
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		}
		cfg.PprofAddr = addr
	}

	return parseResult{cfg: cfg, showVersion: *showVersion, usage: usage}, nil
}
