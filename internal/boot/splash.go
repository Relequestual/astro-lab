// Package boot provides the CLI boot-splash animation.
package boot

import (
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const (
	frameCount    = 10
	frameInterval = 280 * time.Millisecond
)

// Lipgloss styles used across frames.
var (
	hexSt    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF"))           // bright cyan – hex border
	traceSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AFAF"))           // teal – circuit traces & nodes
	starBrt  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true) // bold white – central star
	starDim  = lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB"))           // sky blue – dim/emerging star
	sparkSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AFAF"))           // teal – outer sparkles
)

type tickMsg struct{}

// Model is the Bubble Tea model for the splash animation.
type Model struct {
	frame int
}

// Init starts the first animation tick.
func (m Model) Init() tea.Cmd {
	return tea.Tick(frameInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// Update advances the animation one frame per tick, or quits on any key.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m.frame++
		if m.frame >= frameCount {
			return m, tea.Quit
		}
		return m, tea.Tick(frameInterval, func(time.Time) tea.Msg { return tickMsg{} })
	}
	return m, nil
}

// View renders the current animation frame.
func (m Model) View() string {
	return buildFrame(m.frame)
}

// buildFrame composes the styled ASCII art for frame f.
//
// The hexagon uses this geometry (40 cols wide, 11 rows):
//
//	row 0 : blank / sparkles
//	row 1 :           ___________
//	row 2 :          /           \
//	row 3 :         /             \
//	row 4 :        |               |
//	row 5 :        |  [trace row]  |      ← centre row, col 15 is the star
//	row 6 :        |               |
//	row 7 :         \             /
//	row 8 :          \           /
//	row 9 :           \___________/
//	row 10: blank / sparkles
//
// Hex wall positions: left | at col 7, right | at col 23. Centre col = 15.
func buildFrame(f int) string {
	H := hexSt.Render
	T := traceSt.Render
	S := starBrt.Render
	D := starDim.Render
	K := sparkSt.Render

	empty := strings.Repeat(" ", 40)

	// ── plain hex rows ────────────────────────────────────────────────────────
	hexTop    := "          " + H("___________") + "                   " // 10+11+19 = 40
	diagA1    := "         " + H("/") + "           " + H("\\") + "                  " // 9+1+11+1+18
	diagA2    := "        " + H("/") + "             " + H("\\") + "                 " // 8+1+13+1+17
	wallPlain := "       " + H("|") + "               " + H("|") + "                " // 7+1+15+1+16
	diagB1    := "        " + H("\\") + "             " + H("/") + "                 " // 8+1+13+1+17
	diagB2    := "         " + H("\\") + "           " + H("/") + "                  " // 9+1+11+1+18
	hexBot    := "          " + H("\\") + H("___________") + H("/") + "                 " // 10+1+11+1+17

	// ── hex rows with vertical centre trace ───────────────────────────────────
	// Row 2 – /…·…\  : interior 11 wide (cols 10-20), centre node · at col 15
	diagA1v := "         " + H("/") + "     " + T("·") + "     " + H("\\") + "                  " // 9+1+5+1+5+1+18
	// Row 3 – /…│…\  : interior 13 wide (cols 9-21), │ at col 15
	diagA2v := "        " + H("/") + "      " + T("│") + "      " + H("\\") + "                 " // 8+1+6+1+6+1+17
	// Rows 4,6 – |…│…|  : interior 15 wide (cols 8-22), │ at col 15
	wallVert := "       " + H("|") + "       " + T("│") + "       " + H("|") + "                " // 7+1+7+1+7+1+16
	// Row 7 – \…│…/
	diagB1v := "        " + H("\\") + "      " + T("│") + "      " + H("/") + "                 " // 8+1+6+1+6+1+17
	// Row 8 – \…·…/
	diagB2v := "         " + H("\\") + "     " + T("·") + "     " + H("/") + "                  " // 9+1+5+1+5+1+18

	// ── centre (trace) rows – 7+1+7+1+7+1+7+9 = 40 ───────────────────────────
	// Traces only, crossroads node at centre
	ctrTrace := T("── · ──") + H("|") + T("── · ──·── · ──") + H("|") + T("── · ──") + "         "
	// Dim star (emerging)
	ctrDim := T("── · ──") + H("|") + T("── · ──") + D("✦") + T("── · ──") + H("|") + T("── · ──") + "         "
	// Bright star (final)
	ctrStar := T("── · ──") + H("|") + T("── · ──") + S("★") + T("── · ──") + H("|") + T("── · ──") + "         "

	// ── sparkle decoration rows ───────────────────────────────────────────────
	// Two ✦ positioned wide of the hex  2+1+26+1+10 = 40
	sparkRow := "  " + K("✦") + "                          " + K("✦") + "          "

	// ── assemble 11 rows ──────────────────────────────────────────────────────
	var rows [11]string
	switch f {
	case 0: // hex top cap only
		rows = [11]string{empty, hexTop, diagA1, empty, empty, empty, empty, empty, empty, empty, empty}
	case 1: // full plain hex outline
		rows = [11]string{empty, hexTop, diagA1, diagA2, wallPlain, wallPlain, wallPlain, diagB1, diagB2, hexBot, empty}
	case 2: // vertical trace through the centre
		rows = [11]string{empty, hexTop, diagA1v, diagA2v, wallVert, wallVert, wallVert, diagB1v, diagB2v, hexBot, empty}
	case 3: // horizontal + vertical traces, crossroads node
		rows = [11]string{empty, hexTop, diagA1v, diagA2v, wallVert, ctrTrace, wallVert, diagB1v, diagB2v, hexBot, empty}
	case 4: // dim star emerges
		rows = [11]string{empty, hexTop, diagA1v, diagA2v, wallVert, ctrDim, wallVert, diagB1v, diagB2v, hexBot, empty}
	case 5: // bright star
		rows = [11]string{empty, hexTop, diagA1v, diagA2v, wallVert, ctrStar, wallVert, diagB1v, diagB2v, hexBot, empty}
	case 6: // sparkles appear
		rows = [11]string{sparkRow, hexTop, diagA1v, diagA2v, wallVert, ctrStar, wallVert, diagB1v, diagB2v, hexBot, sparkRow}
	case 7: // sparkles dim (twinkle off)
		rows = [11]string{empty, hexTop, diagA1v, diagA2v, wallVert, ctrStar, wallVert, diagB1v, diagB2v, hexBot, empty}
	case 8: // sparkles back on
		rows = [11]string{sparkRow, hexTop, diagA1v, diagA2v, wallVert, ctrStar, wallVert, diagB1v, diagB2v, hexBot, sparkRow}
	case 9: // hold final frame
		rows = [11]string{sparkRow, hexTop, diagA1v, diagA2v, wallVert, ctrStar, wallVert, diagB1v, diagB2v, hexBot, sparkRow}
	default:
		return empty
	}
	return strings.Join(rows[:], "\n")
}

// Run plays the boot splash animation.
// It is a no-op when stdout is not an interactive terminal.
func Run() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}
	p := tea.NewProgram(Model{})
	_, _ = p.Run()
}
