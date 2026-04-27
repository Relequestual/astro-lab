package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/models"
)

// detailModel shows full metadata for a single repository.
type detailModel struct {
	repo        models.Repository
	listNames   []string
	width       int
	height      int
	actionIndex int
}

var detailActions = []string{"Add to list (a)", "Copy URL (o)", "Back (esc)"}

func newDetailModel(repo models.Repository, listNames []string) detailModel {
	return detailModel{
		repo:      repo,
		listNames: listNames,
	}
}

func (m detailModel) Init() tea.Cmd {
	return nil
}

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.actionIndex > 0 {
				m.actionIndex--
			}
		case "down", "j":
			if m.actionIndex < len(detailActions)-1 {
				m.actionIndex++
			}
		case "a":
			return m, func() tea.Msg {
				return showListPickerMsg{repoID: m.repo.ID, repoName: m.repo.NameWithOwner}
			}
		case "o":
			url := m.repo.URL
			return m, func() tea.Msg {
				if err := clipboard.WriteAll(url); err != nil {
					return statusMsg{text: "Clipboard unavailable: " + url, isError: true}
				}
				return statusMsg{text: "Copied to clipboard: " + url}
			}
		case "esc":
			return m, func() tea.Msg { return backMsg{} }
		case "enter":
			switch m.actionIndex {
			case 0:
				return m, func() tea.Msg {
					return showListPickerMsg{repoID: m.repo.ID, repoName: m.repo.NameWithOwner}
				}
			case 1:
				url := m.repo.URL
				return m, func() tea.Msg {
					if err := clipboard.WriteAll(url); err != nil {
						return statusMsg{text: "Clipboard unavailable: " + url, isError: true}
					}
					return statusMsg{text: "Copied to clipboard: " + url}
				}
			case 2:
				return m, func() tea.Msg { return backMsg{} }
			}
		}
	}
	return m, nil
}

func (m detailModel) View() string {
	var b strings.Builder
	r := m.repo

	b.WriteString(titleStyle.Render("📦 "+r.NameWithOwner) + "\n\n")

	row := func(label, value string) {
		b.WriteString("  " + labelStyle.Render(label) + valueStyle.Render(value) + "\n")
	}

	row("Name", r.NameWithOwner)
	row("Description", r.Description)
	row("Language", r.Language)
	row("Stars", fmt.Sprintf("%d", r.StargazerCount))
	row("Forks", fmt.Sprintf("%d", r.ForkCount))
	row("URL", r.URL)
	if !r.StarredAt.IsZero() {
		row("Starred", r.StarredAt.Format("2006-01-02"))
	}
	row("ID", r.ID)

	b.WriteString("\n")

	// List memberships
	if len(m.listNames) > 0 {
		b.WriteString("  " + labelStyle.Render("Lists") + strings.Join(m.listNames, ", ") + "\n")
	} else {
		b.WriteString("  " + labelStyle.Render("Lists") + mutedStyle.Render("(none)") + "\n")
	}

	b.WriteString("\n")

	// Actions
	for i, action := range detailActions {
		if i == m.actionIndex {
			b.WriteString("  " + selectedStyle.Render("▸ "+action) + "\n")
		} else {
			b.WriteString("    " + mutedStyle.Render(action) + "\n")
		}
	}

	return b.String()
}
