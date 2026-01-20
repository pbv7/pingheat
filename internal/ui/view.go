package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/pbv7/pingheat/internal/ui/colors"
)

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Stats line
	b.WriteString(m.renderStats())
	b.WriteString("\n")

	// Heatmap
	b.WriteString(m.renderHeatmap())

	// Status bar
	b.WriteString(m.renderStatusBar())

	// Help overlay (rendered on top if shown)
	if m.showHelp {
		return m.renderHelpOverlay(b.String())
	}

	return b.String()
}

// renderHeader renders the title bar.
func (m Model) renderHeader() string {
	title := TitleStyle.Render("pingheat")
	target := TargetStyle.Render(m.config.Target)
	return fmt.Sprintf("%s %s", title, target)
}

// renderStats renders the statistics lines.
func (m Model) renderStats() string {
	if m.stats.TotalSamples == 0 {
		return LabelStyle.Render("Waiting for data...")
	}

	// First line: basic stats
	line1 := []string{
		fmt.Sprintf("%s %s",
			LabelStyle.Render("Sent:"),
			ValueStyle.Render(fmt.Sprintf("%d", m.stats.TotalSamples))),
	}

	// Loss percentage with color coding
	lossStyle := GoodValueStyle
	if m.stats.LossPercent > 0 {
		lossStyle = WarnValueStyle
	}
	if m.stats.LossPercent > 5 {
		lossStyle = BadValueStyle
	}
	line1 = append(line1, fmt.Sprintf("%s %s",
		LabelStyle.Render("Loss:"),
		lossStyle.Render(fmt.Sprintf("%.1f%%", m.stats.LossPercent))))

	// RTT stats (only if we have successful pings)
	if m.stats.TotalSamples > m.stats.TotalTimeouts {
		line1 = append(line1,
			fmt.Sprintf("%s %s",
				LabelStyle.Render("Min:"),
				m.colorizeRTT(m.stats.MinRTT)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("Avg:"),
				m.colorizeRTT(m.stats.AvgRTT)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("Max:"),
				m.colorizeRTT(m.stats.MaxRTT)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("σ:"),
				m.colorizeRTT(m.stats.StdDev)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("Jitter:"),
				m.colorizeRTT(m.stats.Jitter)),
		)
	}

	// Second line: percentiles and instability
	var line2 []string

	if m.stats.TotalSuccess > 0 {
		// Percentiles
		line2 = append(line2,
			fmt.Sprintf("%s %s",
				LabelStyle.Render("p50:"),
				m.colorizeRTTMs(m.stats.Percentiles.P50)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("p90:"),
				m.colorizeRTTMs(m.stats.Percentiles.P90)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("p95:"),
				m.colorizeRTTMs(m.stats.Percentiles.P95)),
			fmt.Sprintf("%s %s",
				LabelStyle.Render("p99:"),
				m.colorizeRTTMs(m.stats.Percentiles.P99)),
		)
	}

	// Instability patterns
	if m.stats.LossBursts > 0 {
		line2 = append(line2, fmt.Sprintf("%s %s",
			LabelStyle.Render("Outages:"),
			BadValueStyle.Render(fmt.Sprintf("%d", m.stats.LossBursts))))
	}

	if m.stats.LongestTimeout > 0 {
		line2 = append(line2, fmt.Sprintf("%s %s",
			LabelStyle.Render("MaxDrop:"),
			BadValueStyle.Render(fmt.Sprintf("%d", m.stats.LongestTimeout))))
	}

	if m.stats.BrownoutBursts > 0 {
		line2 = append(line2, fmt.Sprintf("%s %s",
			LabelStyle.Render("Brownouts:"),
			WarnValueStyle.Render(fmt.Sprintf("%d", m.stats.BrownoutBursts))))
	}

	// Current streak indicator
	if m.stats.CurrentStreak < -1 {
		line2 = append(line2, fmt.Sprintf("%s %s",
			LabelStyle.Render("Streak:"),
			BadValueStyle.Render(fmt.Sprintf("-%d timeout", -m.stats.CurrentStreak))))
	} else if m.stats.InBrownout {
		line2 = append(line2, fmt.Sprintf("%s %s",
			LabelStyle.Render("Status:"),
			WarnValueStyle.Render("BROWNOUT")))
	}

	result := strings.Join(line1, "  ")
	if len(line2) > 0 {
		result += "\n" + strings.Join(line2, "  ")
	}
	return result
}

// colorizeRTTMs returns a styled RTT string from milliseconds value.
func (m Model) colorizeRTTMs(ms float64) string {
	color := colors.ClassifyMs(ms)
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(fmt.Sprintf("%.1fms", ms))
}

// colorizeRTT returns a styled RTT string.
func (m Model) colorizeRTT(d time.Duration) string {
	ms := float64(d.Microseconds()) / 1000.0
	color := colors.ClassifyMs(ms)
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(fmt.Sprintf("%.1fms", ms))
}

