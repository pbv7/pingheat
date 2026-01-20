package metrics

import (
	"math"
	"testing"
	"time"
)

func TestPercentileCalculator_Basic(t *testing.T) {
	p := NewPercentileCalculator()

	// Add values 1-100
	for i := 1; i <= 100; i++ {
		p.AddMs(float64(i))
	}

	if p.Count() != 100 {
		t.Errorf("Count = %d, want 100", p.Count())
	}

	// P50 should be around 50
	p50 := p.P50()
	if math.Abs(p50-50.5) > 1 {
		t.Errorf("P50 = %f, want ~50.5", p50)
	}

	// P99 should be around 99
	p99 := p.P99()
	if math.Abs(p99-99.01) > 1 {
		t.Errorf("P99 = %f, want ~99", p99)
	}
}

func TestPercentileCalculator_AddDuration(t *testing.T) {
	p := NewPercentileCalculator()

	p.Add(10 * time.Millisecond)
	p.Add(20 * time.Millisecond)
	p.Add(30 * time.Millisecond)

	// Median should be 20ms
	p50 := p.P50()
	if p50 != 20 {
		t.Errorf("P50 = %f, want 20", p50)
	}
}

func TestPercentileCalculator_Empty(t *testing.T) {
	p := NewPercentileCalculator()

	if p.P50() != 0 {
		t.Errorf("P50 of empty set = %f, want 0", p.P50())
	}
}

func TestPercentileCalculator_Reset(t *testing.T) {
	p := NewPercentileCalculator()

	p.AddMs(10)
	p.AddMs(20)

	p.Reset()

	if p.Count() != 0 {
		t.Errorf("Count after reset = %d, want 0", p.Count())
	}
}

func TestPercentileCalculator_GetPercentiles(t *testing.T) {
	p := NewPercentileCalculator()

	for i := 1; i <= 100; i++ {
		p.AddMs(float64(i))
	}

	pcts := p.GetPercentiles()

	if pcts.P50 == 0 || pcts.P90 == 0 || pcts.P95 == 0 || pcts.P99 == 0 {
		t.Error("GetPercentiles returned zero values")
	}

	// Verify ordering
	if pcts.P50 >= pcts.P90 || pcts.P90 >= pcts.P95 || pcts.P95 >= pcts.P99 {
		t.Errorf("Percentiles not in order: P50=%f, P90=%f, P95=%f, P99=%f",
			pcts.P50, pcts.P90, pcts.P95, pcts.P99)
	}
}
