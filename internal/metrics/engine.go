package metrics

import (
	"math"
	"sync"
	"time"

	"github.com/pbv7/pingheat/internal/types"
)

// Thresholds for brownout detection
const (
	BrownoutThresholdMs = 200 // RTT > 200ms is considered brownout
)

// Stats holds computed metrics.
type Stats struct {
	// Sample counts
	TotalSamples  int
	TotalTimeouts int
	TotalSuccess  int

	// Loss and availability
	LossPercent  float64
	AvailPercent float64 // 100 - LossPercent

	// RTT statistics (in time.Duration)
	MinRTT  time.Duration
	MaxRTT  time.Duration
	AvgRTT  time.Duration
	StdDev  time.Duration // Standard deviation
	Jitter  time.Duration // Mean absolute deviation between consecutive samples
	LastRTT time.Duration // Most recent RTT

	// RTT statistics in milliseconds (for display/export)
	MinRTTMs   float64
	MaxRTTMs   float64
	AvgRTTMs   float64
	StdDevMs   float64
	JitterMs   float64
	LastRTTMs  float64
	VarianceMs float64 // Variance in ms²

	// Streaks
	CurrentStreak  int // Positive = success streak, negative = timeout streak
	LongestSuccess int
	LongestTimeout int

	// Percentiles
	Percentiles Percentiles

	// Outage and instability patterns
	LossBursts      int // Number of separate timeout burst events
	BrownoutSamples int // Number of high-latency samples (> 200ms)
	BrownoutBursts  int // Number of brownout events (transitions to high latency)
	InBrownout      bool // Currently in brownout state

	// Timing
	StartTime        time.Time
	LastSuccessTime  time.Time
	LastTimeoutTime  time.Time
	TimeSinceTimeout time.Duration // Time since last timeout (0 if never timed out)
	UptimeSeconds    float64       // Seconds since monitoring started
}

// Engine computes metrics from ping samples.
type Engine struct {
	mu sync.RWMutex

	totalSamples   int
	totalTimeouts  int
	minRTT         time.Duration
	maxRTT         time.Duration
	sumRTT         time.Duration
	sumRTTSquares  float64 // Sum of RTT² in microseconds² for variance calculation
	lastRTT        time.Duration
	sumJitter      time.Duration
	jitterCount    int
	currentStreak  int
	longestSuccess int
	longestTimeout int
	percentiles    *PercentileCalculator

	// Outage tracking
	lossBursts      int  // Number of timeout burst events
	inTimeoutBurst  bool // Currently in a timeout burst
	brownoutSamples int  // Count of high-latency samples
	brownoutBursts  int  // Number of brownout events
	inBrownout      bool // Currently in brownout

	// Timing
	startTime       time.Time
	lastSuccessTime time.Time
	lastTimeoutTime time.Time
}

// NewEngine creates a new metrics engine.
func NewEngine() *Engine {
	return &Engine{
		minRTT:      time.Duration(math.MaxInt64),
		percentiles: NewPercentileCalculator(),
		startTime:   time.Now(),
	}
}

// Add processes a new ping sample.
func (e *Engine) Add(sample types.Sample) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.totalSamples++

	if sample.Timeout {
		e.totalTimeouts++
		e.lastTimeoutTime = sample.Timestamp

		// Track loss bursts (new burst when transitioning from success to timeout)
		if !e.inTimeoutBurst {
			e.lossBursts++
			e.inTimeoutBurst = true
		}

		// Exit brownout on timeout
		e.inBrownout = false

		// Update streak
		if e.currentStreak > 0 {
			e.currentStreak = -1
		} else {
			e.currentStreak--
		}
		if -e.currentStreak > e.longestTimeout {
			e.longestTimeout = -e.currentStreak
		}
		return
	}

	// Successful ping
	e.lastSuccessTime = sample.Timestamp
	e.inTimeoutBurst = false // End timeout burst on success
	rtt := sample.RTT

	// Check for brownout (high latency)
	rttMs := float64(rtt.Microseconds()) / 1000.0
	if rttMs > BrownoutThresholdMs {
		e.brownoutSamples++
		if !e.inBrownout {
			e.brownoutBursts++
			e.inBrownout = true
		}
	} else {
		e.inBrownout = false
	}

	if rtt < e.minRTT {
		e.minRTT = rtt
	}
	if rtt > e.maxRTT {
		e.maxRTT = rtt
	}
	e.sumRTT += rtt

	// Track sum of squares for variance/stddev (in microseconds)
	rttUs := float64(rtt.Microseconds())
	e.sumRTTSquares += rttUs * rttUs

	// Calculate jitter (variation from last RTT)
	if e.lastRTT > 0 {
		diff := rtt - e.lastRTT
		if diff < 0 {
			diff = -diff
		}
		e.sumJitter += diff
		e.jitterCount++
	}
	e.lastRTT = rtt

	// Update streak
	if e.currentStreak < 0 {
		e.currentStreak = 1
	} else {
		e.currentStreak++
	}
	if e.currentStreak > e.longestSuccess {
		e.longestSuccess = e.currentStreak
	}

	// Add to percentile calculator
	e.percentiles.Add(rtt)
}

