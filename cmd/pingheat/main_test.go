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
