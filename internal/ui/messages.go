package ui

import (
	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/pbv7/pingheat/internal/ping"
)

// SampleMsg is sent when a new ping sample is received.
type SampleMsg struct {
	Sample ping.Sample
}

// MetricsMsg is sent when metrics are updated.
type MetricsMsg struct {
	Stats metrics.Stats
}

// StatusMsg is sent to update the status bar message.
type StatusMsg struct {
	Message string
	IsError bool
}

// TickMsg is sent periodically to trigger UI updates.
type TickMsg struct{}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Err error
}

// WindowSizeMsg is sent when the terminal is resized.
type WindowSizeMsg struct {
	Width  int
	Height int
}
