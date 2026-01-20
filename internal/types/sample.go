package types

import "time"

// Sample represents a single ping measurement.
type Sample struct {
	Timestamp time.Time
	Sequence  int
	RTT       time.Duration
	Timeout   bool
}

// IsTimeout returns true if this sample represents a timeout.
func (s Sample) IsTimeout() bool {
	return s.Timeout
}

// RTTMs returns the RTT in milliseconds.
func (s Sample) RTTMs() float64 {
	if s.Timeout {
		return -1
	}
	return float64(s.RTT.Microseconds()) / 1000.0
}
