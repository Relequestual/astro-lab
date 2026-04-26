package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
)

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{"seconds", 30 * time.Second, "just now"},
		{"minutes", 5 * time.Minute, "5m ago"},
		{"hours", 3 * time.Hour, "3h ago"},
		{"days", 48 * time.Hour, "2d ago"},
		{"zero", 0, "just now"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanDuration(tt.dur)
			if got != tt.want {
				t.Errorf("humanDuration(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world", 8, "hello w…"},
		{"very_short_max", "hello", 1, "…"},
		{"zero_max", "hello", 0, "…"},
		{"empty", "", 5, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name   string
		target string
		query  string
		want   bool
	}{
		{"empty_query", "anything", "", true},
		{"exact", "hello", "hello", true},
		{"subsequence", "hello world", "hlo", true},
		{"case_insensitive", "Hello World", "hw", true},
		{"no_match", "hello", "xyz", false},
		{"query_longer", "hi", "hello", false},
		{"single_char", "test", "t", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyMatch(tt.target, tt.query)
			if got != tt.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.target, tt.query, got, tt.want)
			}
		})
	}
}

func TestVisibleWindow(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		total      int
		windowSize int
		wantStart  int
		wantEnd    int
	}{
		{"fits", 0, 5, 10, 0, 5},
		{"start", 2, 100, 10, 0, 10},
		{"middle", 50, 100, 10, 45, 55},
		{"end", 98, 100, 10, 90, 100},
		{"exact_size", 0, 10, 10, 0, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := visibleWindow(tt.cursor, tt.total, tt.windowSize)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("visibleWindow(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.cursor, tt.total, tt.windowSize, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestBuildDiff(t *testing.T) {
	t.Run("add_and_remove", func(t *testing.T) {
		before := []string{"a", "b", "c"}
		after := []string{"b", "c", "d"}
		diff := buildDiff(before, after)

		if len(diff.Added) != 1 || diff.Added[0] != "d" {
			t.Errorf("Added = %v, want [d]", diff.Added)
		}
		if len(diff.Removed) != 1 || diff.Removed[0] != "a" {
			t.Errorf("Removed = %v, want [a]", diff.Removed)
		}
	})

	t.Run("no_change", func(t *testing.T) {
		before := []string{"a", "b"}
		after := []string{"a", "b"}
		diff := buildDiff(before, after)

		if len(diff.Added) != 0 {
			t.Errorf("Added = %v, want []", diff.Added)
		}
		if len(diff.Removed) != 0 {
			t.Errorf("Removed = %v, want []", diff.Removed)
		}
	})

	t.Run("empty_before", func(t *testing.T) {
		diff := buildDiff(nil, []string{"a", "b"})
		if len(diff.Added) != 2 {
			t.Errorf("Added = %v, want [a b]", diff.Added)
		}
		if len(diff.Removed) != 0 {
			t.Errorf("Removed = %v, want []", diff.Removed)
		}
	})

	t.Run("empty_after", func(t *testing.T) {
		diff := buildDiff([]string{"a", "b"}, nil)
		if len(diff.Added) != 0 {
			t.Errorf("Added = %v, want []", diff.Added)
		}
		if len(diff.Removed) != 2 {
			t.Errorf("Removed = %v, want [a b]", diff.Removed)
		}
	})
}

func TestUpdateLocalMemberships(t *testing.T) {
	t.Run("add_to_new_list", func(t *testing.T) {
		md := &storage.MembershipsData{
			ListToRepos: map[string][]string{
				"list1": {"repo1"},
			},
			RepoToLists: map[string][]string{
				"repo1": {"list1"},
			},
		}

		updateLocalMemberships(md, "repo1", []string{"list1", "list2"})

		if got := md.RepoToLists["repo1"]; len(got) != 2 {
			t.Errorf("RepoToLists[repo1] = %v, want 2 entries", got)
		}
		if got := md.ListToRepos["list2"]; len(got) != 1 || got[0] != "repo1" {
			t.Errorf("ListToRepos[list2] = %v, want [repo1]", got)
		}
	})

	t.Run("remove_from_list", func(t *testing.T) {
		md := &storage.MembershipsData{
			ListToRepos: map[string][]string{
				"list1": {"repo1", "repo2"},
				"list2": {"repo1"},
			},
			RepoToLists: map[string][]string{
				"repo1": {"list1", "list2"},
				"repo2": {"list1"},
			},
		}

		updateLocalMemberships(md, "repo1", []string{"list1"})

		if got := md.RepoToLists["repo1"]; len(got) != 1 || got[0] != "list1" {
			t.Errorf("RepoToLists[repo1] = %v, want [list1]", got)
		}
		if got := md.ListToRepos["list2"]; len(got) != 0 {
			t.Errorf("ListToRepos[list2] = %v, want []", got)
		}
		// repo2 should be unaffected
		if got := md.RepoToLists["repo2"]; len(got) != 1 || got[0] != "list1" {
			t.Errorf("RepoToLists[repo2] = %v, want [list1]", got)
		}
	})

	t.Run("empty_new_lists", func(t *testing.T) {
		md := &storage.MembershipsData{
			ListToRepos: map[string][]string{
				"list1": {"repo1"},
			},
			RepoToLists: map[string][]string{
				"repo1": {"list1"},
			},
		}

		updateLocalMemberships(md, "repo1", nil)

		if got := md.RepoToLists["repo1"]; got != nil {
			t.Errorf("RepoToLists[repo1] = %v, want nil", got)
		}
		if got := md.ListToRepos["list1"]; len(got) != 0 {
			t.Errorf("ListToRepos[list1] = %v, want []", got)
		}
	})
}

func TestRateLimitString(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := rateLimitString(nil); got != "" {
			t.Errorf("rateLimitString(nil) = %q, want empty", got)
		}
	})
	t.Run("normal", func(t *testing.T) {
		rl := &models.RateLimit{Remaining: 4500, Limit: 5000}
		if got := rateLimitString(rl); got != "4500/5000" {
			t.Errorf("rateLimitString = %q, want 4500/5000", got)
		}
	})
}

