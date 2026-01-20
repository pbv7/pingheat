package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pbv7/pingheat/internal/app"
	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/pkg/version"
)

func main() {
	cfg := config.DefaultConfig()

	// CLI flags
	interval := flag.Duration("i", cfg.Interval, "Ping interval")
	historySize := flag.Int("history", cfg.HistorySize, "History buffer size (samples)")
	exporterAddr := flag.String("exporter", "", "Enable Prometheus exporter on address (e.g., :9090)")
	pprofAddr := flag.String("pprof", "", "Enable pprof server on address (e.g., :6060)")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help on startup")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <target>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "pingheat - Network latency heatmap visualizer\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s google.com                    # Ping google.com with default settings\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -i 500ms 8.8.8.8              # Ping every 500ms\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -exporter :9090 1.1.1.1       # Enable Prometheus metrics on :9090\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -pprof :6060 google.com       # Enable pprof server on :6060\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Println("pingheat", version.Info())
		os.Exit(0)
	}

	// Get target from positional argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: target host required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate interval
	if *interval < 100*time.Millisecond {
		fmt.Fprintf(os.Stderr, "Error: interval must be at least 100ms\n")
		os.Exit(1)
	}

	cfg.Target = args[0]
	cfg.Interval = *interval
	cfg.HistorySize = *historySize
	cfg.ShowHelp = *showHelp

	// Enable exporter if address provided
	if *exporterAddr != "" {
		cfg.ExporterEnabled = true
		cfg.ExporterAddr = *exporterAddr
	}

	// Enable pprof if address provided
	if *pprofAddr != "" {
		cfg.PprofEnabled = true
		cfg.PprofAddr = *pprofAddr
	}

	// Run application
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
