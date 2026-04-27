package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
)

// KeyBinding describes a single key binding for display.
type KeyBinding struct {
	Key  string
	Desc string
}

// Screen identifies which screen is active.
type Screen int

const (
	ScreenAuth Screen = iota
	ScreenDashboard
	ScreenLists
	ScreenReposInList
	ScreenRepoDetail
	ScreenAllStars
)

// humanDuration returns a human-friendly "time ago" string.
func humanDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// rateLimitString formats a rate limit for display.
func rateLimitString(rl *models.RateLimit) string {
	if rl == nil {
		return ""
	}
	return fmt.Sprintf("%d/%d", rl.Remaining, rl.Limit)
}

// truncate a string to maxLen runes, adding "…" if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

// fuzzyMatch checks if query is a subsequence of target (case-insensitive).
// Operates on runes for correct Unicode handling.
func fuzzyMatch(target, query string) bool {
	if query == "" {
		return true
	}
	t := []rune(strings.ToLower(target))
	q := []rune(strings.ToLower(query))
	ti := 0
	for qi := 0; qi < len(q); qi++ {
		found := false
		for ; ti < len(t); ti++ {
			if t[ti] == q[qi] {
				ti++
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// visibleWindow returns the start/end indices for a scrolling window.
func visibleWindow(cursor, total, windowSize int) (int, int) {
	if total <= windowSize {
		return 0, total
	}
	half := windowSize / 2
	start := cursor - half
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > total {
		end = total
		start = end - windowSize
	}
	return start, end
}

// repoSlice converts a map of repos to a slice.
func repoSlice(m map[string]models.Repository) []models.Repository {
	repos := make([]models.Repository, 0, len(m))
	for _, r := range m {
		repos = append(repos, r)
	}
	return repos
}

// listSlice converts a map of lists to a slice.
func listSlice(m map[string]models.StarList) []models.StarList {
	lists := make([]models.StarList, 0, len(m))
	for _, l := range m {
		lists = append(lists, l)
	}
	return lists
}

// buildDiff computes which lists were added/removed between two sets of list IDs.
func buildDiff(before, after []string) models.MoveDiff {
	beforeSet := make(map[string]bool, len(before))
	for _, id := range before {
		beforeSet[id] = true
	}
	afterSet := make(map[string]bool, len(after))
	for _, id := range after {
		afterSet[id] = true
	}
	var added, removed []string
	for _, id := range after {
		if !beforeSet[id] {
			added = append(added, id)
		}
	}
	for _, id := range before {
		if !afterSet[id] {
			removed = append(removed, id)
		}
	}
	return models.MoveDiff{Before: before, After: after, Added: added, Removed: removed}
}

// updateLocalMemberships updates the in-memory membership data after a mutation.
func updateLocalMemberships(md *storage.MembershipsData, repoID string, newListIDs []string) {
	oldLists := md.RepoToLists[repoID]
	for _, lid := range oldLists {
		filtered := md.ListToRepos[lid][:0]
		for _, rid := range md.ListToRepos[lid] {
			if rid != repoID {
				filtered = append(filtered, rid)
			}
		}
		md.ListToRepos[lid] = filtered
	}
	md.RepoToLists[repoID] = newListIDs
	for _, lid := range newListIDs {
		md.ListToRepos[lid] = append(md.ListToRepos[lid], repoID)
	}
}

// headerView renders the persistent context bar at the top of every screen.
func headerView(width int, login string, rateLimit string, lastSync time.Time) string {
	left := "⭐ Astrometrics Lab"

	var parts []string
	if login != "" {
		parts = append(parts, fmt.Sprintf("👤 %s", login))
	}
	if rateLimit != "" {
		parts = append(parts, fmt.Sprintf("⚡ %s", rateLimit))
	}
	if !lastSync.IsZero() {
		parts = append(parts, fmt.Sprintf("🔄 %s", humanDuration(time.Since(lastSync))))
	}
	right := strings.Join(parts, "  ")

	leftWidth := lipgloss.Width(headerStyle.Render(left))
	rightWidth := lipgloss.Width(right)
	spacer := ""
	if gap := width - leftWidth - rightWidth - 2; gap > 0 {
		spacer = strings.Repeat(" ", gap)
	}
	bar := left + spacer + right
	return headerStyle.Width(width).Render(bar)
}

// footerView renders the persistent footer bar at the bottom of every screen.
func footerView(width int, bindings []KeyBinding, status string, isError bool) string {
	var hints []string
	for _, b := range bindings {
		hint := footerKeyStyle.Render(b.Key) + " " + footerDescStyle.Render(b.Desc)
		hints = append(hints, hint)
	}
	left := strings.Join(hints, footerDescStyle.Render("  │  "))

	right := ""
	if status != "" {
		if isError {
			right = errorStyle.Render("✗ " + status)
		} else {
			right = successStyle.Render("✓ " + status)
		}
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacer := ""
	if gap := width - leftWidth - rightWidth - 2; gap > 0 {
		spacer = strings.Repeat(" ", gap)
	}
	bar := left + spacer + right
	return footerStyle.Width(width).Render(bar)
}

// globalBindings returns the key bindings available on every screen.
func globalBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "?", Desc: "help"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

// helpView renders a full-screen help overlay with all keybindings.
func helpView(width, height int) string {
	sections := []struct {
		title    string
		bindings []KeyBinding
	}{
		{"Global", []KeyBinding{
			{Key: "?", Desc: "Toggle this help"},
			{Key: "ctrl+c", Desc: "Quit"},
			{Key: "q", Desc: "Quit"},
			{Key: "esc", Desc: "Back / close overlay"},
			{Key: "u", Desc: "Undo last action"},
		}},
		{"Dashboard", []KeyBinding{
			{Key: "←/→", Desc: "Move between buttons"},
			{Key: "enter", Desc: "Select button"},
		}},
		{"Lists", []KeyBinding{
			{Key: "↑/↓", Desc: "Navigate"},
			{Key: "enter", Desc: "View repos"},
			{Key: "/", Desc: "Search"},
			{Key: "n", Desc: "New list"},
			{Key: "r", Desc: "Rename"},
			{Key: "d", Desc: "Delete list"},
		}},
		{"Stars / Repos in List", []KeyBinding{
			{Key: "↑/↓", Desc: "Navigate"},
			{Key: "enter", Desc: "View detail"},
			{Key: "space", Desc: "Toggle select"},
			{Key: "/", Desc: "Search"},
			{Key: "a", Desc: "Add to list(s)"},
			{Key: "tab", Desc: "Cycle sort"},
			{Key: "f", Desc: "Filter by membership"},
		}},
		{"Repos in List (extra)", []KeyBinding{
			{Key: "r", Desc: "Remove from list"},
			{Key: "m", Desc: "Move to list"},
		}},
		{"Repo Detail", []KeyBinding{
			{Key: "esc", Desc: "Back"},
			{Key: "o", Desc: "Open in browser"},
			{Key: "c", Desc: "Copy URL"},
			{Key: "a", Desc: "Add to list"},
		}},
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Keyboard Shortcuts") + "\n")
	for _, sec := range sections {
		b.WriteString(helpSectionStyle.Render(sec.title) + "\n")
		for _, kb := range sec.bindings {
			b.WriteString(helpKeyStyle.Render(kb.Key) + helpDescStyle.Render(kb.Desc) + "\n")
		}
	}
	b.WriteString("\n" + mutedStyle.Render("Press ? or Esc to close"))

	content := b.String()
	maxW := width - 4
	if maxW > 60 {
		maxW = 60
	}
	maxH := height - 4
	overlay := helpOverlayStyle.MaxWidth(maxW).MaxHeight(maxH).Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)
}
