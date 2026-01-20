package ui

import "github.com/charmbracelet/lipgloss"

// Style definitions for the UI
var (
	// Header styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5F5FD7")).
			Padding(0, 1)

	TargetStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))

	// Stats styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	GoodValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	WarnValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	BadValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Background(lipgloss.Color("#1A1A1A")).
				Padding(0, 1)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5F5FD7")).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	// Heatmap border
	HeatmapBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444"))

	// Help overlay
	HelpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#5F5FD7")).
				Padding(1, 2).
				Background(lipgloss.Color("#1A1A1A"))
)
