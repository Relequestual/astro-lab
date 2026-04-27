package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Relequestual/astro-lab/internal/models"
)

// listPickerModel is a modal overlay for multi-selecting star lists.
type listPickerModel struct {
	active      bool
	repoID      string
	repoName    string
	lists       []models.StarList
	checked     map[string]bool
	cursor      int
	previousIDs []string
	width       int
	height      int
}

func newListPickerModel() listPickerModel {
	return listPickerModel{
		checked: make(map[string]bool),
	}
}

func (m *listPickerModel) show(repoID, repoName string, lists []models.StarList, currentListIDs []string) {
	m.active = true
	m.repoID = repoID
	m.repoName = repoName
	m.lists = lists
	m.cursor = 0
	// Copy to avoid aliasing the shared membership slice
	m.previousIDs = make([]string, len(currentListIDs))
	copy(m.previousIDs, currentListIDs)
	m.checked = make(map[string]bool, len(currentListIDs))
	for _, id := range currentListIDs {
		m.checked[id] = true
	}
}

func (m *listPickerModel) hide() {
	m.active = false
}

func (m listPickerModel) Init() tea.Cmd {
	return nil
}

func (m listPickerModel) Update(msg tea.Msg) (listPickerModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.lists)-1 {
				m.cursor++
			}
		case " ", "x":
			if m.cursor >= 0 && m.cursor < len(m.lists) {
				id := m.lists[m.cursor].ID
				m.checked[id] = !m.checked[id]
				if !m.checked[id] {
					delete(m.checked, id)
				}
			}
		case "enter":
			// Build selected IDs in stable list order
			selected := make([]string, 0, len(m.checked))
			for _, l := range m.lists {
				if m.checked[l.ID] {
					selected = append(selected, l.ID)
				}
			}
			rid := m.repoID
			rname := m.repoName
			prev := m.previousIDs
			m.active = false
			return m, func() tea.Msg {
				return listPickerConfirmedMsg{
					repoID:      rid,
					repoName:    rname,
					selectedIDs: selected,
					previousIDs: prev,
				}
			}
		case "esc":
			m.active = false
			return m, nil
		}
	}
	return m, nil
}

func (m listPickerModel) View() string {
	if !m.active {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Select lists for %s\n\n", focusedStyle.Render(m.repoName)))

	if len(m.lists) == 0 {
		b.WriteString(mutedStyle.Render("  No lists available.") + "\n")
		b.WriteString(mutedStyle.Render("  Press Esc to close.") + "\n")
	}

	availH := m.height - 6
	if availH < 1 {
		availH = 1
	}
	start, end := visibleWindow(m.cursor, len(m.lists), availH)

	for i := start; i < end; i++ {
		l := m.lists[i]
		check := "[ ]"
		if m.checked[l.ID] {
			check = "[x]"
		}
		prefix := "  "
		if i == m.cursor {
			prefix = selectedStyle.Render("▸ ")
			b.WriteString(prefix + focusedStyle.Render(check+" "+l.Name) + "\n")
		} else {
			b.WriteString(prefix + check + " " + l.Name + "\n")
		}
	}

	b.WriteString("\n" + mutedStyle.Render("Space toggle • Enter confirm • Esc cancel"))

	maxW := m.width - 4
	if maxW < 30 {
		maxW = 30
	} else if maxW > 50 {
		maxW = 50
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		dialogStyle.MaxWidth(maxW).Render(b.String()))
}