// Stats returns the current computed metrics.
func (e *Engine) Stats() Stats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	successCount := e.totalSamples - e.totalTimeouts

	stats := Stats{
		TotalSamples:    e.totalSamples,
		TotalTimeouts:   e.totalTimeouts,
		TotalSuccess:    successCount,
		CurrentStreak:   e.currentStreak,
		LongestSuccess:  e.longestSuccess,
		LongestTimeout:  e.longestTimeout,
		LossBursts:      e.lossBursts,
		BrownoutSamples: e.brownoutSamples,
		BrownoutBursts:  e.brownoutBursts,
		InBrownout:      e.inBrownout,
		StartTime:       e.startTime,
		UptimeSeconds:   time.Since(e.startTime).Seconds(),
	}

	if e.totalSamples > 0 {
		stats.LossPercent = float64(e.totalTimeouts) / float64(e.totalSamples) * 100
		stats.AvailPercent = 100 - stats.LossPercent
	}

	if successCount > 0 {
		stats.MinRTT = e.minRTT
		stats.MaxRTT = e.maxRTT
		stats.AvgRTT = e.sumRTT / time.Duration(successCount)
		stats.LastRTT = e.lastRTT
		stats.Percentiles = e.percentiles.GetPercentiles()

		// Calculate variance and standard deviation
		// Variance = E[X²] - (E[X])²
		n := float64(successCount)
		meanUs := float64(e.sumRTT.Microseconds()) / n
		varianceUs := (e.sumRTTSquares / n) - (meanUs * meanUs)
		if varianceUs < 0 {
			varianceUs = 0 // Handle floating point errors
		}
		stdDevUs := math.Sqrt(varianceUs)
		stats.StdDev = time.Duration(stdDevUs) * time.Microsecond

		// Convert to milliseconds for display
		stats.MinRTTMs = float64(e.minRTT.Microseconds()) / 1000.0
		stats.MaxRTTMs = float64(e.maxRTT.Microseconds()) / 1000.0
		stats.AvgRTTMs = float64(stats.AvgRTT.Microseconds()) / 1000.0
		stats.StdDevMs = stdDevUs / 1000.0
		stats.VarianceMs = varianceUs / 1000000.0 // Convert µs² to ms²
		stats.LastRTTMs = float64(e.lastRTT.Microseconds()) / 1000.0

		stats.LastSuccessTime = e.lastSuccessTime
	}

	if e.jitterCount > 0 {
		stats.Jitter = e.sumJitter / time.Duration(e.jitterCount)
		stats.JitterMs = float64(stats.Jitter.Microseconds()) / 1000.0
	}

	if !e.lastTimeoutTime.IsZero() {
		stats.LastTimeoutTime = e.lastTimeoutTime
		stats.TimeSinceTimeout = time.Since(e.lastTimeoutTime)
	}

	return stats
}

// Reset clears all metrics.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.totalSamples = 0
	e.totalTimeouts = 0
	e.minRTT = time.Duration(math.MaxInt64)
	e.maxRTT = 0
	e.sumRTT = 0
	e.sumRTTSquares = 0
	e.lastRTT = 0
	e.sumJitter = 0
	e.jitterCount = 0
	e.currentStreak = 0
	e.longestSuccess = 0
	e.longestTimeout = 0
	e.lossBursts = 0
	e.inTimeoutBurst = false
	e.brownoutSamples = 0
	e.brownoutBursts = 0
	e.inBrownout = false
	e.percentiles.Reset()
	e.startTime = time.Now()
	e.lastSuccessTime = time.Time{}
	e.lastTimeoutTime = time.Time{}
}