// renderHeatmap renders the main heatmap grid.
func (m Model) renderHeatmap() string {
	cols, rows := m.GridDimensions()
	if cols <= 0 || rows <= 0 {
		return ""
	}

	samples := m.VisibleSamples()
	sampleIdx := 0

	var grid strings.Builder

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			if sampleIdx < len(samples) {
				sample := samples[sampleIdx]
				char := colors.HeatmapChar(sample.Timeout)

				var color lipgloss.Color
				if sample.Timeout {
					color = colors.ColorTimeout
				} else {
					color = colors.Classify(sample.RTT)
				}

				style := lipgloss.NewStyle().Foreground(color)
				grid.WriteString(style.Render(char))
				sampleIdx++
			} else {
				// Empty cell
				grid.WriteString(" ")
			}
		}
		if row < rows-1 {
			grid.WriteString("\n")
		}
	}

	// Apply border
	return HeatmapBorderStyle.Render(grid.String()) + "\n"
}

// renderStatusBar renders the status bar at the bottom.
func (m Model) renderStatusBar() string {
	// Left side: status message or scroll info
	var left string
	if m.statusMsg != "" {
		if m.statusErr {
			left = StatusErrorStyle.Render(m.statusMsg)
		} else {
			left = StatusBarStyle.Render(m.statusMsg)
		}
	} else {
		scrollInfo := ""
		if m.CanScrollUp() || m.CanScrollDown() {
			scrollInfo = fmt.Sprintf("Scroll: %d", m.scrollPos)
		}
		left = StatusBarStyle.Render(scrollInfo)
	}

	// Right side: help hint
	right := StatusBarStyle.Render("Press ? for help")

	// Calculate padding
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	padding := m.width - leftLen - rightLen
	if padding < 1 {
		padding = 1
	}

	return left + strings.Repeat(" ", padding) + right
}

// renderHelpOverlay renders the help overlay on top of the main view.
func (m Model) renderHelpOverlay(base string) string {
	help := m.renderHelp()

	// Center the help overlay
	helpWidth := lipgloss.Width(help)
	helpHeight := lipgloss.Height(help)

	x := (m.width - helpWidth) / 2
	y := (m.height - helpHeight) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Overlay the help on top of the base view
	return placeOverlay(x, y, help, base)
}

// renderHelp renders the help content.
func (m Model) renderHelp() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Scroll up (older)"},
		{"↓/j", "Scroll down (newer)"},
		{"PgUp", "Page up"},
		{"PgDn", "Page down"},
		{"Home/g", "Go to oldest"},
		{"End/G", "Go to newest"},
		{"c", "Clear history"},
		{"?/h", "Toggle help"},
		{"q", "Quit"},
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	for _, k := range keys {
		b.WriteString(HelpKeyStyle.Render(fmt.Sprintf("%8s", k.key)))
		b.WriteString("  ")
		b.WriteString(HelpDescStyle.Render(k.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(LabelStyle.Render("Legend: "))
	b.WriteString(lipgloss.NewStyle().Foreground(colors.ColorExcellent).Render("█"))
	b.WriteString(" <30ms ")
	b.WriteString(lipgloss.NewStyle().Foreground(colors.ColorGood).Render("█"))
	b.WriteString(" <80ms ")
	b.WriteString(lipgloss.NewStyle().Foreground(colors.ColorFair).Render("█"))
	b.WriteString(" <150ms ")
	b.WriteString(lipgloss.NewStyle().Foreground(colors.ColorPoor).Render("█"))
	b.WriteString(" <300ms ")
	b.WriteString(lipgloss.NewStyle().Foreground(colors.ColorBad).Render("█"))
	b.WriteString(" >300ms ")
	b.WriteString(lipgloss.NewStyle().Foreground(colors.ColorTimeout).Render("█"))
	b.WriteString(" timeout")

	return HelpOverlayStyle.Render(b.String())
}

// placeOverlay places an overlay string on top of a background string.
func placeOverlay(x, y int, overlay, background string) string {
	bgLines := strings.Split(background, "\n")
	ovLines := strings.Split(overlay, "\n")

	for i, ovLine := range ovLines {
		bgIdx := y + i
		if bgIdx < 0 || bgIdx >= len(bgLines) {
			continue
		}

		bgLine := bgLines[bgIdx]
		bgRunes := []rune(bgLine)
		ovRunes := []rune(ovLine)

		// Expand background line if needed
		for len(bgRunes) < x+len(ovRunes) {
			bgRunes = append(bgRunes, ' ')
		}

		// Overlay the text
		for j, r := range ovRunes {
			if x+j >= 0 && x+j < len(bgRunes) {
				bgRunes[x+j] = r
			}
		}

		bgLines[bgIdx] = string(bgRunes)
	}

	return strings.Join(bgLines, "\n")
}
