package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Relequestual/astro-lab/internal/models"
)

// previewModel shows a dry-run diff before confirming a mutation.
type previewModel struct {
	active        bool
	repoName      string
	diff          models.MoveDiff
	listNames     map[string]string
	confirmed     bool
	resolved      bool
	width, height int
}

func newPreviewModel() previewModel {
	return previewModel{
		listNames: make(map[string]string),
	}
}

func (m *previewModel) show(repoName string, diff models.MoveDiff, listNames map[string]string) {
	m.active = true
	m.repoName = repoName
	m.diff = diff
	m.listNames = listNames
	m.confirmed = false
	m.resolved = false
}

func (m *previewModel) hide() {
	m.active = false
	m.resolved = false
	m.confirmed = false
}

func (m previewModel) Init() tea.Cmd {
	return nil
}

func (m previewModel) Update(msg tea.Msg) (previewModel, tea.Cmd) {
	if !m.active || m.resolved {
		return m, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y", "Y", "enter":
			m.confirmed = true
			m.resolved = true
		case "n", "N", "esc":
			m.confirmed = false
			m.resolved = true
		}
	}
	return m, nil
}

func (m previewModel) View() string {
	if !m.active {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Preview changes for %s\n\n", focusedStyle.Render(m.repoName)))

	if len(m.diff.Added) == 0 && len(m.diff.Removed) == 0 {
		b.WriteString(mutedStyle.Render("  No changes.") + "\n")
	}

	for _, id := range m.diff.Added {
		name := id
		if n, ok := m.listNames[id]; ok {
			name = n
		}
		b.WriteString(previewAddStyle.Render("  + "+name) + "\n")
	}
	for _, id := range m.diff.Removed {
		name := id
		if n, ok := m.listNames[id]; ok {
			name = n
		}
		b.WriteString(previewRemoveStyle.Render("  - "+name) + "\n")
	}

	b.WriteString("\n" + mutedStyle.Render("[y]es  [n]o"))

	w, h := m.width, m.height
	if w < 50 {
		w = 50
	}
	if h < 12 {
		h = 12
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(b.String()))
}
