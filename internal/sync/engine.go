package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
)

// SyncPhase describes the current phase of a sync operation
type SyncPhase string

const (
	PhaseStars       SyncPhase = "stars"
	PhaseLists       SyncPhase = "lists"
	PhaseMemberships SyncPhase = "memberships"
)

// SyncProgress reports progress during a sync operation
type SyncProgress struct {
	Phase   SyncPhase
	Fetched int
	Total   int
}

// SyncProgressFunc is called during sync to report progress
type SyncProgressFunc func(SyncProgress)

// Engine handles sync operations between GitHub and local storage
type Engine struct {
	client *github.Client
	store  *storage.Store
}

// NewEngine creates a new sync engine
func NewEngine(client *github.Client, store *storage.Store) *Engine {
	return &Engine{client: client, store: store}
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	NewStars     int  `json:"newStars"`
	UpdatedLists int  `json:"updatedLists"`
	RemovedStars int  `json:"removedStars"`
	TotalStars   int  `json:"totalStars"`
	TotalLists   int  `json:"totalLists"`
	FullSync     bool `json:"fullSync"`
}

func (r SyncResult) String() string {
	mode := "delta"
	if r.FullSync {
		mode = "full"
	}
	return fmt.Sprintf("Sync complete (%s): %d new stars, %d removed, %d lists updated. Total: %d stars, %d lists.",
		mode, r.NewStars, r.RemovedStars, r.UpdatedLists, r.TotalStars, r.TotalLists)
}

// Delta performs an incremental sync since last sync time
func (e *Engine) Delta(ctx context.Context, onProgress SyncProgressFunc) (*SyncResult, error) {
	meta, err := e.store.LoadMetadata()
	if err != nil {
		return nil, fmt.Errorf("loading metadata: %w", err)
	}

	// If no previous sync, do a full sync
	if meta.LastSyncedAt.IsZero() {
		return e.Full(ctx, onProgress)
	}

	return e.deltaSync(ctx, meta, onProgress)
}

func (e *Engine) deltaSync(ctx context.Context, meta *models.Metadata, onProgress SyncProgressFunc) (*SyncResult, error) {
	result := &SyncResult{}

	// Fetch new stars since last sync
	var starProgress github.ProgressFunc
	if onProgress != nil {
		starProgress = func(fetched, total int) {
			onProgress(SyncProgress{Phase: PhaseStars, Fetched: fetched, Total: total})
		}
	}
	newStars, err := e.client.FetchStarredRepos(ctx, meta.LastSyncedAt, starProgress)
	if err != nil {
		return nil, fmt.Errorf("fetching new stars: %w", err)
	}

	// Update local stars
	starsData, err := e.store.LoadStars()
	if err != nil {
		return nil, fmt.Errorf("loading stars: %w", err)
	}

	for _, star := range newStars {
		if _, exists := starsData.ByRepoID[star.ID]; !exists {
			result.NewStars++
		}
		starsData.ByRepoID[star.ID] = star
	}
	result.TotalStars = len(starsData.ByRepoID)

	if err := e.store.SaveStars(starsData); err != nil {
		return nil, fmt.Errorf("saving stars: %w", err)
	}

	// Sync lists
	listsResult, err := e.syncLists(ctx)
	if err != nil {
		return nil, fmt.Errorf("syncing lists: %w", err)
	}
	result.UpdatedLists = listsResult.UpdatedLists
	result.TotalLists = listsResult.TotalLists

	// Sync memberships for changed lists
	if err := e.syncMemberships(ctx, meta); err != nil {
		return nil, fmt.Errorf("syncing memberships: %w", err)
	}

	// Update metadata
	meta.LastSyncedAt = time.Now().UTC()
	if err := e.store.SaveMetadata(meta); err != nil {
		return nil, fmt.Errorf("saving metadata: %w", err)
	}

	return result, nil
}
