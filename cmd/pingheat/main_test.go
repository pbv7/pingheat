package main

import (
	"errors"
	"testing"
)

func TestParseArgsMissingTarget(t *testing.T) {
	_, err := parseArgs([]string{}, "pingheat")
	if !errors.Is(err, errMissingTarget) {
		t.Fatalf("expected errMissingTarget, got %v", err)
	}
}

func TestParseArgsIntervalTooShort(t *testing.T) {
	_, err := parseArgs([]string{"-i", "50ms", "example.com"}, "pingheat")
	if !errors.Is(err, errIntervalTooShort) {
		t.Fatalf("expected errIntervalTooShort, got %v", err)
	}
}

func TestParseArgsIntervalLongForm(t *testing.T) {
	res, err := parseArgs([]string{"-interval", "500ms", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.cfg.Interval.Milliseconds() != 500 {
		t.Fatalf("expected Interval 500ms, got %v", res.cfg.Interval)
	}
}

func TestParseArgsIntervalLongFormTooShort(t *testing.T) {
	_, err := parseArgs([]string{"-interval", "50ms", "example.com"}, "pingheat")
	if !errors.Is(err, errIntervalTooShort) {
		t.Fatalf("expected errIntervalTooShort, got %v", err)
	}
}

func TestParseArgsBothIntervalFlags(t *testing.T) {
	// When both flags are set, -interval should take precedence
	res, err := parseArgs([]string{"-i", "200ms", "-interval", "300ms", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.cfg.Interval.Milliseconds() != 300 {
		t.Fatalf("expected Interval 300ms (from -interval flag), got %v", res.cfg.Interval)
	}
}

func TestParseArgsIntervalPrecedenceWithDefaultValue(t *testing.T) {
	// Edge case: -interval should take precedence even when explicitly set to default value
	// This tests the bug fix where we use flag.Visit() instead of comparing to defaults
	res, err := parseArgs([]string{"-i", "200ms", "-interval", "1s", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.cfg.Interval.Milliseconds() != 1000 {
		t.Fatalf("expected Interval 1s (from -interval flag), got %v", res.cfg.Interval)
	}
}

func TestParseArgsShowVersion(t *testing.T) {
	res, err := parseArgs([]string{"-version"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.showVersion {
		t.Fatalf("expected showVersion true")
	}
}

func TestParseArgsShowHelp(t *testing.T) {
	res, err := parseArgs([]string{"-help", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.cfg.ShowHelp {
		t.Fatalf("expected ShowHelp true")
	}
}

func TestParseArgsExporter(t *testing.T) {
	res, err := parseArgs([]string{"-exporter", ":9090", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.cfg.ExporterEnabled {
		t.Fatalf("expected ExporterEnabled true")
	}
	if res.cfg.ExporterAddr != ":9090" {
		t.Fatalf("expected ExporterAddr :9090, got %q", res.cfg.ExporterAddr)
	}
}

func TestParseArgsPprofNormalization(t *testing.T) {
	res, err := parseArgs([]string{"-pprof", ":6060", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.cfg.PprofEnabled {
		t.Fatalf("expected PprofEnabled true")
	}
	if res.cfg.PprofAddr != "127.0.0.1:6060" {
		t.Fatalf("expected PprofAddr 127.0.0.1:6060, got %q", res.cfg.PprofAddr)
	}
}

func TestParseArgsPprofExplicitHost(t *testing.T) {
	res, err := parseArgs([]string{"-pprof", "0.0.0.0:6060", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.cfg.PprofAddr != "0.0.0.0:6060" {
		t.Fatalf("expected PprofAddr 0.0.0.0:6060, got %q", res.cfg.PprofAddr)
	}
}
