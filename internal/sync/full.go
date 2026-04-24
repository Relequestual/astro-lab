package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
)

// Full performs a complete sync, replacing all local data
func (e *Engine) Full(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{FullSync: true}

	// Fetch all stars
	allStars, err := e.client.FetchStarredRepos(ctx, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("fetching all stars: %w", err)
	}

	starsData := &storage.StarsData{
		ByRepoID: make(map[string]models.Repository),
	}
	for _, star := range allStars {
		starsData.ByRepoID[star.ID] = star
	}
	result.NewStars = len(allStars)
	result.TotalStars = len(allStars)

	if err := e.store.SaveStars(starsData); err != nil {
		return nil, fmt.Errorf("saving stars: %w", err)
	}

	// Fetch all lists
	allLists, err := e.client.FetchLists(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching all lists: %w", err)
	}

	listsData := &storage.ListsData{
		ByListID: make(map[string]models.StarList),
	}
	for _, l := range allLists {
		listsData.ByListID[l.ID] = l
	}
	result.UpdatedLists = len(allLists)
	result.TotalLists = len(allLists)

	if err := e.store.SaveLists(listsData); err != nil {
		return nil, fmt.Errorf("saving lists: %w", err)
	}

	// Fetch all memberships
	memberships := &storage.MembershipsData{
		ListToRepos: make(map[string][]string),
		RepoToLists: make(map[string][]string),
	}

	for _, list := range allLists {
		items, err := e.client.FetchListItems(ctx, list.ID)
		if err != nil {
			return nil, fmt.Errorf("fetching items for list %s: %w", list.Name, err)
		}

		repoIDs := make([]string, len(items))
		for i, item := range items {
			repoIDs[i] = item.ID
		}
		memberships.ListToRepos[list.ID] = repoIDs

		for _, item := range items {
			memberships.RepoToLists[item.ID] = append(memberships.RepoToLists[item.ID], list.ID)
		}
	}

	if err := e.store.SaveMemberships(memberships); err != nil {
		return nil, fmt.Errorf("saving memberships: %w", err)
	}

	// Load or create metadata
	meta, err := e.store.LoadMetadata()
	if err != nil {
		return nil, fmt.Errorf("loading metadata: %w", err)
	}

	// Detect removed stars on full sync
	oldStars, _ := e.store.LoadStars()
	if oldStars != nil {
		for id := range oldStars.ByRepoID {
			if _, exists := starsData.ByRepoID[id]; !exists {
				result.RemovedStars++
			}
		}
	}

	now := time.Now().UTC()
	meta.LastSyncedAt = now
	meta.LastFullSyncAt = now
	if err := e.store.SaveMetadata(meta); err != nil {
		return nil, fmt.Errorf("saving metadata: %w", err)
	}

	return result, nil
}
