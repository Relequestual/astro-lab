package tui

import (
	"strings"
	"testing"

	"github.com/Relequestual/astro-lab/internal/models"
)

func TestPreviewViewAddRemove(t *testing.T) {
	m := newPreviewModel()
	m.show("org/repo", models.MoveDiff{
		Added:   []string{"list-1"},
		Removed: []string{"list-2"},
	}, map[string]string{
		"list-1": "Go Projects",
		"list-2": "Old Stuff",
	})

	view := m.View()
	if !strings.Contains(view, "org/repo") {
		t.Error("View should contain repo name")
	}
	if !strings.Contains(view, "+ Go Projects") {
		t.Error("View should show added list")
	}
	if !strings.Contains(view, "- Old Stuff") {
		t.Error("View should show removed list")
	}
}

func TestPreviewViewEmpty(t *testing.T) {
	m := newPreviewModel()
	m.show("org/repo", models.MoveDiff{}, nil)

	view := m.View()
	if !strings.Contains(view, "No changes") {
		t.Error("View should indicate no changes")
	}
}

func TestPreviewViewInactive(t *testing.T) {
	m := newPreviewModel()
	view := m.View()
	if view != "" {
		t.Errorf("View when inactive should be empty, got %q", view)
	}
}

func TestPreviewViewNoChanges(t *testing.T) {
	m := newPreviewModel()
	m.show("org/repo", models.MoveDiff{
		Before: []string{"a"},
		After:  []string{"a"},
	}, nil)

	view := m.View()
	if !strings.Contains(view, "No changes") {
		t.Error("View with same before/after should show no changes")
	}
}
