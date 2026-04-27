package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// syncOverlayModel shows sync progress as a modal overlay.
type syncOverlayModel struct {
	active   bool
	spinner  spinner.Model
	phase    string
	fetched  int
	total    int
	cancelFn context.CancelFunc
}

func newSyncOverlayModel() syncOverlayModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return syncOverlayModel{spinner: sp}
}

func (m *syncOverlayModel) start(cancel context.CancelFunc) {
	m.active = true
	m.cancelFn = cancel
	m.phase = "Starting..."
	m.fetched = 0
	m.total = 0
}

func (m *syncOverlayModel) stop() {
	m.active = false
	if m.cancelFn != nil {
		m.cancelFn()
		m.cancelFn = nil
	}
}

func (m syncOverlayModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m syncOverlayModel) Update(msg tea.Msg) (syncOverlayModel, tea.Cmd) {
	switch msg := msg.(type) {
	case syncProgressMsg:
		m.phase = string(msg.progress.Phase)
		m.fetched = msg.progress.Fetched
		m.total = msg.progress.Total
		return m, nil

	case spinner.TickMsg:
		if m.active {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.active && msg.String() == "esc" {
			m.stop()
			return m, func() tea.Msg {
				return statusMsg{text: "Sync cancelled", isError: false}
			}
		}
	}
	return m, nil
}

func (m syncOverlayModel) View() string {
	if !m.active {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.spinner.View() + " Syncing...\n\n")
	b.WriteString(fmt.Sprintf("  Phase: %s\n", m.phase))
	if m.total > 0 {
		b.WriteString(fmt.Sprintf("  Progress: %d / %d\n", m.fetched, m.total))
	}
	b.WriteString("\n" + mutedStyle.Render("Esc to cancel"))

	return lipgloss.Place(60, 10, lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(b.String()))
}

