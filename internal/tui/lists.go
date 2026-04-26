package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/models"
)

type listInputMode int

const (
	listInputNone   listInputMode = iota
	listInputCreate
	listInputRename
)

// listsModel manages the star lists panel.
type listsModel struct {
	lists        []models.StarList
	filtered     []models.StarList
	cursor       int
	width        int
	height       int
	searching    bool
	search       searchModel
	confirming   bool
	confirm      confirmModel
	inputMode    listInputMode
	textInput    textinput.Model
	renameListID string
}

func newListsModel() listsModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 80
	return listsModel{
		search:    newSearchModel("Search lists..."),
		textInput: ti,
	}
}

func (m *listsModel) setLists(lists []models.StarList) {
	// Sort by name for stable ordering (lists come from a map)
	sort.Slice(lists, func(i, j int) bool {
		return lists[i].Name < lists[j].Name
	})
	m.lists = lists
	m.applyFilter()
}

func (m *listsModel) applyFilter() {
	if m.search.Query() == "" {
		m.filtered = m.lists
	} else {
		m.filtered = nil
		for _, l := range m.lists {
			if fuzzyMatch(l.Name, m.search.Query()) {
				m.filtered = append(m.filtered, l)
			}
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *listsModel) selectedList() (models.StarList, bool) {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		return m.filtered[m.cursor], true
	}
	return models.StarList{}, false
}

func (m listsModel) Init() tea.Cmd {
	return nil
}

func (m listsModel) Update(msg tea.Msg) (listsModel, tea.Cmd) {
	// Confirmation dialog takes priority
	if m.confirming {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		if m.confirm.resolved {
			m.confirming = false
			if m.confirm.confirmed {
				if l, ok := m.selectedList(); ok {
					return m, func() tea.Msg { return listDeletedMsg{listID: l.ID} }
				}
			}
		}
		return m, cmd
	}

	// Text input for create/rename
	if m.inputMode != listInputNone {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				val := m.textInput.Value()
				if val == "" {
					m.inputMode = listInputNone
					return m, nil
				}
				mode := m.inputMode
				renameID := m.renameListID
				m.inputMode = listInputNone
				m.textInput.SetValue("")
				if mode == listInputCreate {
					return m, func() tea.Msg {
						return listCreatedMsg{list: models.StarList{Name: val}}
					}
				}
				return m, func() tea.Msg {
					return listUpdatedMsg{listID: renameID, newName: val}
				}
			case "esc":
				m.inputMode = listInputNone
				m.textInput.SetValue("")
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	// Search mode
	if m.searching {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.search.Reset()
				m.applyFilter()
				return m, nil
			case "enter":
				m.searching = false
				m.search.input.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	// Normal mode key handling
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "enter":
			if l, ok := m.selectedList(); ok {
				return m, func() tea.Msg {
					return navigateMsg{screen: ScreenReposInList, listID: l.ID, listName: l.Name}
				}
			}
		case "/":
			m.searching = true
			m.search.Focus()
			return m, nil
		case "n":
			m.inputMode = listInputCreate
			m.textInput.Placeholder = "New list name..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, nil
		case "r":
			if l, ok := m.selectedList(); ok {
				m.inputMode = listInputRename
				m.renameListID = l.ID
				m.textInput.Placeholder = l.Name
				m.textInput.SetValue(l.Name)
				m.textInput.Focus()
				return m, nil
			}
		case "d":
			if l, ok := m.selectedList(); ok {
				m.confirming = true
				m.confirm = newConfirmModel(fmt.Sprintf("Delete list %q?", l.Name))
				return m, nil
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = max(0, len(m.filtered)-1)
		}
	}
	return m, nil
}

func (m listsModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📋 Star Lists") + "\n")

	if m.confirming {
		b.WriteString(m.confirm.View())
		return b.String()
	}

	if m.inputMode != listInputNone {
		label := "Create"
		if m.inputMode == listInputRename {
			label = "Rename"
		}
		b.WriteString(fmt.Sprintf("\n  %s: %s\n", label, m.textInput.View()))
		return b.String()
	}

	if m.searching {
		b.WriteString("  " + m.search.View() + "\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString("\n  " + mutedStyle.Render("No lists found.") + "\n")
		b.WriteString("  " + mutedStyle.Render("Press n to create one.") + "\n")
		return b.String()
	}

	availH := m.height - 6
	if availH < 1 {
		availH = 10
	}
	start, end := visibleWindow(m.cursor, len(m.filtered), availH)

	b.WriteString("\n")
	for i := start; i < end; i++ {
		l := m.filtered[i]
		prefix := "  "
		if i == m.cursor {
			prefix = selectedStyle.Render("▸ ")
		}
		name := l.Name
		if l.IsPrivate {
			name += " 🔒"
		}
		count := mutedStyle.Render(fmt.Sprintf("(%d repos)", l.ItemCount))
		if i == m.cursor {
			b.WriteString(prefix + selectedStyle.Render(name) + " " + count + "\n")
		} else {
			b.WriteString(prefix + name + " " + count + "\n")
		}
	}

	b.WriteString("\n  " + mutedStyle.Render("↑/↓ navigate • Enter view • / search • n new • r rename • d delete"))
	return b.String()
}
