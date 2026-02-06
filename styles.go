package cssgen

import "github.com/charmbracelet/lipgloss"

// Terminal styles for consistent output formatting across reporters.
// Lipgloss automatically degrades colors based on terminal capabilities.
var (
	// StyleCyan is used for file locations, section headers, and statistics headers.
	StyleCyan = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	// StyleRed is used for error sections and build failure messages.
	StyleRed = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	// StyleYellow is used for warning sections and caret indicators.
	StyleYellow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	// StyleGreen is used for quick wins, recommendations, and success messages.
	StyleGreen = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	// StyleGray is used for linter names and hints.
	StyleGray = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// RenderStyle applies a lipgloss style to text when colors are enabled.
// When useColors is false, the text is returned unmodified.
func RenderStyle(style lipgloss.Style, text string, useColors bool) string {
	if !useColors {
		return text
	}
	return style.Render(text)
}
