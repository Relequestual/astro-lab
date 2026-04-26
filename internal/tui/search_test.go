package tui

import "testing"

func TestSearchModelNewAndReset(t *testing.T) {
	s := newSearchModel("Search...")

	if s.Query() != "" {
		t.Errorf("initial Query() = %q, want empty", s.Query())
	}

	s.input.SetValue("hello")
	s.query = s.input.Value()
	if s.Query() != "hello" {
		t.Errorf("Query() = %q, want %q", s.Query(), "hello")
	}

	s.Reset()
	if s.Query() != "" {
		t.Errorf("Query() after Reset = %q, want empty", s.Query())
	}
}

func TestSearchModelFuzzyMatchIntegration(t *testing.T) {
	tests := []struct {
		name   string
		target string
		query  string
		want   bool
	}{
		{"empty_query_matches_all", "anything", "", true},
		{"exact", "react", "react", true},
		{"subsequence", "charmbracelet/bubbletea", "cbt", true},
		{"case_insensitive", "TypeScript", "ts", true},
		{"no_match", "golang", "xyz", false},
		{"partial_match", "astro-lab", "al", true},
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
