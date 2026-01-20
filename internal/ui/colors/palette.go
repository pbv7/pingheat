package colors

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RTT thresholds in milliseconds
const (
	ThresholdExcellent = 30   // 0-30ms: Green
	ThresholdGood      = 80   // 30-80ms: Light Green
	ThresholdFair      = 150  // 80-150ms: Yellow
	ThresholdPoor      = 300  // 150-300ms: Orange
	// >300ms: Red
)

// Colors for different RTT ranges
var (
	ColorExcellent = lipgloss.Color("#00FF00") // Green
	ColorGood      = lipgloss.Color("#7FFF00") // Light Green
	ColorFair      = lipgloss.Color("#FFFF00") // Yellow
	ColorPoor      = lipgloss.Color("#FF8C00") // Orange
	ColorBad       = lipgloss.Color("#FF0000") // Red
	ColorTimeout   = lipgloss.Color("#8B008B") // Dark Magenta - stands out but flows with heatmap
)

// Background colors (dimmer versions)
var (
	BGExcellent = lipgloss.Color("#004400")
	BGGood      = lipgloss.Color("#224400")
	BGFair      = lipgloss.Color("#444400")
	BGPoor      = lipgloss.Color("#442200")
	BGBad       = lipgloss.Color("#440000")
	BGTimeout   = lipgloss.Color("#222222")
)

// Classify returns the color classification for an RTT duration.
func Classify(rtt time.Duration) lipgloss.Color {
	ms := float64(rtt.Microseconds()) / 1000.0
	return ClassifyMs(ms)
}

// ClassifyMs returns the color classification for an RTT in milliseconds.
func ClassifyMs(ms float64) lipgloss.Color {
	switch {
	case ms < 0:
		return ColorTimeout
	case ms <= ThresholdExcellent:
		return ColorExcellent
	case ms <= ThresholdGood:
		return ColorGood
	case ms <= ThresholdFair:
		return ColorFair
	case ms <= ThresholdPoor:
		return ColorPoor
	default:
		return ColorBad
	}
}

// ClassifyBG returns the background color for an RTT duration.
func ClassifyBG(rtt time.Duration) lipgloss.Color {
	ms := float64(rtt.Microseconds()) / 1000.0
	return ClassifyBGMs(ms)
}

// ClassifyBGMs returns the background color for an RTT in milliseconds.
func ClassifyBGMs(ms float64) lipgloss.Color {
	switch {
	case ms < 0:
		return BGTimeout
	case ms <= ThresholdExcellent:
		return BGExcellent
	case ms <= ThresholdGood:
		return BGGood
	case ms <= ThresholdFair:
		return BGFair
	case ms <= ThresholdPoor:
		return BGPoor
	default:
		return BGBad
	}
}

// HeatmapChar returns a character representing the RTT level.
// Uses filled block (█) for all states to maintain visual flow.
func HeatmapChar(timeout bool) string {
	return "█"
}

// ForTimeout returns true if the value represents a timeout.
func ForTimeout(ms float64) bool {
	return ms < 0
}
