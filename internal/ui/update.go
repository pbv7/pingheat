package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case SampleMsg:
		m.samples.Push(msg.Sample)
		m.lastUpdate = time.Now()
		return m, m.listenForSamples()

	case MetricsMsg:
		m.stats = msg.Stats
		return m, m.listenForMetrics()

	case StatusMsg:
		m.statusMsg = msg.Message
		m.statusErr = msg.IsError
		return m, nil

	case TickMsg:
		return m, m.tick()

	case ErrorMsg:
		m.statusMsg = msg.Err.Error()
		m.statusErr = true
		return m, nil
	}

	return m, nil
}

// handleKeypress processes keyboard input.
func (m Model) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "?", "h":
		m.showHelp = !m.showHelp
		return m, nil

	case "c":
		// Clear samples and reset scroll
		m.samples.Clear()
		m.scrollPos = 0
		m.statusMsg = "Cleared"
		m.statusErr = false
		return m, nil

	case "up", "k":
		if m.CanScrollUp() {
			m.scrollPos++
		}
		return m, nil

	case "down", "j":
		if m.CanScrollDown() {
			m.scrollPos--
		}
		return m, nil

	case "pgup":
		_, rows := m.GridDimensions()
		for i := 0; i < rows && m.CanScrollUp(); i++ {
			m.scrollPos++
		}
		return m, nil

	case "pgdown":
		_, rows := m.GridDimensions()
		for i := 0; i < rows && m.CanScrollDown(); i++ {
			m.scrollPos--
		}
		return m, nil

	case "home", "g":
		// Scroll to oldest
		cols, rows := m.GridDimensions()
		visibleCount := cols * rows
		maxScroll := m.samples.Len() - visibleCount
		if maxScroll > 0 {
			m.scrollPos = maxScroll
		}
		return m, nil

	case "end", "G":
		// Scroll to newest
		m.scrollPos = 0
		return m, nil

	case "esc":
		if m.showHelp {
			m.showHelp = false
		}
		return m, nil
	}

	return m, nil
}
