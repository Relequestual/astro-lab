package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	message  string
	quitting bool
}

func NewModel() model {
	return model{
		message: "Welcome to Astrometrics Lab! Press q to quit.",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}
	return fmt.Sprintf("\n  %s\n\n", m.message)
}

// Run starts the Bubble Tea application
func Run() error {
	p := tea.NewProgram(NewModel())
	_, err := p.Run()
	return err
}
