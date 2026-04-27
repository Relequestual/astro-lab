package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/models"
)

func testLists() []models.StarList {
	return []models.StarList{
		{ID: "l1", Name: "Go Projects"},
		{ID: "l2", Name: "JS Tools"},
		{ID: "l3", Name: "DevOps"},
	}
}

func TestListPickerShowPreChecks(t *testing.T) {
	m := newListPickerModel()
	m.show("r1", "org/repo", testLists(), []string{"l2"})

	if !m.active {
		t.Error("expected active after show")
	}
	if !m.checked["l2"] {
		t.Error("l2 should be pre-checked")
	}
	if m.checked["l1"] {
		t.Error("l1 should not be pre-checked")
	}
}

func TestListPickerToggle(t *testing.T) {
	m := newListPickerModel()
	m.show("r1", "org/repo", testLists(), nil)

	// Toggle first item
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.checked["l1"] {
		t.Error("l1 should be checked after space")
	}

	// Toggle again to uncheck
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if m.checked["l1"] {
		t.Error("l1 should be unchecked after second space")
	}
}

func TestListPickerNavigate(t *testing.T) {
	m := newListPickerModel()
	m.show("r1", "org/repo", testLists(), nil)

	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("cursor after second down = %d, want 2", m.cursor)
	}

	// Should not go past the end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("cursor should not exceed list length, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 1 {
		t.Errorf("cursor after up = %d, want 1", m.cursor)
	}
}

func TestListPickerConfirm(t *testing.T) {
	m := newListPickerModel()
	m.show("r1", "org/repo", testLists(), []string{"l1"})

	// Check l3
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	// Confirm
	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.active {
		t.Error("should be inactive after confirm")
	}

	if cmd == nil {
		t.Fatal("expected a command after confirm")
	}

	msg := cmd()
	confirmed, ok := msg.(listPickerConfirmedMsg)
	if !ok {
		t.Fatalf("expected listPickerConfirmedMsg, got %T", msg)
	}
	if confirmed.repoID != "r1" {
		t.Errorf("repoID = %q, want %q", confirmed.repoID, "r1")
	}

	// Should contain l1 (pre-checked) and l3 (newly checked)
	idSet := make(map[string]bool)
	for _, id := range confirmed.selectedIDs {
		idSet[id] = true
	}
	if !idSet["l1"] || !idSet["l3"] {
		t.Errorf("selectedIDs = %v, want l1 and l3", confirmed.selectedIDs)
	}
}

func TestListPickerEsc(t *testing.T) {
	m := newListPickerModel()
	m.show("r1", "org/repo", testLists(), nil)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.active {
		t.Error("should be inactive after esc")
	}
}

func TestListPickerView(t *testing.T) {
	m := newListPickerModel()
	m.width = 60
	m.height = 20
	m.show("r1", "org/repo", testLists(), []string{"l2"})

	view := m.View()
	if !strings.Contains(view, "org/repo") {
		t.Error("View should contain repo name")
	}
	if !strings.Contains(view, "Go Projects") {
		t.Error("View should contain list names")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("View should show checked item")
	}
	if !strings.Contains(view, "[ ]") {
		t.Error("View should show unchecked items")
	}
}

func TestListPickerViewInactive(t *testing.T) {
	m := newListPickerModel()
	view := m.View()
	if view != "" {
		t.Errorf("View when inactive should be empty, got %q", view)
	}
}
