package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/models"
)

type sortColumn int

const (
	sortByName     sortColumn = iota
	sortByLanguage
	sortByStars
	sortByForks
	sortByDate
)

type membershipFilter int

const (
	filterAll          membershipFilter = iota
	filterNotInAnyList
)

var sortColumnNames = [...]string{"Name", "Language", "Stars", "Forks", "Date"}

// starsModel shows all starred repos with sorting, filtering, and batch selection.
type starsModel struct {
	repos       []models.Repository
	filtered    []models.Repository
	cursor      int
	width       int
	height      int
	sortCol     sortColumn
	sortAsc     bool
	selected    map[string]bool
	searching   bool
	search      searchModel
	memFilter   membershipFilter
	repoToLists map[string][]string
}

func newStarsModel() starsModel {
	return starsModel{
		sortCol:     sortByName,
		sortAsc:     true,
		selected:    make(map[string]bool),
		search:      newSearchModel("Search repos..."),
		repoToLists: make(map[string][]string),
	}
}

func (m *starsModel) setRepos(repos []models.Repository) {
	m.repos = repos
	m.sortAndFilter()
}

func (m *starsModel) sortAndFilter() {
	// Filter
	var result []models.Repository
	for _, r := range m.repos {
		if m.memFilter == filterNotInAnyList {
			if lists, ok := m.repoToLists[r.ID]; ok && len(lists) > 0 {
				continue
			}
		}
		if m.search.Query() != "" {
			target := r.NameWithOwner + " " + r.Description + " " + r.Language
			if !fuzzyMatch(target, m.search.Query()) {
				continue
			}
		}
		result = append(result, r)
	}

	// Sort
	col := m.sortCol
	asc := m.sortAsc
	sort.Slice(result, func(i, j int) bool {
		var less bool
		switch col {
		case sortByLanguage:
			less = result[i].Language < result[j].Language
		case sortByStars:
			less = result[i].StargazerCount < result[j].StargazerCount
		case sortByForks:
			less = result[i].ForkCount < result[j].ForkCount
		case sortByDate:
			less = result[i].StarredAt.Before(result[j].StarredAt)
		default:
			less = result[i].NameWithOwner < result[j].NameWithOwner
		}
		if !asc {
			less = !less
		}
		return less
	})

	m.filtered = result
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *starsModel) selectedRepo() (models.Repository, bool) {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		return m.filtered[m.cursor], true
	}
	return models.Repository{}, false
}

func (m starsModel) Init() tea.Cmd {
	return nil
}

func (m starsModel) Update(msg tea.Msg) (starsModel, tea.Cmd) {
	// Search mode
	if m.searching {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.search.Reset()
				m.sortAndFilter()
				return m, nil
			case "enter":
				m.searching = false
				m.search.input.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.sortAndFilter()
		return m, cmd
	}

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
			if r, ok := m.selectedRepo(); ok {
				return m, func() tea.Msg {
					return navigateMsg{screen: ScreenRepoDetail, repoID: r.ID}
				}
			}
		case "/":
			m.searching = true
			m.search.Focus()
			return m, nil
		case "tab":
			m.sortCol = (m.sortCol + 1) % sortColumn(len(sortColumnNames))
			m.sortAndFilter()
		case "1":
			m.toggleSort(sortByName)
		case "2":
			m.toggleSort(sortByLanguage)
		case "3":
			m.toggleSort(sortByStars)
		case "4":
			m.toggleSort(sortByForks)
		case "5":
			m.toggleSort(sortByDate)
		case " ":
			if r, ok := m.selectedRepo(); ok {
				if m.selected[r.ID] {
					delete(m.selected, r.ID)
				} else {
					m.selected[r.ID] = true
				}
			}
		case "a":
			if len(m.selected) > 0 {
				// Batch mode: use first selected
				for id := range m.selected {
					return m, func() tea.Msg {
						return showListPickerMsg{repoID: id}
					}
				}
			}
			if r, ok := m.selectedRepo(); ok {
				return m, func() tea.Msg {
					return showListPickerMsg{repoID: r.ID, repoName: r.NameWithOwner}
				}
			}
		case "f":
			m.memFilter = (m.memFilter + 1) % 2
			m.sortAndFilter()
		case "pgup":
			m.cursor -= 10
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "pgdown":
			m.cursor += 10
			if m.cursor >= len(m.filtered) {
				m.cursor = max(0, len(m.filtered)-1)
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = max(0, len(m.filtered)-1)
		}
	}
	return m, nil
}

func (m *starsModel) toggleSort(col sortColumn) {
	if m.sortCol == col {
		m.sortAsc = !m.sortAsc
	} else {
		m.sortCol = col
		m.sortAsc = true
	}
	m.sortAndFilter()
}

// RenderTable renders the star table for use by both stars and repos_in_list views.
func (m starsModel) RenderTable(title string, width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(title) + "\n")

	if m.searching {
		b.WriteString("  " + m.search.View() + "\n")
	}

	// Filter indicator
	if m.memFilter == filterNotInAnyList {
		b.WriteString("  " + warningStyle.Render("Filter: not in any list") + "\n")
	}

	// Column headers
	headers := make([]string, len(sortColumnNames))
	for i, name := range sortColumnNames {
		if sortColumn(i) == m.sortCol {
			arrow := "▲"
			if !m.sortAsc {
				arrow = "▼"
			}
			headers[i] = focusedStyle.Render(name + " " + arrow)
		} else {
			headers[i] = mutedStyle.Render(name)
		}
	}
	b.WriteString("  " + strings.Join(headers, "  ") + "\n")

	if len(m.filtered) == 0 {
		b.WriteString("\n  " + mutedStyle.Render("No repos found.") + "\n")
		return b.String()
	}

	selCount := len(m.selected)
	if selCount > 0 {
		b.WriteString("  " + focusedStyle.Render(fmt.Sprintf("%d selected", selCount)) + "\n")
	}

	availH := height - 8
	if availH < 1 {
		availH = 10
	}
	start, end := visibleWindow(m.cursor, len(m.filtered), availH)

	nameW := 30
	descW := width - 75
	if descW < 10 {
		descW = 10
	}

	for i := start; i < end; i++ {
		r := m.filtered[i]
		prefix := "  "
		if m.selected[r.ID] {
			prefix = focusedStyle.Render("● ")
		}
		if i == m.cursor {
			prefix = selectedStyle.Render("▸ ")
		}

		name := truncate(r.NameWithOwner, nameW)
		desc := truncate(r.Description, descW)
		lang := fmt.Sprintf("%-10s", truncate(r.Language, 10))
		stars := fmt.Sprintf("%5d⭐", r.StargazerCount)
		forks := fmt.Sprintf("%4d🍴", r.ForkCount)

		line := fmt.Sprintf("%-*s %s %s %s %s", nameW, name, desc, lang, stars, forks)
		if i == m.cursor {
			b.WriteString(prefix + selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(prefix + line + "\n")
		}
	}

	b.WriteString("\n  " + mutedStyle.Render(
		fmt.Sprintf("%d repos • ↑/↓ navigate • Space select • / search • Tab sort • a add to list • f filter",
			len(m.filtered))))
	return b.String()
}

func (m starsModel) View() string {
	return m.RenderTable("⭐ All Stars", m.width, m.height)
}
