package sync

import (
	"context"
	"fmt"

	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
)

type listsResult struct {
	UpdatedLists int
	TotalLists   int
}

func (e *Engine) syncLists(ctx context.Context) (*listsResult, error) {
	remoteLists, err := e.client.FetchLists(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching lists: %w", err)
	}

	localLists, err := e.store.LoadLists()
	if err != nil {
		return nil, fmt.Errorf("loading local lists: %w", err)
	}

	result := &listsResult{TotalLists: len(remoteLists)}

	newListData := &storage.ListsData{
		ByListID: make(map[string]models.StarList),
	}

	for _, l := range remoteLists {
		existing, found := localLists.ByListID[l.ID]
		if !found || existing.UpdatedAt != l.UpdatedAt || existing.ItemCount != l.ItemCount {
			result.UpdatedLists++
		}
		newListData.ByListID[l.ID] = l
	}

	if err := e.store.SaveLists(newListData); err != nil {
		return nil, fmt.Errorf("saving lists: %w", err)
	}

	return result, nil
}

func (e *Engine) syncMemberships(ctx context.Context, meta *models.Metadata) error {
	listsData, err := e.store.LoadLists()
	if err != nil {
		return fmt.Errorf("loading lists: %w", err)
	}

	memberships, err := e.store.LoadMemberships()
	if err != nil {
		return fmt.Errorf("loading memberships: %w", err)
	}

	for listID, list := range listsData.ByListID {
		// For delta sync, only refresh lists that changed
		if !meta.LastSyncedAt.IsZero() {
			if list.UpdatedAt.Before(meta.LastSyncedAt) && list.LastAddedAt.Before(meta.LastSyncedAt) {
				continue
			}
		}

		items, err := e.client.FetchListItems(ctx, listID)
		if err != nil {
			return fmt.Errorf("fetching items for list %s: %w", list.Name, err)
		}

		// Update list->repo mapping
		repoIDs := make([]string, len(items))
		for i, item := range items {
			repoIDs[i] = item.ID
		}
		memberships.ListToRepos[listID] = repoIDs

		// Rebuild repo->list mapping for affected repos
		for _, repoID := range repoIDs {
			lists := memberships.RepoToLists[repoID]
			if !containsString(lists, listID) {
				memberships.RepoToLists[repoID] = append(lists, listID)
			}
		}
	}

	return e.store.SaveMemberships(memberships)
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
