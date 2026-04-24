package mutation

import (
	"reflect"
	"testing"
)

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		desired []string
		added   []string
		removed []string
	}{
		{
			name:    "no changes",
			current: []string{"a", "b"},
			desired: []string{"a", "b"},
			added:   nil,
			removed: nil,
		},
		{
			name:    "add only",
			current: []string{"a"},
			desired: []string{"a", "b", "c"},
			added:   []string{"b", "c"},
			removed: nil,
		},
		{
			name:    "remove only",
			current: []string{"a", "b", "c"},
			desired: []string{"a"},
			added:   nil,
			removed: []string{"b", "c"},
		},
		{
			name:    "add and remove",
			current: []string{"a", "b"},
			desired: []string{"b", "c"},
			added:   []string{"c"},
			removed: []string{"a"},
		},
		{
			name:    "empty to many",
			current: nil,
			desired: []string{"a", "b"},
			added:   []string{"a", "b"},
			removed: nil,
		},
		{
			name:    "many to empty",
			current: []string{"a", "b"},
			desired: nil,
			added:   nil,
			removed: []string{"a", "b"},
		},
		{
			name:    "both empty",
			current: nil,
			desired: nil,
			added:   nil,
			removed: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed := ComputeDiff(tt.current, tt.desired)
			if !reflect.DeepEqual(added, tt.added) {
				t.Errorf("added = %v, want %v", added, tt.added)
			}
			if !reflect.DeepEqual(removed, tt.removed) {
				t.Errorf("removed = %v, want %v", removed, tt.removed)
			}
		})
	}
}

func TestComputeDiff_Idempotent(t *testing.T) {
	// Running the same desired state should produce no changes
	current := []string{"a", "b", "c"}
	added, removed := ComputeDiff(current, current)
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("expected no changes for identical inputs, got added=%v removed=%v", added, removed)
	}
}
