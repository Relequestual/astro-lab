package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/models"
)

// reposInListModel displays repos within a specific star list.
type reposInListModel struct {
	listID     string
	listName   string
	stars      starsModel
	confirming bool
	confirm    confirmModel
}

func newReposInListModel(listID, listName string) reposInListModel {
	return reposInListModel{
		listID:   listID,
		listName: listName,
		stars:    newStarsModel(),
	}
}

func (m *reposInListModel) setRepos(repos []models.Repository) {
	m.stars.setRepos(repos)
}

func (m reposInListModel) Init() tea.Cmd {
	return nil
}

func (m reposInListModel) Update(msg tea.Msg) (reposInListModel, tea.Cmd) {
	// Confirmation dialog takes priority
	if m.confirming {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		if m.confirm.resolved {
			m.confirming = false
			if m.confirm.confirmed {
				if r, ok := m.stars.selectedRepo(); ok {
					rid := r.ID
					rname := r.NameWithOwner
					lname := m.listName
					return m, func() tea.Msg {
						return showListPickerMsg{repoID: rid, repoName: rname + " (remove from " + lname + ")"}
					}
				}
			}
		}
		return m, cmd
	}

	// Check for special keys before delegating to stars
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "r":
			if !m.stars.searching {
				if r, ok := m.stars.selectedRepo(); ok {
					m.confirming = true
					m.confirm = newConfirmModel(fmt.Sprintf("Remove %q from %q?", r.NameWithOwner, m.listName))
					m.confirm.width, m.confirm.height = m.stars.width, m.stars.height
					return m, nil
				}
			}
		case "m", "a":
			if !m.stars.searching {
				if r, ok := m.stars.selectedRepo(); ok {
					rid := r.ID
					rname := r.NameWithOwner
					return m, func() tea.Msg {
						return showListPickerMsg{repoID: rid, repoName: rname}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.stars, cmd = m.stars.Update(msg)
	return m, cmd
}

func (m reposInListModel) View() string {
	if m.confirming {
		return m.confirm.View()
	}
	title := fmt.Sprintf("📋 %s", m.listName)
	return m.stars.RenderTable(title, m.stars.width, m.stars.height)
}
