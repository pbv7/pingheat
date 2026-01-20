package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Target != "" {
		t.Fatalf("Target=%q, want empty", cfg.Target)
	}
	if cfg.Interval <= 0 {
		t.Fatalf("Interval=%v, want > 0", cfg.Interval)
	}
	if cfg.HistorySize <= 0 {
		t.Fatalf("HistorySize=%d, want > 0", cfg.HistorySize)
	}
	if cfg.MetricsBufferSize <= 0 {
		t.Fatalf("MetricsBufferSize=%d, want > 0", cfg.MetricsBufferSize)
	}
	if cfg.ExporterEnabled {
		t.Fatalf("ExporterEnabled=true, want false")
	}
	if cfg.ExporterAddr == "" {
		t.Fatalf("ExporterAddr empty, want default")
	}
	if cfg.PprofEnabled {
		t.Fatalf("PprofEnabled=true, want false")
	}
	if cfg.PprofAddr == "" {
		t.Fatalf("PprofAddr empty, want default")
	}
	if cfg.ShowHelp {
		t.Fatalf("ShowHelp=true, want false")
	}
}
