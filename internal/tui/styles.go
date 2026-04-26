package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("99")  // purple
	colorSecondary = lipgloss.Color("39")  // blue
	colorMuted     = lipgloss.Color("241") // grey
	colorSuccess   = lipgloss.Color("78")  // green
	colorWarning   = lipgloss.Color("220") // yellow
	colorDanger    = lipgloss.Color("196") // red
	colorHighlight = lipgloss.Color("212") // pink

	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	footerKeyStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(colorHighlight).
			Bold(true)

	footerDescStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).Bold(true)

	focusedStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).Bold(true)

	mutedStyle = lipgloss.NewStyle().Foreground(colorMuted)

	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).Bold(true).
			Padding(0, 0, 1, 0)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("238")).
			Padding(0, 3).MarginRight(1)

	activeButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Padding(0, 3).MarginRight(1).Bold(true)

	successStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	warningStyle = lipgloss.NewStyle().Foreground(colorWarning)
	errorStyle   = lipgloss.NewStyle().Foreground(colorDanger)

	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).Bold(true).Width(12)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).Bold(true).
				MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).Bold(true).Width(16)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorWarning).
			Padding(1, 2).Width(50)

	previewAddStyle    = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	previewRemoveStyle = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
)
