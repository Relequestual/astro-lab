package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

var detailActions = []string{"Add to list (a)", "Open in browser (o)", "Copy URL (c)", "Back (esc)"}

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
			return m, openBrowserCmd(url)
		case "c":
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
				return m, openBrowserCmd(url)
			case 2:
				url := m.repo.URL
				return m, func() tea.Msg {
					if err := clipboard.WriteAll(url); err != nil {
						return statusMsg{text: "Clipboard unavailable: " + url, isError: true}
					}
					return statusMsg{text: "Copied to clipboard: " + url}
				}
			case 3:
				return m, func() tea.Msg { return backMsg{} }
			}
		}
	}
	return m, nil
}

// openBrowserCmd returns a tea.Cmd that opens the given URL in the default browser.
func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default: // linux, freebsd, etc.
			cmd = exec.Command("xdg-open", url)
		}
		if err := cmd.Start(); err != nil {
			return statusMsg{text: "Could not open browser: " + err.Error(), isError: true}
		}
		return statusMsg{text: "Opened in browser: " + url}
	}
}

func (m detailModel) View() string {
	var b strings.Builder
	r := m.repo

	b.WriteString(titleStyle.Render("📦 "+r.NameWithOwner) + "\n\n")

	row := func(label, value string) {
		b.WriteString("  " + labelStyle.Render(label) + valueStyle.Render(value) + "\n")
	}

	row("Name", r.NameWithOwner)
	if r.Description != "" {
		// Wrap description to fit the available width after the label column.
		descWidth := m.width - 20 // 2 indent + 16 label width + 2 margin
		if descWidth < 30 {
			descWidth = 30
		}
		wrapped := lipgloss.NewStyle().Width(descWidth).Render(r.Description)
		b.WriteString("  " + labelStyle.Render("Description") + wrapped + "\n")
	}
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
