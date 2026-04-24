package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type progressUpdate struct {
	text string
}

type progressDone struct {
	err error
}

type progressModel struct {
	spinner spinner.Model
	text    string
	done    bool
}

func newProgressModel(title string) progressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return progressModel{
		spinner: s,
		text:    title,
	}
}

func (m progressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressUpdate:
		m.text = msg.text
		return m, nil
	case progressDone:
		m.done = true
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m progressModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.text)
}

// runWithProgress runs an action while showing a spinner with dynamically updating text.
// The action receives a tea.Program it can use to send progressUpdate messages.
func runWithProgress(title string, action func(p *tea.Program) error) error {
	m := newProgressModel(title)
	p := tea.NewProgram(m)

	var actionErr error
	go func() {
		actionErr = action(p)
		p.Send(progressDone{})
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("progress spinner: %w", err)
	}
	return actionErr
}
