package metrics

import (
	"sort"
	"time"
)

// PercentileCalculator computes percentiles from RTT samples.
type PercentileCalculator struct {
	values []float64
	sorted bool
}

// NewPercentileCalculator creates a new percentile calculator.
func NewPercentileCalculator() *PercentileCalculator {
	return &PercentileCalculator{
		values: make([]float64, 0, 1024),
	}
}

// Add adds a new RTT value (in milliseconds).
func (p *PercentileCalculator) Add(rtt time.Duration) {
	p.values = append(p.values, float64(rtt.Microseconds())/1000.0)
	p.sorted = false
}

// AddMs adds a new RTT value already in milliseconds.
func (p *PercentileCalculator) AddMs(ms float64) {
	p.values = append(p.values, ms)
	p.sorted = false
}

// Reset clears all values.
func (p *PercentileCalculator) Reset() {
	p.values = p.values[:0]
	p.sorted = false
}

// Count returns the number of values.
func (p *PercentileCalculator) Count() int {
	return len(p.values)
}

// ensureSorted sorts the values if needed.
func (p *PercentileCalculator) ensureSorted() {
	if !p.sorted && len(p.values) > 0 {
		sort.Float64s(p.values)
		p.sorted = true
	}
}

// Percentile returns the value at the given percentile (0-100).
func (p *PercentileCalculator) Percentile(pct float64) float64 {
	if len(p.values) == 0 {
		return 0
	}

	p.ensureSorted()

	if pct <= 0 {
		return p.values[0]
	}
	if pct >= 100 {
		return p.values[len(p.values)-1]
	}

	// Linear interpolation
	rank := (pct / 100.0) * float64(len(p.values)-1)
	lower := int(rank)
	upper := lower + 1
	if upper >= len(p.values) {
		return p.values[len(p.values)-1]
	}

	frac := rank - float64(lower)
	return p.values[lower] + frac*(p.values[upper]-p.values[lower])
}

// P50 returns the 50th percentile (median).
func (p *PercentileCalculator) P50() float64 {
	return p.Percentile(50)
}

// P90 returns the 90th percentile.
func (p *PercentileCalculator) P90() float64 {
	return p.Percentile(90)
}

// P95 returns the 95th percentile.
func (p *PercentileCalculator) P95() float64 {
	return p.Percentile(95)
}

// P99 returns the 99th percentile.
func (p *PercentileCalculator) P99() float64 {
	return p.Percentile(99)
}

// Percentiles returns common percentiles as a struct.
type Percentiles struct {
	P50 float64
	P90 float64
	P95 float64
	P99 float64
}

// GetPercentiles returns all common percentiles.
func (p *PercentileCalculator) GetPercentiles() Percentiles {
	return Percentiles{
		P50: p.P50(),
		P90: p.P90(),
		P95: p.P95(),
		P99: p.P99(),
	}
}
