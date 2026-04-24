package mutation

import (
	"context"
	"fmt"

	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
)

// MoveEngine handles list membership mutations
type MoveEngine struct {
	client *github.Client
	store  *storage.Store
}

// NewMoveEngine creates a new move engine
func NewMoveEngine(client *github.Client, store *storage.Store) *MoveEngine {
	return &MoveEngine{client: client, store: store}
}

// MoveResult represents the outcome of a move operation
type MoveResult struct {
	RepoID   string          `json:"repoId"`
	RepoName string          `json:"repoName"`
	Diff     models.MoveDiff `json:"diff"`
	Applied  bool            `json:"applied"`
	DryRun   bool            `json:"dryRun"`
}

// Plan computes the move diff without executing
func (m *MoveEngine) Plan(ctx context.Context, repoID string, repoName string, desiredListIDs []string) (*MoveResult, error) {
	// Read current memberships from local cache
	memberships, err := m.store.LoadMemberships()
	if err != nil {
		return nil, fmt.Errorf("loading memberships: %w", err)
	}

	currentListIDs := memberships.RepoToLists[repoID]
	added, removed := ComputeDiff(currentListIDs, desiredListIDs)

	return &MoveResult{
		RepoID:   repoID,
		RepoName: repoName,
		Diff: models.MoveDiff{
			Before:  currentListIDs,
			After:   desiredListIDs,
			Added:   added,
			Removed: removed,
		},
		DryRun: true,
	}, nil
}

// Apply executes the move operation
func (m *MoveEngine) Apply(ctx context.Context, repoID string, repoName string, desiredListIDs []string) (*MoveResult, error) {
	result, err := m.Plan(ctx, repoID, repoName, desiredListIDs)
	if err != nil {
		return nil, err
	}

	if err := m.client.UpdateUserListsForItem(ctx, repoID, desiredListIDs); err != nil {
		return nil, fmt.Errorf("updating list memberships: %w", err)
	}

	// Update local memberships cache
	memberships, err := m.store.LoadMemberships()
	if err != nil {
		return nil, fmt.Errorf("loading memberships for update: %w", err)
	}

	// Remove repo from old lists
	for _, listID := range result.Diff.Removed {
		repos := memberships.ListToRepos[listID]
		memberships.ListToRepos[listID] = removeString(repos, repoID)
	}
	// Add repo to new lists
	for _, listID := range result.Diff.Added {
		memberships.ListToRepos[listID] = append(memberships.ListToRepos[listID], repoID)
	}
	memberships.RepoToLists[repoID] = desiredListIDs

	if err := m.store.SaveMemberships(memberships); err != nil {
		return nil, fmt.Errorf("saving memberships: %w", err)
	}

	result.Applied = true
	result.DryRun = false
	return result, nil
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != s {
			result = append(result, v)
		}
	}
	return result
}
