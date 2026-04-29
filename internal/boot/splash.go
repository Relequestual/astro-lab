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
	frameCount      = 10
	frameInterval   = 280 * time.Millisecond
	// frameWidth is the visual width of every rendered row before centering.
	// 65 is odd so the hex is perfectly symmetric: 6 leading = 6 trailing spaces
	// on wall rows, and the centre trace lands on col 32 = (65-1)/2 exactly.
	frameWidth      = 65
	// defaultTermWidth is used when terminal width cannot be detected.
	defaultTermWidth = 80
)

// Lipgloss styles used across frames.
var (
	hexSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF"))            // bright cyan – hex border
	traceSt = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AFAF"))            // teal – circuit traces & nodes
	starBrt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true) // bold white – central star
	starDim = lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB"))            // sky blue – dim/emerging star
	sparkSt = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AFAF"))            // teal – outer sparkles
	titleSt = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Bold(true) // bold cyan – hex title
)

type tickMsg struct{}

// Model is the Bubble Tea model for the splash animation.
type Model struct {
	frame     int
	termWidth int
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
	return buildFrame(m.frame, m.termWidth)
}

// buildFrame composes the styled ASCII art for frame f, centering it within
// termWidth columns.
//
// Hexagon geometry (frameWidth=65 visual cols × 21 rows before centering):
//
//	row 0  : blank / sparkles
//	row 1  :            /─────────────────────────────────────────\
//	row 2  :           /                                           \
//	row 3  :          /                                             \
//	row 4  :         /                                               \
//	row 5  :        /                                                 \
//	row 6  :       /                                                   \
//	row 7  :       |     _   ___ _____ ___  ___    _      _   ___  |  ← art row 0
//	row 8  :       |    /_\ / __|_   _| _ \/ _ \  | |    /_\ | _ )|  ← art row 1
//	row 9  :       |   / _ \\__ \ | | |   / (_) | | |__ / _ \| _ \|  ← art row 2
//	row 10 :       |  /_/ \_\___/ |_| |_|_\\___/  |____/_/ \_\___/|  ← art row 3
//	row 11 :       |  ──── · ──── · ──── · ────★──── · ──── · ──── · ──  |  ← star
//	row 12 :       |            ·            │            ·            |
//	row 13 :       |                      [wall]                       |
//	row 14 :       \                                                   /
//	row 15 :        \                                                 /
//	row 16 :         \                                               /
//	row 17 :          \                                             /
//	row 18 :           \                                           /
//	row 19 :            \─────────────────────────────────────────/
//	row 20 : blank / sparkles
//
// Left wall | at col 6, right wall | at col 58. Centre col = 32.
// Frame is 65 cols wide (odd): every wall row has 6 leading = 6 trailing spaces,
// so the hex is perfectly symmetric and the centre trace is at the exact mid-col.
func buildFrame(f, termWidth int) string {
	H := hexSt.Render
	T := traceSt.Render
	S := starBrt.Render
	D := starDim.Render
	K := sparkSt.Render
	L := titleSt.Render

	empty := strings.Repeat(" ", frameWidth)

	// ── plain hex rows (each row is exactly frameWidth=65 visual cols) ─────────
	//
	// Top bar: / at col 11, 41 underscores, \ at col 53 → 11+(1+41+1)+11 = 65
	hexTopBar := "           " + H("/"+strings.Repeat("_", 41)+"\\") + "           "
	// Five diagonal rows (top): / moves left one col per row, \ moves right.
	diagA1 := "          " + H("/") + strings.Repeat(" ", 43) + H("\\") + "          " // 10+1+43+1+10=65
	diagA2 := "         " + H("/") + strings.Repeat(" ", 45) + H("\\") + "         "  // 9+1+45+1+9=65
	diagA3 := "        " + H("/") + strings.Repeat(" ", 47) + H("\\") + "        "   // 8+1+47+1+8=65
	diagA4 := "       " + H("/") + strings.Repeat(" ", 49) + H("\\") + "       "    // 7+1+49+1+7=65
	diagA5 := "      " + H("/") + strings.Repeat(" ", 51) + H("\\") + "      "     // 6+1+51+1+6=65
	// Wall rows: | at col 6 and col 58, interior 51 chars.
	wallPlain := "      " + H("|") + strings.Repeat(" ", 51) + H("|") + "      " // 6+1+51+1+6=65
	// Five diagonal rows (bottom): \ moves right per row, / moves left.
	diagB5 := "      " + H("\\") + strings.Repeat(" ", 51) + H("/") + "      "     // 6+1+51+1+6=65
	diagB4 := "       " + H("\\") + strings.Repeat(" ", 49) + H("/") + "       "    // 7+1+49+1+7=65
	diagB3 := "        " + H("\\") + strings.Repeat(" ", 47) + H("/") + "        "   // 8+1+47+1+8=65
	diagB2 := "         " + H("\\") + strings.Repeat(" ", 45) + H("/") + "         "  // 9+1+45+1+9=65
	diagB1 := "          " + H("\\") + strings.Repeat(" ", 43) + H("/") + "          " // 10+1+43+1+10=65
	// Bottom bar: \ at col 11, 41 underscores, / at col 53 → 11+(1+41+1)+11 = 65
	hexBotBar := "           " + H("\\"+strings.Repeat("_", 41)+"/") + "           "

	// ── diagonal rows with centre vertical trace (· or │ at col 32) ───────────
	// col 32 is the exact centre of the 65-col frame; spaces adjusted per row.
	// diagA1v: / at 10, · at 32, \ at 54 → 10+1+21+1+21+1+10=65
	diagA1v := "          " + H("/") + strings.Repeat(" ", 21) + T("·") + strings.Repeat(" ", 21) + H("\\") + "          "
	// diagA2v: / at 9, │ at 32, \ at 55 → 9+1+22+1+22+1+9=65
	diagA2v := "         " + H("/") + strings.Repeat(" ", 22) + T("│") + strings.Repeat(" ", 22) + H("\\") + "         "
	// diagA3v: / at 8, │ at 32, \ at 56 → 8+1+23+1+23+1+8=65
	diagA3v := "        " + H("/") + strings.Repeat(" ", 23) + T("│") + strings.Repeat(" ", 23) + H("\\") + "        "
	// diagA4v: / at 7, │ at 32, \ at 57 → 7+1+24+1+24+1+7=65
	diagA4v := "       " + H("/") + strings.Repeat(" ", 24) + T("│") + strings.Repeat(" ", 24) + H("\\") + "       "
	// diagA5v: / at 6, │ at 32, \ at 58 → 6+1+25+1+25+1+6=65
	diagA5v := "      " + H("/") + strings.Repeat(" ", 25) + T("│") + strings.Repeat(" ", 25) + H("\\") + "      "
	// wallVert: | at 6, │ at 32, | at 58 → 6+1+25+1+25+1+6=65
	wallVert := "      " + H("|") + strings.Repeat(" ", 25) + T("│") + strings.Repeat(" ", 25) + H("|") + "      "
	// Bottom diagonal trace rows (symmetric to top).
	diagB5v := "      " + H("\\") + strings.Repeat(" ", 25) + T("│") + strings.Repeat(" ", 25) + H("/") + "      "
	diagB4v := "       " + H("\\") + strings.Repeat(" ", 24) + T("│") + strings.Repeat(" ", 24) + H("/") + "       "
	diagB3v := "        " + H("\\") + strings.Repeat(" ", 23) + T("│") + strings.Repeat(" ", 23) + H("/") + "        "
	diagB2v := "         " + H("\\") + strings.Repeat(" ", 22) + T("│") + strings.Repeat(" ", 22) + H("/") + "         "
	// diagB1v: \ at 10, · at 32, / at 54 → 10+1+21+1+21+1+10=65
	diagB1v := "          " + H("\\") + strings.Repeat(" ", 21) + T("·") + strings.Repeat(" ", 21) + H("/") + "          "

	// ── secondary circuit trace rows (nodes at cols 19 & 45, junction at 32) ──
	// Interior layout: 12 sp | node(1) | 12 trace | junction(1) | 12 trace | node(1) | 12 sp = 51
	//
	// wallSec2: node pillars + centre │ → 6+1+12+1+12+1+12+1+12+1+6=65
	wallSec2 := "      " + H("|") +
		strings.Repeat(" ", 12) + T("·") +
		strings.Repeat(" ", 12) + T("│") +
		strings.Repeat(" ", 12) + T("·") +
		strings.Repeat(" ", 12) +
		H("|") + "      "

	// ── centre (star/trace) rows – interior 51 chars, star/node at col 32 ─────
	// Each arm: ──── · ──── · ──── · ────  (4+3+4+3+4+3+4 = 25 visual chars)
	ctrArm := strings.Repeat("─", 4) + " · " +
		strings.Repeat("─", 4) + " · " +
		strings.Repeat("─", 4) + " · " +
		strings.Repeat("─", 4)
	// Traces only, crossroads ┼ at centre → 6+1+(25+1+25)+1+6=65
	ctrTrace := "      " + H("|") + T(ctrArm+"┼"+ctrArm) + H("|") + "      "
	// Dim star emerging → 6+1+25+1+25+1+6=65
	ctrDim := "      " + H("|") + T(ctrArm) + D("✦") + T(ctrArm) + H("|") + "      "
	// Bright star → 6+1+25+1+25+1+6=65
	ctrStar := "      " + H("|") + T(ctrArm) + S("★") + T(ctrArm) + H("|") + "      "

	// ── "ASTRO LAB" ASCII art title (figlet "small" font, 4 rows) ───────────────
	// Each interior is exactly 51 chars: 4 leading + 43-or-44 body + 4-or-3 trailing.
	wallArt0 := "      " + H("|") + L("    "+"   _   ___ _____ ___  ___    _      _   ___"+"    ") + H("|") + "      "
	wallArt1 := "      " + H("|") + L("    "+"  /_\\ / __|_   _| _ \\/ _ \\  | |    /_\\ | _ )"+"   ") + H("|") + "      "
	wallArt2 := "      " + H("|") + L("    "+" / _ \\\\__ \\ | | |   / (_) | | |__ / _ \\| _ \\"+"   ") + H("|") + "      "
	wallArt3 := "      " + H("|") + L("    "+"/_/ \\_\\___/ |_| |_|_\\\\___/  |____/_/ \\_\\___/"+"   ") + H("|") + "      "

	// ── sparkle decoration rows ────────────────────────────────────────────────
	// Two ✦ at cols 2 and 62 → 2+1+59+1+2=65
	sparkRow := "  " + K("✦") + strings.Repeat(" ", 59) + K("✦") + "  "

	// ── assemble 21 rows ───────────────────────────────────────────────────────
	var rows [21]string
	switch f {
	case 0: // top bar cap only
		rows = [21]string{
			empty, hexTopBar,
			empty, empty, empty, empty, empty,
			empty, empty, empty, empty, empty, empty, empty,
			empty, empty, empty, empty, empty,
			empty, empty,
		}
	case 1: // full plain hex outline
		rows = [21]string{
			empty, hexTopBar,
			diagA1, diagA2, diagA3, diagA4, diagA5,
			wallArt0, wallArt1, wallArt2, wallArt3, wallPlain, wallPlain, wallPlain,
			diagB5, diagB4, diagB3, diagB2, diagB1,
			hexBotBar, empty,
		}
	case 2: // centre vertical trace
		rows = [21]string{
			empty, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, wallVert, wallVert, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, empty,
		}
	case 3: // secondary traces + horizontal crossroads
		rows = [21]string{
			empty, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrTrace, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, empty,
		}
	case 4: // dim star emerges
		rows = [21]string{
			empty, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrDim, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, empty,
		}
	case 5: // bright star
		rows = [21]string{
			empty, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrStar, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, empty,
		}
	case 6: // sparkles appear
		rows = [21]string{
			sparkRow, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrStar, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, sparkRow,
		}
	case 7: // sparkles dim (twinkle off)
		rows = [21]string{
			empty, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrStar, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, empty,
		}
	case 8: // sparkles back on
		rows = [21]string{
			sparkRow, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrStar, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, sparkRow,
		}
	case 9: // hold final frame
		rows = [21]string{
			sparkRow, hexTopBar,
			diagA1v, diagA2v, diagA3v, diagA4v, diagA5v,
			wallArt0, wallArt1, wallArt2, wallArt3, ctrStar, wallSec2, wallVert,
			diagB5v, diagB4v, diagB3v, diagB2v, diagB1v,
			hexBotBar, sparkRow,
		}
	default:
		return empty
	}

	// ── apply terminal centering ───────────────────────────────────────────────
	indent := 0
	if termWidth > frameWidth {
		indent = (termWidth - frameWidth) / 2
	}
	lines := rows[:]
	if indent > 0 {
		prefix := strings.Repeat(" ", indent)
		for i, row := range lines {
			lines[i] = prefix + row
		}
	}
	return strings.Join(lines, "\n")
}

// Run plays the boot splash animation.
// It is a no-op when stdout is not an interactive terminal.
func Run() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		w = defaultTermWidth
	}
	p := tea.NewProgram(Model{termWidth: w})
	_, _ = p.Run()
}
