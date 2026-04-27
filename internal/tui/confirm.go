package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// confirmModel presents a yes/no confirmation dialog.
type confirmModel struct {
	message       string
	confirmed     bool
	resolved      bool
	width, height int
}

func newConfirmModel(message string) confirmModel {
	return confirmModel{message: message}
}

func (c confirmModel) Init() tea.Cmd {
	return nil
}

func (c confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if c.resolved {
		return c, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y", "Y", "enter":
			c.confirmed = true
			c.resolved = true
		case "n", "N", "esc":
			c.confirmed = false
			c.resolved = true
		}
	}
	return c, nil
}

func (c confirmModel) View() string {
	prompt := c.message + "\n\n" +
		mutedStyle.Render("[y]es  [n]o")
	w, h := c.width, c.height
	if w < 50 {
		w = 50
	}
	if h < 5 {
		h = 5
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(prompt))
}