func TestHeaderView(t *testing.T) {
	result := headerView(80, "testuser", "4500/5000", time.Time{})
	if !strings.Contains(result, "Astrometrics Lab") {
		t.Error("headerView should contain app name")
	}
	if !strings.Contains(result, "testuser") {
		t.Error("headerView should contain login")
	}
	if !strings.Contains(result, "4500/5000") {
		t.Error("headerView should contain rate limit")
	}
}

func TestHeaderViewEmptyFields(t *testing.T) {
	result := headerView(80, "", "", time.Time{})
	if !strings.Contains(result, "Astrometrics Lab") {
		t.Error("headerView should contain app name even with empty fields")
	}
}

func TestFooterView(t *testing.T) {
	bindings := []KeyBinding{{Key: "q", Desc: "quit"}}
	result := footerView(80, bindings, "", false)
	if !strings.Contains(result, "q") {
		t.Error("footerView should contain key binding")
	}
	if !strings.Contains(result, "quit") {
		t.Error("footerView should contain binding description")
	}
}

func TestFooterViewWithStatus(t *testing.T) {
	result := footerView(80, nil, "Sync complete", false)
	if !strings.Contains(result, "Sync complete") {
		t.Error("footerView should contain status text")
	}
}

func TestFooterViewWithError(t *testing.T) {
	result := footerView(80, nil, "Connection failed", true)
	if !strings.Contains(result, "Connection failed") {
		t.Error("footerView should contain error text")
	}
}

func TestGlobalBindings(t *testing.T) {
	bindings := globalBindings()
	if len(bindings) == 0 {
		t.Fatal("globalBindings should return at least one binding")
	}
	keys := make(map[string]bool)
	for _, b := range bindings {
		keys[b.Key] = true
	}
	for _, want := range []string{"?", "q"} {
		if !keys[want] {
			t.Errorf("globalBindings missing key %q", want)
		}
	}
}

func TestErrMsg(t *testing.T) {
	t.Run("nil_error", func(t *testing.T) {
		msg := errMsg{err: nil}
		if got := msg.Error(); got != "unknown error" {
			t.Errorf("errMsg{nil}.Error() = %q, want %q", got, "unknown error")
		}
	})
	t.Run("real_error", func(t *testing.T) {
		msg := errMsg{err: errForTest("boom")}
		if got := msg.Error(); got != "boom" {
			t.Errorf("errMsg.Error() = %q, want %q", got, "boom")
		}
	})
}

type errForTest string

func (e errForTest) Error() string { return string(e) }

func TestHelpView(t *testing.T) {
	result := helpView(80, 80)
	if !strings.Contains(result, "Keyboard Shortcuts") {
		t.Error("helpView should contain title")
	}
	if !strings.Contains(result, "Global") {
		t.Error("helpView should contain Global section")
	}
	if !strings.Contains(result, "Press ? or Esc to close") {
		t.Error("helpView should contain close hint")
	}
}

func TestScreenConstants(t *testing.T) {
	// Verify screen enum values are distinct
	screens := []Screen{ScreenAuth, ScreenDashboard, ScreenLists, ScreenReposInList, ScreenRepoDetail, ScreenAllStars}
	seen := make(map[Screen]bool)
	for _, s := range screens {
		if seen[s] {
			t.Errorf("duplicate Screen value: %d", s)
		}
		seen[s] = true
	}
}

func TestRepoSlice(t *testing.T) {
	m := map[string]models.Repository{
		"r1": {ID: "r1", NameWithOwner: "org/repo1"},
		"r2": {ID: "r2", NameWithOwner: "org/repo2"},
	}
	got := repoSlice(m)
	if len(got) != 2 {
		t.Errorf("repoSlice returned %d items, want 2", len(got))
	}
}

func TestListSlice(t *testing.T) {
	m := map[string]models.StarList{
		"l1": {ID: "l1", Name: "list1"},
		"l2": {ID: "l2", Name: "list2"},
	}
	got := listSlice(m)
	if len(got) != 2 {
		t.Errorf("listSlice returned %d items, want 2", len(got))
	}
}
