package sync

import (
	"testing"
)

func TestContainsString(t *testing.T) {
	tests := []struct {
		slice    []string
		s        string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{nil, "a", false},
		{[]string{}, "a", false},
	}

	for _, tt := range tests {
		if got := containsString(tt.slice, tt.s); got != tt.expected {
			t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.s, got, tt.expected)
		}
	}
}

func TestSyncResult_String(t *testing.T) {
	r := SyncResult{
		NewStars:     5,
		UpdatedLists: 2,
		RemovedStars: 1,
		TotalStars:   100,
		TotalLists:   10,
		FullSync:     false,
	}

	s := r.String()
	if s == "" {
		t.Error("expected non-empty string")
	}

	r.FullSync = true
	s = r.String()
	if s == "" {
		t.Error("expected non-empty string for full sync")
	}
}
