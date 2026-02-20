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

// --- Interval Validation Tests ---

func TestParseArgsIntervalTooLong(t *testing.T) {
	_, err := parseArgs([]string{"-i", "2h", "example.com"}, "pingheat")
	if !errors.Is(err, errIntervalTooLong) {
		t.Fatalf("expected errIntervalTooLong, got %v", err)
	}
}

func TestParseArgsIntervalLongFormTooLong(t *testing.T) {
	_, err := parseArgs([]string{"-interval", "90m", "example.com"}, "pingheat")
	if !errors.Is(err, errIntervalTooLong) {
		t.Fatalf("expected errIntervalTooLong, got %v", err)
	}
}

func TestParseArgsIntervalMaxAllowed(t *testing.T) {
	res, err := parseArgs([]string{"-i", "1h", "example.com"}, "pingheat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.cfg.Interval.Hours() != 1 {
		t.Fatalf("expected Interval 1h, got %v", res.cfg.Interval)
	}
}

// --- Target Format Validation Tests ---

func TestParseArgsInvalidTarget(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{"spaces in hostname", "google .com"},
		{"shell metacharacter ampersand", "google.com && echo pwned"},
		{"shell metacharacter pipe", "google.com | cat"},
		{"shell metacharacter semicolon", "google.com; ls"},
		{"invalid characters", "google!.com"},
		{"double dots", "google..com"},
		{"trailing hyphen", "google.com-"},
		{"underscore in hostname", "google_com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseArgs([]string{tt.target}, "pingheat")
			if !errors.Is(err, errInvalidTarget) {
				t.Errorf("expected errInvalidTarget for %q, got %v", tt.target, err)
			}
		})
	}
}

func TestParseArgsValidTargets(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{"simple hostname", "google.com"},
		{"subdomain", "www.google.com"},
		{"multiple subdomains", "api.dev.example.com"},
		{"hostname with hyphen", "my-server.example.com"},
		{"single label hostname", "localhost"},
		{"IPv4 address", "8.8.8.8"},
		{"IPv4 private", "192.168.1.1"},
		{"IPv6 full", "2001:4860:4860::8888"},
		{"IPv6 loopback", "::1"},
		{"IPv6 with brackets", "[::1]"},
		{"IPv6 full with brackets", "[2001:db8::1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseArgs([]string{tt.target}, "pingheat")
			if err != nil {
				t.Errorf("unexpected error for %q: %v", tt.target, err)
			}
			if res.cfg.Target != tt.target {
				t.Errorf("expected target %q, got %q", tt.target, res.cfg.Target)
			}
		})
	}
}

// --- Port Validation Tests ---

func TestParseArgsExporterInvalidPort(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{"port zero", ":0"},
		{"port too high", ":99999"},
		{"negative port", ":-1"},
		{"malformed port", ":abc"},
		{"no port", "localhost"},
		{"malformed address", "::invalid::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseArgs([]string{"-exporter", tt.addr, "example.com"}, "pingheat")
			if err == nil {
				t.Errorf("expected error for exporter address %q, got nil", tt.addr)
			}
		})
	}
}

func TestParseArgsPprofInvalidPort(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{"port zero", ":0"},
		{"port too high", ":100000"},
		{"negative port", ":-5"},
		{"malformed port", ":xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseArgs([]string{"-pprof", tt.addr, "example.com"}, "pingheat")
			if err == nil {
				t.Errorf("expected error for pprof address %q, got nil", tt.addr)
			}
		})
	}
}

func TestParseArgsExporterValidPorts(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{"port only", ":9090"},
		{"localhost with port", "localhost:9090"},
		{"IPv4 with port", "0.0.0.0:9090"},
		{"IPv6 with port", "[::1]:9090"},
		{"privileged port 80", ":80"},
		{"privileged port 443", ":443"},
		{"max port", ":65535"},
		{"min port", ":1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseArgs([]string{"-exporter", tt.addr, "example.com"}, "pingheat")
			if err != nil {
				t.Errorf("unexpected error for exporter address %q: %v", tt.addr, err)
			}
			if !res.cfg.ExporterEnabled {
				t.Errorf("expected ExporterEnabled true")
			}
			if res.cfg.ExporterAddr != tt.addr {
				t.Errorf("expected ExporterAddr %q, got %q", tt.addr, res.cfg.ExporterAddr)
			}
		})
	}
}

func TestParseArgsPprofValidPorts(t *testing.T) {
	tests := []struct {
		name         string
		addr         string
		expectedAddr string // After normalization
	}{
		{"port only", ":6060", "127.0.0.1:6060"},
		{"explicit localhost", "localhost:6060", "localhost:6060"},
		{"explicit IPv4", "0.0.0.0:6060", "0.0.0.0:6060"},
		{"IPv6", "[::1]:6060", "[::1]:6060"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseArgs([]string{"-pprof", tt.addr, "example.com"}, "pingheat")
			if err != nil {
				t.Errorf("unexpected error for pprof address %q: %v", tt.addr, err)
			}
			if !res.cfg.PprofEnabled {
				t.Errorf("expected PprofEnabled true")
			}
			if res.cfg.PprofAddr != tt.expectedAddr {
				t.Errorf("expected PprofAddr %q, got %q", tt.expectedAddr, res.cfg.PprofAddr)
			}
		})
	}
}

func TestParseArgsPprofPortValidationBeforeNormalization(t *testing.T) {
	// Verify that port validation happens BEFORE localhost normalization
	// This test ensures :0 is rejected even though it would be normalized to 127.0.0.1:0
	_, err := parseArgs([]string{"-pprof", ":0", "example.com"}, "pingheat")
	if err == nil {
		t.Fatalf("expected error for pprof port 0, got nil")
	}
}
