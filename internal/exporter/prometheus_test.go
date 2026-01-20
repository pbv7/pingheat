package exporter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestExporterUpdateMetrics(t *testing.T) {
	e := NewExporter(":0", "target")
	stats := metrics.Stats{
		TotalSamples:    2,
		TotalSuccess:    2,
		TotalTimeouts:   0,
		LossPercent:     0,
		AvailPercent:    100,
		CurrentStreak:   2,
		LongestSuccess:  2,
		LongestTimeout:  0,
		LossBursts:      1,
		BrownoutSamples: 1,
		BrownoutBursts:  1,
		InBrownout:      true,
		UptimeSeconds:   10,
		MinRTTMs:        1.1,
		AvgRTTMs:        2.2,
		MaxRTTMs:        3.3,
		StdDevMs:        0.5,
		VarianceMs:      0.25,
		JitterMs:        0.2,
		LastRTTMs:       3.3,
		Percentiles: metrics.Percentiles{
			P50: 2.2,
			P90: 3.0,
			P95: 3.5,
			P99: 4.0,
		},
	}

	e.Update(stats)

	if v := testutil.ToFloat64(e.pingSentTotal.WithLabelValues("target")); v != 2 {
		t.Fatalf("pingSentTotal=%v, want 2", v)
	}
	if v := testutil.ToFloat64(e.pingSuccessTotal.WithLabelValues("target")); v != 2 {
		t.Fatalf("pingSuccessTotal=%v, want 2", v)
	}
	if v := testutil.ToFloat64(e.pingTimeoutTotal.WithLabelValues("target")); v != 0 {
		t.Fatalf("pingTimeoutTotal=%v, want 0", v)
	}
	if v := testutil.ToFloat64(e.pingInBrownout.WithLabelValues("target")); v != 1 {
		t.Fatalf("pingInBrownout=%v, want 1", v)
	}
	if v := testutil.ToFloat64(e.pingLastRTTMs.WithLabelValues("target")); v != 3.3 {
		t.Fatalf("pingLastRTTMs=%v, want 3.3", v)
	}
	if v := testutil.ToFloat64(e.pingUp.WithLabelValues("target")); v != 1 {
		t.Fatalf("pingUp=%v, want 1", v)
	}

	stats.TotalSamples = 3
	stats.TotalTimeouts = 1
	stats.CurrentStreak = -1
	stats.InBrownout = false
	e.Update(stats)

	if v := testutil.ToFloat64(e.pingSentTotal.WithLabelValues("target")); v != 3 {
		t.Fatalf("pingSentTotal=%v, want 3", v)
	}
	if v := testutil.ToFloat64(e.pingTimeoutTotal.WithLabelValues("target")); v != 1 {
		t.Fatalf("pingTimeoutTotal=%v, want 1", v)
	}
	if v := testutil.ToFloat64(e.pingLastRTTMs.WithLabelValues("target")); v != -1 {
		t.Fatalf("pingLastRTTMs=%v, want -1", v)
	}
	if v := testutil.ToFloat64(e.pingUp.WithLabelValues("target")); v != 0 {
		t.Fatalf("pingUp=%v, want 0", v)
	}
}

func TestExporterServerHandlersAndTimeouts(t *testing.T) {
	e := NewExporter("127.0.0.1:9090", "target")
	reg := prometheus.NewRegistry()
	e.register(reg)
	server := e.newServer(reg)
	e.Update(metrics.Stats{TotalSamples: 1, TotalSuccess: 1})

	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout=%v, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("ReadTimeout=%v, want 10s", server.ReadTimeout)
	}
	if server.WriteTimeout != 10*time.Second {
		t.Fatalf("WriteTimeout=%v, want 10s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout=%v, want 60s", server.IdleTimeout)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d, want 200", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "OK" {
		t.Fatalf("health body=%q, want OK", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status=%d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "pingheat_ping_sent_total") {
		t.Fatalf("metrics output missing pingheat_ping_sent_total")
	}
}
