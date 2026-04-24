package mutation

import (
	"sort"
)

// ComputeDiff computes the difference between current and desired list memberships
func ComputeDiff(currentListIDs, desiredListIDs []string) (added, removed []string) {
	currentSet := toSet(currentListIDs)
	desiredSet := toSet(desiredListIDs)

	for id := range desiredSet {
		if !currentSet[id] {
			added = append(added, id)
		}
	}
	for id := range currentSet {
		if !desiredSet[id] {
			removed = append(removed, id)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	return
}

func toSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set
}
