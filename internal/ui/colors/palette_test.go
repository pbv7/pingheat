package colors

import (
	"testing"
	"time"
)

func TestClassifyMsThresholds(t *testing.T) {
	if ClassifyMs(-1) != ColorTimeout {
		t.Fatalf("expected timeout color for negative ms")
	}
	if ClassifyMs(0) != ColorExcellent {
		t.Fatalf("expected excellent color for 0ms")
	}
	if ClassifyMs(ThresholdExcellent) != ColorExcellent {
		t.Fatalf("expected excellent color at threshold")
	}
	if ClassifyMs(ThresholdExcellent+1) != ColorGood {
		t.Fatalf("expected good color above excellent threshold")
	}
	if ClassifyMs(ThresholdGood+1) != ColorFair {
		t.Fatalf("expected fair color above good threshold")
	}
	if ClassifyMs(ThresholdFair+1) != ColorPoor {
		t.Fatalf("expected poor color above fair threshold")
	}
	if ClassifyMs(ThresholdPoor+1) != ColorBad {
		t.Fatalf("expected bad color above poor threshold")
	}
}

func TestClassifyAndBackground(t *testing.T) {
	if Classify(50*time.Millisecond) != ColorGood {
		t.Fatalf("expected good color for 50ms")
	}
	if ClassifyBG(50*time.Millisecond) != BGGood {
		t.Fatalf("expected good background for 50ms")
	}
}

func TestForTimeout(t *testing.T) {
	if !ForTimeout(-1) {
		t.Fatalf("expected timeout for negative ms")
	}
	if ForTimeout(0) {
		t.Fatalf("expected no timeout for zero ms")
	}
}
