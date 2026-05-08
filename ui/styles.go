package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen   = lipgloss.Color("#50FA7B")
	colorYellow  = lipgloss.Color("#F1FA8C")
	colorRed     = lipgloss.Color("#FF5555")
	colorMagenta = lipgloss.Color("#FF79C6")
	colorGray    = lipgloss.Color("#6272A4")
	colorBright  = lipgloss.Color("#F8F8F2")
	colorAccent  = lipgloss.Color("#BD93F9")
	colorBgAlt   = lipgloss.Color("#44475A")

	styleTitle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleGray  = lipgloss.NewStyle().Foreground(colorGray)
	styleBold  = lipgloss.NewStyle().Foreground(colorBright).Bold(true)

	styleHeaderBar = lipgloss.NewStyle().
			Foreground(colorBright).
			Bold(true).
			Padding(0, 1)

	styleFilterLabel = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleSuccess     = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleErrorMsg    = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	styleExportBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 3)

	styleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorGray).
				BorderBottom(true)

	styleTableCell     = lipgloss.NewStyle().Foreground(colorBright)
	styleTableSelected = lipgloss.NewStyle().Background(colorBgAlt).Foreground(colorBright).Bold(true)

	styleStatus2xx = lipgloss.NewStyle().Foreground(colorGreen)
	styleStatus3xx = lipgloss.NewStyle().Foreground(colorYellow)
	styleStatus4xx = lipgloss.NewStyle().Foreground(colorRed)
	styleStatus5xx = lipgloss.NewStyle().Foreground(colorMagenta)
)

func statusStyle(code int, errStr string) lipgloss.Style {
	if errStr != "" {
		return styleStatus4xx
	}
	switch {
	case code >= 500:
		return styleStatus5xx
	case code >= 400:
		return styleStatus4xx
	case code >= 300:
		return styleStatus3xx
	case code >= 200:
		return styleStatus2xx
	}
	return styleGray
}
