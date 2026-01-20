package exporter

import (
	"context"
	"net/http"
	"sync"

	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Exporter exports ping metrics to Prometheus.
type Exporter struct {
	addr   string
	target string
	server *http.Server

	mu    sync.RWMutex
	stats metrics.Stats

	// Prometheus metrics - Counters
	pingSentTotal    *prometheus.CounterVec
	pingSuccessTotal *prometheus.CounterVec
	pingTimeoutTotal *prometheus.CounterVec

	// Gauges - Latency
	pingLatencyMs  *prometheus.GaugeVec
	pingStdDevMs   *prometheus.GaugeVec
	pingVarianceMs *prometheus.GaugeVec
	pingJitterMs   *prometheus.GaugeVec
	pingLastRTTMs  *prometheus.GaugeVec

	// Gauges - Percentiles
	pingLatencyP50Ms *prometheus.GaugeVec
	pingLatencyP90Ms *prometheus.GaugeVec
	pingLatencyP95Ms *prometheus.GaugeVec
	pingLatencyP99Ms *prometheus.GaugeVec

	// Gauges - Availability
	pingLossPercent  *prometheus.GaugeVec
	pingAvailPercent *prometheus.GaugeVec

	// Gauges - Streaks
	pingCurrentStreak  *prometheus.GaugeVec
	pingLongestSuccess *prometheus.GaugeVec
	pingLongestTimeout *prometheus.GaugeVec

	// Gauges - Instability patterns
	pingLossBursts      *prometheus.GaugeVec
	pingBrownoutSamples *prometheus.GaugeVec
	pingBrownoutBursts  *prometheus.GaugeVec
	pingInBrownout      *prometheus.GaugeVec

	// Gauges - Timing
	pingUptimeSeconds *prometheus.GaugeVec

	// Info - for "up" logic
	pingUp *prometheus.GaugeVec
}

// NewExporter creates a new Prometheus exporter.
func NewExporter(addr, target string) *Exporter {
	e := &Exporter{
		addr:   addr,
		target: target,
	}

	labels := []string{"target"}

	// Counters
	e.pingSentTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pingheat_ping_sent_total",
		Help: "Total number of ping packets sent",
	}, labels)

	e.pingSuccessTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pingheat_ping_success_total",
		Help: "Total number of successful ping responses",
	}, labels)

	e.pingTimeoutTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pingheat_ping_timeout_total",
		Help: "Total number of ping timeouts",
	}, labels)

	// Latency gauges
	e.pingLatencyMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_latency_ms",
		Help: "Ping latency in milliseconds (min, avg, max)",
	}, append(labels, "stat"))

	e.pingStdDevMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_stddev_ms",
		Help: "Standard deviation of ping latency in milliseconds",
	}, labels)

	e.pingVarianceMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_variance_ms2",
		Help: "Variance of ping latency in milliseconds squared",
	}, labels)

	e.pingJitterMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_jitter_ms",
		Help: "Ping jitter (mean absolute deviation) in milliseconds",
	}, labels)

	e.pingLastRTTMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_last_rtt_ms",
		Help: "Most recent ping RTT in milliseconds (-1 if last was timeout)",
	}, labels)

	// Percentile gauges
	e.pingLatencyP50Ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_latency_p50_ms",
		Help: "50th percentile (median) latency in milliseconds",
	}, labels)

	e.pingLatencyP90Ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_latency_p90_ms",
		Help: "90th percentile latency in milliseconds",
	}, labels)

	e.pingLatencyP95Ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_latency_p95_ms",
		Help: "95th percentile latency in milliseconds",
	}, labels)

	e.pingLatencyP99Ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_latency_p99_ms",
		Help: "99th percentile latency in milliseconds",
	}, labels)

	// Availability gauges
	e.pingLossPercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_loss_percent",
		Help: "Packet loss percentage (0-100)",
	}, labels)

	e.pingAvailPercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_availability_percent",
		Help: "Availability percentage (0-100)",
	}, labels)

	// Streak gauges
	e.pingCurrentStreak = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_current_streak",
		Help: "Current streak (positive=success, negative=timeout)",
	}, labels)

	e.pingLongestSuccess = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_longest_success_streak",
		Help: "Longest consecutive successful pings",
	}, labels)

	e.pingLongestTimeout = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_longest_timeout_streak",
		Help: "Longest consecutive timeout streak",
	}, labels)

	// Instability pattern gauges
	e.pingLossBursts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_loss_bursts_total",
		Help: "Number of separate packet loss burst events (outages)",
	}, labels)

	e.pingBrownoutSamples = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_brownout_samples_total",
		Help: "Total number of high-latency samples (>200ms)",
	}, labels)

	e.pingBrownoutBursts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_brownout_bursts_total",
		Help: "Number of brownout events (transitions to high latency)",
	}, labels)

	e.pingInBrownout = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_in_brownout",
		Help: "Currently in brownout state (1=yes, 0=no)",
	}, labels)

	// Timing gauges
	e.pingUptimeSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_uptime_seconds",
		Help: "Seconds since monitoring started",
	}, labels)

	// Up gauge for alerting
	e.pingUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pingheat_ping_up",
		Help: "Target is reachable (1=up, 0=down based on last ping)",
	}, labels)

	return e
}

