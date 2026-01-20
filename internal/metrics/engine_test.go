package metrics

import (
	"testing"
	"time"

	"github.com/pbv7/pingheat/internal/types"
)

func TestEngine_Basic(t *testing.T) {
	e := NewEngine()

	// Add some samples
	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 20 * time.Millisecond})
	e.Add(types.Sample{RTT: 30 * time.Millisecond})

	stats := e.Stats()

	if stats.TotalSamples != 3 {
		t.Errorf("TotalSamples = %d, want 3", stats.TotalSamples)
	}

	if stats.TotalTimeouts != 0 {
		t.Errorf("TotalTimeouts = %d, want 0", stats.TotalTimeouts)
	}

	if stats.MinRTT != 10*time.Millisecond {
		t.Errorf("MinRTT = %v, want 10ms", stats.MinRTT)
	}

	if stats.MaxRTT != 30*time.Millisecond {
		t.Errorf("MaxRTT = %v, want 30ms", stats.MaxRTT)
	}

	if stats.AvgRTT != 20*time.Millisecond {
		t.Errorf("AvgRTT = %v, want 20ms", stats.AvgRTT)
	}
}

func TestEngine_Timeouts(t *testing.T) {
	e := NewEngine()

	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{Timeout: true})
	e.Add(types.Sample{RTT: 20 * time.Millisecond})
	e.Add(types.Sample{Timeout: true})

	stats := e.Stats()

	if stats.TotalSamples != 4 {
		t.Errorf("TotalSamples = %d, want 4", stats.TotalSamples)
	}

	if stats.TotalTimeouts != 2 {
		t.Errorf("TotalTimeouts = %d, want 2", stats.TotalTimeouts)
	}

	if stats.LossPercent != 50 {
		t.Errorf("LossPercent = %f, want 50", stats.LossPercent)
	}
}

func TestEngine_Streaks(t *testing.T) {
	e := NewEngine()

	// 3 successes
	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 10 * time.Millisecond})

	stats := e.Stats()
	if stats.CurrentStreak != 3 {
		t.Errorf("CurrentStreak = %d, want 3", stats.CurrentStreak)
	}
	if stats.LongestSuccess != 3 {
		t.Errorf("LongestSuccess = %d, want 3", stats.LongestSuccess)
	}

	// 2 timeouts
	e.Add(types.Sample{Timeout: true})
	e.Add(types.Sample{Timeout: true})

	stats = e.Stats()
	if stats.CurrentStreak != -2 {
		t.Errorf("CurrentStreak = %d, want -2", stats.CurrentStreak)
	}
	if stats.LongestTimeout != 2 {
		t.Errorf("LongestTimeout = %d, want 2", stats.LongestTimeout)
	}

	// Back to success
	e.Add(types.Sample{RTT: 10 * time.Millisecond})

	stats = e.Stats()
	if stats.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %d, want 1", stats.CurrentStreak)
	}
}

func TestEngine_Jitter(t *testing.T) {
	e := NewEngine()

	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 20 * time.Millisecond})
	e.Add(types.Sample{RTT: 15 * time.Millisecond})

	stats := e.Stats()

	// Jitter is average of |20-10| and |15-20| = (10 + 5) / 2 = 7.5ms
	expectedJitter := 7500 * time.Microsecond
	if stats.Jitter != expectedJitter {
		t.Errorf("Jitter = %v, want %v", stats.Jitter, expectedJitter)
	}
}

func TestEngine_StdDev(t *testing.T) {
	e := NewEngine()

	// Values: 10, 20, 30 ms
	// Mean = 20ms
	// Variance = ((10-20)² + (20-20)² + (30-20)²) / 3 = (100 + 0 + 100) / 3 = 66.67
	// StdDev = sqrt(66.67) ≈ 8.16 ms
	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 20 * time.Millisecond})
	e.Add(types.Sample{RTT: 30 * time.Millisecond})

	stats := e.Stats()

	// Check StdDev is approximately 8.16ms (allow 0.1ms tolerance)
	expectedStdDev := 8165 * time.Microsecond // ~8.165ms
	diff := stats.StdDev - expectedStdDev
	if diff < 0 {
		diff = -diff
	}
	if diff > 100*time.Microsecond {
		t.Errorf("StdDev = %v, want ~%v", stats.StdDev, expectedStdDev)
	}

	// Check variance is approximately 66.67 ms²
	if stats.VarianceMs < 66 || stats.VarianceMs > 67 {
		t.Errorf("VarianceMs = %v, want ~66.67", stats.VarianceMs)
	}
}

func TestEngine_Availability(t *testing.T) {
	e := NewEngine()

	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{Timeout: true})

	stats := e.Stats()

	if stats.AvailPercent != 75 {
		t.Errorf("AvailPercent = %f, want 75", stats.AvailPercent)
	}

	if stats.TotalSuccess != 3 {
		t.Errorf("TotalSuccess = %d, want 3", stats.TotalSuccess)
	}
}

func TestEngine_Reset(t *testing.T) {
	e := NewEngine()

	e.Add(types.Sample{RTT: 10 * time.Millisecond})
	e.Add(types.Sample{Timeout: true})

	e.Reset()

	stats := e.Stats()
	if stats.TotalSamples != 0 {
		t.Errorf("TotalSamples after reset = %d, want 0", stats.TotalSamples)
	}
}
