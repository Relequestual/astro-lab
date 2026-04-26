package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// dashboardButton represents a selectable action on the dashboard.
type dashboardButton struct {
	label  string
	action tea.Msg
}

// dashboardModel is the home hub screen.
type dashboardModel struct {
	selected   int
	buttons    []dashboardButton
	totalStars int
	totalLists int
	lastSync   string
	needsSync  bool
}

func newDashboardModel() dashboardModel {
	return dashboardModel{
		buttons: []dashboardButton{
			{label: "Browse Lists", action: navigateMsg{screen: ScreenLists}},
			{label: "Browse All Stars", action: navigateMsg{screen: ScreenAllStars}},
			{label: "Sync Now", action: syncStartMsg{full: false}},
			{label: "Full Sync", action: syncStartMsg{full: true}},
		},
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return nil
}

func (m dashboardModel) Update(msg tea.Msg) (dashboardModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "left", "h":
			if m.selected > 0 {
				m.selected--
			}
		case "right", "l":
			if m.selected < len(m.buttons)-1 {
				m.selected++
			}
		case "enter":
			if m.selected >= 0 && m.selected < len(m.buttons) {
				return m, func() tea.Msg { return m.buttons[m.selected].action }
			}
		}
	}
	return m, nil
}

func (m dashboardModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🏠 Dashboard") + "\n\n")

	// Stats
	b.WriteString(fmt.Sprintf("  ⭐ Stars: %d    📋 Lists: %d", m.totalStars, m.totalLists))
	if m.lastSync != "" {
		b.WriteString(fmt.Sprintf("    🔄 Last sync: %s", m.lastSync))
	}
	b.WriteString("\n")

	if m.needsSync {
		b.WriteString("  " + warningStyle.Render("No data yet — run a sync to get started.") + "\n")
	}

	b.WriteString("\n  ")

	// Buttons
	for i, btn := range m.buttons {
		if i == m.selected {
			b.WriteString(activeButtonStyle.Render(btn.label))
		} else {
			b.WriteString(buttonStyle.Render(btn.label))
		}
	}

	b.WriteString("\n\n  " + mutedStyle.Render("←/→ navigate • Enter select"))

	return b.String()
}
