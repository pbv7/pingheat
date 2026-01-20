package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/pbv7/pingheat/internal/ping"
)

func newTestModel() Model {
	cfg := config.DefaultConfig()
	return NewModel(cfg, make(chan ping.Sample), make(chan metrics.Stats))
}

func TestGridDimensions(t *testing.T) {
	model := newTestModel()
	model.width = 10
	model.height = 10
	cols, rows := model.GridDimensions()
	if cols != 6 || rows != 3 {
		t.Fatalf("GridDimensions = (%d,%d), want (6,3)", cols, rows)
	}

	model.width = 2
	model.height = 2
	cols, rows = model.GridDimensions()
	if cols != 1 || rows != 1 {
		t.Fatalf("GridDimensions min = (%d,%d), want (1,1)", cols, rows)
	}
}

func TestVisibleSamplesAndScroll(t *testing.T) {
	model := newTestModel()
	model.width = 10
	model.height = 10

	for i := 1; i <= 30; i++ {
		model.samples.Push(ping.Sample{Sequence: i})
	}

	visible := model.VisibleSamples()
	if len(visible) != 18 {
		t.Fatalf("VisibleSamples len=%d, want 18", len(visible))
	}
	if visible[0].Sequence != 13 || visible[len(visible)-1].Sequence != 30 {
		t.Fatalf("VisibleSamples range=%d..%d, want 13..30", visible[0].Sequence, visible[len(visible)-1].Sequence)
	}

	model.scrollPos = 5
	visible = model.VisibleSamples()
	if visible[0].Sequence != 8 || visible[len(visible)-1].Sequence != 25 {
		t.Fatalf("VisibleSamples scroll range=%d..%d, want 8..25", visible[0].Sequence, visible[len(visible)-1].Sequence)
	}

	model.scrollPos = 20
	visible = model.VisibleSamples()
	if visible[0].Sequence != 1 || visible[len(visible)-1].Sequence != 18 {
		t.Fatalf("VisibleSamples clamped range=%d..%d, want 1..18", visible[0].Sequence, visible[len(visible)-1].Sequence)
	}
}

func TestCanScrollUpDown(t *testing.T) {
	model := newTestModel()
	model.width = 10
	model.height = 10

	for i := 1; i <= 30; i++ {
		model.samples.Push(ping.Sample{Sequence: i})
	}

	model.scrollPos = 0
	if !model.CanScrollUp() || model.CanScrollDown() {
		t.Fatalf("expected CanScrollUp=true, CanScrollDown=false at scrollPos=0")
	}

	model.scrollPos = 5
	if !model.CanScrollUp() || !model.CanScrollDown() {
		t.Fatalf("expected CanScrollUp=true, CanScrollDown=true at scrollPos=5")
	}

	model.scrollPos = 12
	if model.CanScrollUp() || !model.CanScrollDown() {
		t.Fatalf("expected CanScrollUp=false, CanScrollDown=true at max scroll")
	}
}

func TestColorizeRTTFormatting(t *testing.T) {
	model := newTestModel()
	out := model.colorizeRTTMs(12.34)
	if !strings.Contains(out, "12.3ms") {
		t.Fatalf("expected formatted RTT in output, got %q", out)
	}

	out = model.colorizeRTT(12*time.Millisecond + 340*time.Microsecond)
	if !strings.Contains(out, "12.3ms") {
		t.Fatalf("expected formatted RTT in output, got %q", out)
	}
}

func TestRenderStatsWaiting(t *testing.T) {
	model := newTestModel()
	model.stats.TotalSamples = 0
	out := model.renderStats()
	if !strings.Contains(out, "Waiting for data") {
		t.Fatalf("expected waiting message, got %q", out)
	}
}

func TestPlaceOverlay(t *testing.T) {
	background := "12345\nabcde"
	overlay := "XX\nYY"
	out := placeOverlay(1, 0, overlay, background)
	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "1XX45" || lines[1] != "aYYde" {
		t.Fatalf("unexpected overlay result: %q / %q", lines[0], lines[1])
	}
}

func TestRenderHelpOverlay(t *testing.T) {
	model := newTestModel()
	model.width = 40
	model.height = 10

	base := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"
	out := model.renderHelpOverlay(base)
	if !strings.Contains(out, "Keyboard Shortcuts") {
		t.Fatalf("expected help overlay in output")
	}
}

func TestRenderStatusBar(t *testing.T) {
	model := newTestModel()
	model.width = 40

	model.statusMsg = "OK"
	out := model.renderStatusBar()
	if !strings.Contains(out, "OK") {
		t.Fatalf("expected status message")
	}

	model.statusMsg = ""
	model.scrollPos = 1
	out = model.renderStatusBar()
	if !strings.Contains(out, "Scroll: 1") {
		t.Fatalf("expected scroll info")
	}
}
