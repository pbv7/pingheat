package config

import "time"

// Config holds all configuration options for pingheat.
type Config struct {
	// Target host to ping
	Target string

	// Ping interval
	Interval time.Duration

	// Display history length in samples
	HistorySize int

	// Metrics buffer size
	MetricsBufferSize int

	// Prometheus exporter settings
	ExporterEnabled bool
	ExporterAddr    string

	// pprof server settings
	PprofEnabled bool
	PprofAddr    string

	// UI settings
	ShowHelp bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Target:            "",
		Interval:          time.Second,
		HistorySize:       30000,
		MetricsBufferSize: 120000,
		ExporterEnabled:   false,
		ExporterAddr:      ":9090",
		PprofEnabled:      false,
		PprofAddr:         ":6060",
		ShowHelp:          false,
	}
}
