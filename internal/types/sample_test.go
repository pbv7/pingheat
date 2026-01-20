package types

import (
	"testing"
	"time"
)

func TestRTTMs(t *testing.T) {
	sample := Sample{RTT: 1500 * time.Microsecond}
	if got := sample.RTTMs(); got != 1.5 {
		t.Fatalf("RTTMs() = %v, want 1.5", got)
	}

	sample.Timeout = true
	if got := sample.RTTMs(); got != -1 {
		t.Fatalf("RTTMs() timeout = %v, want -1", got)
	}
}
