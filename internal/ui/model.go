package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pbv7/pingheat/internal/buffer"
	"github.com/pbv7/pingheat/internal/config"
	"github.com/pbv7/pingheat/internal/metrics"
	"github.com/pbv7/pingheat/internal/ping"
)

// Model is the Bubble Tea model for the UI.
type Model struct {
	// Configuration
	config config.Config

	// Data
	samples *buffer.RingBuffer[ping.Sample]
	stats   metrics.Stats

	// UI state
	width      int
	height     int
	scrollPos  int
	showHelp   bool
	statusMsg  string
	statusErr  bool
	quitting   bool
	lastUpdate time.Time

	// Channels for receiving data
	sampleChan  <-chan ping.Sample
	metricsChan <-chan metrics.Stats
}

// NewModel creates a new UI model.
func NewModel(cfg config.Config, sampleChan <-chan ping.Sample, metricsChan <-chan metrics.Stats) Model {
	return Model{
		config:      cfg,
		samples:     buffer.NewRingBuffer[ping.Sample](cfg.HistorySize),
		sampleChan:  sampleChan,
		metricsChan: metricsChan,
		showHelp:    cfg.ShowHelp,
		lastUpdate:  time.Now(),
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForSamples(),
		m.listenForMetrics(),
		m.tick(),
	)
}

// listenForSamples returns a command that waits for samples.
func (m Model) listenForSamples() tea.Cmd {
	return func() tea.Msg {
		sample, ok := <-m.sampleChan
		if !ok {
			return nil
		}
		return SampleMsg{Sample: sample}
	}
}

// listenForMetrics returns a command that waits for metrics updates.
func (m Model) listenForMetrics() tea.Cmd {
	return func() tea.Msg {
		stats, ok := <-m.metricsChan
		if !ok {
			return nil
		}
		return MetricsMsg{Stats: stats}
	}
}

// tick returns a command that triggers periodic updates.
func (m Model) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// SetSize sets the terminal size.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GridDimensions returns the heatmap grid dimensions.
func (m Model) GridDimensions() (cols, rows int) {
	// Reserve space for header (1 line), stats (2 lines), status bar (1 line), and borders (2 lines)
	availableHeight := m.height - 7
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Each cell is 1 character wide, reserve 2 for borders
	availableWidth := m.width - 4
	if availableWidth < 1 {
		availableWidth = 1
	}

	return availableWidth, availableHeight
}

// VisibleSamples returns the samples currently visible in the heatmap.
func (m Model) VisibleSamples() []ping.Sample {
	cols, rows := m.GridDimensions()
	visibleCount := cols * rows

	totalSamples := m.samples.Len()
	if totalSamples == 0 {
		return nil
	}

	// Calculate the start index based on scroll position
	maxScroll := totalSamples - visibleCount
	if maxScroll < 0 {
		maxScroll = 0
	}

	startIdx := maxScroll - m.scrollPos
	if startIdx < 0 {
		startIdx = 0
	}

	endIdx := startIdx + visibleCount
	if endIdx > totalSamples {
		endIdx = totalSamples
	}

	return m.samples.GetRange(startIdx, endIdx-1)
}

// CanScrollUp returns true if scrolling up is possible.
func (m Model) CanScrollUp() bool {
	cols, rows := m.GridDimensions()
	visibleCount := cols * rows
	maxScroll := m.samples.Len() - visibleCount
	return m.scrollPos < maxScroll
}

// CanScrollDown returns true if scrolling down is possible.
func (m Model) CanScrollDown() bool {
	return m.scrollPos > 0
}