// Start starts the Prometheus HTTP server.
func (e *Exporter) Start(ctx context.Context) error {
	// Register metrics
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		e.pingSentTotal,
		e.pingSuccessTotal,
		e.pingTimeoutTotal,
		e.pingLatencyMs,
		e.pingStdDevMs,
		e.pingVarianceMs,
		e.pingJitterMs,
		e.pingLastRTTMs,
		e.pingLatencyP50Ms,
		e.pingLatencyP90Ms,
		e.pingLatencyP95Ms,
		e.pingLatencyP99Ms,
		e.pingLossPercent,
		e.pingAvailPercent,
		e.pingCurrentStreak,
		e.pingLongestSuccess,
		e.pingLongestTimeout,
		e.pingLossBursts,
		e.pingBrownoutSamples,
		e.pingBrownoutBursts,
		e.pingInBrownout,
		e.pingUptimeSeconds,
		e.pingUp,
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	e.server = &http.Server{
		Addr:    e.addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		e.server.Shutdown(context.Background())
	}()

	err := e.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Update updates the exported metrics.
func (e *Exporter) Update(stats metrics.Stats) {
	e.mu.Lock()
	defer e.mu.Unlock()

	prevStats := e.stats
	e.stats = stats

	// Update counters (incremental)
	if stats.TotalSamples > prevStats.TotalSamples {
		e.pingSentTotal.WithLabelValues(e.target).Add(float64(stats.TotalSamples - prevStats.TotalSamples))
	}
	if stats.TotalSuccess > prevStats.TotalSuccess {
		e.pingSuccessTotal.WithLabelValues(e.target).Add(float64(stats.TotalSuccess - prevStats.TotalSuccess))
	}
	if stats.TotalTimeouts > prevStats.TotalTimeouts {
		e.pingTimeoutTotal.WithLabelValues(e.target).Add(float64(stats.TotalTimeouts - prevStats.TotalTimeouts))
	}

	// Update availability gauges
	e.pingLossPercent.WithLabelValues(e.target).Set(stats.LossPercent)
	e.pingAvailPercent.WithLabelValues(e.target).Set(stats.AvailPercent)

	// Update streak gauges
	e.pingCurrentStreak.WithLabelValues(e.target).Set(float64(stats.CurrentStreak))
	e.pingLongestSuccess.WithLabelValues(e.target).Set(float64(stats.LongestSuccess))
	e.pingLongestTimeout.WithLabelValues(e.target).Set(float64(stats.LongestTimeout))

	// Update instability pattern gauges
	e.pingLossBursts.WithLabelValues(e.target).Set(float64(stats.LossBursts))
	e.pingBrownoutSamples.WithLabelValues(e.target).Set(float64(stats.BrownoutSamples))
	e.pingBrownoutBursts.WithLabelValues(e.target).Set(float64(stats.BrownoutBursts))
	if stats.InBrownout {
		e.pingInBrownout.WithLabelValues(e.target).Set(1)
	} else {
		e.pingInBrownout.WithLabelValues(e.target).Set(0)
	}

	// Update uptime
	e.pingUptimeSeconds.WithLabelValues(e.target).Set(stats.UptimeSeconds)

	// Update "up" status based on current streak
	// Up = 1 if last ping was successful (positive streak), 0 if in timeout streak
	if stats.CurrentStreak > 0 {
		e.pingUp.WithLabelValues(e.target).Set(1)
	} else {
		e.pingUp.WithLabelValues(e.target).Set(0)
	}

	// Update latency gauges (only if we have successful pings)
	if stats.TotalSuccess > 0 {
		e.pingLatencyMs.WithLabelValues(e.target, "min").Set(stats.MinRTTMs)
		e.pingLatencyMs.WithLabelValues(e.target, "avg").Set(stats.AvgRTTMs)
		e.pingLatencyMs.WithLabelValues(e.target, "max").Set(stats.MaxRTTMs)

		e.pingStdDevMs.WithLabelValues(e.target).Set(stats.StdDevMs)
		e.pingVarianceMs.WithLabelValues(e.target).Set(stats.VarianceMs)
		e.pingJitterMs.WithLabelValues(e.target).Set(stats.JitterMs)

		// LastRTT: set to actual value if up, -1 if currently in timeout
		if stats.CurrentStreak > 0 {
			e.pingLastRTTMs.WithLabelValues(e.target).Set(stats.LastRTTMs)
		} else {
			e.pingLastRTTMs.WithLabelValues(e.target).Set(-1)
		}

		e.pingLatencyP50Ms.WithLabelValues(e.target).Set(stats.Percentiles.P50)
		e.pingLatencyP90Ms.WithLabelValues(e.target).Set(stats.Percentiles.P90)
		e.pingLatencyP95Ms.WithLabelValues(e.target).Set(stats.Percentiles.P95)
		e.pingLatencyP99Ms.WithLabelValues(e.target).Set(stats.Percentiles.P99)
	}
}
